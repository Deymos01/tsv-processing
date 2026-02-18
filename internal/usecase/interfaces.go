package usecase

import (
	"context"

	"github.com/Deymos01/tsv-processing/internal/domain"
)

// MessageUseCase defines operations on messages.
type MessageUseCase interface {
	// GetByUnitGUID returns paginated messages and total count for a given unit_guid.
	GetByUnitGUID(ctx context.Context, filter domain.MessageFilter) ([]domain.Message, int, error)
}

// FileUseCase defines operations on processed files.
type FileUseCase interface {
	// ProcessFile operates the full pipeline for a single TSV file:
	// parse → save to DB → generate output files.
	ProcessFile(ctx context.Context, filePath string) error

	// RegisterFile registers a new file in the DB with status "pending".
	// Returns ErrAlreadyExists if the file is already tracked.
	RegisterFile(ctx context.Context, fileName string) (*domain.ProcessedFile, error)

	// GetTrackedFileNames returns the set of all file names known to the system.
	GetTrackedFileNames(ctx context.Context) (map[string]struct{}, error)
}
