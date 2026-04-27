package cli

import "github.com/spf13/cobra"

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "envoy",
		Short: "Manage and sync .env files across environments using encrypted vaults",
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

	return root
}
