package cli

import "github.com/spf13/cobra"

// NewRootCmd builds the top-level cobra command and registers all sub-commands.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "envoy",
		Short: "Manage and sync .env files across environments using encrypted vaults",
	}

	root.AddCommand(
		NewInitCmd(),
		NewPushCmd(),
		NewPullCmd(),
		NewDiffCmd(),
		NewListCmd(),
		NewSetCmd(),
		NewGetCmd(),
		NewDeleteCmd(),
		NewImportCmd(),
		NewExportCmd(),
		NewCopyCmd(),
		NewRotateCmd(),
		NewRenameCmd(),
	)

	return root
}
