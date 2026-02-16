package cli

import "github.com/spf13/cobra"

func newToolCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tool",
		Short: "Manage tool binaries",
		Long: `Manage external tool binaries required by Forge.

Forge depends on several external tools (templ, sqlc, tailwind, atlas)
but downloads them automatically as needed. Use these commands to
manually manage the tool binaries.`,
	}

	// Register subcommands
	cmd.AddCommand(newToolSyncCmd())

	return cmd
}
