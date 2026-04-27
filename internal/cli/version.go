package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version is the current version of envoy-cli.
	// These are set at build time via ldflags.
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

// NewVersionCmd returns a cobra command that prints version information.
func NewVersionCmd() *cobra.Command {
	var short bool

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version of envoy-cli",
		Long:  `Print the current version, git commit, and build date of envoy-cli.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if short {
				fmt.Fprintln(cmd.OutOrStdout(), Version)
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "envoy-cli version %s\n", Version)
			fmt.Fprintf(cmd.OutOrStdout(), "  commit:     %s\n", Commit)
			fmt.Fprintf(cmd.OutOrStdout(), "  build date: %s\n", BuildDate)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&short, "short", "s", false, "Print only the version number")

	return cmd
}
