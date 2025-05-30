package cli

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/octasoft-ltd/gmail-exporter/internal/cleaner"
)

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Archive or delete processed emails from source account",
	Long: `Archive or delete emails that have been successfully exported/imported.
Use with caution when deleting emails.

Use --limit to process only a specific number of messages, which is useful for testing
the cleanup process with a small number of messages before running a full cleanup.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Build cleanup configuration from flags
		cleanupConfig, err := buildCleanupConfig(cmd)
		if err != nil {
			return fmt.Errorf("failed to build cleanup config: %w", err)
		}

		// Create cleaner
		cl, err := cleaner.New(cleanupConfig)
		if err != nil {
			return fmt.Errorf("failed to create cleaner: %w", err)
		}

		// Run cleanup
		logrus.WithFields(logrus.Fields{
			"action":      cleanupConfig.Action,
			"filter_file": cleanupConfig.FilterFile,
			"dry_run":     cleanupConfig.DryRun,
			"limit":       cleanupConfig.Limit,
		}).Info("Starting email cleanup")

		result, err := cl.Cleanup()
		if err != nil {
			return fmt.Errorf("cleanup failed: %w", err)
		}

		// Display results
		if result.DryRun {
			fmt.Printf("DRY RUN - Cleanup simulation completed!\n")
		} else {
			fmt.Printf("Cleanup completed successfully!\n")
		}
		fmt.Printf("Total emails found: %d\n", result.TotalFound)
		fmt.Printf("Total emails %s: %d\n", result.Action+"d", result.TotalProcessed)
		fmt.Printf("Action: %s\n", result.Action)
		fmt.Printf("Duration: %s\n", result.Duration)

		if result.TotalFailed > 0 {
			fmt.Printf("Failed operations: %d (see log for details)\n", result.TotalFailed)
		}

		return nil
	},
}

func init() {
	cleanupCmd.Flags().String("action", "archive", "Action to perform (archive, delete)")
	cleanupCmd.Flags().String("filter-file", "", "File containing list of processed email IDs")
	cleanupCmd.Flags().Bool("dry-run", false, "Show what would be done without actually doing it")
	cleanupCmd.Flags().IntP("limit", "l", 0, "Limit the number of messages to process (0 = no limit, useful for testing)")
}

func buildCleanupConfig(cmd *cobra.Command) (*cleaner.Config, error) {
	config := &cleaner.Config{
		CredentialsFile: viper.GetString("credentials_file"),
		TokenFile:       viper.GetString("token_file"),
	}

	// Get flags
	if action, _ := cmd.Flags().GetString("action"); action != "" {
		config.Action = action
	}
	if filterFile, _ := cmd.Flags().GetString("filter-file"); filterFile != "" {
		config.FilterFile = filterFile
	}
	if dryRun, _ := cmd.Flags().GetBool("dry-run"); dryRun {
		config.DryRun = dryRun
	}
	if limit, _ := cmd.Flags().GetInt("limit"); limit > 0 {
		config.Limit = limit
	}

	// Validate required fields
	if config.FilterFile == "" {
		return nil, fmt.Errorf("filter file is required")
	}

	return config, nil
}
