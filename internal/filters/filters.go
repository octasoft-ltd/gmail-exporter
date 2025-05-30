package filters

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Config represents email filtering configuration
type Config struct {
	// Basic filters
	To            string `json:"to,omitempty"`
	From          string `json:"from,omitempty"`
	Subject       string `json:"subject,omitempty"`
	IncludesWords string `json:"includes_words,omitempty"`
	ExcludesWords string `json:"excludes_words,omitempty"`

	// Size filters (in bytes)
	SizeGreaterThan int64 `json:"size_greater_than,omitempty"`
	SizeLessThan    int64 `json:"size_less_than,omitempty"`

	// Date filters
	DateWithin time.Duration `json:"date_within,omitempty"`
	DateAfter  *time.Time    `json:"date_after,omitempty"`
	DateBefore *time.Time    `json:"date_before,omitempty"`

	// Boolean filters
	HasAttachment *bool `json:"has_attachment,omitempty"`
	ExcludeChats  bool  `json:"exclude_chats,omitempty"`

	// Labels and search scope
	Labels      string `json:"labels,omitempty"`
	SearchScope string `json:"search_scope,omitempty"`
}

// BuildGmailQuery converts the filter configuration to a Gmail search query
func (c *Config) BuildGmailQuery() string {
	var parts []string

	// Basic filters
	if c.To != "" {
		parts = append(parts, fmt.Sprintf("to:%s", c.To))
	}
	if c.From != "" {
		parts = append(parts, fmt.Sprintf("from:%s", c.From))
	}
	if c.Subject != "" {
		parts = append(parts, fmt.Sprintf("subject:(%s)", c.Subject))
	}
	if c.IncludesWords != "" {
		parts = append(parts, c.IncludesWords)
	}
	if c.ExcludesWords != "" {
		words := strings.Fields(c.ExcludesWords)
		for _, word := range words {
			parts = append(parts, fmt.Sprintf("-%s", word))
		}
	}

	// Size filters
	if c.SizeGreaterThan > 0 {
		parts = append(parts, fmt.Sprintf("size:%d", c.SizeGreaterThan))
	}
	if c.SizeLessThan > 0 {
		parts = append(parts, fmt.Sprintf("-size:%d", c.SizeLessThan))
	}

	// Date filters
	if c.DateWithin > 0 {
		days := int(c.DateWithin.Hours() / 24)
		parts = append(parts, fmt.Sprintf("newer_than:%dd", days))
	}
	if c.DateAfter != nil {
		parts = append(parts, fmt.Sprintf("after:%s", c.DateAfter.Format("2006/01/02")))
	}
	if c.DateBefore != nil {
		parts = append(parts, fmt.Sprintf("before:%s", c.DateBefore.Format("2006/01/02")))
	}

	// Boolean filters
	if c.HasAttachment != nil {
		if *c.HasAttachment {
			parts = append(parts, "has:attachment")
		} else {
			parts = append(parts, "-has:attachment")
		}
	}
	if c.ExcludeChats {
		parts = append(parts, "-in:chats")
	}

	// Labels
	if c.Labels != "" {
		labels := strings.Split(c.Labels, ",")
		for _, label := range labels {
			label = strings.TrimSpace(label)
			if label != "" {
				parts = append(parts, fmt.Sprintf("label:%s", label))
			}
		}
	}

	// Search scope
	if c.SearchScope != "" && c.SearchScope != "all_mail" {
		parts = append(parts, fmt.Sprintf("in:%s", c.SearchScope))
	}

	return strings.Join(parts, " ")
}

// Validate checks if the filter configuration is valid
func (c *Config) Validate() error {
	// Check for conflicting size filters
	if c.SizeGreaterThan > 0 && c.SizeLessThan > 0 && c.SizeGreaterThan >= c.SizeLessThan {
		return fmt.Errorf("size-greater-than must be less than size-less-than")
	}

	// Check for conflicting date filters
	if c.DateAfter != nil && c.DateBefore != nil && c.DateAfter.After(*c.DateBefore) {
		return fmt.Errorf("date-after must be before date-before")
	}

	// Check for conflicting attachment filters
	// Attachment filter conflicts are handled in the CLI layer

	// Validate search scope
	validScopes := []string{"all_mail", "inbox", "sent", "drafts", "spam", "trash"}
	if c.SearchScope != "" {
		valid := false
		for _, scope := range validScopes {
			if c.SearchScope == scope {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid search scope: %s (valid: %s)", c.SearchScope, strings.Join(validScopes, ", "))
		}
	}

	return nil
}

// ParseSize parses size strings like "5MB", "1GB", etc.
func ParseSize(sizeStr string) (int64, error) {
	sizeStr = strings.ToUpper(strings.TrimSpace(sizeStr))

	// Extract number and unit
	var numStr string
	var unit string

	for i, r := range sizeStr {
		if r >= '0' && r <= '9' || r == '.' {
			numStr += string(r)
		} else {
			unit = sizeStr[i:]
			break
		}
	}

	if numStr == "" {
		return 0, fmt.Errorf("invalid size format: %s", sizeStr)
	}

	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number in size: %s", numStr)
	}

	var multiplier int64
	switch unit {
	case "", "B":
		multiplier = 1
	case "KB":
		multiplier = 1024
	case "MB":
		multiplier = 1024 * 1024
	case "GB":
		multiplier = 1024 * 1024 * 1024
	case "TB":
		multiplier = 1024 * 1024 * 1024 * 1024
	default:
		return 0, fmt.Errorf("invalid size unit: %s (valid: B, KB, MB, GB, TB)", unit)
	}

	return int64(num * float64(multiplier)), nil
}

// ParseDuration parses duration strings like "30d", "1w", "6m"
func ParseDuration(durationStr string) (time.Duration, error) {
	durationStr = strings.ToLower(strings.TrimSpace(durationStr))

	if len(durationStr) < 2 {
		return 0, fmt.Errorf("invalid duration format: %s", durationStr)
	}

	numStr := durationStr[:len(durationStr)-1]
	unit := durationStr[len(durationStr)-1:]

	num, err := strconv.Atoi(numStr)
	if err != nil {
		return 0, fmt.Errorf("invalid number in duration: %s", numStr)
	}

	var duration time.Duration
	switch unit {
	case "d":
		duration = time.Duration(num) * 24 * time.Hour
	case "w":
		duration = time.Duration(num) * 7 * 24 * time.Hour
	case "m":
		duration = time.Duration(num) * 30 * 24 * time.Hour // Approximate month
	case "y":
		duration = time.Duration(num) * 365 * 24 * time.Hour // Approximate year
	case "h":
		duration = time.Duration(num) * time.Hour
	default:
		return 0, fmt.Errorf("invalid duration unit: %s (valid: h, d, w, m, y)", unit)
	}

	return duration, nil
}
