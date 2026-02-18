package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Deymos01/tsv-processing/internal/domain"
)

// FileRepo implements repository.FileRepository.
type FileRepo struct {
	pool *pgxpool.Pool
}

// NewFileRepo creates a new FileRepo.
func NewFileRepo(pool *pgxpool.Pool) *FileRepo {
	return &FileRepo{pool: pool}
}

// Create inserts a new file record with status "pending".
// Returns ErrAlreadyExists if a record with the same file_name exists.
func (r *FileRepo) Create(ctx context.Context, fileName string) (*domain.ProcessedFile, error) {
	const q = `
		INSERT INTO processed_files (file_name, status)
		VALUES ($1, $2)
		ON CONFLICT (file_name) DO NOTHING
		RETURNING id, file_name, status, error_detail, processed_at, created_at`

	row := r.pool.QueryRow(ctx, q, fileName, domain.FileStatusPending)

	pf, err := scanProcessedFile(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("create processed file: %w", domain.ErrAlreadyExists)
		}
		return nil, fmt.Errorf("create processed file: %w", err)
	}

	return pf, nil
}

// GetByName fetches a processed file record by its file name.
func (r *FileRepo) GetByName(ctx context.Context, fileName string) (*domain.ProcessedFile, error) {
	const q = `
		SELECT id, file_name, status, error_detail, processed_at, created_at
		FROM processed_files
		WHERE file_name = $1`

	row := r.pool.QueryRow(ctx, q, fileName)

	pf, err := scanProcessedFile(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("get processed file: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("get processed file: %w", err)
	}

	return pf, nil
}

// GetAllFileNames returns all tracked file names as a set.
func (r *FileRepo) GetAllFileNames(ctx context.Context) (map[string]struct{}, error) {
	const q = `SELECT file_name FROM processed_files`

	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("get all file names: %w", err)
	}
	defer rows.Close()

	result := make(map[string]struct{})
	for rows.Next() {
		var name string
		if err = rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scan file name: %w", err)
		}
		result[name] = struct{}{}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating file names: %w", err)
	}

	return result, nil
}

// UpdateStatus sets the status (and optionally error_detail) of a processed file.
func (r *FileRepo) UpdateStatus(
	ctx context.Context,
	id int64,
	status domain.FileStatus,
	errDetail string,
) error {
	var processedAt *time.Time
	if status == domain.FileStatusDone || status == domain.FileStatusError {
		t := time.Now()
		processedAt = &t
	}

	const q = `
		UPDATE processed_files
		SET status       = $1,
		    error_detail = $2,
		    processed_at = $3
		WHERE id = $4`

	tag, err := r.pool.Exec(ctx, q, status, errDetail, processedAt, id)
	if err != nil {
		return fmt.Errorf("update file status: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("update file status: %w", domain.ErrNotFound)
	}

	return nil
}

// scanProcessedFile scans a single row into a ProcessedFile domain model.
func scanProcessedFile(row pgx.Row) (*domain.ProcessedFile, error) {
	var pf domain.ProcessedFile
	var processedAt *time.Time

	err := row.Scan(
		&pf.ID,
		&pf.FileName,
		&pf.Status,
		&pf.ErrorDetail,
		&processedAt,
		&pf.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if processedAt != nil {
		pf.ProcessedAt = *processedAt
	}

	return &pf, nil
}
