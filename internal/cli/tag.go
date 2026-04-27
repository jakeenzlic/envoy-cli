package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

// NewTagCmd returns a command that lists keys matching a given prefix tag pattern.
// Tags are inferred from key naming conventions: TAG_KEY_NAME → tag "TAG".
func NewTagCmd() *cobra.Command {
	var env string

	cmd := &cobra.Command{
		Use:   "tag [tag]",
		Short: "List keys grouped by tag prefix",
		Long:  "Displays keys that share a common prefix (tag). If no tag is given, lists all discovered tags.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadCfgForCmd(cmd)
			if err != nil {
				return err
			}

			passphrase, err := passphraseForCmd(cfg, env)
			if err != nil {
				return err
			}

			vaultPath := vaultPathForEnv(cfg, env)
			v, err := openVaultForEnv(vaultPath, passphrase)
			if err != nil {
				return fmt.Errorf("open vault: %w", err)
			}

			keys := v.Keys().All()
			sort.Strings(keys)

			if len(args) == 0 {
				return printAllTags(cmd, keys)
			}

			filter := strings.ToUpper(args[0])
			return printTaggedKeys(cmd, keys, filter)
		},
	}

	cmd.Flags().StringVarP(&env, "env", "e", "local", "Target environment")
	return cmd
}

func extractTag(key string) string {
	parts := strings.SplitN(key, "_", 2)
	if len(parts) < 2 {
		return ""
	}
	return parts[0]
}

func printAllTags(cmd *cobra.Command, keys []string) error {
	seen := map[string]int{}
	for _, k := range keys {
		if t := extractTag(k); t != "" {
			seen[t]++
		}
	}
	if len(seen) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "no tags found")
		return nil
	}
	tags := make([]string, 0, len(seen))
	for t := range seen {
		tags = append(tags, t)
	}
	sort.Strings(tags)
	for _, t := range tags {
		fmt.Fprintf(cmd.OutOrStdout(), "%s (%d keys)\n", t, seen[t])
	}
	return nil
}

func printTaggedKeys(cmd *cobra.Command, keys []string, tag string) error {
	matched := []string{}
	prefix := tag + "_"
	for _, k := range keys {
		if strings.HasPrefix(k, prefix) {
			matched = append(matched, k)
		}
	}
	if len(matched) == 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "no keys found for tag %q\n", tag)
		return nil
	}
	for _, k := range matched {
		fmt.Fprintln(cmd.OutOrStdout(), k)
	}
	return nil
}
