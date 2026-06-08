package types

import (
	"fmt"
	"time"
)

type Document struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	ContentHash Hash      `json:"content_hash"`
	Owner       string    `json:"owner"`
	CreatedAt   time.Time `json:"created_at"`
	FileSize    int64     `json:"file_size"`
	MimeType    string    `json:"mime_type"`
}

func NewDocument(title string, contentHash Hash, owner string, fileSize int64, mimeType string) *Document {
	return &Document{
		ID:          generateID(),
		Title:       title,
		ContentHash: contentHash,
		Owner:       owner,
		CreatedAt:   time.Now(),
		FileSize:    fileSize,
		MimeType:    mimeType,
	}
}

func (d *Document) Validate() error {
	if d.Title == "" {
		return fmt.Errorf("заголовок не может быть пустым")
	}

	if d.ContentHash.IsZero() {
		return fmt.Errorf("хэш содержимого не может быть равен нулю")
	}

	if d.Owner == "" {
		return fmt.Errorf("владелец не может быть пустым")
	}

	if d.FileSize <= 0 {
		return fmt.Errorf("размер файла должен быть положительным")
	}

	return nil
}

func generateID() string {
	return fmt.Sprintf("doc_%d", time.Now().UnixNano())
}
