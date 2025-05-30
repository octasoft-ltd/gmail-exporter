package metrics

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// Collector handles metrics collection and export
type Collector struct {
	operation string
	startTime time.Time
	data      *Data

	// Prometheus metrics
	emailsProcessed   prometheus.CounterVec
	bytesProcessed    prometheus.Counter
	operationDuration prometheus.Histogram
}

// Data represents the metrics data structure
type Data struct {
	Operation   string        `json:"operation"`
	StartTime   time.Time     `json:"start_time"`
	EndTime     *time.Time    `json:"end_time,omitempty"`
	Duration    time.Duration `json:"duration_seconds"`
	Emails      EmailMetrics  `json:"emails"`
	Performance Performance   `json:"performance"`
	Failures    []Failure     `json:"failures,omitempty"`
}

// EmailMetrics represents email-related metrics
type EmailMetrics struct {
	TotalMatched  int   `json:"total_matched"`
	TotalExported int   `json:"total_exported"`
	TotalFailed   int   `json:"total_failed"`
	TotalSize     int64 `json:"total_size_bytes"`
}

// Performance represents performance metrics
type Performance struct {
	EmailsPerSecond float64 `json:"emails_per_second"`
	BytesPerSecond  float64 `json:"bytes_per_second"`
}

// Failure represents a failed operation
type Failure struct {
	EmailID   string    `json:"email_id"`
	Error     string    `json:"error"`
	Timestamp time.Time `json:"timestamp"`
}

// NewCollector creates a new metrics collector
func NewCollector(operation string) *Collector {
	// Create a new registry for this collector to avoid conflicts in tests
	registry := prometheus.NewRegistry()

	emailsProcessed := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gmail_exporter_emails_total",
			Help: "Total number of emails processed",
		},
		[]string{"operation", "status"},
	)

	bytesProcessed := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "gmail_exporter_bytes_total",
			Help: "Total number of bytes processed",
		},
	)

	operationDuration := prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "gmail_exporter_duration_seconds",
			Help:    "Time taken for operation",
			Buckets: prometheus.DefBuckets,
		},
	)

	// Register metrics with the local registry
	registry.MustRegister(emailsProcessed, bytesProcessed, operationDuration)

	return &Collector{
		operation: operation,
		data: &Data{
			Operation: operation,
			Emails:    EmailMetrics{},
			Failures:  make([]Failure, 0),
		},
		emailsProcessed:   *emailsProcessed,
		bytesProcessed:    bytesProcessed,
		operationDuration: operationDuration,
	}
}

// Start marks the beginning of an operation
func (c *Collector) Start() {
	c.startTime = time.Now()
	c.data.StartTime = c.startTime
	logrus.WithField("operation", c.operation).Debug("Started metrics collection")
}

// RecordEmailsProcessed records the number of emails processed
func (c *Collector) RecordEmailsProcessed(exported, failed int) {
	c.data.Emails.TotalExported = exported
	c.data.Emails.TotalFailed = failed

	// Update Prometheus metrics
	c.emailsProcessed.WithLabelValues(c.operation, "success").Add(float64(exported))
	c.emailsProcessed.WithLabelValues(c.operation, "failed").Add(float64(failed))

	logrus.WithFields(logrus.Fields{
		"exported": exported,
		"failed":   failed,
	}).Debug("Recorded email processing metrics")
}

// RecordBytesProcessed records the number of bytes processed
func (c *Collector) RecordBytesProcessed(bytes int64) {
	c.data.Emails.TotalSize = bytes
	c.bytesProcessed.Add(float64(bytes))

	logrus.WithField("bytes", bytes).Debug("Recorded bytes processed")
}

// RecordDuration records the operation duration
func (c *Collector) RecordDuration(duration time.Duration) {
	endTime := time.Now()
	c.data.EndTime = &endTime
	c.data.Duration = duration

	// Calculate performance metrics
	if duration.Seconds() > 0 {
		totalEmails := float64(c.data.Emails.TotalExported + c.data.Emails.TotalFailed)
		c.data.Performance.EmailsPerSecond = totalEmails / duration.Seconds()
		c.data.Performance.BytesPerSecond = float64(c.data.Emails.TotalSize) / duration.Seconds()
	}

	c.operationDuration.Observe(duration.Seconds())

	logrus.WithField("duration", duration).Debug("Recorded operation duration")
}

// RecordFailure records a failed operation
func (c *Collector) RecordFailure(emailID, errorMsg string) {
	failure := Failure{
		EmailID:   emailID,
		Error:     errorMsg,
		Timestamp: time.Now(),
	}
	c.data.Failures = append(c.data.Failures, failure)

	logrus.WithFields(logrus.Fields{
		"email_id": emailID,
		"error":    errorMsg,
	}).Debug("Recorded failure")
}

// SetTotalMatched sets the total number of emails matched
func (c *Collector) SetTotalMatched(total int) {
	c.data.Emails.TotalMatched = total
	logrus.WithField("total_matched", total).Debug("Set total matched emails")
}

// Save saves the metrics to a file in JSON format
func (c *Collector) Save(filename string) error {
	data, err := json.MarshalIndent(c.data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	if err := os.WriteFile(filename, data, 0o600); err != nil {
		return fmt.Errorf("failed to write metrics file: %w", err)
	}

	logrus.WithField("filename", filename).Info("Saved metrics to file")
	return nil
}

// SavePrometheus saves the metrics in Prometheus format
func (c *Collector) SavePrometheus(filename string) error {
	// This is a simplified implementation
	// In a real implementation, you would use the Prometheus registry to export metrics
	prometheusData := fmt.Sprintf(`# HELP gmail_exporter_emails_total Total number of emails processed
# TYPE gmail_exporter_emails_total counter
gmail_exporter_emails_total{operation="%s",status="success"} %d
gmail_exporter_emails_total{operation="%s",status="failed"} %d

# HELP gmail_exporter_bytes_total Total number of bytes processed
# TYPE gmail_exporter_bytes_total counter
gmail_exporter_bytes_total %d

# HELP gmail_exporter_duration_seconds Time taken for operation
# TYPE gmail_exporter_duration_seconds histogram
gmail_exporter_duration_seconds_bucket{operation="%s",le="300"} %d
gmail_exporter_duration_seconds_bucket{operation="%s",le="600"} %d
gmail_exporter_duration_seconds_bucket{operation="%s",le="+Inf"} %d
gmail_exporter_duration_seconds_sum{operation="%s"} %.2f
gmail_exporter_duration_seconds_count{operation="%s"} 1
`,
		c.operation, c.data.Emails.TotalExported,
		c.operation, c.data.Emails.TotalFailed,
		c.data.Emails.TotalSize,
		c.operation, getBucketCount(c.data.Duration, 300*time.Second),
		c.operation, getBucketCount(c.data.Duration, 600*time.Second),
		c.operation, 1,
		c.operation, c.data.Duration.Seconds(),
		c.operation,
	)

	if err := os.WriteFile(filename, []byte(prometheusData), 0o600); err != nil {
		return fmt.Errorf("failed to write Prometheus metrics file: %w", err)
	}

	logrus.WithField("filename", filename).Info("Saved Prometheus metrics to file")
	return nil
}

// GetData returns the current metrics data
func (c *Collector) GetData() *Data {
	return c.data
}

// getBucketCount returns 1 if duration is less than or equal to bucket, 0 otherwise
func getBucketCount(duration, bucket time.Duration) int {
	if duration <= bucket {
		return 1
	}
	return 0
}

// Summary returns a human-readable summary of the metrics
func (c *Collector) Summary() string {
	if c.data.EndTime == nil {
		return fmt.Sprintf("Operation '%s' in progress (started: %s)",
			c.operation, c.data.StartTime.Format("2006-01-02 15:04:05"))
	}

	return fmt.Sprintf(`Operation Summary:
  Operation: %s
  Duration: %s
  Emails Matched: %d
  Emails Exported: %d
  Emails Failed: %d
  Total Size: %s
  Performance: %.2f emails/sec, %s/sec`,
		c.operation,
		c.data.Duration,
		c.data.Emails.TotalMatched,
		c.data.Emails.TotalExported,
		c.data.Emails.TotalFailed,
		FormatBytes(c.data.Emails.TotalSize),
		c.data.Performance.EmailsPerSecond,
		FormatBytes(int64(c.data.Performance.BytesPerSecond)),
	)
}

// FormatBytes formats bytes in human-readable format
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
