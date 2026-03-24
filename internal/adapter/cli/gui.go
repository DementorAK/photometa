package cli

import (
	"github.com/DementorAK/photometa/internal/port"

	"github.com/DementorAK/photometa/internal/adapter/gui"

	"github.com/spf13/cobra"
)

func NewGUICmd(service port.ImageAnalyzer, logger port.Logger) *cobra.Command {
	return &cobra.Command{
		Use:     "gui",
		Aliases: []string{"g"},
		Short:   "Start the graphical user interface",
		RunE: func(cmd *cobra.Command, args []string) error {
			g := gui.NewGUI(service)
			g.Start()
			return nil
		},
	}
}
