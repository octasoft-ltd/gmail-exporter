package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"gmail-exporter/internal/cleaner"
)

func TestScanExportsDirectory(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "generate_filter_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test email files
	testFiles := []struct {
		filename string
		content  string
	}{
		{"125288cd4bd52814.eml", "test email content"},
		{"125289498b7ee74e.eml", "another email"},
		{"invalid-id.eml", "should be skipped"},
		{"125289e540f06a4e.json", "json email"},
		{"12d7a4179949125c.mbox", "mbox email"},
		{"not-an-email.txt", "should be ignored"},
		{"metrics.json", "should be ignored"},
	}

	for _, tf := range testFiles {
		filePath := filepath.Join(tempDir, tf.filename)
		err := os.WriteFile(filePath, []byte(tf.content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", tf.filename, err)
		}
	}

	// Scan the directory
	processedEmails, err := scanExportsDirectory(tempDir)
	if err != nil {
		t.Fatalf("scanExportsDirectory failed: %v", err)
	}

	// Verify results
	expectedCount := 4 // Only valid Gmail message IDs should be included
	if len(processedEmails) != expectedCount {
		t.Errorf("Expected %d processed emails, got %d", expectedCount, len(processedEmails))
	}

	// Check that valid IDs are included
	expectedIDs := map[string]bool{
		"125288cd4bd52814": true,
		"125289498b7ee74e": true,
		"125289e540f06a4e": true,
		"12d7a4179949125c": true,
	}

	for _, email := range processedEmails {
		if !expectedIDs[email.ID] {
			t.Errorf("Unexpected email ID: %s", email.ID)
		}
		if email.Size == 0 {
			t.Errorf("Email %s should have non-zero size", email.ID)
		}
		if email.Processed.IsZero() {
			t.Errorf("Email %s should have processed timestamp", email.ID)
		}
	}
}

func TestIsValidGmailMessageID(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		expected bool
	}{
		{
			name:     "valid hex ID",
			id:       "125288cd4bd52814",
			expected: true,
		},
		{
			name:     "valid uppercase hex ID",
			id:       "125288CD4BD52814",
			expected: true,
		},
		{
			name:     "valid mixed case hex ID",
			id:       "125288Cd4bD52814",
			expected: true,
		},
		{
			name:     "too short",
			id:       "123",
			expected: false,
		},
		{
			name:     "too long",
			id:       "125288cd4bd52814125288cd4bd52814",
			expected: false,
		},
		{
			name:     "contains invalid characters",
			id:       "125288cd4bd5281g",
			expected: false,
		},
		{
			name:     "contains special characters",
			id:       "125288cd-4bd52814",
			expected: false,
		},
		{
			name:     "empty string",
			id:       "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidGmailMessageID(tt.id)
			if result != tt.expected {
				t.Errorf("isValidGmailMessageID(%s) = %v, want %v", tt.id, result, tt.expected)
			}
		})
	}
}

func TestGenerateFilterIntegration(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "generate_filter_integration_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create subdirectories like real exports
	subDirs := []string{"SENT", "INBOX", "IMPORTANT"}
	for _, subDir := range subDirs {
		err := os.MkdirAll(filepath.Join(tempDir, subDir), 0755)
		if err != nil {
			t.Fatalf("Failed to create subdir %s: %v", subDir, err)
		}
	}

	// Create test email files in subdirectories
	testEmails := []struct {
		dir      string
		filename string
	}{
		{"SENT", "125288cd4bd52814.eml"},
		{"SENT", "125289498b7ee74e.eml"},
		{"INBOX", "125289e540f06a4e.eml"},
		{"IMPORTANT", "12d7a4179949125c.eml"},
	}

	for _, te := range testEmails {
		filePath := filepath.Join(tempDir, te.dir, te.filename)
		err := os.WriteFile(filePath, []byte("test email content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filePath, err)
		}
	}

	// Scan the directory
	processedEmails, err := scanExportsDirectory(tempDir)
	if err != nil {
		t.Fatalf("scanExportsDirectory failed: %v", err)
	}

	// Verify all emails were found
	if len(processedEmails) != len(testEmails) {
		t.Errorf("Expected %d processed emails, got %d", len(testEmails), len(processedEmails))
	}

	// Test writing and reading the filter file
	filterFile := filepath.Join(tempDir, "processed_emails.json")
	data, err := json.MarshalIndent(processedEmails, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal processed emails: %v", err)
	}

	err = os.WriteFile(filterFile, data, 0644)
	if err != nil {
		t.Fatalf("Failed to write filter file: %v", err)
	}

	// Verify the file can be read back
	var loadedEmails []cleaner.ProcessedEmail
	fileData, err := os.ReadFile(filterFile)
	if err != nil {
		t.Fatalf("Failed to read filter file: %v", err)
	}

	err = json.Unmarshal(fileData, &loadedEmails)
	if err != nil {
		t.Fatalf("Failed to unmarshal filter file: %v", err)
	}

	if len(loadedEmails) != len(processedEmails) {
		t.Errorf("Expected %d loaded emails, got %d", len(processedEmails), len(loadedEmails))
	}
}
