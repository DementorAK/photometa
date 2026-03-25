package cli

import (
	"fmt"

	"github.com/DementorAK/photometa/internal/port"

	"github.com/DementorAK/photometa/internal/adapter/http"

	"github.com/spf13/cobra"
)

// NewServerCmd creates a Cobra command for starting the HTTP server mode.
// It accepts an ImageAnalyzer service to handle the core logic.
func NewServerCmd(service port.ImageAnalyzer, logger port.Logger) *cobra.Command {
	var portFlag string

	cmd := &cobra.Command{
		Use:     "server",
		Aliases: []string{"s"},
		Short:   "Start the metadata analysis web server",
		RunE: func(_ *cobra.Command, _ []string) error {
			srv := http.NewServer(service)
			logger.Info("Starting server", "port", portFlag)
			if err := srv.Start(portFlag); err != nil {
				return fmt.Errorf("server error: %w", err)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&portFlag, "port", "p", "8080", "Port for web server")

	return cmd
}
