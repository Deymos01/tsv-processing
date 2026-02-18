package worker

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"

	"github.com/Deymos01/tsv-processing/internal/usecase"
)

// Scanner periodically watches the input directory for new .tsv files
// and enqueues them into the worker pool.
type Scanner struct {
	inputDir string
	interval time.Duration
	pool     *Pool
	fileUC   usecase.FileUseCase
	log      *zap.Logger
}

// NewScanner creates a new Scanner.
func NewScanner(
	inputDir string,
	interval time.Duration,
	pool *Pool,
	fileUC usecase.FileUseCase,
	log *zap.Logger,
) *Scanner {
	return &Scanner{
		inputDir: inputDir,
		interval: interval,
		pool:     pool,
		fileUC:   fileUC,
		log:      log,
	}
}

// Run starts the scanner loop. It scans immediately on start, then on each tick.
// Exits when ctx is cancelled.
func (s *Scanner) Run(ctx context.Context) {
	s.log.Info("scanner started",
		zap.String("input_dir", s.inputDir),
		zap.Duration("interval", s.interval),
	)

	// Scan immediately on startup, then on each tick.
	s.scan(ctx)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.scan(ctx)
		case <-ctx.Done():
			s.log.Info("scanner stopped")
			return
		}
	}
}

// scan reads the input directory, compares against already-tracked files,
// and enqueues any new .tsv files.
func (s *Scanner) scan(ctx context.Context) {
	s.log.Debug("scanning input directory")

	tracked, err := s.fileUC.GetTrackedFileNames(ctx)
	if err != nil {
		s.log.Error("failed to get tracked file names", zap.Error(err))
		return
	}

	entries, err := os.ReadDir(s.inputDir)
	if err != nil {
		s.log.Error("failed to read input directory",
			zap.String("dir", s.inputDir),
			zap.Error(err),
		)
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".tsv" {
			continue
		}
		if _, ok := tracked[entry.Name()]; ok {
			// Already known — skip regardless of current status.
			continue
		}

		fullPath := filepath.Join(s.inputDir, entry.Name())
		job := Job{FilePath: fullPath}

		s.log.Info("new file detected, enqueuing", zap.String("file", entry.Name()))

		// Register the file immediately.
		if _, err = s.fileUC.RegisterFile(ctx, entry.Name()); err != nil {
			s.log.Error("failed to register file",
				zap.String("file", entry.Name()),
				zap.Error(err),
			)
			continue
		}

		if !s.pool.Enqueue(ctx, job) {
			// Context was cancelled
			s.log.Warn("enqueue failed, stopping scan")
			return
		}
	}
}
