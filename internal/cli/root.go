package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile  string
	logLevel string
	logFile  string
	verbose  bool

	// Version information
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gmail-exporter",
	Short: "A powerful CLI tool for exporting, importing, and managing Gmail messages",
	Long: `Gmail Exporter is a comprehensive tool for managing Gmail messages with advanced
filtering capabilities, OAuth authentication, and comprehensive metrics collection.

Features:
- OAuth authentication with Gmail API
- Advanced email filtering (To, From, Subject, Size, Date, Labels, etc.)
- Export emails to local filesystem with optional label-based organization
- Import/forward emails to another Gmail account
- Archive or delete processed emails
- Generate filter files from existing exports for cleanup operations
- Comprehensive metrics in JSON and Prometheus formats
- Progress tracking and resumable operations
- Parallel and serial processing options`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		initLogging()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

// SetVersion sets the version information
func SetVersion(v, c, d string) {
	version = v
	commit = c
	date = d
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gmail-exporter.yaml)")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVar(&logFile, "log-file", "", "log file path (default: stderr)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Bind flags to viper
	if err := viper.BindPFlag("log_level", rootCmd.PersistentFlags().Lookup("log-level")); err != nil {
		logrus.WithError(err).Fatal("Failed to bind log-level flag")
	}
	if err := viper.BindPFlag("log_file", rootCmd.PersistentFlags().Lookup("log-file")); err != nil {
		logrus.WithError(err).Fatal("Failed to bind log-file flag")
	}
	if err := viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose")); err != nil {
		logrus.WithError(err).Fatal("Failed to bind verbose flag")
	}

	// Add subcommands
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(cleanupCmd)
	rootCmd.AddCommand(workflowCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(generateFilterCmd)
	rootCmd.AddCommand(versionCmd)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".gmail-exporter" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".gmail-exporter")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// Set default values
	viper.SetDefault("credentials_file", filepath.Join(os.Getenv("HOME"), ".gmail-exporter", "credentials.json"))
	viper.SetDefault("token_file", filepath.Join(os.Getenv("HOME"), ".gmail-exporter", "token.json"))
	viper.SetDefault("output_dir", "./exports")
	viper.SetDefault("parallel_workers", 3)
	viper.SetDefault("organize_by_labels", false)
	viper.SetDefault("filters.exclude_chats", true)
	viper.SetDefault("filters.search_scope", "all_mail")
	viper.SetDefault("metrics.enabled", true)
	viper.SetDefault("metrics.format", "json")
	viper.SetDefault("metrics.output_file", "metrics.json")
	viper.SetDefault("log_level", "info")

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		logrus.WithField("config_file", viper.ConfigFileUsed()).Debug("Using config file")
	}
}

// initLogging configures the logging system
func initLogging() {
	// Set log level
	level := viper.GetString("log_level")
	if verbose {
		level = "debug"
	}

	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logrus.WithError(err).Warn("Invalid log level, using info")
		logLevel = logrus.InfoLevel
	}
	logrus.SetLevel(logLevel)

	// Set log format
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})

	// Set log output
	logFile := viper.GetString("log_file")
	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
		if err != nil {
			logrus.WithError(err).Warn("Failed to open log file, using stderr")
		} else {
			logrus.SetOutput(file)
		}
	}
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Gmail Exporter %s\n", version)
		if commit != "unknown" {
			fmt.Printf("Commit: %s\n", commit)
		}
		if date != "unknown" {
			fmt.Printf("Built: %s\n", date)
		}
	},
}
