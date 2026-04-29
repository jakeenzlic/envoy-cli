package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

type Snapshot struct {
	Env       string            `json:"env"`
	CreatedAt time.Time         `json:"created_at"`
	Label     string            `json:"label,omitempty"`
	Keys      map[string]string `json:"keys"`
}

func snapshotDir(configPath string) string {
	return filepath.Join(filepath.Dir(configPath), ".snapshots")
}

func NewSnapshotCmd() *cobra.Command {
	var label string

	cmd := &cobra.Command{
		Use:   "snapshot <env>",
		Short: "Save a named snapshot of an environment vault",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			env := args[0]
			cfg, err := loadCfgForCmd(cmd)
			if err != nil {
				return err
			}
			pass, err := passphraseForCmd(cmd, cfg, env)
			if err != nil {
				return err
			}
			vaultPath := vaultPathForEnv(cfg, env)
			v, err := openVaultForEnv(vaultPath, pass)
			if err != nil {
				return fmt.Errorf("open vault: %w", err)
			}
			keys := v.Keys()
			data := make(map[string]string, len(keys))
			for _, k := range keys {
				val, _ := v.Get(k)
				data[k] = val
			}
			snap := Snapshot{
				Env:       env,
				CreatedAt: time.Now().UTC(),
				Label:     label,
				Keys:      data,
			}
			cfgPath, _ := cmd.Flags().GetString("config")
			dir := snapshotDir(cfgPath)
			if err := os.MkdirAll(dir, 0700); err != nil {
				return err
			}
			ts := snap.CreatedAt.Format("20060102T150405")
			fileName := fmt.Sprintf("%s_%s.json", env, ts)
			path := filepath.Join(dir, fileName)
			f, err := os.Create(path)
			if err != nil {
				return err
			}
			defer f.Close()
			enc := json.NewEncoder(f)
			enc.SetIndent("", "  ")
			if err := enc.Encode(snap); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "snapshot saved: %s\n", path)
			return nil
		},
	}
	cmd.Flags().StringVar(&label, "label", "", "optional human-readable label for the snapshot")
	return cmd
}
