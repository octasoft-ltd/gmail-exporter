package importer

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/api/gmail/v1"

	"gmail-exporter/internal/auth"
	"gmail-exporter/internal/metrics"
)

// Config represents the importer configuration
type Config struct {
	CredentialsFile string `json:"credentials_file"`
	TokenFile       string `json:"token_file"`
	InputDir        string `json:"input_dir"`
	ParallelWorkers int    `json:"parallel_workers"`
	PreserveDates   bool   `json:"preserve_dates"`
	Limit           int    `json:"limit"`
}

// Result represents the import operation result
type Result struct {
	TotalFound    int           `json:"total_found"`
	TotalImported int           `json:"total_imported"`
	TotalFailed   int           `json:"total_failed"`
	TotalSize     int64         `json:"total_size"`
	Duration      time.Duration `json:"duration"`
	Failures      []Failure     `json:"failures,omitempty"`
}

// Failure represents a failed import operation
type Failure struct {
	FilePath  string    `json:"file_path"`
	Error     string    `json:"error"`
	Timestamp time.Time `json:"timestamp"`
}

// Importer handles email import operations
type Importer struct {
	config        *Config
	authenticator *auth.Authenticator
	gmailService  *gmail.Service
	metrics       *metrics.Collector
}

// New creates a new importer instance
func New(config *Config) (*Importer, error) {
	// Validate configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Create authenticator
	authenticator, err := auth.NewAuthenticator(config.CredentialsFile, config.TokenFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create authenticator: %w", err)
	}

	// Get Gmail service
	gmailService, err := authenticator.GetGmailService()
	if err != nil {
		return nil, fmt.Errorf("failed to get Gmail service: %w", err)
	}

	// Create metrics collector
	metricsCollector := metrics.NewCollector("import")

	return &Importer{
		config:        config,
		authenticator: authenticator,
		gmailService:  gmailService,
		metrics:       metricsCollector,
	}, nil
}

// Import performs the email import operation
func (i *Importer) Import() (*Result, error) {
	startTime := time.Now()
	i.metrics.Start()

	logrus.WithFields(logrus.Fields{
		"input_dir": i.config.InputDir,
		"limit":     i.config.Limit,
	}).Info("Starting email import")

	// Find email files
	emailFiles, err := i.findEmailFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to find email files: %w", err)
	}

	logrus.WithField("count", len(emailFiles)).Info("Found email files to import")

	// Apply limit if specified
	if i.config.Limit > 0 && len(emailFiles) > i.config.Limit {
		emailFiles = emailFiles[:i.config.Limit]
		logrus.WithField("limited_count", len(emailFiles)).Info("Limited number of files to process")
	}

	// Import emails
	result, err := i.importEmails(emailFiles)
	if err != nil {
		return nil, fmt.Errorf("failed to import emails: %w", err)
	}

	// Calculate duration
	result.Duration = time.Since(startTime)
	result.TotalFound = len(emailFiles)

	// Record metrics
	i.metrics.RecordEmailsProcessed(result.TotalImported, result.TotalFailed)
	i.metrics.RecordBytesProcessed(result.TotalSize)
	i.metrics.RecordDuration(result.Duration)
	i.metrics.SetTotalMatched(result.TotalFound)

	// Save metrics
	metricsPath := filepath.Join(filepath.Dir(i.config.InputDir), "import_metrics.json")
	if err := i.metrics.Save(metricsPath); err != nil {
		logrus.WithError(err).Warn("Failed to save metrics")
	}

	logrus.WithFields(logrus.Fields{
		"total_found":    result.TotalFound,
		"total_imported": result.TotalImported,
		"total_failed":   result.TotalFailed,
		"duration":       result.Duration,
	}).Info("Import completed")

	return result, nil
}

// findEmailFiles finds all email files in the input directory
func (i *Importer) findEmailFiles() ([]string, error) {
	var emailFiles []string

	err := filepath.WalkDir(i.config.InputDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		// Check for supported email file extensions
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".eml" || ext == ".json" || ext == ".mbox" {
			emailFiles = append(emailFiles, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return emailFiles, nil
}

// importEmails imports the specified email files
func (i *Importer) importEmails(emailFiles []string) (*Result, error) {
	result := &Result{
		Failures: make([]Failure, 0),
	}

	// Create worker pool for parallel processing
	if i.config.ParallelWorkers <= 0 {
		i.config.ParallelWorkers = 1
	}

	jobs := make(chan string, len(emailFiles))
	results := make(chan importResult, len(emailFiles))

	// Start workers
	var wg sync.WaitGroup
	for w := 0; w < i.config.ParallelWorkers; w++ {
		wg.Add(1)
		go i.importWorker(jobs, results, &wg)
	}

	// Send jobs
	for _, filePath := range emailFiles {
		jobs <- filePath
	}
	close(jobs)

	// Wait for workers to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results with progress indicator
	processed := 0
	total := len(emailFiles)
	for importRes := range results {
		processed++

		if importRes.Error != nil {
			result.TotalFailed++
			result.Failures = append(result.Failures, Failure{
				FilePath:  importRes.FilePath,
				Error:     importRes.Error.Error(),
				Timestamp: time.Now(),
			})
			logrus.WithError(importRes.Error).WithField("file_path", importRes.FilePath).Error("Failed to import email")
		} else {
			result.TotalImported++
			result.TotalSize += importRes.Size
		}

		// Show progress
		fmt.Printf("\rProgress: %d of %d messages imported (%.1f%%)",
			result.TotalImported, total, float64(processed)/float64(total)*100)
	}
	fmt.Println() // New line after progress

	return result, nil
}

// importResult represents the result of importing a single email
type importResult struct {
	FilePath string
	Size     int64
	Error    error
}

// importWorker is a worker function for importing emails in parallel
func (i *Importer) importWorker(jobs <-chan string, results chan<- importResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for filePath := range jobs {
		size, err := i.importSingleEmail(filePath)
		results <- importResult{
			FilePath: filePath,
			Size:     size,
			Error:    err,
		}
	}
}

// importSingleEmail imports a single email file
func (i *Importer) importSingleEmail(filePath string) (int64, error) {
	// Read the email file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to read file: %w", err)
	}

	// Determine file type and process accordingly
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".eml":
		return i.importEMLFile(data)
	case ".json":
		return i.importJSONFile(data)
	case ".mbox":
		return i.importMboxFile(data)
	default:
		return 0, fmt.Errorf("unsupported file type: %s", ext)
	}
}

// importEMLFile imports an EML format email
func (i *Importer) importEMLFile(data []byte) (int64, error) {
	// Create a Gmail message from the EML data
	message := &gmail.Message{
		Raw: encodeBase64URL(data),
	}

	// Import the message (does not send, just adds to mailbox)
	_, err := i.gmailService.Users.Messages.Import("me", message).Do()
	if err != nil {
		return 0, fmt.Errorf("failed to import message: %w", err)
	}

	return int64(len(data)), nil
}

// importJSONFile imports a JSON format email
func (i *Importer) importJSONFile(data []byte) (int64, error) {
	// Parse the JSON to extract the raw email data
	var emailData struct {
		Raw string `json:"raw"`
	}

	if err := json.Unmarshal(data, &emailData); err != nil {
		return 0, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Create a Gmail message
	message := &gmail.Message{
		Raw: emailData.Raw,
	}

	// Import the message (does not send, just adds to mailbox)
	_, err := i.gmailService.Users.Messages.Import("me", message).Do()
	if err != nil {
		return 0, fmt.Errorf("failed to import message: %w", err)
	}

	return int64(len(data)), nil
}

// importMboxFile imports an mbox format email
func (i *Importer) importMboxFile(data []byte) (int64, error) {
	// For mbox files, we need to parse the format and extract individual messages
	// This is a simplified implementation - in practice, you'd want a proper mbox parser
	message := &gmail.Message{
		Raw: encodeBase64URL(data),
	}

	// Import the message (does not send, just adds to mailbox)
	_, err := i.gmailService.Users.Messages.Import("me", message).Do()
	if err != nil {
		return 0, fmt.Errorf("failed to import message: %w", err)
	}

	return int64(len(data)), nil
}

// validateConfig validates the importer configuration
func validateConfig(config *Config) error {
	if config.InputDir == "" {
		return fmt.Errorf("input directory is required")
	}

	if _, err := os.Stat(config.InputDir); os.IsNotExist(err) {
		return fmt.Errorf("input directory does not exist: %s", config.InputDir)
	}

	if config.ParallelWorkers < 0 {
		return fmt.Errorf("parallel workers must be >= 0")
	}

	if config.Limit < 0 {
		return fmt.Errorf("limit must be >= 0")
	}

	return nil
}

// encodeBase64URL encodes data in base64url format for Gmail API
func encodeBase64URL(data []byte) string {
	encoded := base64.URLEncoding.EncodeToString(data)
	return strings.TrimRight(encoded, "=")
}
