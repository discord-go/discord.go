package rest

import "bytes"

// AttachmentBuilder is a small Discord.js-style wrapper around a multipart
// file. Use Build to pass the upload to a bot context response helper.
type AttachmentBuilder struct {
	file File
}

// NewAttachmentBuilder loads a file from disk.
func NewAttachmentBuilder(path string) (*AttachmentBuilder, error) {
	file, err := FileFromPath(path)
	if err != nil {
		return nil, err
	}
	return &AttachmentBuilder{file: file}, nil
}

// NewAttachmentBuilderFromBytes creates an attachment from generated content.
func NewAttachmentBuilderFromBytes(name string, content []byte) *AttachmentBuilder {
	return &AttachmentBuilder{file: FileFromBytes(name, bytes.Clone(content))}
}

// SetName changes the filename sent to Discord.
func (b *AttachmentBuilder) SetName(name string) *AttachmentBuilder {
	if b != nil && name != "" {
		b.file.Name = name
	}
	return b
}

// Build returns the multipart upload.
func (b *AttachmentBuilder) Build() File {
	if b == nil {
		return File{}
	}
	return b.file
}
