package cleaner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/api/gmail/v1"

	"gmail-exporter/internal/auth"
	"gmail-exporter/internal/metrics"
)

// Config represents the cleaner configuration
type Config struct {
	CredentialsFile string `json:"credentials_file"`
	TokenFile       string `json:"token_file"`
	Action          string `json:"action"` // "archive" or "delete"
	FilterFile      string `json:"filter_file"`
	DryRun          bool   `json:"dry_run"`
	Limit           int    `json:"limit"`
}

// Result represents the cleanup operation result
type Result struct {
	TotalFound     int           `json:"total_found"`
	TotalProcessed int           `json:"total_processed"`
	TotalFailed    int           `json:"total_failed"`
	Duration       time.Duration `json:"duration"`
	Action         string        `json:"action"`
	DryRun         bool          `json:"dry_run"`
	Failures       []Failure     `json:"failures,omitempty"`
}

// Failure represents a failed cleanup operation
type Failure struct {
	EmailID   string    `json:"email_id"`
	Error     string    `json:"error"`
	Timestamp time.Time `json:"timestamp"`
}

// ProcessedEmail represents an email that was processed during export/import
type ProcessedEmail struct {
	ID        string    `json:"id"`
	Subject   string    `json:"subject,omitempty"`
	From      string    `json:"from,omitempty"`
	Date      time.Time `json:"date,omitempty"`
	Size      int64     `json:"size,omitempty"`
	Processed time.Time `json:"processed"`
}

// Cleaner handles email cleanup operations
type Cleaner struct {
	config        *Config
	authenticator *auth.Authenticator
	gmailService  *gmail.Service
	metrics       *metrics.Collector
}

// New creates a new cleaner instance
func New(config *Config) (*Cleaner, error) {
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
	metricsCollector := metrics.NewCollector("cleanup")

	return &Cleaner{
		config:        config,
		authenticator: authenticator,
		gmailService:  gmailService,
		metrics:       metricsCollector,
	}, nil
}

// Cleanup performs the email cleanup operation
func (c *Cleaner) Cleanup() (*Result, error) {
	startTime := time.Now()
	c.metrics.Start()

	logrus.WithFields(logrus.Fields{
		"action":      c.config.Action,
		"filter_file": c.config.FilterFile,
		"dry_run":     c.config.DryRun,
		"limit":       c.config.Limit,
	}).Info("Starting email cleanup")

	// Load processed emails from filter file
	processedEmails, err := c.loadProcessedEmails()
	if err != nil {
		return nil, fmt.Errorf("failed to load processed emails: %w", err)
	}

	logrus.WithField("count", len(processedEmails)).Info("Found processed emails to clean up")

	// Apply limit if specified
	if c.config.Limit > 0 && len(processedEmails) > c.config.Limit {
		processedEmails = processedEmails[:c.config.Limit]
		logrus.WithField("limited_count", len(processedEmails)).Info("Limited number of emails to process")
	}

	// Perform cleanup
	result, err := c.cleanupEmails(processedEmails)
	if err != nil {
		return nil, fmt.Errorf("failed to cleanup emails: %w", err)
	}

	// Calculate duration
	result.Duration = time.Since(startTime)
	result.TotalFound = len(processedEmails)
	result.Action = c.config.Action
	result.DryRun = c.config.DryRun

	// Record metrics
	c.metrics.RecordEmailsProcessed(result.TotalProcessed, result.TotalFailed)
	c.metrics.RecordDuration(result.Duration)
	c.metrics.SetTotalMatched(result.TotalFound)

	// Save metrics
	metricsPath := filepath.Join(filepath.Dir(c.config.FilterFile), "cleanup_metrics.json")
	if err := c.metrics.Save(metricsPath); err != nil {
		logrus.WithError(err).Warn("Failed to save metrics")
	}

	logrus.WithFields(logrus.Fields{
		"total_found":     result.TotalFound,
		"total_processed": result.TotalProcessed,
		"total_failed":    result.TotalFailed,
		"action":          result.Action,
		"dry_run":         result.DryRun,
		"duration":        result.Duration,
	}).Info("Cleanup completed")

	return result, nil
}

// loadProcessedEmails loads the list of processed emails from the filter file
func (c *Cleaner) loadProcessedEmails() ([]ProcessedEmail, error) {
	data, err := os.ReadFile(c.config.FilterFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read filter file: %w", err)
	}

	var processedEmails []ProcessedEmail
	if err := json.Unmarshal(data, &processedEmails); err != nil {
		return nil, fmt.Errorf("failed to parse filter file: %w", err)
	}

	return processedEmails, nil
}

// cleanupEmails performs cleanup on the specified emails
func (c *Cleaner) cleanupEmails(processedEmails []ProcessedEmail) (*Result, error) {
	result := &Result{
		Failures: make([]Failure, 0),
	}

	// Process emails with progress indicator
	total := len(processedEmails)
	for i, email := range processedEmails {
		err := c.cleanupSingleEmail(email.ID)

		if err != nil {
			result.TotalFailed++
			result.Failures = append(result.Failures, Failure{
				EmailID:   email.ID,
				Error:     err.Error(),
				Timestamp: time.Now(),
			})
			logrus.WithError(err).WithField("email_id", email.ID).Error("Failed to cleanup email")
		} else {
			result.TotalProcessed++
		}

		// Show progress
		processed := i + 1
		fmt.Printf("\rProgress: %d of %d messages %s (%.1f%%)",
			processed, total, c.getActionVerb(), float64(processed)/float64(total)*100)
	}
	fmt.Println() // New line after progress

	return result, nil
}

// cleanupSingleEmail performs cleanup on a single email
func (c *Cleaner) cleanupSingleEmail(emailID string) error {
	if c.config.DryRun {
		logrus.WithFields(logrus.Fields{
			"email_id": emailID,
			"action":   c.config.Action,
		}).Info("DRY RUN: Would perform cleanup action")
		return nil
	}

	switch c.config.Action {
	case "archive":
		return c.archiveEmail(emailID)
	case "delete":
		return c.deleteEmail(emailID)
	default:
		return fmt.Errorf("unsupported action: %s", c.config.Action)
	}
}

// archiveEmail archives a single email
func (c *Cleaner) archiveEmail(emailID string) error {
	// Remove the INBOX label to archive the email
	modifyRequest := &gmail.ModifyMessageRequest{
		RemoveLabelIds: []string{"INBOX"},
	}

	_, err := c.gmailService.Users.Messages.Modify("me", emailID, modifyRequest).Do()
	if err != nil {
		return fmt.Errorf("failed to archive email: %w", err)
	}

	return nil
}

// deleteEmail deletes a single email
func (c *Cleaner) deleteEmail(emailID string) error {
	err := c.gmailService.Users.Messages.Delete("me", emailID).Do()
	if err != nil {
		return fmt.Errorf("failed to delete email: %w", err)
	}

	return nil
}

// getActionVerb returns the appropriate verb for the action
func (c *Cleaner) getActionVerb() string {
	switch c.config.Action {
	case "archive":
		return "archived"
	case "delete":
		return "deleted"
	default:
		return "processed"
	}
}

// validateConfig validates the cleaner configuration
func validateConfig(config *Config) error {
	if config.Action == "" {
		config.Action = "archive" // Default action
	}

	if config.Action != "archive" && config.Action != "delete" {
		return fmt.Errorf("action must be 'archive' or 'delete', got: %s", config.Action)
	}

	if config.FilterFile == "" {
		return fmt.Errorf("filter file is required")
	}

	if _, err := os.Stat(config.FilterFile); os.IsNotExist(err) {
		return fmt.Errorf("filter file does not exist: %s", config.FilterFile)
	}

	if config.Limit < 0 {
		return fmt.Errorf("limit must be >= 0")
	}

	return nil
}
