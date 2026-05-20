package file

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	domain "github.com/companyofcreators/file-service/internal/domain/file"
)

type DownloadUseCase struct {
	repo    domain.FileRepository
	storage domain.ObjectStorage
	ttl     time.Duration
}

func NewDownloadUseCase(repo domain.FileRepository, storage domain.ObjectStorage, ttl time.Duration) *DownloadUseCase {
	return &DownloadUseCase{
		repo:    repo,
		storage: storage,
		ttl:     ttl,
	}
}

type DownloadInput struct {
	FileID uuid.UUID
}

type DownloadOutput struct {
	File         *domain.File
	PresignedURL string
	ThumbnailURL string
}

func (uc *DownloadUseCase) Execute(ctx context.Context, input DownloadInput) (*DownloadOutput, error) {
	f, err := uc.repo.FindByID(ctx, input.FileID)
	if err != nil {
		return nil, fmt.Errorf("failed to find file: %w", err)
	}

	expiry := uc.ttl

	presignedURL, err := uc.storage.GetPresignedURL(ctx, f.ObjectKey, expiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	var thumbnailURL string
	if domain.IsImage(f.MimeType) {
		thumbnailURL, _ = uc.storage.GetThumbnailURL(ctx, f.ObjectKey, expiry)
	}

	return &DownloadOutput{
		File:         f,
		PresignedURL: presignedURL,
		ThumbnailURL: thumbnailURL,
	}, nil
}
