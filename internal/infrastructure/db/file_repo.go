package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	domain "github.com/companyofcreators/file-service/internal/domain/file"
)

type FileRepository struct {
	db *sqlx.DB
}

func NewFileRepository(db *sqlx.DB) *FileRepository {
	return &FileRepository{db: db}
}

func (r *FileRepository) Create(ctx context.Context, f *domain.File) error {
	query := `
		INSERT INTO files (id, owner_id, bucket, object_key, mime_type, size, file_type, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		f.ID,
		f.OwnerID,
		f.Bucket,
		f.ObjectKey,
		f.MimeType,
		f.Size,
		string(f.FileType),
		f.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert file record: %w", err)
	}

	return nil
}

func (r *FileRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.File, error) {
	query := `
		SELECT id, owner_id, bucket, object_key, mime_type, size, file_type, created_at
		FROM files
		WHERE id = $1
	`

	var f domain.File
	var fileType string
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&f.ID,
		&f.OwnerID,
		&f.Bucket,
		&f.ObjectKey,
		&f.MimeType,
		&f.Size,
		&fileType,
		&f.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to query file by id: %w", err)
	}

	f.FileType = domain.FileType(fileType)

	return &f, nil
}

func (r *FileRepository) ListByOwner(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]*domain.File, int, error) {
	countQuery := `SELECT COUNT(*) FROM files WHERE owner_id = $1`
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, ownerID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count files: %w", err)
	}

	query := `
		SELECT id, owner_id, bucket, object_key, mime_type, size, file_type, created_at
		FROM files
		WHERE owner_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, ownerID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list files: %w", err)
	}
	defer rows.Close()

	var files []*domain.File
	for rows.Next() {
		var f domain.File
		var fileType string
		if err := rows.Scan(
			&f.ID,
			&f.OwnerID,
			&f.Bucket,
			&f.ObjectKey,
			&f.MimeType,
			&f.Size,
			&fileType,
			&f.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan file row: %w", err)
		}
		f.FileType = domain.FileType(fileType)
		files = append(files, &f)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating file rows: %w", err)
	}

	return files, total, nil
}

func (r *FileRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM files WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete file record: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}
