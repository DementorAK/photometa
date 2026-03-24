package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/DementorAK/photometa/internal/platform/locale"
	"github.com/DementorAK/photometa/internal/platform/version"
	"github.com/DementorAK/photometa/internal/port"

	"github.com/DementorAK/photometa/internal/domain"

	"github.com/spf13/cobra"
)

var (
	pathFlag   string
	localeFlag string
)

func NewRootCmd(service port.ImageAnalyzer, logger port.Logger) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "photometa [path]",
		Short: "Photo Metadata Viewer",
		Long:  "A tool for analyzing image metadata via CLI, GUI, or Web Server.",
		// This line is CRITICAL: it tells Cobra to allow positional arguments
		// and pass them to the Run function instead of failing with "unknown command".
		Args: cobra.ArbitraryArgs,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// If --locale was passed with a value, set the locale.
			if localeFlag != "" {
				locale.SetLocale(localeFlag)
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check for Stdin Pipe
			stat, _ := os.Stdin.Stat()
			hasPipe := (stat.Mode() & os.ModeCharDevice) == 0

			// CLI Mode (when pipe, path flag, or file arguments are provided)
			if hasPipe || pathFlag != "" || len(args) > 0 {
				return runCLI(cmd, service, logger, pathFlag, args, hasPipe)
			}

			// If no flags or args, show help
			return cmd.Help()
		},
	}

	// Register Subcommands
	rootCmd.AddCommand(NewServerCmd(service, logger))
	rootCmd.AddCommand(NewGUICmd(service, logger))
	rootCmd.AddCommand(NewLocalesCmd())
	rootCmd.AddCommand(NewVersionCmd())

	// Root-specific flags
	rootCmd.Flags().StringVarP(&pathFlag, "path", "p", "", "Directory to scan")

	// Persistent locale flag (available on root and all subcommands).
	rootCmd.PersistentFlags().StringVarP(&localeFlag, "locale", "l", "", "Set display language")

	return rootCmd
}

// NewLocalesCmd creates a cobra command for listing available locales.
func NewLocalesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "locales",
		Short: "List available display languages",
		RunE: func(cmd *cobra.Command, args []string) error {
			return encodeJSON(locale.GetLocales())
		},
	}
}

// NewVersionCmd creates a cobra command for displaying version information.
func NewVersionCmd() *cobra.Command {
	var jsonFlag bool

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			if jsonFlag {
				info := map[string]string{
					"version": version.Version,
					"commit":  version.Commit,
					"date":    version.Date,
				}
				return encodeJSON(info)
			}
			fmt.Printf("Version: %s", version.Version)
			if version.Date != "" {
				fmt.Printf(" from %s", version.Date)
			}
			fmt.Println()
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonFlag, "json", false, "Output version as JSON")

	return cmd
}

func runCLI(cmd *cobra.Command, s port.ImageAnalyzer, logger port.Logger, path string, args []string, hasPipe bool) error {
	var results []domain.ImageFile

	logger.Info("Starting CLI analysis", "hasPipe", hasPipe, "pathFlag", path, "argsCount", len(args))

	// 1. Handle Pipe (Stdin)
	if hasPipe {
		img, err := s.AnalyzeStream(cmd.Context(), os.Stdin, "<stdin>", 0)
		if err != nil {
			return fmt.Errorf("reading stdin: %w", err)
		}
		logger.Info("Analyzed stream from stdin", "format", img.Metadata.Format, "size", img.Metadata.FileSize)
		// If ONLY pipe, return single object immediately as before
		if path == "" && len(args) == 0 {
			return encodeJSON(img)
		}
		results = append(results, *img)
	}

	// 2. Handle Path Flag
	if path != "" {
		imgs, err := s.ScanDirectory(cmd.Context(), path)
		if err != nil {
			return fmt.Errorf("scanning directory %s: %w", path, err)
		}
		logger.Info("Scanned directory from flag", "path", path, "imagesFound", len(imgs))
		results = append(results, imgs...)
	}

	// 3. Handle Positional Args
	for _, arg := range args {
		info, err := os.Stat(arg)
		if err != nil {
			logger.Warn("Error accessing argument", "arg", arg, "error", err)
			continue
		}

		if info.IsDir() {
			imgs, err := s.ScanDirectory(cmd.Context(), arg)
			if err != nil {
				logger.Warn("Error scanning directory", "path", arg, "error", err)
				continue
			}
			results = append(results, imgs...)
		} else {
			img, err := s.AnalyzeFile(cmd.Context(), arg)
			if err != nil {
				logger.Warn("Error analyzing file", "path", arg, "error", err)
				continue
			}
			results = append(results, *img)
		}
	}

	logger.Info("CLI analysis complete", "totalResults", len(results))

	// 4. Output logic
	if len(results) == 0 {
		return nil
	}

	// If the user provided EXACTLY one positional argument that turned out to be a file,
	// return a single object instead of an array of one for better UX.
	if len(results) == 1 && path == "" && len(args) == 1 && !hasPipe {
		// Check if the single arg was a file (not a dir that happened to have 1 file)
		firstArgInfo, err := os.Stat(args[0])
		if err == nil && !firstArgInfo.IsDir() {
			return encodeJSON(results[0])
		}
	}

	return encodeJSON(results)
}

func encodeJSON(v interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(v); err != nil {
		return fmt.Errorf("encoding JSON: %w", err)
	}
	return nil
}
