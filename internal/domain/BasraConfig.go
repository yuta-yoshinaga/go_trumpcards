//go:build !js || !wasm || extra3

package domain

import "encoding/json"

// BasraCpuDifficulty は CPU の難易度レベル。
type BasraCpuDifficulty int

// Basra の CPU 難易度定数
const (
	// BasraCpuDifficultyEasy 低難易度 (合法手からランダム)
	BasraCpuDifficultyEasy BasraCpuDifficulty = iota
	// BasraCpuDifficultyNormal 中難易度 (捕獲を優先)
	BasraCpuDifficultyNormal
	// BasraCpuDifficultyHard 高難易度 (バスラ/高得点札を狙う)
	BasraCpuDifficultyHard
)

// BasraDifficultyNames 難易度名マップ
var BasraDifficultyNames = map[BasraCpuDifficulty]string{
	BasraCpuDifficultyEasy:   "Easy",
	BasraCpuDifficultyNormal: "Normal",
	BasraCpuDifficultyHard:   "Hard",
}

// BasraConfig はバスラ (Basra / Bastra) のローカルルール設定。
type BasraConfig struct {
	// CpuDifficulty CPU 難易度
	CpuDifficulty BasraCpuDifficulty `json:"cd"`
}

// DefaultBasraConfig はデフォルトのローカルルール設定を返す。
//   - プレイヤー数: 4 (1 human + 3 CPU, 個人戦)
//   - デッキ: 52 枚
//   - 手札 4 枚配り / 場札 4 枚
//   - CPU 難易度: 3 段階
func DefaultBasraConfig() BasraConfig {
	return BasraConfig{
		CpuDifficulty: BasraCpuDifficultyNormal,
	}
}

// Validate は設定値のドメインバリデーションを行う。
func (c BasraConfig) Validate() error {
	return ValidateRange(
		"CPU difficulty",
		int(c.CpuDifficulty),
		int(BasraCpuDifficultyEasy),
		int(BasraCpuDifficultyHard),
	)
}

// basraConfigJSON is the JSON wire format for BasraConfig.
type basraConfigJSON struct {
	CpuDifficulty BasraCpuDifficulty `json:"cd"`
}

// MarshalJSON implements json.Marshaler.
func (c BasraConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(basraConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *BasraConfig) UnmarshalJSON(data []byte) error {
	var j basraConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	*c = BasraConfig(j)
	return nil
}
