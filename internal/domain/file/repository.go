package file

import (
	"context"

	"github.com/google/uuid"
)

type FileRepository interface {
	Create(ctx context.Context, f *File) error
	FindByID(ctx context.Context, id uuid.UUID) (*File, error)
	ListByOwner(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]*File, int, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
