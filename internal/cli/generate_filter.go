package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/octasoft-ltd/gmail-exporter/internal/cleaner"
)

var generateFilterCmd = &cobra.Command{
	Use:   "generate-filter",
	Short: "Generate a filter file from exported emails directory",
	Long: `Generate a filter file containing processed email IDs from an exports directory.
This file can then be used with the cleanup command to archive or delete the processed emails.

The command scans the exports directory for .eml files and extracts the Gmail message IDs
from the filenames to create a processed_emails.json file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		inputDir, _ := cmd.Flags().GetString("input-dir")
		outputFile, _ := cmd.Flags().GetString("output-file")

		if inputDir == "" {
			return fmt.Errorf("input-dir is required")
		}

		if outputFile == "" {
			outputFile = filepath.Join(inputDir, "processed_emails.json")
		}

		logrus.WithFields(logrus.Fields{
			"input_dir":   inputDir,
			"output_file": outputFile,
		}).Info("Generating filter file from exports directory")

		// Scan for email files and extract IDs
		processedEmails, err := scanExportsDirectory(inputDir)
		if err != nil {
			return fmt.Errorf("failed to scan exports directory: %w", err)
		}

		if len(processedEmails) == 0 {
			return fmt.Errorf("no email files found in directory: %s", inputDir)
		}

		// Write the filter file
		data, err := json.MarshalIndent(processedEmails, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal processed emails: %w", err)
		}

		if err := os.WriteFile(outputFile, data, 0o600); err != nil {
			return fmt.Errorf("failed to write filter file: %w", err)
		}

		fmt.Printf("Filter file generated successfully!\n")
		fmt.Printf("Output file: %s\n", outputFile)
		fmt.Printf("Total emails: %d\n", len(processedEmails))
		fmt.Printf("\nYou can now use this file with the cleanup command:\n")
		fmt.Printf("  ./gmail-exporter cleanup --filter-file %s --action archive --dry-run\n", outputFile)

		return nil
	},
}

func init() {
	generateFilterCmd.Flags().StringP("input-dir", "i", "", "Input directory containing exported emails")
	generateFilterCmd.Flags().StringP("output-file", "o", "", "Output filter file path (default: input-dir/processed_emails.json)")
	if err := generateFilterCmd.MarkFlagRequired("input-dir"); err != nil {
		logrus.WithError(err).Fatal("Failed to mark input-dir flag as required")
	}
}

// scanExportsDirectory scans the exports directory and extracts email IDs from filenames
func scanExportsDirectory(inputDir string) ([]cleaner.ProcessedEmail, error) {
	var processedEmails []cleaner.ProcessedEmail
	now := time.Now()

	err := filepath.WalkDir(inputDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		// Check for email files
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".eml" && ext != ".json" && ext != ".mbox" {
			return nil
		}

		// Extract email ID from filename
		filename := d.Name()
		emailID := strings.TrimSuffix(filename, ext)

		// Validate that it looks like a Gmail message ID (hexadecimal)
		if !isValidGmailMessageID(emailID) {
			logrus.WithField("filename", filename).Debug("Skipping file with invalid Gmail message ID format")
			return nil
		}

		// Get file info for additional metadata
		fileInfo, err := d.Info()
		if err != nil {
			logrus.WithError(err).WithField("path", path).Warn("Failed to get file info")
			// Continue processing, just without the metadata
		}

		processedEmail := cleaner.ProcessedEmail{
			ID:        emailID,
			Processed: now,
		}

		// Add file size if available
		if fileInfo != nil {
			processedEmail.Size = fileInfo.Size()
		}

		processedEmails = append(processedEmails, processedEmail)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return processedEmails, nil
}

// isValidGmailMessageID checks if a string looks like a valid Gmail message ID
func isValidGmailMessageID(id string) bool {
	if len(id) < 10 || len(id) > 20 {
		return false
	}

	// Gmail message IDs are typically hexadecimal
	for _, char := range id {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
			return false
		}
	}

	return true
}
