package generator

import "github.com/Deymos01/tsv-processing/internal/domain"

// OutputGenerator defines output file generation operations.
type OutputGenerator interface {
	// GenerateReport creates an RTF file named <unitGUID>.rtf containing all messages.
	GenerateReport(unitGUID string, messages []domain.Message) error

	// GenerateErrorReport creates an RTF file describing a parse failure.
	GenerateErrorReport(sourceFile string, parseErr error) error
}
