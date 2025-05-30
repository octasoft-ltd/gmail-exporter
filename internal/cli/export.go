package cli

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/octasoft-ltd/gmail-exporter/internal/exporter"
	"github.com/octasoft-ltd/gmail-exporter/internal/filters"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export emails from Gmail",
	Long: `Export emails from Gmail based on specified filters.
Supports all Gmail search operators and additional filtering options.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Build filter configuration from flags
		filterConfig, err := buildFilterConfig(cmd)
		if err != nil {
			return fmt.Errorf("failed to build filter config: %w", err)
		}

		// Build export configuration
		exportConfig, err := buildExportConfig(cmd)
		if err != nil {
			return fmt.Errorf("failed to build export config: %w", err)
		}

		// Create exporter
		exp, err := exporter.New(exportConfig)
		if err != nil {
			return fmt.Errorf("failed to create exporter: %w", err)
		}

		// Run export
		logrus.WithFields(logrus.Fields{
			"output_dir": exportConfig.OutputDir,
			"filters":    filterConfig,
		}).Info("Starting email export")

		result, err := exp.Export(filterConfig)
		if err != nil {
			return fmt.Errorf("export failed: %w", err)
		}

		// Display results
		fmt.Printf("Export completed successfully!\n")
		fmt.Printf("Total emails matched: %d\n", result.TotalMatched)
		fmt.Printf("Total emails exported: %d\n", result.TotalExported)
		fmt.Printf("Total size: %s\n", formatBytes(result.TotalSize))
		fmt.Printf("Duration: %s\n", result.Duration)
		fmt.Printf("Output directory: %s\n", exportConfig.OutputDir)

		if result.TotalFailed > 0 {
			fmt.Printf("Failed exports: %d (see log for details)\n", result.TotalFailed)
		}

		return nil
	},
}

func init() {
	// Filter flags
	exportCmd.Flags().String("to", "", "Recipient email address")
	exportCmd.Flags().String("from", "", "Sender email address")
	exportCmd.Flags().String("subject", "", "Subject contains text")
	exportCmd.Flags().String("includes-words", "", "Email body contains words (space-separated)")
	exportCmd.Flags().String("excludes-words", "", "Email body excludes words (space-separated)")
	exportCmd.Flags().String("size-greater-than", "", "Email size greater than (e.g., 5MB)")
	exportCmd.Flags().String("size-less-than", "", "Email size less than (e.g., 10MB)")
	exportCmd.Flags().String("date-within", "", "Date within period (e.g., 30d, 1w, 6m)")
	exportCmd.Flags().String("date-after", "", "After specific date (YYYY-MM-DD)")
	exportCmd.Flags().String("date-before", "", "Before specific date (YYYY-MM-DD)")
	exportCmd.Flags().Bool("has-attachment", false, "Has attachments")
	exportCmd.Flags().Bool("no-attachment", false, "No attachments")
	exportCmd.Flags().Bool("exclude-chats", true, "Exclude chat messages")
	exportCmd.Flags().String("labels", "", "Specific labels (comma-separated)")
	exportCmd.Flags().String("search-scope", "all_mail", "Search scope (all_mail, inbox, sent, drafts, spam, trash)")

	// Export configuration flags
	exportCmd.Flags().StringP("output-dir", "o", "", "Output directory for exported emails")
	exportCmd.Flags().Bool("organize-by-labels", false, "Organize exported emails by labels in folder structure")
	exportCmd.Flags().Int("parallel-workers", 0, "Number of parallel workers (0 = use config default)")
	exportCmd.Flags().Bool("include-attachments", true, "Include email attachments in export")
	exportCmd.Flags().Bool("compress-exports", false, "Compress exported emails")
	exportCmd.Flags().String("format", "eml", "Export format (eml, mbox, json)")
	exportCmd.Flags().Bool("resume", false, "Resume a previous export")
	exportCmd.Flags().String("state-file", "", "State file for resumable operations")
	exportCmd.Flags().IntP("limit", "l", 0, "Limit the number of messages to process (0 = no limit, useful for testing)")

	// Bind flags to viper
	if err := viper.BindPFlag("output_dir", exportCmd.Flags().Lookup("output-dir")); err != nil {
		logrus.WithError(err).Fatal("Failed to bind output-dir flag")
	}
	if err := viper.BindPFlag("organize_by_labels", exportCmd.Flags().Lookup("organize-by-labels")); err != nil {
		logrus.WithError(err).Fatal("Failed to bind organize-by-labels flag")
	}
	if err := viper.BindPFlag("parallel_workers", exportCmd.Flags().Lookup("parallel-workers")); err != nil {
		logrus.WithError(err).Fatal("Failed to bind parallel-workers flag")
	}
}

func buildFilterConfig(cmd *cobra.Command) (*filters.Config, error) {
	config := &filters.Config{}

	// Basic filters
	if to, _ := cmd.Flags().GetString("to"); to != "" {
		config.To = to
	}
	if from, _ := cmd.Flags().GetString("from"); from != "" {
		config.From = from
	}
	if subject, _ := cmd.Flags().GetString("subject"); subject != "" {
		config.Subject = subject
	}
	if includes, _ := cmd.Flags().GetString("includes-words"); includes != "" {
		config.IncludesWords = includes
	}
	if excludes, _ := cmd.Flags().GetString("excludes-words"); excludes != "" {
		config.ExcludesWords = excludes
	}

	// Size filters
	if sizeGreater, _ := cmd.Flags().GetString("size-greater-than"); sizeGreater != "" {
		size, err := filters.ParseSize(sizeGreater)
		if err != nil {
			return nil, fmt.Errorf("invalid size-greater-than: %w", err)
		}
		config.SizeGreaterThan = size
	}
	if sizeLess, _ := cmd.Flags().GetString("size-less-than"); sizeLess != "" {
		size, err := filters.ParseSize(sizeLess)
		if err != nil {
			return nil, fmt.Errorf("invalid size-less-than: %w", err)
		}
		config.SizeLessThan = size
	}

	// Date filters
	if dateWithin, _ := cmd.Flags().GetString("date-within"); dateWithin != "" {
		duration, err := filters.ParseDuration(dateWithin)
		if err != nil {
			return nil, fmt.Errorf("invalid date-within: %w", err)
		}
		config.DateWithin = duration
	}
	if dateAfter, _ := cmd.Flags().GetString("date-after"); dateAfter != "" {
		date, err := time.Parse("2006-01-02", dateAfter)
		if err != nil {
			return nil, fmt.Errorf("invalid date-after format (use YYYY-MM-DD): %w", err)
		}
		config.DateAfter = &date
	}
	if dateBefore, _ := cmd.Flags().GetString("date-before"); dateBefore != "" {
		date, err := time.Parse("2006-01-02", dateBefore)
		if err != nil {
			return nil, fmt.Errorf("invalid date-before format (use YYYY-MM-DD): %w", err)
		}
		config.DateBefore = &date
	}

	// Boolean filters
	if hasAttachment, _ := cmd.Flags().GetBool("has-attachment"); hasAttachment {
		config.HasAttachment = &hasAttachment
	}
	if noAttachment, _ := cmd.Flags().GetBool("no-attachment"); noAttachment {
		falseVal := false
		config.HasAttachment = &falseVal
	}
	if excludeChats, _ := cmd.Flags().GetBool("exclude-chats"); excludeChats {
		config.ExcludeChats = excludeChats
	}

	// Labels and search scope
	if labels, _ := cmd.Flags().GetString("labels"); labels != "" {
		config.Labels = labels
	}
	if searchScope, _ := cmd.Flags().GetString("search-scope"); searchScope != "" {
		config.SearchScope = searchScope
	}

	return config, nil
}

func buildExportConfig(cmd *cobra.Command) (*exporter.Config, error) {
	config := &exporter.Config{
		CredentialsFile:  viper.GetString("credentials_file"),
		TokenFile:        viper.GetString("token_file"),
		OutputDir:        viper.GetString("output_dir"),
		OrganizeByLabels: viper.GetBool("organize_by_labels"),
		ParallelWorkers:  viper.GetInt("parallel_workers"),
	}

	// Override with command flags if provided
	if outputDir, _ := cmd.Flags().GetString("output-dir"); outputDir != "" {
		config.OutputDir = outputDir
	}
	if organizeByLabels, _ := cmd.Flags().GetBool("organize-by-labels"); organizeByLabels {
		config.OrganizeByLabels = organizeByLabels
	}
	if parallelWorkers, _ := cmd.Flags().GetInt("parallel-workers"); parallelWorkers > 0 {
		config.ParallelWorkers = parallelWorkers
	}
	if includeAttachments, _ := cmd.Flags().GetBool("include-attachments"); !includeAttachments {
		config.IncludeAttachments = includeAttachments
	} else {
		config.IncludeAttachments = true
	}
	if compressExports, _ := cmd.Flags().GetBool("compress-exports"); compressExports {
		config.CompressExports = compressExports
	}
	if format, _ := cmd.Flags().GetString("format"); format != "" {
		config.Format = format
	}
	if resume, _ := cmd.Flags().GetBool("resume"); resume {
		config.Resume = resume
	}
	if stateFile, _ := cmd.Flags().GetString("state-file"); stateFile != "" {
		config.StateFile = stateFile
	}
	if limit, _ := cmd.Flags().GetInt("limit"); limit > 0 {
		config.Limit = limit
	}

	// Validate required fields
	if config.OutputDir == "" {
		return nil, fmt.Errorf("output directory is required")
	}

	return config, nil
}

func formatBytes(bytes int64) string {
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
