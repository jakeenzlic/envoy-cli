package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ImportEnvFile reads a .env file and loads its key-value pairs into the vault.
// Lines starting with '#' and empty lines are ignored.
// Inline comments (after a value) are stripped.
func (v *Vault) ImportEnvFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("vault: cannot open env file: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("vault: invalid syntax at line %d: %q", lineNum, line)
		}
		key := strings.TrimSpace(parts[0])
		value := stripInlineComment(strings.TrimSpace(parts[1]))
		value = strings.Trim(value, `"'`)
		v.Set(key, value)
	}
	return scanner.Err()
}

// ExportEnvFile writes the vault entries to a .env file.
func (v *Vault) ExportEnvFile(path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("vault: cannot write env file: %w", err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for k, val := range v.Entries {
		if _, err := fmt.Fprintf(w, "%s=%q\n", k, val); err != nil {
			return err
		}
	}
	return w.Flush()
}

// stripInlineComment removes trailing inline comments from a value.
// e.g. "myvalue # this is a comment" -> "myvalue"
func stripInlineComment(value string) string {
	if idx := strings.Index(value, " #"); idx != -1 {
		return strings.TrimSpace(value[:idx])
	}
	return value
}
