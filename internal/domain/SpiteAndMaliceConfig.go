package domain

import "encoding/json"

// SpiteAndMaliceCpuDifficulty CPUの難易度
type SpiteAndMaliceCpuDifficulty int

// Spite & Malice の難易度定数
const (
	// SpiteAndMaliceCpuDifficultyEasy 見つけた順に1枚だけ出す
	SpiteAndMaliceCpuDifficultyEasy SpiteAndMaliceCpuDifficulty = iota
	// SpiteAndMaliceCpuDifficultyNormal ゴール優先で貪欲に出し切る
	SpiteAndMaliceCpuDifficultyNormal
	// SpiteAndMaliceCpuDifficultyHard 貪欲 + 相手の進行を妨害するディスカード戦略
	SpiteAndMaliceCpuDifficultyHard
)

// Spite & Malice のゴールパイルサイズ範囲
const (
	// SpiteAndMaliceGoalSizeMin ゴールパイル最小枚数
	SpiteAndMaliceGoalSizeMin = 5
	// SpiteAndMaliceGoalSizeMax ゴールパイル最大枚数
	SpiteAndMaliceGoalSizeMax = 30
	// SpiteAndMaliceGoalSizeDefault ゴールパイル既定枚数
	SpiteAndMaliceGoalSizeDefault = 20
)

// SpiteAndMaliceConfig Spite & Malice ゲーム設定
type SpiteAndMaliceConfig struct {
	GoalSize      int
	CpuDifficulty SpiteAndMaliceCpuDifficulty
}

// DefaultSpiteAndMaliceConfig デフォルト設定を返す
func DefaultSpiteAndMaliceConfig() SpiteAndMaliceConfig {
	return SpiteAndMaliceConfig{
		GoalSize:      SpiteAndMaliceGoalSizeDefault,
		CpuDifficulty: SpiteAndMaliceCpuDifficultyNormal,
	}
}

// Validate 設定値の検証
func (c SpiteAndMaliceConfig) Validate() error {
	if err := ValidateRange("goal size", c.GoalSize, SpiteAndMaliceGoalSizeMin, SpiteAndMaliceGoalSizeMax); err != nil {
		return err
	}
	return ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(SpiteAndMaliceCpuDifficultyEasy), int(SpiteAndMaliceCpuDifficultyHard))
}

// spiteAndMaliceConfigJSON is the JSON wire format for SpiteAndMaliceConfig.
type spiteAndMaliceConfigJSON struct {
	GoalSize      int                         `json:"gs"`
	CpuDifficulty SpiteAndMaliceCpuDifficulty `json:"cd"`
}

// MarshalJSON implements json.Marshaler.
func (c SpiteAndMaliceConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(spiteAndMaliceConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *SpiteAndMaliceConfig) UnmarshalJSON(data []byte) error {
	var j spiteAndMaliceConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	c.GoalSize = j.GoalSize
	c.CpuDifficulty = j.CpuDifficulty
	return nil
}
