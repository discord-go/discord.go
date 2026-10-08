package rest

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/discord-go/discord.go/interactions"
	"github.com/discord-go/discord.go/messages"
	"github.com/discord-go/discord.go/snowflake"
)

func TestAttachmentMetadataUsesFileIndexes(t *testing.T) {
	files := []File{{Name: "a.txt"}, {Name: "b.png"}}
	got := AttachmentMetadata(files)
	if len(got) != 2 {
		t.Fatalf("len = %d", len(got))
	}
	if got[0].ID != 0 || got[0].Filename != "a.txt" {
		t.Errorf("got[0] = %+v", got[0])
	}
	if got[1].ID != 1 || got[1].Filename != "b.png" {
		t.Errorf("got[1] = %+v", got[1])
	}
}

func TestEditMessageParamsAttachmentsJSON(t *testing.T) {
	existing := []messages.AttachmentParams{
		{ID: 123456789012345678},
		messages.NewAttachmentParams(0, "screenshot.png"),
	}
	params := EditMessageParams{Attachments: &existing}
	raw, err := json.Marshal(params)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	list, ok := decoded["attachments"].([]any)
	if !ok || len(list) != 2 {
		t.Fatalf("attachments = %#v", decoded["attachments"])
	}
	first := list[0].(map[string]any)
	if len(first) != 1 || first["id"] != "123456789012345678" {
		t.Errorf("first = %#v", first)
	}
	second := list[1].(map[string]any)
	if second["id"] != "0" || second["filename"] != "screenshot.png" {
		t.Errorf("second = %#v", second)
	}
	if _, exists := second["url"]; exists {
		t.Error("request attachments must not carry response-only fields such as url")
	}
}

func TestExecuteWebhookParamsAttachmentsJSON(t *testing.T) {
	params := ExecuteWebhookParams{Attachments: []messages.AttachmentParams{
		{ID: 7, Filename: "x.png", IsSpoiler: true},
	}}
	raw, err := json.Marshal(params)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"is_spoiler":true`) {
		t.Errorf("json = %s", raw)
	}
	if strings.Contains(string(raw), `"proxy_url"`) {
		t.Errorf("json = %s", raw)
	}
}

func TestExecuteWebhookParamsBuilderAddAttachment(t *testing.T) {
	params := NewExecuteWebhookParamsBuilder().
		AddAttachment(messages.NewAttachmentParams(0, "a.txt")).
		Build()
	if len(params.Attachments) != 1 {
		t.Fatalf("len = %d", len(params.Attachments))
	}
	if params.Attachments[0].Filename != "a.txt" {
		t.Errorf("attachment = %+v", params.Attachments[0])
	}
}

func TestInteractionCallbackDataAttachmentsJSON(t *testing.T) {
	data := interactions.InteractionCallbackData{
		Attachments: []messages.AttachmentParams{{ID: 5, Filename: "clip.mp4"}},
	}
	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"attachments":[{"id":"5","filename":"clip.mp4"}]`) {
		t.Errorf("json = %s", raw)
	}
}

func TestWithAttachmentMetadataAppendsDescriptors(t *testing.T) {
	response := interactions.InteractionResponse{
		Data: &interactions.InteractionCallbackData{
			Attachments: []messages.AttachmentParams{{ID: 9, Filename: "keep.txt"}},
		},
	}
	got := withAttachmentMetadata(response, []File{{Name: "new.txt"}})
	if len(got.Data.Attachments) != 2 {
		t.Fatalf("len = %d", len(got.Data.Attachments))
	}
	if got.Data.Attachments[0].ID != 9 || got.Data.Attachments[1].ID != 0 {
		t.Errorf("attachments = %+v", got.Data.Attachments)
	}
}

func TestWithEditAttachmentMetadataPrependsExisting(t *testing.T) {
	existing := []messages.AttachmentParams{{ID: 11}}
	params := EditMessageParams{Attachments: &existing}
	got := withEditAttachmentMetadata(params, []File{{Name: "f.txt"}})
	if got.Attachments == nil || len(*got.Attachments) != 2 {
		t.Fatalf("attachments = %v", got.Attachments)
	}
	list := *got.Attachments
	if list[0].ID != 11 || list[1].ID != 0 || list[1].Filename != "f.txt" {
		t.Errorf("attachments = %+v", list)
	}
}

func TestMaxAttachmentSizeNoneRaisedTo20MiB(t *testing.T) {
	const twentyMiB = 20 * 1024 * 1024
	if MaxAttachmentSizeNone != twentyMiB {
		t.Errorf("MaxAttachmentSizeNone = %d, want %d", MaxAttachmentSizeNone, twentyMiB)
	}
	if err := ValidateFilesSize([][]byte{make([]byte, twentyMiB)}, 0); err != nil {
		t.Errorf("20 MiB should be accepted: %v", err)
	}
	if err := ValidateFilesSize([][]byte{make([]byte, twentyMiB+1)}, 0); err == nil {
		t.Error("20 MiB + 1 should be rejected")
	}
}

func TestAttachmentMetadataSnowflakeIndexRoundTrip(t *testing.T) {
	got := AttachmentMetadata([]File{{Name: "one.bin"}})
	raw, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	var decoded []messages.AttachmentParams
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if len(decoded) != 1 || decoded[0].ID != snowflake.ID(0) {
		t.Errorf("decoded = %+v", decoded)
	}
}
