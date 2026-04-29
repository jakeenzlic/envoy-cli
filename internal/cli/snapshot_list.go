package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

func NewSnapshotListCmd() *cobra.Command {
	var env string

	cmd := &cobra.Command{
		Use:   "snapshot-list",
		Short: "List saved snapshots, optionally filtered by environment",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgPath, _ := cmd.Flags().GetString("config")
			dir := snapshotDir(cfgPath)
			entries, err := os.ReadDir(dir)
			if err != nil {
				if os.IsNotExist(err) {
					fmt.Fprintln(cmd.OutOrStdout(), "no snapshots found")
					return nil
				}
				return err
			}
			var snaps []Snapshot
			for _, e := range entries {
				if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
					continue
				}
				path := filepath.Join(dir, e.Name())
				f, err := os.Open(path)
				if err != nil {
					continue
				}
				var s Snapshot
				if err := json.NewDecoder(f).Decode(&s); err != nil {
					f.Close()
					continue
				}
				f.Close()
				if env == "" || s.Env == env {
					snaps = append(snaps, s)
				}
			}
			if len(snaps) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no snapshots found")
				return nil
			}
			sort.Slice(snaps, func(i, j int) bool {
				return snaps[i].CreatedAt.Before(snaps[j].CreatedAt)
			})
			for _, s := range snaps {
				label := s.Label
				if label == "" {
					label = "(no label)"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "[%s] env=%s keys=%d  %s\n",
					s.CreatedAt.Format("2006-01-02 15:04:05"), s.Env, len(s.Keys), label)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&env, "env", "", "filter by environment name")
	return cmd
}
