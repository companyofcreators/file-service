package http

import (
	"time"

	"github.com/google/uuid"

	domain "github.com/companyofcreators/file-service/internal/domain/file"
)

type UploadResponseDTO struct {
	Success bool             `json:"success"`
	Data    *FileResponseDTO `json:"data,omitempty"`
	Error   string           `json:"error,omitempty"`
}

type FileResponseDTO struct {
	ID        uuid.UUID `json:"id"`
	URL       string    `json:"url"`
	MimeType  string    `json:"mime_type"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
}

type FileDetailDTO struct {
	ID        uuid.UUID `json:"id"`
	OwnerID   uuid.UUID `json:"owner_id"`
	Bucket    string    `json:"bucket"`
	ObjectKey string    `json:"object_key"`
	MimeType  string    `json:"mime_type"`
	Size      int64     `json:"size"`
	FileType  string    `json:"file_type"`
	CreatedAt time.Time `json:"created_at"`
}

type DownloadResponseDTO struct {
	Success bool             `json:"success"`
	Data    *DownloadDataDTO `json:"data,omitempty"`
	Error   string           `json:"error,omitempty"`
}

type DownloadDataDTO struct {
	FileID       uuid.UUID `json:"file_id"`
	PresignedURL string    `json:"presigned_url"`
	ThumbnailURL string    `json:"thumbnail_url,omitempty"`
	MimeType     string    `json:"mime_type"`
	Size         int64     `json:"size"`
}

type ListFilesResponseDTO struct {
	Success bool              `json:"success"`
	Data    *ListFilesDataDTO `json:"data,omitempty"`
	Error   string            `json:"error,omitempty"`
}

type ListFilesDataDTO struct {
	Files  []*FileDetailDTO `json:"files"`
	Total  int              `json:"total"`
	Limit  int              `json:"limit"`
	Offset int              `json:"offset"`
}

type ErrorResponseDTO struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

func fileToResponseDTO(f *domain.File, baseURL string) *FileResponseDTO {
	return &FileResponseDTO{
		ID:        f.ID,
		URL:       baseURL + "/internal/files/" + f.ID.String() + "/download",
		MimeType:  f.MimeType,
		Size:      f.Size,
		CreatedAt: f.CreatedAt,
	}
}

func fileToDetailDTO(f *domain.File) *FileDetailDTO {
	return &FileDetailDTO{
		ID:        f.ID,
		OwnerID:   f.OwnerID,
		Bucket:    f.Bucket,
		ObjectKey: f.ObjectKey,
		MimeType:  f.MimeType,
		Size:      f.Size,
		FileType:  string(f.FileType),
		CreatedAt: f.CreatedAt,
	}
}
