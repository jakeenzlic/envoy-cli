package cli

import (
	"github.com/spf13/cobra"
)

// NewRootCmd builds the top-level cobra command and registers all sub-commands.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "envoy",
		Short: "Manage and sync .env files across environments using encrypted vaults",
		Long: `envoy is a CLI tool for managing .env files.

It stores secrets in encrypted vaults so they can be safely versioned or
shared, and provides push / pull commands to sync values between your local
workstation and remote environments (staging, production, …).`,
		SilenceUsage: true,
	}

	root.AddCommand(
		NewInitCmd(),
		NewPushCmd(),
		NewPullCmd(),
	)

	return root
}
