package filters

import (
	"testing"
	"time"
)

func TestConfig_BuildGmailQuery(t *testing.T) {
	// Helper function to create time pointers
	timePtr := func(s string) *time.Time {
		t, _ := time.Parse("2006-01-02", s)
		return &t
	}

	// Helper function to create bool pointers
	boolPtr := func(b bool) *bool {
		return &b
	}

	tests := []struct {
		name     string
		config   Config
		expected string
	}{
		{
			name: "basic to filter",
			config: Config{
				To: "user@example.com",
			},
			expected: "to:user@example.com",
		},
		{
			name: "basic from filter",
			config: Config{
				From: "sender@example.com",
			},
			expected: "from:sender@example.com",
		},
		{
			name: "subject filter",
			config: Config{
				Subject: "Important Meeting",
			},
			expected: "subject:(Important Meeting)",
		},
		{
			name: "multiple filters",
			config: Config{
				To:      "user@example.com",
				From:    "sender@example.com",
				Subject: "Invoice",
			},
			expected: "to:user@example.com from:sender@example.com subject:(Invoice)",
		},
		{
			name: "has attachment",
			config: Config{
				HasAttachment: boolPtr(true),
			},
			expected: "has:attachment",
		},
		{
			name: "no attachment",
			config: Config{
				HasAttachment: boolPtr(false),
			},
			expected: "-has:attachment",
		},
		{
			name: "exclude chats",
			config: Config{
				ExcludeChats: true,
			},
			expected: "-in:chats",
		},
		{
			name: "includes words",
			config: Config{
				IncludesWords: "urgent important",
			},
			expected: "urgent important",
		},
		{
			name: "excludes words",
			config: Config{
				ExcludesWords: "spam promotional",
			},
			expected: "-spam -promotional",
		},
		{
			name: "size greater than",
			config: Config{
				SizeGreaterThan: 5242880, // 5MB in bytes
			},
			expected: "size:5242880",
		},
		{
			name: "size less than",
			config: Config{
				SizeLessThan: 10485760, // 10MB in bytes
			},
			expected: "-size:10485760",
		},
		{
			name: "date after",
			config: Config{
				DateAfter: timePtr("2024-01-01"),
			},
			expected: "after:2024/01/01",
		},
		{
			name: "date before",
			config: Config{
				DateBefore: timePtr("2024-12-31"),
			},
			expected: "before:2024/12/31",
		},
		{
			name: "labels",
			config: Config{
				Labels: "important,work",
			},
			expected: "label:important label:work",
		},
		{
			name: "search scope inbox",
			config: Config{
				SearchScope: "inbox",
			},
			expected: "in:inbox",
		},
		{
			name: "complex query",
			config: Config{
				To:              "user@example.com",
				From:            "sender@example.com",
				Subject:         "Invoice",
				HasAttachment:   boolPtr(true),
				SizeGreaterThan: 1048576, // 1MB in bytes
				DateAfter:       timePtr("2024-01-01"),
				Labels:          "important",
				ExcludeChats:    true,
			},
			expected: "to:user@example.com from:sender@example.com subject:(Invoice) size:1048576 after:2024/01/01 has:attachment -in:chats label:important",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.BuildGmailQuery()
			if result != tt.expected {
				t.Errorf("BuildGmailQuery() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	// Helper function to create time pointers
	timePtr := func(s string) *time.Time {
		t, _ := time.Parse("2006-01-02", s)
		return &t
	}

	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				To: "user@example.com",
			},
			wantErr: false,
		},
		{
			name: "conflicting size filters",
			config: Config{
				SizeGreaterThan: 10485760, // 10MB
				SizeLessThan:    5242880,  // 5MB
			},
			wantErr: true,
		},
		{
			name: "conflicting date filters",
			config: Config{
				DateAfter:  timePtr("2024-12-31"),
				DateBefore: timePtr("2024-01-01"),
			},
			wantErr: true,
		},
		{
			name: "valid size formats",
			config: Config{
				SizeGreaterThan: 5242880,  // 5MB
				SizeLessThan:    10485760, // 10MB
			},
			wantErr: false,
		},
		{
			name: "invalid search scope",
			config: Config{
				SearchScope: "invalid_scope",
			},
			wantErr: true,
		},
		{
			name: "valid search scope",
			config: Config{
				SearchScope: "inbox",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseSize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
		wantErr  bool
	}{
		{
			name:     "megabytes",
			input:    "5MB",
			expected: 5242880, // 5 * 1024 * 1024
			wantErr:  false,
		},
		{
			name:     "gigabytes",
			input:    "2GB",
			expected: 2147483648, // 2 * 1024 * 1024 * 1024
			wantErr:  false,
		},
		{
			name:     "kilobytes",
			input:    "500KB",
			expected: 512000, // 500 * 1024
			wantErr:  false,
		},
		{
			name:     "bytes",
			input:    "1024B",
			expected: 1024,
			wantErr:  false,
		},
		{
			name:     "lowercase",
			input:    "5mb",
			expected: 5242880,
			wantErr:  false,
		},
		{
			name:     "invalid format",
			input:    "invalid",
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "no unit",
			input:    "1024",
			expected: 1024,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseSize(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("ParseSize() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Duration
		wantErr  bool
	}{
		{
			name:     "days",
			input:    "30d",
			expected: 30 * 24 * time.Hour,
			wantErr:  false,
		},
		{
			name:     "weeks",
			input:    "2w",
			expected: 2 * 7 * 24 * time.Hour,
			wantErr:  false,
		},
		{
			name:     "months",
			input:    "6m",
			expected: 6 * 30 * 24 * time.Hour,
			wantErr:  false,
		},
		{
			name:     "years",
			input:    "1y",
			expected: 365 * 24 * time.Hour,
			wantErr:  false,
		},
		{
			name:     "invalid format",
			input:    "invalid",
			expected: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDuration(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDuration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("ParseDuration() = %v, want %v", result, tt.expected)
			}
		})
	}
}
