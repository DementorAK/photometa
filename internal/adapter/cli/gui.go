package cli

import (
	"github.com/DementorAK/photometa/internal/port"

	"github.com/DementorAK/photometa/internal/adapter/gui"

	"github.com/spf13/cobra"
)

func NewGUICmd(service port.ImageAnalyzer, _ port.Logger) *cobra.Command {
	return &cobra.Command{
		Use:     "gui",
		Aliases: []string{"g"},
		Short:   "Start the graphical user interface",
		RunE: func(_ *cobra.Command, _ []string) error {
			g := gui.NewGUI(service)
			g.Start()
			return nil
		},
	}
}
