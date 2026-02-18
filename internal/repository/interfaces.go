package repository

import (
	"context"

	"github.com/Deymos01/tsv-processor/internal/domain"
)

// MessageRepository defines persistence operations for messages.
type MessageRepository interface {
	// BulkInsert saves a batch of messages in a single transaction.
	BulkInsert(ctx context.Context, messages []domain.Message) error

	// GetByUnitGUID returns paginated messages filtered by unit_guid.
	GetByUnitGUID(ctx context.Context, filter domain.MessageFilter) ([]domain.Message, int, error)

	// GetDistinctUnitGUIDs returns all unique unit_guid values present in a given file.
	GetDistinctUnitGUIDs(ctx context.Context, sourceFile string) ([]string, error)
}

// FileRepository defines persistence operations for tracking processed files.
type FileRepository interface {
	// Create registers a new file with status "pending".
	Create(ctx context.Context, fileName string) (*domain.ProcessedFile, error)

	// GetByName returns a file record by its file name.
	GetByName(ctx context.Context, fileName string) (*domain.ProcessedFile, error)

	// GetAllFileNames returns the set of all tracked file names.
	// Used by the scanner to skip already-known files.
	GetAllFileNames(ctx context.Context) (map[string]struct{}, error)

	// UpdateStatus updates the file status and optionally sets error_detail.
	UpdateStatus(ctx context.Context, id int64, status domain.FileStatus, errDetail string) error
}
