package file

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domain "github.com/companyofcreators/file-service/internal/domain/file"
)

type DeleteUseCase struct {
	repo     domain.FileRepository
	storage  domain.ObjectStorage
	producer EventProducer
}

func NewDeleteUseCase(repo domain.FileRepository, storage domain.ObjectStorage, producer EventProducer) *DeleteUseCase {
	return &DeleteUseCase{
		repo:     repo,
		storage:  storage,
		producer: producer,
	}
}

type DeleteInput struct {
	FileID      uuid.UUID
	RequesterID uuid.UUID
	IsAdmin     bool
}

func (uc *DeleteUseCase) Execute(ctx context.Context, input DeleteInput) error {
	f, err := uc.repo.FindByID(ctx, input.FileID)
	if err != nil {
		return fmt.Errorf("failed to find file: %w", err)
	}

	// Ownership check: owner can delete, admin can delete any
	if !input.IsAdmin && f.OwnerID != input.RequesterID {
		return domain.ErrForbidden
	}

	// Delete from MinIO
	if err := uc.storage.Delete(ctx, f.ObjectKey); err != nil {
		return fmt.Errorf("%w: %v", domain.ErrDeleteFailed, err)
	}

	// Delete metadata from DB
	if err := uc.repo.Delete(ctx, input.FileID); err != nil {
		return fmt.Errorf("failed to delete file metadata: %w", err)
	}

	// Publish event
	event := FileDeletedEvent{
		FileID:    f.ID.String(),
		OwnerID:   f.OwnerID.String(),
		Bucket:    f.Bucket,
		ObjectKey: f.ObjectKey,
		Timestamp: f.CreatedAt.Unix(),
	}
	if err := uc.producer.PublishFileDeleted(ctx, event); err != nil {
		// Non-fatal
		_ = err
	}

	return nil
}
