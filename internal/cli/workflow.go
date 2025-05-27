package cli

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var workflowCmd = &cobra.Command{
	Use:   "workflow",
	Short: "Run complete export, import, and cleanup workflow",
	Long: `Run a complete workflow that exports emails, forwards them to another account,
and optionally archives or deletes the original emails.

Use --limit to process only a specific number of messages in each step, which is useful 
for testing the complete workflow with a small number of messages before running a full workflow.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		if limit > 0 {
			logrus.WithField("limit", limit).Info("Workflow will be limited to specified number of messages per step")
		}

		logrus.Info("Workflow command not yet implemented")
		return fmt.Errorf("workflow command not yet implemented")
	},
}

func init() {
	// Inherit flags from other commands
	workflowCmd.Flags().String("to", "", "Recipient email address to filter")
	workflowCmd.Flags().String("destination", "", "Destination email address for forwarding")
	workflowCmd.Flags().String("cleanup-action", "archive", "Cleanup action (archive, delete, none)")
	workflowCmd.Flags().StringP("output-dir", "o", "./exports", "Output directory for exported emails")
	workflowCmd.Flags().Int("parallel-workers", 3, "Number of parallel workers")
	workflowCmd.Flags().Bool("dry-run", false, "Show what would be done without actually doing it")
	workflowCmd.Flags().IntP("limit", "l", 0, "Limit the number of messages to process in each step (0 = no limit, useful for testing)")
}
