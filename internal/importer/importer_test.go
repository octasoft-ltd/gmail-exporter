package importer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "importer_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := &Config{
		CredentialsFile: "test_credentials.json",
		TokenFile:       "test_token.json",
		InputDir:        tempDir,
		ParallelWorkers: 3,
		PreserveDates:   true,
		Limit:           0,
	}

	// This will fail because we don't have valid credentials, but we can test validation
	_, err = New(config)
	if err == nil {
		t.Error("Expected error for invalid credentials file")
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name: "valid config",
			config: &Config{
				InputDir:        ".",
				ParallelWorkers: 3,
				Limit:           0,
			},
			expectError: false,
		},
		{
			name: "missing input dir",
			config: &Config{
				ParallelWorkers: 3,
			},
			expectError: true,
		},
		{
			name: "negative parallel workers",
			config: &Config{
				InputDir:        ".",
				ParallelWorkers: -1,
			},
			expectError: true,
		},
		{
			name: "negative limit",
			config: &Config{
				InputDir:        ".",
				ParallelWorkers: 3,
				Limit:           -1,
			},
			expectError: true,
		},
		{
			name: "non-existent input dir",
			config: &Config{
				InputDir:        "/non/existent/path",
				ParallelWorkers: 3,
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

func TestFindEmailFiles(t *testing.T) {
	// Create temporary directory with test files
	tempDir, err := os.MkdirTemp("", "importer_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	testFiles := []string{
		"email1.eml",
		"email2.json",
		"email3.mbox",
		"not_email.txt",
		"document.pdf",
	}

	for _, filename := range testFiles {
		filePath := filepath.Join(tempDir, filename)
		err := os.WriteFile(filePath, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// Create subdirectory with more files
	subDir := filepath.Join(tempDir, "subdir")
	err = os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	subFile := filepath.Join(subDir, "email4.eml")
	err = os.WriteFile(subFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create sub file: %v", err)
	}

	// Create importer instance
	importer := &Importer{
		config: &Config{
			InputDir: tempDir,
		},
	}

	// Test finding email files
	emailFiles, err := importer.findEmailFiles()
	if err != nil {
		t.Fatalf("Failed to find email files: %v", err)
	}

	// Should find 4 email files (3 in root + 1 in subdir)
	expectedCount := 4
	if len(emailFiles) != expectedCount {
		t.Errorf("Expected %d email files, got %d", expectedCount, len(emailFiles))
	}

	// Check that only email files are included
	for _, filePath := range emailFiles {
		ext := filepath.Ext(filePath)
		if ext != ".eml" && ext != ".json" && ext != ".mbox" {
			t.Errorf("Unexpected file extension found: %s", ext)
		}
	}
}

func TestEncodeBase64URL(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "simple text",
			input:    []byte("hello world"),
			expected: "aGVsbG8gd29ybGQ",
		},
		{
			name:     "empty input",
			input:    []byte(""),
			expected: "",
		},
		{
			name:     "binary data",
			input:    []byte{0x00, 0x01, 0x02, 0x03},
			expected: "AAECAw",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := encodeBase64URL(tt.input)
			if result != tt.expected {
				t.Errorf("encodeBase64URL() = %s, want %s", result, tt.expected)
			}
		})
	}
}
