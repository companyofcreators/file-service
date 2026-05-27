package file

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domain "github.com/companyofcreators/file-service/internal/domain/file"
)

type GetFileUseCase struct {
	repo domain.FileRepository
}

func NewGetFileUseCase(repo domain.FileRepository) *GetFileUseCase {
	return &GetFileUseCase{
		repo: repo,
	}
}

type GetFileInput struct {
	FileID uuid.UUID
}

func (uc *GetFileUseCase) Execute(ctx context.Context, input GetFileInput) (*domain.File, error) {
	f, err := uc.repo.FindByID(ctx, input.FileID)
	if err != nil {
		return nil, fmt.Errorf("failed to find file: %w", err)
	}

	return f, nil
}

type ListFilesUseCase struct {
	repo domain.FileRepository
}

func NewListFilesUseCase(repo domain.FileRepository) *ListFilesUseCase {
	return &ListFilesUseCase{
		repo: repo,
	}
}

type ListFilesInput struct {
	OwnerID uuid.UUID
	Limit   int
	Offset  int
}

type ListFilesOutput struct {
	Files []*domain.File
	Total int
}

func (uc *ListFilesUseCase) Execute(ctx context.Context, input ListFilesInput) (*ListFilesOutput, error) {
	if input.Limit <= 0 {
		input.Limit = 20
	}
	if input.Limit > 100 {
		input.Limit = 100
	}
	if input.Offset < 0 {
		input.Offset = 0
	}

	files, total, err := uc.repo.ListByOwner(ctx, input.OwnerID, input.Limit, input.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	return &ListFilesOutput{
		Files: files,
		Total: total,
	}, nil
}
