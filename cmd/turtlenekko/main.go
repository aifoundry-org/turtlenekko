package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/aifoundry-org/turtlenekko/internal/benchmark"
	"github.com/aifoundry-org/turtlenekko/internal/config"
	"github.com/aifoundry-org/turtlenekko/internal/formatter"
	"github.com/spf13/cobra"
)

// Version information (set by build flags)
var (
	Version   = "dev"
	BuildTime = "unknown"
)

// parseLogLevel converts a string log level to slog.Level
func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo // Default to info
	}
}

// setupLogger configures the global slog logger
func setupLogger(level string, output io.Writer) {
	logLevel := parseLogLevel(level)

	handler := slog.NewTextHandler(output, &slog.HandlerOptions{
		Level: logLevel,
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)

	slog.Debug("Logger initialized", "level", level)
}

func main() {
	var configPath string
	var resultsLogPath string
	var outputFormat string
	var logLevel string

	rootCmd := &cobra.Command{
		Use:   "turtlenekko",
		Short: "Turtlenekko - LLM Performance Measurement Tool",
		Long:  "A tool for measuring performance of local LLMs using chat completion endpoints",
	}

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new configuration file",
		Run: func(cmd *cobra.Command, args []string) {
			configFile := "config.yaml"
			if len(args) > 0 {
				configFile = args[0]
			}

			// Check if file already exists
			if _, err := os.Stat(configFile); err == nil {
				slog.Error("Configuration file already exists", "path", configFile)
				os.Exit(1)
			}

			// Write default config
			if err := config.WriteDefaultConfig(configFile); err != nil {
				slog.Error("Failed to write configuration file", "error", err, "path", configFile)
				os.Exit(1)
			}

			slog.Info("Configuration file created successfully", "path", configFile)
		},
	}

	benchmarkCmd := &cobra.Command{
		Use:   "benchmark",
		Short: "Run a benchmark against an LLM",
		Run: func(cmd *cobra.Command, args []string) {
			// Load configuration from YAML file
			cfg, err := config.Load(configPath)
			if err != nil {
				if os.IsNotExist(err) {
					slog.Info("You can create a new config file with the example above")
					slog.Info("Or run: turtlenekko init to create a default config file")
					os.Exit(1)
				}
				slog.Error("Error loading configuration", "error", err)
				os.Exit(1)
			}

			// Create results log file
			resultsFile, err := os.Create(resultsLogPath)
			if err != nil {
				slog.Error("Error creating results log file", "error", err, "path", resultsLogPath)
				os.Exit(1)
			}
			defer resultsFile.Close()

			// Run matrix benchmarks
			matrixResults, err := benchmark.RunMatrix(cfg.Driver, nil, cfg.Matrix)
			if err != nil {
				slog.Error("Matrix benchmark failed", "error", err)
				fmt.Fprintf(resultsFile, "Matrix benchmark failed: %v\n", err)
				os.Exit(1)
			}

			// Format and print results based on the selected format
			switch outputFormat {
			case "json":
				if err := formatter.FormatJSON(matrixResults); err != nil {
					slog.Error("Error formatting JSON", "error", err)
				}
			case "text":
				formatter.FormatText(matrixResults)
			case "csv":
				formatter.FormatCSV(matrixResults)
			default:
				slog.Warn("Unknown format, using text format", "format", outputFormat)
				formatter.FormatText(matrixResults)
			}

			// Always write detailed results to the log file
			formatter.WriteToFile(resultsFile, matrixResults)

			slog.Info("Results have been saved", "path", resultsLogPath)
		},
	}

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "info", "Log level (debug, info, warn, error)")

	// Benchmark command flags
	benchmarkCmd.Flags().StringVarP(&configPath, "config", "c", "config.yaml", "Path to configuration file")
	benchmarkCmd.Flags().StringVarP(&resultsLogPath, "results", "r", "results.log", "Path to results log file")
	benchmarkCmd.Flags().StringVarP(&outputFormat, "format", "f", "json", "Output format (csv, text, json)")

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Turtlenekko version %s (built at %s)\n", Version, BuildTime)
		},
	}

	rootCmd.AddCommand(benchmarkCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(versionCmd)

	// Initialize the logger before executing commands
	cobra.OnInitialize(func() {
		setupLogger(logLevel, os.Stderr)
	})

	if err := rootCmd.Execute(); err != nil {
		slog.Error("Command execution failed", "error", err)
		os.Exit(1)
	}
}

