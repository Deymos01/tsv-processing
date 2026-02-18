package worker

// Job represents a single unit of work to be processed by the pool.
type Job struct {
	// FilePath is the absolute path to the TSV file to process.
	FilePath string
}
