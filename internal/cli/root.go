package cli

import (
	"github.com/spf13/cobra"
)

// NewRootCmd builds and returns the root cobra command with all sub-commands registered.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "envoy",
		Short: "Manage and sync .env files across environments using encrypted vaults",
		Long: `envoy-cli helps you securely manage environment variables across
local, staging, and production environments using encrypted vaults.`,
	}

	root.AddCommand(NewInitCmd())
	root.AddCommand(NewPushCmd())
	root.AddCommand(NewPullCmd())
	root.AddCommand(NewDiffCmd())

	return root
}
