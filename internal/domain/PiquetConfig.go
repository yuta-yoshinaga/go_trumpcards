//go:build !js || !wasm || extra8

package domain

import "encoding/json"

// PiquetDealsPerPartie 1パルティ (試合) のディール数 (古典的な6ディール)
const PiquetDealsPerPartie = 6

// PiquetConfig Piquet ゲーム設定
type PiquetConfig struct {
	DealsPerPartie int // 1パルティのディール数 (default = 6)
}

// DefaultPiquetConfig returns the default configuration
func DefaultPiquetConfig() PiquetConfig {
	return PiquetConfig{
		DealsPerPartie: PiquetDealsPerPartie,
	}
}

// Validate domain validation for the configuration
func (c PiquetConfig) Validate() error {
	if err := ValidateMin("deals per partie", c.DealsPerPartie, 1); err != nil {
		return err
	}
	return nil
}

// piquetConfigJSON is the JSON wire format for PiquetConfig.
type piquetConfigJSON struct {
	DealsPerPartie int `json:"dp"`
}

// MarshalJSON implements json.Marshaler.
func (c PiquetConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(piquetConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *PiquetConfig) UnmarshalJSON(data []byte) error {
	var j piquetConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	c.DealsPerPartie = j.DealsPerPartie
	return nil
}
