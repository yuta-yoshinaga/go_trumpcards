//go:build !js || !wasm || extra

package domain

import "encoding/json"

// HachiHachiCpuDifficulty は CPU の難易度レベル。
type HachiHachiCpuDifficulty int

// Hachi-Hachi の CPU 難易度定数
const (
	// HachiHachiCpuDifficultyEasy 低難易度 (合法手からランダム)
	HachiHachiCpuDifficultyEasy HachiHachiCpuDifficulty = iota
	// HachiHachiCpuDifficultyNormal 中難易度 (捕獲価値を優先)
	HachiHachiCpuDifficultyNormal
	// HachiHachiCpuDifficultyHard 高難易度 (高得点札を優先)
	HachiHachiCpuDifficultyHard
)

// HachiHachiDifficultyNames 難易度名マップ
var HachiHachiDifficultyNames = map[HachiHachiCpuDifficulty]string{
	HachiHachiCpuDifficultyEasy:   "Easy",
	HachiHachiCpuDifficultyNormal: "Normal",
	HachiHachiCpuDifficultyHard:   "Hard",
}

// HachiHachiTargetRoundsMin / Max は対戦ラウンド数の許容範囲。
const (
	HachiHachiTargetRoundsMin = 1
	HachiHachiTargetRoundsMax = 12
)

// HachiHachiConfig は八八 (Hachi-Hachi) のローカルルール設定。
type HachiHachiConfig struct {
	// CpuDifficulty CPU 難易度
	CpuDifficulty HachiHachiCpuDifficulty `json:"cd"`
	// TargetRounds この回数のラウンドを戦って累計得点で勝敗を決する。
	TargetRounds int `json:"tr"`
}

// DefaultHachiHachiConfig はデフォルトのローカルルール設定を返す。
//   - プレイヤー数: 3 (1 human + 2 CPU)
//   - デッキ: 花札 48 枚 (12 か月 × 4)
//   - 手札 7 枚 / 場 6 枚 / 山札 21 枚
//   - 対戦ラウンド数: 3
//   - CPU 難易度: 3 段階
func DefaultHachiHachiConfig() HachiHachiConfig {
	return HachiHachiConfig{
		CpuDifficulty: HachiHachiCpuDifficultyNormal,
		TargetRounds:  3,
	}
}

// Validate は設定値のドメインバリデーションを行う。
func (c HachiHachiConfig) Validate() error {
	if err := ValidateRange(
		"CPU difficulty",
		int(c.CpuDifficulty),
		int(HachiHachiCpuDifficultyEasy),
		int(HachiHachiCpuDifficultyHard),
	); err != nil {
		return err
	}
	return ValidateRange("target rounds", c.TargetRounds, HachiHachiTargetRoundsMin, HachiHachiTargetRoundsMax)
}

// hachihachiConfigJSON is the JSON wire format for HachiHachiConfig.
type hachihachiConfigJSON struct {
	CpuDifficulty HachiHachiCpuDifficulty `json:"cd"`
	TargetRounds  int                     `json:"tr"`
}

// MarshalJSON implements json.Marshaler.
func (c HachiHachiConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(hachihachiConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *HachiHachiConfig) UnmarshalJSON(data []byte) error {
	var j hachihachiConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	*c = HachiHachiConfig(j)
	return nil
}
