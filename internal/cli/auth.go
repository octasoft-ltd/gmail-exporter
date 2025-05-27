package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"gmail-exporter/internal/auth"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage Gmail API authentication",
	Long:  `Commands for setting up and managing Gmail API authentication using OAuth 2.0.`,
}

var authSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Set up Gmail API credentials",
	Long: `Set up Gmail API credentials by providing the credentials JSON file
downloaded from Google Cloud Console.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		credentialsFile, _ := cmd.Flags().GetString("credentials-file")
		if credentialsFile == "" {
			return fmt.Errorf("credentials file is required")
		}

		// Ensure the credentials file exists
		if _, err := os.Stat(credentialsFile); os.IsNotExist(err) {
			return fmt.Errorf("credentials file does not exist: %s", credentialsFile)
		}

		// Create the config directory
		configDir := filepath.Dir(viper.GetString("credentials_file"))
		if err := os.MkdirAll(configDir, 0700); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		// Copy credentials file to config location
		targetPath := viper.GetString("credentials_file")
		if err := copyFile(credentialsFile, targetPath); err != nil {
			return fmt.Errorf("failed to copy credentials file: %w", err)
		}

		logrus.WithField("path", targetPath).Info("Credentials file set up successfully")
		fmt.Printf("Credentials file set up at: %s\n", targetPath)
		fmt.Println("Run 'gmail-exporter auth login' to authenticate with Gmail.")

		return nil
	},
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Gmail API",
	Long:  `Authenticate with Gmail API using OAuth 2.0 flow.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		credentialsFile := viper.GetString("credentials_file")
		tokenFile := viper.GetString("token_file")

		authenticator, err := auth.NewAuthenticator(credentialsFile, tokenFile)
		if err != nil {
			return fmt.Errorf("failed to create authenticator: %w", err)
		}

		if err := authenticator.Authenticate(); err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}

		fmt.Println("Authentication successful!")
		return nil
	},
}

var authRefreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refresh authentication token",
	Long:  `Refresh the authentication token if it has expired.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		credentialsFile := viper.GetString("credentials_file")
		tokenFile := viper.GetString("token_file")

		authenticator, err := auth.NewAuthenticator(credentialsFile, tokenFile)
		if err != nil {
			return fmt.Errorf("failed to create authenticator: %w", err)
		}

		if err := authenticator.RefreshToken(); err != nil {
			return fmt.Errorf("token refresh failed: %w", err)
		}

		fmt.Println("Token refreshed successfully!")
		return nil
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check authentication status",
	Long:  `Check the current authentication status and token validity.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		credentialsFile := viper.GetString("credentials_file")
		tokenFile := viper.GetString("token_file")

		authenticator, err := auth.NewAuthenticator(credentialsFile, tokenFile)
		if err != nil {
			return fmt.Errorf("failed to create authenticator: %w", err)
		}

		status, err := authenticator.GetStatus()
		if err != nil {
			return fmt.Errorf("failed to get status: %w", err)
		}

		fmt.Printf("Authentication Status: %s\n", status.Status)
		if status.TokenExpiry != nil {
			fmt.Printf("Token Expires: %s\n", status.TokenExpiry.Format("2006-01-02 15:04:05"))
		}
		if status.Email != "" {
			fmt.Printf("Authenticated Email: %s\n", status.Email)
		}

		return nil
	},
}

func init() {
	// Add subcommands
	authCmd.AddCommand(authSetupCmd)
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authRefreshCmd)
	authCmd.AddCommand(authStatusCmd)

	// Setup command flags
	authSetupCmd.Flags().StringP("credentials-file", "c", "", "Path to credentials JSON file from Google Cloud Console")
	authSetupCmd.MarkFlagRequired("credentials-file")
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = destFile.ReadFrom(sourceFile)
	return err
}
