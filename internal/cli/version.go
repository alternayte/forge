package cli

import (
	"fmt"

	"github.com/forge-framework/forge/internal/ui"
	"github.com/spf13/cobra"
)

var (
	// Version information set via ldflags at build time
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			// Format: forge version dev (none) built unknown
			fmt.Printf("%s version %s %s built %s\n",
				ui.BoldStyle.Render("forge"),
				Version,
				ui.DimStyle.Render(fmt.Sprintf("(%s)", Commit)),
				ui.DimStyle.Render(Date),
			)
		},
	}
}
