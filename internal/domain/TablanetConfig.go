//go:build !js || !wasm || extra3

package domain

import "encoding/json"

// TablanetCpuDifficulty は CPU の難易度レベル。
type TablanetCpuDifficulty int

// Tablanet の CPU 難易度定数
const (
	// TablanetCpuDifficultyEasy 低難易度 (合法手からランダム)
	TablanetCpuDifficultyEasy TablanetCpuDifficulty = iota
	// TablanetCpuDifficultyNormal 中難易度 (捕獲を優先)
	TablanetCpuDifficultyNormal
	// TablanetCpuDifficultyHard 高難易度 (タブラ/高得点札を狙う)
	TablanetCpuDifficultyHard
)

// TablanetDifficultyNames 難易度名マップ
var TablanetDifficultyNames = map[TablanetCpuDifficulty]string{
	TablanetCpuDifficultyEasy:   "Easy",
	TablanetCpuDifficultyNormal: "Normal",
	TablanetCpuDifficultyHard:   "Hard",
}

// TablanetConfig はタブラネット (Tablanet / Bastra) のローカルルール設定。
type TablanetConfig struct {
	// CpuDifficulty CPU 難易度
	CpuDifficulty TablanetCpuDifficulty `json:"cd"`
}

// DefaultTablanetConfig はデフォルトのローカルルール設定を返す。
//   - プレイヤー数: 4 (1 human + 3 CPU, 個人戦)
//   - デッキ: 52 枚
//   - 手札 4 枚配り / 場札 4 枚
//   - CPU 難易度: 3 段階
func DefaultTablanetConfig() TablanetConfig {
	return TablanetConfig{
		CpuDifficulty: TablanetCpuDifficultyNormal,
	}
}

// Validate は設定値のドメインバリデーションを行う。
func (c TablanetConfig) Validate() error {
	return ValidateRange(
		"CPU difficulty",
		int(c.CpuDifficulty),
		int(TablanetCpuDifficultyEasy),
		int(TablanetCpuDifficultyHard),
	)
}

// tablanetConfigJSON is the JSON wire format for TablanetConfig.
type tablanetConfigJSON struct {
	CpuDifficulty TablanetCpuDifficulty `json:"cd"`
}

// MarshalJSON implements json.Marshaler.
func (c TablanetConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(tablanetConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *TablanetConfig) UnmarshalJSON(data []byte) error {
	var j tablanetConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	*c = TablanetConfig(j)
	return nil
}
