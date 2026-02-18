package usecase

import (
	"context"
	"fmt"
	"path/filepath"

	"go.uber.org/zap"

	"github.com/Deymos01/tsv-processing/internal/domain"
	"github.com/Deymos01/tsv-processing/internal/repository"
	"github.com/Deymos01/tsv-processing/internal/usecase/generator"
	"github.com/Deymos01/tsv-processing/internal/usecase/parser"
)

// fileUseCase implements FileUseCase.
type fileUseCase struct {
	fileRepo    repository.FileRepository
	messageRepo repository.MessageRepository
	generator   generator.OutputGenerator
	log         *zap.Logger
}

// NewFileUseCase constructs a new fileUseCase.
func NewFileUseCase(
	fileRepo repository.FileRepository,
	messageRepo repository.MessageRepository,
	gen generator.OutputGenerator,
	log *zap.Logger,
) FileUseCase {
	return &fileUseCase{
		fileRepo:    fileRepo,
		messageRepo: messageRepo,
		generator:   gen,
		log:         log,
	}
}

// RegisterFile registers a file in the DB with status "pending".
func (uc *fileUseCase) RegisterFile(
	ctx context.Context,
	fileName string,
) (*domain.ProcessedFile, error) {
	pf, err := uc.fileRepo.Create(ctx, fileName)
	if err != nil {
		return nil, fmt.Errorf("register file %q: %w", fileName, err)
	}
	return pf, nil
}

// GetTrackedFileNames returns the set of all tracked file names.
func (uc *fileUseCase) GetTrackedFileNames(ctx context.Context) (map[string]struct{}, error) {
	names, err := uc.fileRepo.GetAllFileNames(ctx)
	if err != nil {
		return nil, fmt.Errorf("get tracked file names: %w", err)
	}
	return names, nil
}

// ProcessFile runs the full pipeline for one TSV file:
//  1. Mark file as "processing"
//  2. Parse TSV rows
//  3. Save messages to DB
//  4. Generate one RTF output file per unique unit_guid
//  5. Mark file as "done" (or "error" on failure)
func (uc *fileUseCase) ProcessFile(ctx context.Context, filePath string) error {
	fileName := filepath.Base(filePath)
	log := uc.log.With(zap.String("file", fileName))

	// Fetch the file record.
	pf, err := uc.fileRepo.GetByName(ctx, fileName)
	if err != nil {
		return fmt.Errorf("process file: fetch record: %w", err)
	}

	// Mark as processing.
	if err = uc.fileRepo.UpdateStatus(ctx, pf.ID, domain.FileStatusProcessing, ""); err != nil {
		return fmt.Errorf("process file: mark processing: %w", err)
	}

	log.Info("processing started")

	// Parse TSV. Capture parse errors explicitly.
	messages, parseErr := parser.ParseTSV(filePath, fileName)

	if len(messages) > 0 {
		if err = uc.messageRepo.BulkInsert(ctx, messages); err != nil {
			// DB insert failure is fatal.
			detail := fmt.Sprintf("bulk insert failed: %s", err.Error())
			_ = uc.fileRepo.UpdateStatus(ctx, pf.ID, domain.FileStatusError, detail)
			return fmt.Errorf("process file: %w", err)
		}
		log.Info("messages saved", zap.Int("count", len(messages)))
	}

	if parseErr != nil {
		// Parsing failed on some (or all) rows.
		log.Warn("parse errors encountered", zap.Error(parseErr))

		detail := parseErr.Error()
		if err = uc.fileRepo.UpdateStatus(ctx, pf.ID, domain.FileStatusError, detail); err != nil {
			return fmt.Errorf("process file: mark error: %w", err)
		}

		// Generate an error report RTF.
		if genErr := uc.generator.GenerateErrorReport(fileName, parseErr); genErr != nil {
			log.Error("failed to generate error report", zap.Error(genErr))
		}

		return fmt.Errorf("process file %q: %w", fileName, parseErr)
	}

	// Generate one output RTF per unique unit_guid found in this file.
	guids, err := uc.messageRepo.GetDistinctUnitGUIDs(ctx, fileName)
	if err != nil {
		detail := fmt.Sprintf("get distinct unit_guids failed: %s", err.Error())
		_ = uc.fileRepo.UpdateStatus(ctx, pf.ID, domain.FileStatusError, detail)
		return fmt.Errorf("process file: %w", err)
	}

	for _, guid := range guids {
		msgs, _, err := uc.messageRepo.GetByUnitGUID(ctx, domain.MessageFilter{
			UnitGUID: guid,
			Page:     1,
			Limit:    100_000, // fetch all rows for this guid to build the report
		})
		if err != nil {
			log.Error("failed to fetch messages for guid",
				zap.String("unit_guid", guid),
				zap.Error(err),
			)
			continue
		}

		if err = uc.generator.GenerateReport(guid, msgs); err != nil {
			log.Error("failed to generate report",
				zap.String("unit_guid", guid),
				zap.Error(err),
			)
		}
	}

	if err = uc.fileRepo.UpdateStatus(ctx, pf.ID, domain.FileStatusDone, ""); err != nil {
		return fmt.Errorf("process file: mark done: %w", err)
	}

	log.Info("processing finished", zap.Int("unit_guids", len(guids)))
	return nil
}
