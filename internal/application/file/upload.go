package file

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	domain "github.com/companyofcreators/file-service/internal/domain/file"
)

type UploadUseCase struct {
	repo     domain.FileRepository
	storage  domain.ObjectStorage
	producer EventProducer
	maxSize  int64
	maxVideo int64
	bucket   string
}

type EventProducer interface {
	PublishFileUploaded(ctx context.Context, event FileUploadedEvent) error
	PublishFileDeleted(ctx context.Context, event FileDeletedEvent) error
}

type FileUploadedEvent struct {
	FileID    string `json:"file_id"`
	OwnerID   string `json:"owner_id"`
	MimeType  string `json:"mime_type"`
	Size      int64  `json:"size"`
	Bucket    string `json:"bucket"`
	ObjectKey string `json:"object_key"`
	Timestamp int64  `json:"timestamp"`
}

type FileDeletedEvent struct {
	FileID    string `json:"file_id"`
	OwnerID   string `json:"owner_id"`
	Bucket    string `json:"bucket"`
	ObjectKey string `json:"object_key"`
	Timestamp int64  `json:"timestamp"`
}

func NewUploadUseCase(repo domain.FileRepository, storage domain.ObjectStorage, producer EventProducer, maxSize, maxVideo int64, bucket string) *UploadUseCase {
	return &UploadUseCase{
		repo:     repo,
		storage:  storage,
		producer: producer,
		maxSize:  maxSize,
		maxVideo: maxVideo,
		bucket:   bucket,
	}
}

type UploadInput struct {
	Reader   io.Reader
	FileName string
	MimeType string
	Size     int64
	FileType domain.FileType
	OwnerID  uuid.UUID
}

type UploadOutput struct {
	File      *domain.File
	ObjectKey string
	Bucket    string
}

func (uc *UploadUseCase) Execute(ctx context.Context, input UploadInput) (*UploadOutput, error) {
	if input.Size == 0 {
		return nil, domain.ErrEmptyFile
	}

	if !domain.IsAllowedMimeType(input.MimeType) {
		return nil, domain.ErrInvalidMimeType
	}

	// Determine max size based on type
	maxAllowed := uc.maxSize
	if domain.VideoMimeTypes[input.MimeType] {
		maxAllowed = uc.maxVideo
	}
	if input.Size > maxAllowed {
		return nil, domain.ErrFileTooLarge
	}

	// Generate unique object key
	ext := filepath.Ext(input.FileName)
	safeName := sanitizeFileName(input.FileName)
	objectKey := fmt.Sprintf("%s/%s/%s-%s%s",
		input.FileType,
		input.OwnerID.String(),
		uuid.New().String(),
		safeName,
		ext,
	)

	// Upload to MinIO
	if err := uc.storage.Upload(ctx, objectKey, input.Reader, input.Size, input.MimeType); err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrUploadFailed, err)
	}

	// Save metadata to PostgreSQL
	f := &domain.File{
		ID:        uuid.New(),
		OwnerID:   input.OwnerID,
		Bucket:    uc.bucket,
		ObjectKey: objectKey,
		MimeType:  input.MimeType,
		Size:      input.Size,
		FileType:  input.FileType,
		CreatedAt: time.Now().UTC(),
	}

	if err := uc.repo.Create(ctx, f); err != nil {
		// Best-effort cleanup of uploaded object
		if derr := uc.storage.Delete(ctx, objectKey); derr != nil {
			slog.Warn("failed to clean up uploaded object after metadata save failure", "object_key", objectKey, "error", derr)
		}
		return nil, fmt.Errorf("failed to save file metadata: %w", err)
	}

	// Publish event
	event := FileUploadedEvent{
		FileID:    f.ID.String(),
		OwnerID:   f.OwnerID.String(),
		MimeType:  f.MimeType,
		Size:      f.Size,
		Bucket:    f.Bucket,
		ObjectKey: f.ObjectKey,
		Timestamp: f.CreatedAt.Unix(),
	}
	if err := uc.producer.PublishFileUploaded(ctx, event); err != nil {
		// Non-fatal: log but don't fail the upload
		slog.Warn("failed to publish file uploaded event", "file_id", f.ID.String(), "error", err)
	}

	return &UploadOutput{
		File:      f,
		ObjectKey: objectKey,
		Bucket:    uc.bucket,
	}, nil
}

func sanitizeFileName(name string) string {
	name = strings.TrimSuffix(name, filepath.Ext(name))
	name = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, name)
	if len(name) > 100 {
		name = name[:100]
	}
	return name
}
