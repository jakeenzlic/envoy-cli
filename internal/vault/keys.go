package vault

import "sort"

// Keys returns a sorted slice of all key names stored in the vault.
func (v *Vault) Keys() []string {
	v.mu.RLock()
	defer v.mu.RUnlock()

	keys := make([]string, 0, len(v.data))
	for k := range v.data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Len returns the number of keys stored in the vault.
func (v *Vault) Len() int {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return len(v.data)
}
