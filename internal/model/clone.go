package model

import "encoding/json"

// CloneConfig returns a deep copy of cfg so callers can safely edit templates or snapshots.
func CloneConfig(cfg *ModelConfig) *ModelConfig {
	if cfg == nil {
		return nil
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		return nil
	}

	var clone ModelConfig
	if err := json.Unmarshal(data, &clone); err != nil {
		return nil
	}
	return &clone
}
