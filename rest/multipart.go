package rest

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	gohttp "net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/discord-go/discord.go/json"
	"github.com/discord-go/discord.go/messages"
	"github.com/discord-go/discord.go/ratelimit"
	"github.com/discord-go/discord.go/snowflake"
)

// File represents a file attachment to be uploaded.
type File struct {
	Name        string
	ContentType string
	Reader      io.Reader
}

// FileFromPath loads a local file into an upload that can be reused safely for
// retries and multipart interaction responses.
func FileFromPath(path string) (File, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return File{}, err
	}
	file := FileFromBytes(filepath.Base(path), content)
	file.ContentType = mime.TypeByExtension(filepath.Ext(path))
	return file, nil
}

// FileFromBytes creates an in-memory upload.
func FileFromBytes(name string, content []byte) File {
	return File{Name: name, Reader: bytes.NewReader(content)}
}

// AttachmentMetadata returns the attachment descriptors required in the JSON
// payload for the supplied multipart files.
func AttachmentMetadata(files []File) []messages.Attachment {
	attachments := make([]messages.Attachment, 0, len(files))
	for index, file := range files {
		attachments = append(attachments, messages.Attachment{ID: snowflake.ID(index), Filename: file.Name})
	}
	return attachments
}

// Max attachment sizes by premium tier
const (
	MaxAttachmentSizeNone  = 8 * 1024 * 1024
	MaxAttachmentSizeTier1 = 25 * 1024 * 1024
	MaxAttachmentSizeTier2 = 50 * 1024 * 1024
	MaxAttachmentSizeTier3 = 100 * 1024 * 1024
)

// ValidateFilesSize checks if the total size of each file is within the limit for the given tier.
func ValidateFilesSize(fileContents [][]byte, premiumTier int) error {
	maxSize := MaxAttachmentSizeNone
	switch premiumTier {
	case 1:
		maxSize = MaxAttachmentSizeTier1
	case 2:
		maxSize = MaxAttachmentSizeTier2
	case 3:
		maxSize = MaxAttachmentSizeTier3
	}

	for i, content := range fileContents {
		if len(content) > maxSize {
			return fmt.Errorf("file %d exceeds maximum attachment size of %d bytes for premium tier %d", i, maxSize, premiumTier)
		}
	}
	return nil
}

// RequestMultipart performs an HTTP request to the Discord API using multipart/form-data.
// It sends the provided payload JSON as `payload_json` and attaches the provided files.
func (c *Client) RequestMultipart(ctx context.Context, method, path string, payload any, files []File, v any) error {
	return c.requestMultipart(ctx, method, path, payload, files, v, true)
}

// RequestMultipartNoAuth performs a multipart request without an Authorization
// header. Discord interaction webhook callbacks use the interaction token in
// the URL and do not require bot-token authorization.
func (c *Client) RequestMultipartNoAuth(ctx context.Context, method, path string, payload any, files []File, v any) error {
	return c.requestMultipart(ctx, method, path, payload, files, v, false)
}

// RequestMultipartForm sends ordinary multipart form fields plus one or more
// files. This is used by Discord endpoints such as sticker creation that do not
// accept payload_json.
func (c *Client) RequestMultipartForm(ctx context.Context, method, path string, fields map[string]string, files []File, v any) error {
	return c.requestMultipartForm(ctx, method, path, fields, "files[0]", files, v)
}

func (c *Client) RequestMultipartFormNamedFile(ctx context.Context, method, path string, fields map[string]string, fileField string, files []File, v any) error {
	return c.requestMultipartForm(ctx, method, path, fields, fileField, files, v)
}

func (c *Client) requestMultipartForm(ctx context.Context, method, path string, fields map[string]string, fileField string, files []File, v any) error {
	contents := make([][]byte, len(files))
	for index, file := range files {
		content, err := io.ReadAll(file.Reader)
		if err != nil {
			return err
		}
		contents[index] = content
	}
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			return err
		}
	}
	for index, file := range files {
		field := fileField
		if len(files) > 1 {
			field = fmt.Sprintf("files[%d]", index)
		}
		part, err := writer.CreateFormFile(field, file.Name)
		if err != nil {
			return err
		}
		if _, err := part.Write(contents[index]); err != nil {
			return err
		}
	}
	if err := writer.Close(); err != nil {
		return err
	}
	req, err := gohttp.NewRequestWithContext(ctx, method, c.BaseURL+path, &body)
	if err != nil {
		return err
	}
	if c.AuthMode != AuthNone && c.Token != "" {
		req.Header.Set("Authorization", string(c.AuthMode)+" "+c.Token)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if reason, ok := ReasonFromContext(ctx); ok {
		req.Header.Set("X-Audit-Log-Reason", url.QueryEscape(reason))
	}
	if err := c.Limiter.Wait(ctx, path); err != nil {
		return err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	info := ratelimit.ParseHeaders(resp.Header)
	c.Limiter.Update(path, info)
	responseBody, readErr := io.ReadAll(resp.Body)
	resp.Body.Close()
	if readErr != nil {
		return readErr
	}
	if resp.StatusCode >= 400 {
		var apiErr APIError
		if err := json.Unmarshal(responseBody, &apiErr); err == nil {
			apiErr.HTTPStatus = resp.StatusCode
			return &apiErr
		}
		return &APIError{HTTPStatus: resp.StatusCode, Message: string(responseBody)}
	}
	if v != nil && len(responseBody) > 0 {
		return json.Unmarshal(responseBody, v)
	}
	return nil
}

func (c *Client) requestMultipart(ctx context.Context, method, path string, payload any, files []File, v any, authenticate bool) error {
	requestURL := c.BaseURL + path

	var payloadBytes []byte
	if payload != nil {
		var err error
		payloadBytes, err = json.Marshal(payload)
		if err != nil {
			return err
		}
	}

	// Read all files into memory so we can retry if needed
	fileContents := make([][]byte, len(files))
	for i, file := range files {
		content, err := io.ReadAll(file.Reader)
		if err != nil {
			return err
		}
		fileContents[i] = content
	}

	bucket := path
	if idx := strings.Index(path, "?"); idx != -1 {
		bucket = path[:idx]
	}

	for {
		var b bytes.Buffer
		w := multipart.NewWriter(&b)

		if payloadBytes != nil {
			if err := w.WriteField("payload_json", string(payloadBytes)); err != nil {
				return err
			}
		}

		for i, file := range files {
			part, err := w.CreateFormFile(fmt.Sprintf("files[%d]", i), file.Name)
			if err != nil {
				return err
			}
			if _, err := io.Copy(part, bytes.NewReader(fileContents[i])); err != nil {
				return err
			}
		}

		if err := w.Close(); err != nil {
			return err
		}

		req, err := gohttp.NewRequestWithContext(ctx, method, requestURL, &b)
		if err != nil {
			return err
		}

		if authenticate && c.AuthMode != AuthNone && c.Token != "" {
			req.Header.Set("Authorization", string(c.AuthMode)+" "+c.Token)
		}
		req.Header.Set("Content-Type", w.FormDataContentType())

		if reason, ok := ReasonFromContext(ctx); ok {
			req.Header.Set("X-Audit-Log-Reason", url.QueryEscape(reason))
		}

		if err := c.Limiter.Wait(ctx, bucket); err != nil {
			return err
		}

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			return err
		}

		info := ratelimit.ParseHeaders(resp.Header)
		c.Limiter.Update(bucket, info)

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return err
		}

		if resp.StatusCode == 429 {
			var rateLimitErr struct {
				RetryAfter float64 `json:"retry_after"`
			}
			waitDuration := info.ResetAfter
			if err := json.Unmarshal(respBody, &rateLimitErr); err == nil && rateLimitErr.RetryAfter > 0 {
				waitDuration = time.Duration(rateLimitErr.RetryAfter * float64(time.Second))
			} else if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
				if parsed, err := strconv.ParseFloat(retryAfter, 64); err == nil {
					waitDuration = time.Duration(parsed * float64(time.Second))
				}
			}

			if waitDuration > 0 {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(waitDuration):
				}
				continue // Retry the request
			}
		}

		if resp.StatusCode >= 400 {
			var apiErr APIError
			if err := json.Unmarshal(respBody, &apiErr); err == nil {
				apiErr.HTTPStatus = resp.StatusCode
				return &apiErr
			}
			return &APIError{
				HTTPStatus: resp.StatusCode,
				Message:    string(respBody),
			}
		}

		if v != nil && len(respBody) > 0 {
			return json.Unmarshal(respBody, v)
		}

		return nil
	}
}
