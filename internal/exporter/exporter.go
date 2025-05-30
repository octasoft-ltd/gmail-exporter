package exporter

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/api/gmail/v1"

	"gmail-exporter/internal/auth"
	"gmail-exporter/internal/filters"
	"gmail-exporter/internal/metrics"
)

// Config represents the exporter configuration
type Config struct {
	CredentialsFile    string `json:"credentials_file"`
	TokenFile          string `json:"token_file"`
	OutputDir          string `json:"output_dir"`
	OrganizeByLabels   bool   `json:"organize_by_labels"`
	ParallelWorkers    int    `json:"parallel_workers"`
	IncludeAttachments bool   `json:"include_attachments"`
	CompressExports    bool   `json:"compress_exports"`
	Format             string `json:"format"`
	Resume             bool   `json:"resume"`
	StateFile          string `json:"state_file"`
	Limit              int    `json:"limit"`
}

// Result represents the export operation result
type Result struct {
	TotalMatched  int           `json:"total_matched"`
	TotalExported int           `json:"total_exported"`
	TotalFailed   int           `json:"total_failed"`
	TotalSize     int64         `json:"total_size"`
	Duration      time.Duration `json:"duration"`
	Failures      []Failure     `json:"failures,omitempty"`
}

// Failure represents a failed export operation
type Failure struct {
	EmailID   string    `json:"email_id"`
	Error     string    `json:"error"`
	Timestamp time.Time `json:"timestamp"`
}

// ProcessedEmail represents an email that was successfully processed during export
type ProcessedEmail struct {
	ID        string    `json:"id"`
	Subject   string    `json:"subject,omitempty"`
	From      string    `json:"from,omitempty"`
	Date      time.Time `json:"date,omitempty"`
	Size      int64     `json:"size,omitempty"`
	Processed time.Time `json:"processed"`
}

// Exporter handles email export operations
type Exporter struct {
	config        *Config
	authenticator *auth.Authenticator
	gmailService  *gmail.Service
	metrics       *metrics.Collector
}

// New creates a new exporter instance
func New(config *Config) (*Exporter, error) {
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
	metricsCollector := metrics.NewCollector("export")

	return &Exporter{
		config:        config,
		authenticator: authenticator,
		gmailService:  gmailService,
		metrics:       metricsCollector,
	}, nil
}

// Export performs the email export operation
func (e *Exporter) Export(filterConfig *filters.Config) (*Result, error) {
	startTime := time.Now()
	e.metrics.Start()

	logrus.WithField("query", filterConfig.BuildGmailQuery()).Info("Starting export with Gmail query")

	// Validate filter configuration
	if err := filterConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid filter configuration: %w", err)
	}

	// Create output directory
	if err := os.MkdirAll(e.config.OutputDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Search for emails
	messageIDs, err := e.searchEmails(filterConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to search emails: %w", err)
	}

	logrus.WithField("count", len(messageIDs)).Info("Found emails matching filter")

	// Apply limit if specified
	if e.config.Limit > 0 && len(messageIDs) > e.config.Limit {
		messageIDs = messageIDs[:e.config.Limit]
		logrus.WithField("limited_count", len(messageIDs)).Info("Limited number of emails to process")
	}

	// Set total matched in metrics
	e.metrics.SetTotalMatched(len(messageIDs))

	// Export emails
	result, err := e.exportEmails(messageIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to export emails: %w", err)
	}

	// Calculate duration
	result.Duration = time.Since(startTime)
	result.TotalMatched = len(messageIDs)

	// Record metrics
	e.metrics.RecordEmailsProcessed(result.TotalExported, result.TotalFailed)
	e.metrics.RecordBytesProcessed(result.TotalSize)
	e.metrics.RecordDuration(result.Duration)

	// Save metrics
	if err := e.metrics.Save(filepath.Join(e.config.OutputDir, "metrics.json")); err != nil {
		logrus.WithError(err).Warn("Failed to save metrics")
	}

	logrus.WithFields(logrus.Fields{
		"total_matched":  result.TotalMatched,
		"total_exported": result.TotalExported,
		"total_failed":   result.TotalFailed,
		"duration":       result.Duration,
	}).Info("Export completed")

	return result, nil
}

// searchEmails searches for emails matching the filter criteria
func (e *Exporter) searchEmails(filterConfig *filters.Config) ([]string, error) {
	query := filterConfig.BuildGmailQuery()

	var messageIDs []string
	pageToken := ""

	for {
		req := e.gmailService.Users.Messages.List("me").Q(query)
		if pageToken != "" {
			req = req.PageToken(pageToken)
		}

		resp, err := req.Do()
		if err != nil {
			return nil, fmt.Errorf("failed to list messages: %w", err)
		}

		for _, message := range resp.Messages {
			messageIDs = append(messageIDs, message.Id)
		}

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return messageIDs, nil
}

// exportEmails exports the specified emails
func (e *Exporter) exportEmails(messageIDs []string) (*Result, error) {
	result := &Result{
		Failures: make([]Failure, 0),
	}

	// Track successfully processed emails for filter file
	var processedEmails []ProcessedEmail

	// Create worker pool for parallel processing
	if e.config.ParallelWorkers <= 0 {
		e.config.ParallelWorkers = 1
	}

	jobs := make(chan string, len(messageIDs))
	results := make(chan exportResult, len(messageIDs))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < e.config.ParallelWorkers; i++ {
		wg.Add(1)
		go e.exportWorker(jobs, results, &wg)
	}

	// Send jobs
	for _, messageID := range messageIDs {
		jobs <- messageID
	}
	close(jobs)

	// Wait for workers to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results with progress indicator
	processed := 0
	total := len(messageIDs)
	for exportRes := range results {
		processed++

		if exportRes.Error != nil {
			result.TotalFailed++
			result.Failures = append(result.Failures, Failure{
				EmailID:   exportRes.MessageID,
				Error:     exportRes.Error.Error(),
				Timestamp: time.Now(),
			})
			logrus.WithError(exportRes.Error).WithField("message_id", exportRes.MessageID).Error("Failed to export email")
		} else {
			result.TotalExported++
			result.TotalSize += exportRes.Size

			// Add to processed emails for filter file
			processedEmails = append(processedEmails, ProcessedEmail{
				ID:        exportRes.MessageID,
				Size:      exportRes.Size,
				Processed: time.Now(),
			})
		}

		// Show progress
		fmt.Printf("\rProgress: %d of %d messages exported (%.1f%%)",
			result.TotalExported, total, float64(processed)/float64(total)*100)
	}
	fmt.Println() // New line after progress

	// Save processed emails filter file
	if len(processedEmails) > 0 {
		if err := e.saveProcessedEmailsFilter(processedEmails); err != nil {
			logrus.WithError(err).Warn("Failed to save processed emails filter file")
		}
	}

	return result, nil
}

// exportResult represents the result of exporting a single email
type exportResult struct {
	MessageID string
	Size      int64
	Error     error
}

// exportWorker is a worker function for exporting emails in parallel
func (e *Exporter) exportWorker(jobs <-chan string, results chan<- exportResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for messageID := range jobs {
		size, err := e.exportSingleEmail(messageID)
		results <- exportResult{
			MessageID: messageID,
			Size:      size,
			Error:     err,
		}
	}
}

// exportSingleEmail exports a single email
func (e *Exporter) exportSingleEmail(messageID string) (int64, error) {
	// Get the full message
	message, err := e.gmailService.Users.Messages.Get("me", messageID).Format("full").Do()
	if err != nil {
		return 0, fmt.Errorf("failed to get message: %w", err)
	}

	// Determine output path
	outputPath, err := e.getOutputPath(message)
	if err != nil {
		return 0, fmt.Errorf("failed to determine output path: %w", err)
	}

	// Export based on format
	var size int64
	switch e.config.Format {
	case "eml":
		size, err = e.exportAsEML(message, outputPath)
	case "json":
		size, err = e.exportAsJSON(message, outputPath)
	case "mbox":
		size, err = e.exportAsMbox(message, outputPath)
	default:
		return 0, fmt.Errorf("unsupported export format: %s", e.config.Format)
	}

	if err != nil {
		return 0, err
	}

	return size, nil
}

// getOutputPath determines the output path for an email
func (e *Exporter) getOutputPath(message *gmail.Message) (string, error) {
	// Create base filename from message ID and timestamp
	filename := fmt.Sprintf("%s.%s", message.Id, e.config.Format)

	if !e.config.OrganizeByLabels {
		return filepath.Join(e.config.OutputDir, filename), nil
	}

	// Organize by labels
	labelDir := "unlabeled"
	if len(message.LabelIds) > 0 {
		// Use the first label for directory structure
		// In a real implementation, you might want to get label names from the API
		labelDir = message.LabelIds[0]
	}

	outputDir := filepath.Join(e.config.OutputDir, labelDir)
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create label directory: %w", err)
	}

	return filepath.Join(outputDir, filename), nil
}

// exportAsEML exports an email in EML format
func (e *Exporter) exportAsEML(message *gmail.Message, outputPath string) (int64, error) {
	// Get the raw message
	rawMessage, err := e.gmailService.Users.Messages.Get("me", message.Id).Format("raw").Do()
	if err != nil {
		return 0, fmt.Errorf("failed to get raw message: %w", err)
	}

	// Decode the raw message
	rawData, err := decodeBase64URL(rawMessage.Raw)
	if err != nil {
		return 0, fmt.Errorf("failed to decode raw message: %w", err)
	}

	// Write to file
	if err := os.WriteFile(outputPath, rawData, 0o600); err != nil {
		return 0, fmt.Errorf("failed to write EML file: %w", err)
	}

	return int64(len(rawData)), nil
}

// exportAsJSON exports an email in JSON format
func (e *Exporter) exportAsJSON(message *gmail.Message, outputPath string) (int64, error) {
	// Convert message to JSON
	jsonData, err := json.MarshalIndent(message, "", "  ")
	if err != nil {
		return 0, fmt.Errorf("failed to marshal message to JSON: %w", err)
	}

	// Write to file
	if err := os.WriteFile(outputPath, jsonData, 0o600); err != nil {
		return 0, fmt.Errorf("failed to write JSON file: %w", err)
	}

	return int64(len(jsonData)), nil
}

// exportAsMbox exports an email in Mbox format
func (e *Exporter) exportAsMbox(message *gmail.Message, outputPath string) (int64, error) {
	// This is a simplified implementation
	// In a real implementation, you would properly format the mbox
	return e.exportAsEML(message, outputPath)
}

// validateConfig validates the exporter configuration
func validateConfig(config *Config) error {
	if config.CredentialsFile == "" {
		return fmt.Errorf("credentials file is required")
	}
	if config.TokenFile == "" {
		return fmt.Errorf("token file is required")
	}
	if config.OutputDir == "" {
		return fmt.Errorf("output directory is required")
	}
	if config.ParallelWorkers < 0 {
		return fmt.Errorf("parallel workers must be >= 0")
	}
	if config.Format == "" {
		config.Format = "eml"
	}

	validFormats := []string{"eml", "json", "mbox"}
	valid := false
	for _, format := range validFormats {
		if config.Format == format {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid format: %s (valid: eml, json, mbox)", config.Format)
	}

	return nil
}

// decodeBase64URL decodes a base64url encoded string
func decodeBase64URL(data string) ([]byte, error) {
	// Add padding if necessary
	switch len(data) % 4 {
	case 2:
		data += "=="
	case 3:
		data += "="
	}

	// Replace URL-safe characters
	data = strings.ReplaceAll(data, "-", "+")
	data = strings.ReplaceAll(data, "_", "/")

	return base64.StdEncoding.DecodeString(data)
}

// saveProcessedEmailsFilter saves the list of processed emails to a filter file
func (e *Exporter) saveProcessedEmailsFilter(processedEmails []ProcessedEmail) error {
	filterFile := filepath.Join(e.config.OutputDir, "processed_emails.json")

	data, err := json.MarshalIndent(processedEmails, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal processed emails: %w", err)
	}

	if err := os.WriteFile(filterFile, data, 0o600); err != nil {
		return fmt.Errorf("failed to write filter file: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"filter_file": filterFile,
		"count":       len(processedEmails),
	}).Info("Saved processed emails filter file")

	return nil
}
