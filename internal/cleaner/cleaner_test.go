package cleaner

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	// Create temporary filter file for testing
	tempDir, err := os.MkdirTemp("", "cleaner_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	filterFile := filepath.Join(tempDir, "processed.json")
	processedEmails := []ProcessedEmail{
		{
			ID:        "test123",
			Subject:   "Test Email",
			From:      "test@example.com",
			Date:      time.Now(),
			Size:      1024,
			Processed: time.Now(),
		},
	}

	data, err := json.Marshal(processedEmails)
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	err = os.WriteFile(filterFile, data, 0644)
	if err != nil {
		t.Fatalf("Failed to write filter file: %v", err)
	}

	config := &Config{
		CredentialsFile: "test_credentials.json",
		TokenFile:       "test_token.json",
		Action:          "archive",
		FilterFile:      filterFile,
		DryRun:          true,
		Limit:           0,
	}

	// This will fail because we don't have valid credentials, but we can test validation
	_, err = New(config)
	if err == nil {
		t.Error("Expected error for invalid credentials file")
	}
}

func TestValidateConfig(t *testing.T) {
	// Create temporary filter file for testing
	tempDir, err := os.MkdirTemp("", "cleaner_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	validFilterFile := filepath.Join(tempDir, "valid.json")
	err = os.WriteFile(validFilterFile, []byte("[]"), 0644)
	if err != nil {
		t.Fatalf("Failed to create valid filter file: %v", err)
	}

	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name: "valid config with archive",
			config: &Config{
				Action:     "archive",
				FilterFile: validFilterFile,
				DryRun:     false,
				Limit:      0,
			},
			expectError: false,
		},
		{
			name: "valid config with delete",
			config: &Config{
				Action:     "delete",
				FilterFile: validFilterFile,
				DryRun:     true,
				Limit:      5,
			},
			expectError: false,
		},
		{
			name: "default action",
			config: &Config{
				FilterFile: validFilterFile,
			},
			expectError: false,
		},
		{
			name: "invalid action",
			config: &Config{
				Action:     "invalid",
				FilterFile: validFilterFile,
			},
			expectError: true,
		},
		{
			name: "missing filter file",
			config: &Config{
				Action: "archive",
			},
			expectError: true,
		},
		{
			name: "non-existent filter file",
			config: &Config{
				Action:     "archive",
				FilterFile: "/non/existent/file.json",
			},
			expectError: true,
		},
		{
			name: "negative limit",
			config: &Config{
				Action:     "archive",
				FilterFile: validFilterFile,
				Limit:      -1,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestLoadProcessedEmails(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "cleaner_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test processed emails
	testEmails := []ProcessedEmail{
		{
			ID:        "email1",
			Subject:   "Test Email 1",
			From:      "sender1@example.com",
			Date:      time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			Size:      1024,
			Processed: time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
		},
		{
			ID:        "email2",
			Subject:   "Test Email 2",
			From:      "sender2@example.com",
			Date:      time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC),
			Size:      2048,
			Processed: time.Date(2024, 1, 2, 11, 0, 0, 0, time.UTC),
		},
	}

	// Write test data to file
	filterFile := filepath.Join(tempDir, "processed.json")
	data, err := json.Marshal(testEmails)
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	err = os.WriteFile(filterFile, data, 0644)
	if err != nil {
		t.Fatalf("Failed to write filter file: %v", err)
	}

	// Create cleaner instance
	cleaner := &Cleaner{
		config: &Config{
			FilterFile: filterFile,
		},
	}

	// Test loading processed emails
	loadedEmails, err := cleaner.loadProcessedEmails()
	if err != nil {
		t.Fatalf("Failed to load processed emails: %v", err)
	}

	// Verify loaded data
	if len(loadedEmails) != len(testEmails) {
		t.Errorf("Expected %d emails, got %d", len(testEmails), len(loadedEmails))
	}

	for i, email := range loadedEmails {
		if email.ID != testEmails[i].ID {
			t.Errorf("Expected ID %s, got %s", testEmails[i].ID, email.ID)
		}
		if email.Subject != testEmails[i].Subject {
			t.Errorf("Expected Subject %s, got %s", testEmails[i].Subject, email.Subject)
		}
		if email.From != testEmails[i].From {
			t.Errorf("Expected From %s, got %s", testEmails[i].From, email.From)
		}
		if email.Size != testEmails[i].Size {
			t.Errorf("Expected Size %d, got %d", testEmails[i].Size, email.Size)
		}
	}
}

func TestLoadProcessedEmails_InvalidJSON(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "cleaner_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Write invalid JSON to file
	filterFile := filepath.Join(tempDir, "invalid.json")
	err = os.WriteFile(filterFile, []byte("invalid json"), 0644)
	if err != nil {
		t.Fatalf("Failed to write invalid filter file: %v", err)
	}

	// Create cleaner instance
	cleaner := &Cleaner{
		config: &Config{
			FilterFile: filterFile,
		},
	}

	// Test loading processed emails - should fail
	_, err = cleaner.loadProcessedEmails()
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestGetActionVerb(t *testing.T) {
	tests := []struct {
		name     string
		action   string
		expected string
	}{
		{
			name:     "archive action",
			action:   "archive",
			expected: "archived",
		},
		{
			name:     "delete action",
			action:   "delete",
			expected: "deleted",
		},
		{
			name:     "unknown action",
			action:   "unknown",
			expected: "processed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleaner := &Cleaner{
				config: &Config{
					Action: tt.action,
				},
			}

			result := cleaner.getActionVerb()
			if result != tt.expected {
				t.Errorf("getActionVerb() = %s, want %s", result, tt.expected)
			}
		})
	}
}
