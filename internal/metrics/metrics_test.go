package metrics

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewCollector(t *testing.T) {
	collector := NewCollector("test_operation")

	if collector.operation != "test_operation" {
		t.Errorf("Expected operation 'test_operation', got %s", collector.operation)
	}

	if collector.data.Operation != "test_operation" {
		t.Errorf("Expected data operation 'test_operation', got %s", collector.data.Operation)
	}

	if collector.data.Failures == nil {
		t.Error("Expected failures slice to be initialized")
	}
}

func TestCollector_Start(t *testing.T) {
	collector := NewCollector("test")

	beforeStart := time.Now()
	collector.Start()
	afterStart := time.Now()

	if collector.startTime.Before(beforeStart) || collector.startTime.After(afterStart) {
		t.Error("Start time not set correctly")
	}

	if collector.data.StartTime.Before(beforeStart) || collector.data.StartTime.After(afterStart) {
		t.Error("Data start time not set correctly")
	}
}

func TestCollector_RecordEmailsProcessed(t *testing.T) {
	collector := NewCollector("test")

	exported := 100
	failed := 5

	collector.RecordEmailsProcessed(exported, failed)

	if collector.data.Emails.TotalExported != exported {
		t.Errorf("Expected %d exported emails, got %d", exported, collector.data.Emails.TotalExported)
	}

	if collector.data.Emails.TotalFailed != failed {
		t.Errorf("Expected %d failed emails, got %d", failed, collector.data.Emails.TotalFailed)
	}
}

func TestCollector_RecordBytesProcessed(t *testing.T) {
	collector := NewCollector("test")

	bytes := int64(1024 * 1024 * 50) // 50MB

	collector.RecordBytesProcessed(bytes)

	if collector.data.Emails.TotalSize != bytes {
		t.Errorf("Expected %d bytes, got %d", bytes, collector.data.Emails.TotalSize)
	}
}

func TestCollector_RecordDuration(t *testing.T) {
	collector := NewCollector("test")
	collector.Start()

	// Set up some test data
	collector.RecordEmailsProcessed(100, 5)
	collector.RecordBytesProcessed(1024 * 1024 * 50) // 50MB

	duration := 5 * time.Minute
	collector.RecordDuration(duration)

	if collector.data.Duration != duration {
		t.Errorf("Expected duration %v, got %v", duration, collector.data.Duration)
	}

	if collector.data.EndTime == nil {
		t.Error("Expected end time to be set")
	}

	// Check performance calculations
	expectedEmailsPerSecond := float64(105) / duration.Seconds() // 100 + 5
	if collector.data.Performance.EmailsPerSecond != expectedEmailsPerSecond {
		t.Errorf("Expected emails per second %.2f, got %.2f",
			expectedEmailsPerSecond, collector.data.Performance.EmailsPerSecond)
	}

	expectedBytesPerSecond := float64(1024*1024*50) / duration.Seconds()
	if collector.data.Performance.BytesPerSecond != expectedBytesPerSecond {
		t.Errorf("Expected bytes per second %.2f, got %.2f",
			expectedBytesPerSecond, collector.data.Performance.BytesPerSecond)
	}
}

func TestCollector_RecordFailure(t *testing.T) {
	collector := NewCollector("test")

	emailID := "test_email_123"
	errorMsg := "test error message"

	beforeRecord := time.Now()
	collector.RecordFailure(emailID, errorMsg)
	afterRecord := time.Now()

	if len(collector.data.Failures) != 1 {
		t.Errorf("Expected 1 failure, got %d", len(collector.data.Failures))
	}

	failure := collector.data.Failures[0]
	if failure.EmailID != emailID {
		t.Errorf("Expected email ID %s, got %s", emailID, failure.EmailID)
	}

	if failure.Error != errorMsg {
		t.Errorf("Expected error %s, got %s", errorMsg, failure.Error)
	}

	if failure.Timestamp.Before(beforeRecord) || failure.Timestamp.After(afterRecord) {
		t.Error("Failure timestamp not set correctly")
	}
}

func TestCollector_SetTotalMatched(t *testing.T) {
	collector := NewCollector("test")

	total := 250
	collector.SetTotalMatched(total)

	if collector.data.Emails.TotalMatched != total {
		t.Errorf("Expected total matched %d, got %d", total, collector.data.Emails.TotalMatched)
	}
}

func TestCollector_Save(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "metrics_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	collector := NewCollector("test")
	collector.Start()
	collector.RecordEmailsProcessed(100, 5)
	collector.RecordBytesProcessed(1024 * 1024 * 50)
	collector.RecordDuration(5 * time.Minute)
	collector.SetTotalMatched(105)

	filename := filepath.Join(tempDir, "metrics.json")
	err = collector.Save(filename)
	if err != nil {
		t.Fatalf("Failed to save metrics: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Error("Metrics file was not created")
	}

	// Read and verify content
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read metrics file: %v", err)
	}

	var savedMetrics Data
	err = json.Unmarshal(data, &savedMetrics)
	if err != nil {
		t.Fatalf("Failed to unmarshal metrics: %v", err)
	}

	if savedMetrics.Operation != "test" {
		t.Errorf("Expected operation 'test', got %s", savedMetrics.Operation)
	}

	if savedMetrics.Emails.TotalMatched != 105 {
		t.Errorf("Expected total matched 105, got %d", savedMetrics.Emails.TotalMatched)
	}

	if savedMetrics.Emails.TotalExported != 100 {
		t.Errorf("Expected total exported 100, got %d", savedMetrics.Emails.TotalExported)
	}

	if savedMetrics.Emails.TotalFailed != 5 {
		t.Errorf("Expected total failed 5, got %d", savedMetrics.Emails.TotalFailed)
	}
}

func TestCollector_SavePrometheus(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "metrics_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	collector := NewCollector("test")
	collector.Start()
	collector.RecordEmailsProcessed(100, 5)
	collector.RecordBytesProcessed(1024 * 1024 * 50)
	collector.RecordDuration(5 * time.Minute)

	filename := filepath.Join(tempDir, "metrics.prom")
	err = collector.SavePrometheus(filename)
	if err != nil {
		t.Fatalf("Failed to save Prometheus metrics: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Error("Prometheus metrics file was not created")
	}

	// Read and verify content contains expected metrics
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read Prometheus metrics file: %v", err)
	}

	content := string(data)

	// Check for expected metric names
	expectedMetrics := []string{
		"gmail_exporter_emails_total",
		"gmail_exporter_bytes_total",
		"gmail_exporter_duration_seconds",
	}

	for _, metric := range expectedMetrics {
		if !contains(content, metric) {
			t.Errorf("Expected metric %s not found in Prometheus output", metric)
		}
	}

	// Check for specific values
	if !contains(content, `gmail_exporter_emails_total{operation="test",status="success"} 100`) {
		t.Error("Expected success count not found in Prometheus output")
	}

	if !contains(content, `gmail_exporter_emails_total{operation="test",status="failed"} 5`) {
		t.Error("Expected failure count not found in Prometheus output")
	}
}

func TestCollector_GetData(t *testing.T) {
	collector := NewCollector("test")
	collector.Start()
	collector.RecordEmailsProcessed(100, 5)

	data := collector.GetData()

	if data.Operation != "test" {
		t.Errorf("Expected operation 'test', got %s", data.Operation)
	}

	if data.Emails.TotalExported != 100 {
		t.Errorf("Expected total exported 100, got %d", data.Emails.TotalExported)
	}

	if data.Emails.TotalFailed != 5 {
		t.Errorf("Expected total failed 5, got %d", data.Emails.TotalFailed)
	}
}

func TestCollector_Summary(t *testing.T) {
	collector := NewCollector("test")

	// Test in-progress summary
	collector.Start()
	summary := collector.Summary()
	if !contains(summary, "in progress") {
		t.Error("Expected 'in progress' in summary for ongoing operation")
	}

	// Test completed summary
	collector.RecordEmailsProcessed(100, 5)
	collector.RecordBytesProcessed(1024 * 1024 * 50) // 50MB
	collector.SetTotalMatched(105)
	collector.RecordDuration(5 * time.Minute)

	summary = collector.Summary()

	expectedParts := []string{
		"Operation: test",
		"Emails Matched: 105",
		"Emails Exported: 100",
		"Emails Failed: 5",
	}

	for _, part := range expectedParts {
		if !contains(summary, part) {
			t.Errorf("Expected '%s' in summary, got: %s", part, summary)
		}
	}
}

func TestGetBucketCount(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		bucket   time.Duration
		expected int
	}{
		{
			name:     "duration less than bucket",
			duration: 4 * time.Minute,
			bucket:   5 * time.Minute,
			expected: 1,
		},
		{
			name:     "duration equal to bucket",
			duration: 5 * time.Minute,
			bucket:   5 * time.Minute,
			expected: 1,
		},
		{
			name:     "duration greater than bucket",
			duration: 6 * time.Minute,
			bucket:   5 * time.Minute,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getBucketCount(tt.duration, tt.bucket)
			if result != tt.expected {
				t.Errorf("getBucketCount() = %d, want %d", result, tt.expected)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
