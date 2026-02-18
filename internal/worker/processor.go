package worker

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/Deymos01/tsv-processing/internal/usecase"
)

// Processor wraps the FileUseCase and executes it for each job.
type Processor struct {
	fileUC usecase.FileUseCase
	log    *zap.Logger
}

// NewProcessor creates a new Processor.
func NewProcessor(fileUC usecase.FileUseCase, log *zap.Logger) *Processor {
	return &Processor{
		fileUC: fileUC,
		log:    log,
	}
}

// Process handles a single Job: registers the file (if needed) and runs the pipeline.
func (p *Processor) Process(ctx context.Context, job Job) error {
	fileName := fileNameFromPath(job.FilePath)

	// Register the file in the DB.
	if _, err := p.fileUC.RegisterFile(ctx, fileName); err != nil {
		if !isDomainErr(err, "already exists") {
			return fmt.Errorf("processor: register file: %w", err)
		}
		p.log.Debug("file already registered, skipping register step",
			zap.String("file", fileName),
		)
	}

	if err := p.fileUC.ProcessFile(ctx, job.FilePath); err != nil {
		return fmt.Errorf("processor: process file: %w", err)
	}

	return nil
}

// fileNameFromPath extracts the base file name from a full path.
func fileNameFromPath(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[i+1:]
		}
	}
	return path
}

// isDomainErr checks wrapped error messages.
func isDomainErr(err error, substr string) bool {
	if err == nil {
		return false
	}
	return contains(err.Error(), substr)
}

// contains is a substring checker.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(substr) == 0 ||
		indexString(s, substr) >= 0)
}

func indexString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
