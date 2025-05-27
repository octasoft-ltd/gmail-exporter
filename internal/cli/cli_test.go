package cli

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
)

func TestRootCommand(t *testing.T) {
	// Test that the root command executes without error
	cmd := &cobra.Command{
		Use: "test-root",
		Run: func(cmd *cobra.Command, args []string) {
			// Do nothing
		},
	}

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Root command execution failed: %v", err)
	}
}

func TestVersionCommand(t *testing.T) {
	// Test that version command exists and can be executed
	if versionCmd == nil {
		t.Error("Version command is nil")
		return
	}

	if versionCmd.Use != "version" {
		t.Errorf("Expected version command Use to be 'version', got '%s'", versionCmd.Use)
	}

	if versionCmd.Short != "Print the version number" {
		t.Errorf("Expected version command Short to be 'Print the version number', got '%s'", versionCmd.Short)
	}
}

func TestExportCommandFlags(t *testing.T) {
	// Test that export command has all expected flags
	expectedFlags := []string{
		"to",
		"from",
		"subject",
		"includes-words",
		"excludes-words",
		"size-greater-than",
		"size-less-than",
		"date-within",
		"date-after",
		"date-before",
		"has-attachment",
		"no-attachment",
		"exclude-chats",
		"labels",
		"search-scope",
		"output-dir",
		"organize-by-labels",
		"parallel-workers",
		"include-attachments",
		"compress-exports",
		"format",
		"resume",
		"state-file",
	}

	for _, flagName := range expectedFlags {
		flag := exportCmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag '%s' not found in export command", flagName)
		}
	}
}

func TestAuthCommandSubcommands(t *testing.T) {
	// Test that auth command has all expected subcommands
	expectedSubcommands := []string{
		"setup",
		"login",
		"refresh",
		"status",
	}

	for _, subcommandName := range expectedSubcommands {
		found := false
		for _, cmd := range authCmd.Commands() {
			if cmd.Name() == subcommandName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected subcommand '%s' not found in auth command", subcommandName)
		}
	}
}

func TestImportCommandFlags(t *testing.T) {
	// Test that import command has all expected flags
	expectedFlags := []string{
		"input-dir",
		"import-credentials",
		"import-token",
		"parallel-workers",
		"preserve-dates",
		"limit",
	}

	for _, flagName := range expectedFlags {
		flag := importCmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag '%s' not found in import command", flagName)
		}
	}

	// Test that limit flag has short form
	limitFlag := importCmd.Flags().Lookup("limit")
	if limitFlag.Shorthand != "l" {
		t.Errorf("Expected limit flag shorthand to be 'l', got '%s'", limitFlag.Shorthand)
	}
}

func TestCleanupCommandFlags(t *testing.T) {
	// Test that cleanup command has all expected flags
	expectedFlags := []string{
		"action",
		"filter-file",
		"dry-run",
		"limit",
	}

	for _, flagName := range expectedFlags {
		flag := cleanupCmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag '%s' not found in cleanup command", flagName)
		}
	}

	// Test that limit flag has short form
	limitFlag := cleanupCmd.Flags().Lookup("limit")
	if limitFlag.Shorthand != "l" {
		t.Errorf("Expected limit flag shorthand to be 'l', got '%s'", limitFlag.Shorthand)
	}
}

func TestBuildFilterConfig(t *testing.T) {
	// Create a test command with flags set
	cmd := &cobra.Command{}
	cmd.Flags().String("to", "", "")
	cmd.Flags().String("from", "", "")
	cmd.Flags().String("subject", "", "")
	cmd.Flags().String("size-greater-than", "", "")
	cmd.Flags().String("date-after", "", "")
	cmd.Flags().Bool("has-attachment", false, "")
	cmd.Flags().Bool("exclude-chats", false, "")
	cmd.Flags().String("labels", "", "")
	cmd.Flags().String("search-scope", "", "")

	// Set some test values
	cmd.Flags().Set("to", "test@example.com")
	cmd.Flags().Set("from", "sender@example.com")
	cmd.Flags().Set("subject", "Test Subject")
	cmd.Flags().Set("size-greater-than", "5MB")
	cmd.Flags().Set("date-after", "2024-01-01")
	cmd.Flags().Set("has-attachment", "true")
	cmd.Flags().Set("exclude-chats", "true")
	cmd.Flags().Set("labels", "important,work")
	cmd.Flags().Set("search-scope", "inbox")

	// Add missing flags that buildFilterConfig expects
	cmd.Flags().String("includes-words", "", "")
	cmd.Flags().String("excludes-words", "", "")
	cmd.Flags().String("size-less-than", "", "")
	cmd.Flags().String("date-within", "", "")
	cmd.Flags().String("date-before", "", "")
	cmd.Flags().Bool("no-attachment", false, "")

	config, err := buildFilterConfig(cmd)
	if err != nil {
		t.Fatalf("buildFilterConfig failed: %v", err)
	}

	// Verify the configuration
	if config.To != "test@example.com" {
		t.Errorf("Expected To 'test@example.com', got '%s'", config.To)
	}

	if config.From != "sender@example.com" {
		t.Errorf("Expected From 'sender@example.com', got '%s'", config.From)
	}

	if config.Subject != "Test Subject" {
		t.Errorf("Expected Subject 'Test Subject', got '%s'", config.Subject)
	}

	if config.SizeGreaterThan != 5242880 { // 5MB in bytes
		t.Errorf("Expected SizeGreaterThan 5242880, got %d", config.SizeGreaterThan)
	}

	if config.HasAttachment == nil || !*config.HasAttachment {
		t.Error("Expected HasAttachment to be true")
	}

	if !config.ExcludeChats {
		t.Error("Expected ExcludeChats to be true")
	}

	if config.Labels != "important,work" {
		t.Errorf("Expected Labels 'important,work', got '%s'", config.Labels)
	}

	if config.SearchScope != "inbox" {
		t.Errorf("Expected SearchScope 'inbox', got '%s'", config.SearchScope)
	}
}

func TestBuildFilterConfig_InvalidSize(t *testing.T) {
	// Create a test command with invalid size
	cmd := &cobra.Command{}
	cmd.Flags().String("size-greater-than", "", "")

	// Add all required flags
	cmd.Flags().String("to", "", "")
	cmd.Flags().String("from", "", "")
	cmd.Flags().String("subject", "", "")
	cmd.Flags().String("includes-words", "", "")
	cmd.Flags().String("excludes-words", "", "")
	cmd.Flags().String("size-less-than", "", "")
	cmd.Flags().String("date-within", "", "")
	cmd.Flags().String("date-after", "", "")
	cmd.Flags().String("date-before", "", "")
	cmd.Flags().Bool("has-attachment", false, "")
	cmd.Flags().Bool("no-attachment", false, "")
	cmd.Flags().Bool("exclude-chats", false, "")
	cmd.Flags().String("labels", "", "")
	cmd.Flags().String("search-scope", "", "")

	cmd.Flags().Set("size-greater-than", "invalid-size")

	_, err := buildFilterConfig(cmd)
	if err == nil {
		t.Error("Expected error for invalid size format")
	}
}

func TestBuildFilterConfig_InvalidDate(t *testing.T) {
	// Create a test command with invalid date
	cmd := &cobra.Command{}
	cmd.Flags().String("date-after", "", "")

	// Add all required flags
	cmd.Flags().String("to", "", "")
	cmd.Flags().String("from", "", "")
	cmd.Flags().String("subject", "", "")
	cmd.Flags().String("includes-words", "", "")
	cmd.Flags().String("excludes-words", "", "")
	cmd.Flags().String("size-greater-than", "", "")
	cmd.Flags().String("size-less-than", "", "")
	cmd.Flags().String("date-within", "", "")
	cmd.Flags().String("date-before", "", "")
	cmd.Flags().Bool("has-attachment", false, "")
	cmd.Flags().Bool("no-attachment", false, "")
	cmd.Flags().Bool("exclude-chats", false, "")
	cmd.Flags().String("labels", "", "")
	cmd.Flags().String("search-scope", "", "")

	cmd.Flags().Set("date-after", "invalid-date")

	_, err := buildFilterConfig(cmd)
	if err == nil {
		t.Error("Expected error for invalid date format")
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{
			name:     "bytes",
			bytes:    512,
			expected: "512 B",
		},
		{
			name:     "kilobytes",
			bytes:    1536, // 1.5 KB
			expected: "1.5 KB",
		},
		{
			name:     "megabytes",
			bytes:    1572864, // 1.5 MB
			expected: "1.5 MB",
		},
		{
			name:     "gigabytes",
			bytes:    1610612736, // 1.5 GB
			expected: "1.5 GB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatBytes(tt.bytes)
			if result != tt.expected {
				t.Errorf("formatBytes(%d) = %s, want %s", tt.bytes, result, tt.expected)
			}
		})
	}
}

func TestInitConfig(t *testing.T) {
	// Save original values
	originalCfgFile := cfgFile
	originalHome := os.Getenv("HOME")

	// Set up test environment
	tempDir, err := os.MkdirTemp("", "cli_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	os.Setenv("HOME", tempDir)
	cfgFile = ""

	// Test initConfig doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("initConfig panicked: %v", r)
		}
		// Restore original values
		cfgFile = originalCfgFile
		os.Setenv("HOME", originalHome)
	}()

	initConfig()
}
