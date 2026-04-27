package cli

import (
	"github.com/spf13/cobra"
)

// NewRootCmd builds and returns the root cobra command with all sub-commands
// registered.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "envoy",
		Short: "Manage and sync .env files across environments",
		Long: `envoy is a CLI tool for managing and syncing .env files across
local, staging, and production environments using encrypted vaults.`,
		SilenceUsage: true,
	}

	root.AddCommand(NewInitCmd())
	root.AddCommand(NewPushCmd())
	root.AddCommand(NewPullCmd())
	root.AddCommand(NewDiffCmd())
	root.AddCommand(NewListCmd())
	root.AddCommand(NewSetCmd())
	root.AddCommand(NewGetCmd())
	root.AddCommand(NewDeleteCmd())
	root.AddCommand(NewImportCmd())
	root.AddCommand(NewExportCmd())
	root.AddCommand(NewRotateCmd())
	root.AddCommand(NewCopyCmd())
	root.AddCommand(NewRenameCmd())
	root.AddCommand(NewAuditCmd())
	root.AddCommand(NewCompareCmd())
	root.AddCommand(NewSearchCmd())
	root.AddCommand(NewEnvCmd())

	return root
}
