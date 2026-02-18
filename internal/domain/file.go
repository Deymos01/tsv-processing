package domain

import "time"

// FileStatus represents the processing state of an input file.
type FileStatus string

const (
	FileStatusPending    FileStatus = "pending"
	FileStatusProcessing FileStatus = "processing"
	FileStatusDone       FileStatus = "done"
	FileStatusError      FileStatus = "error"
)

// ProcessedFile tracks an input TSV file and its processing state.
type ProcessedFile struct {
	ID          int64
	FileName    string
	Status      FileStatus
	ErrorDetail string
	ProcessedAt time.Time // zero value means not yet processed
	CreatedAt   time.Time
}

// IsDone reports whether the file has been successfully processed.
func (f ProcessedFile) IsDone() bool {
	return f.Status == FileStatusDone
}

// HasError reports whether the file processing ended with an error.
func (f ProcessedFile) HasError() bool {
	return f.Status == FileStatusError
}
