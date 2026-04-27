package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

// HistoryEntry records a single mutation event on a vault key.
type HistoryEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Environment string    `json:"environment"`
	Key       string    `json:"key"`
	Action    string    `json:"action"` // set | delete | rotate | import
}

// historyPath returns the path to the history log file.
func historyPath(cfgDir string) string {
	return filepath.Join(cfgDir, ".envoy_history.json")
}

// AppendHistory appends a new entry to the history log.
func AppendHistory(cfgDir, env, key, action string) error {
	entries, _ := loadHistory(cfgDir)
	entries = append(entries, HistoryEntry{
		Timestamp:   time.Now().UTC(),
		Environment: env,
		Key:         key,
		Action:      action,
	})
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(historyPath(cfgDir), data, 0600)
}

func loadHistory(cfgDir string) ([]HistoryEntry, error) {
	data, err := os.ReadFile(historyPath(cfgDir))
	if err != nil {
		if os.IsNotExist(err) {
			return []HistoryEntry{}, nil
		}
		return nil, err
	}
	var entries []HistoryEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

// NewHistoryCmd returns the `history` sub-command.
func NewHistoryCmd() *cobra.Command {
	var envFilter string
	var limit int

	cmd := &cobra.Command{
		Use:   "history",
		Short: "Show recent vault mutation history",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadCfgForCmd(cmd)
			if err != nil {
				return err
			}
			cfgDir := filepath.Dir(cfg.ConfigPath)
			entries, err := loadHistory(cfgDir)
			if err != nil {
				return fmt.Errorf("reading history: %w", err)
			}
			var filtered []HistoryEntry
			for _, e := range entries {
				if envFilter == "" || e.Environment == envFilter {
					filtered = append(filtered, e)
				}
			}
			if len(filtered) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No history found.")
				return nil
			}
			start := 0
			if limit > 0 && len(filtered) > limit {
				start = len(filtered) - limit
			}
			for _, e := range filtered[start:] {
				fmt.Fprintf(cmd.OutOrStdout(), "%s  [%s] %-8s %s\n",
					e.Timestamp.Format(time.RFC3339), e.Environment, e.Action, e.Key)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&envFilter, "env", "", "Filter by environment name")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum number of entries to show (0 = all)")
	cmd.Flags().String("config", "envoy.json", "Path to envoy config file")
	return cmd
}
