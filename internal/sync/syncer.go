package sync

import (
	"fmt"
	"strings"

	"github.com/envoy-cli/internal/vault"
)

// Direction represents the sync direction.
type Direction string

const (
	DirectionPush Direction = "push"
	DirectionPull Direction = "pull"
)

// Environment represents a target environment.
type Environment string

const (
	EnvLocal      Environment = "local"
	EnvStaging    Environment = "staging"
	EnvProduction Environment = "production"
)

// Syncer handles syncing env variables between vaults.
type Syncer struct {
	source *vault.Vault
	target *vault.Vault
}

// New creates a new Syncer from source to target vault.
func New(source, target *vault.Vault) *Syncer {
	return &Syncer{source: source, target: target}
}

// Sync copies all keys from source to target.
// Existing keys in target are overwritten.
func (s *Syncer) Sync() ([]string, error) {
	keys, err := s.source.Keys()
	if err != nil {
		return nil, fmt.Errorf("sync: listing source keys: %w", err)
	}

	synced := make([]string, 0, len(keys))
	for _, key := range keys {
		val, err := s.source.Get(key)
		if err != nil {
			return synced, fmt.Errorf("sync: reading key %q: %w", key, err)
		}
		if err := s.target.Set(key, val); err != nil {
			return synced, fmt.Errorf("sync: writing key %q: %w", key, err)
		}
		synced = append(synced, key)
	}
	return synced, nil
}

// Diff returns keys that differ between source and target.
func (s *Syncer) Diff() ([]string, error) {
	keys, err := s.source.Keys()
	if err != nil {
		return nil, fmt.Errorf("diff: listing source keys: %w", err)
	}

	var diffKeys []string
	for _, key := range keys {
		srcVal, err := s.source.Get(key)
		if err != nil {
			return nil, fmt.Errorf("diff: reading source key %q: %w", key, err)
		}
		tgtVal, err := s.target.Get(key)
		if err != nil || strings.TrimSpace(srcVal) != strings.TrimSpace(tgtVal) {
			diffKeys = append(diffKeys, key)
		}
	}
	return diffKeys, nil
}
