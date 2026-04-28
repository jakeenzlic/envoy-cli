package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// HistoryEntry records a single key mutation.
type HistoryEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Env       string    `json:"env"`
	Key       string    `json:"key"`
	OldValue  string    `json:"old_value"`
	NewValue  string    `json:"new_value"`
}

// HistoryPath returns the path to the history file for the given env,
// stored relative to the config file directory.
func HistoryPath(baseDir, env string) string {
	return filepath.Join(baseDir, ".envoy_history", env+".jsonl")
}

// historyPath is the internal helper used by other CLI commands.
func historyPath(cfgPath, env string) string {
	base := filepath.Dir(cfgPath)
	if env == "" {
		return filepath.Join(base, ".envoy_history")
	}
	return HistoryPath(base, env)
}

// AppendHistory appends a single entry to the history file for the given env.
func AppendHistory(histDir, env, key, oldVal, newVal string) error {
	hp := HistoryPath(histDir, env)
	if err := os.MkdirAll(filepath.Dir(hp), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(hp, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	entry := HistoryEntry{
		Timestamp: time.Now().UTC(),
		Env:       env,
		Key:       key,
		OldValue:  oldVal,
		NewValue:  newVal,
	}
	return json.NewEncoder(f).Encode(entry)
}

// loadHistory reads all history entries for an env from the given base dir.
func loadHistory(baseDir, env string) ([]HistoryEntry, error) {
	hp := HistoryPath(baseDir, env)
	f, err := os.Open(hp)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var entries []HistoryEntry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var e HistoryEntry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			continue
		}
		entries = append(entries, e)
	}
	return entries, scanner.Err()
}

// NewHistoryCmd returns a cobra command for viewing key mutation history.
func NewHistoryCmd() *cobra.Command {
	var (
		cfgPath string
		env     string
		limit   int
	)

	cmd := &cobra.Command{
		Use:   "history",
		Short: "Show mutation history for an environment",
		RunE: func(cmd *cobra.Command, args []string) error {
			histDir := historyPath(cfgPath, "")
			entries, err := loadHistory(histDir, env)
			if err != nil {
				return fmt.Errorf("load history: %w", err)
			}
			if len(entries) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No history found.")
				return nil
			}
			start := 0
			if limit > 0 && len(entries) > limit {
				start = len(entries) - limit
			}
			for _, e := range entries[start:] {
				fmt.Fprintf(cmd.OutOrStdout(), "%s  [%s] %s: %s -> %s\n",
					e.Timestamp.Format(time.RFC3339),
					e.Env, e.Key,
					quotedOrEmpty(e.OldValue),
					quotedOrEmpty(e.NewValue),
				)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&cfgPath, "config", "envoy.json", "path to envoy config")
	cmd.Flags().StringVar(&env, "env", "local", "environment to show history for")
	cmd.Flags().IntVar(&limit, "limit", 0, "max entries to show (0 = all)")
	return cmd
}

func quotedOrEmpty(s string) string {
	if s == "" {
		return "(empty)"
	}
	return strconv.Quote(s)
}
