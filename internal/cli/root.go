package cli

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "forge",
	Short: "Full-stack Go framework with integrated tooling",
	Long: `Forge is a modern full-stack Go framework with schema-driven code generation.
Define your resources once, and Forge generates database models, migrations,
REST APIs, and web interfaces automatically.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	// Persistent flag for config file path
	rootCmd.PersistentFlags().StringP("config", "c", "forge.toml", "config file path")

	// Register subcommands
	rootCmd.AddCommand(newInitCmd())
	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newToolCmd())
	rootCmd.AddCommand(newGenerateCmd())
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}
