package file

import (
	"time"

	"github.com/google/uuid"
)

type File struct {
	ID        uuid.UUID
	OwnerID   uuid.UUID
	Bucket    string
	ObjectKey string
	MimeType  string
	Size      int64
	FileType  FileType
	CreatedAt time.Time
}

type FileType string

const (
	FileTypeAvatar   FileType = "avatar"
	FileTypeOrder    FileType = "order_attachment"
	FileTypeChat     FileType = "chat_attachment"
	FileTypeDocument FileType = "document"
)

var AllowedMimeTypes = map[string]bool{
	"image/jpeg":      true,
	"image/png":       true,
	"image/webp":      true,
	"image/gif":       true,
	"video/mp4":       true,
	"application/pdf": true,
}

var VideoMimeTypes = map[string]bool{
	"video/mp4": true,
}

var ImageMimeTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
	"image/gif":  true,
}

func IsAllowedMimeType(mimeType string) bool {
	return AllowedMimeTypes[mimeType]
}

func IsImage(mimeType string) bool {
	return ImageMimeTypes[mimeType]
}

func ParseFileType(s string) (FileType, error) {
	ft := FileType(s)
	switch ft {
	case FileTypeAvatar, FileTypeOrder, FileTypeChat, FileTypeDocument:
		return ft, nil
	default:
		return "", ErrInvalidFileType
	}
}
