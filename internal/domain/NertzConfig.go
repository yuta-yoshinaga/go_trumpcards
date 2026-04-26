package domain

import (
	"encoding/json"
	"fmt"
)

// NertzCpuDifficulty CPUの難易度
type NertzCpuDifficulty int

// Nertzの難易度定数
const (
	// NertzCpuDifficultyEasy 1tickあたり1手のみ。優先度の低い手も含めて先頭から実行。
	NertzCpuDifficultyEasy NertzCpuDifficulty = iota
	// NertzCpuDifficultyNormal 1tickあたり最大3手。貪欲にナッツパイル/ファウンデーション優先。
	NertzCpuDifficultyNormal
	// NertzCpuDifficultyHard 1tickあたり最大5手。先読みスコアで最適手を選択。
	NertzCpuDifficultyHard
)

// Nertz のプレイヤー数定数
const (
	// NertzPlayerCntMin 最小プレイヤー数 (人間 + CPU)
	NertzPlayerCntMin = 2
	// NertzPlayerCntMax 最大プレイヤー数
	NertzPlayerCntMax = 6
	// NertzPlayerCntDefault 既定プレイヤー数 (人間 1 + CPU 3)
	NertzPlayerCntDefault = 4
)

// Nertz の目標スコア定数
const (
	// NertzTargetScoreMin 最小目標スコア
	NertzTargetScoreMin = 25
	// NertzTargetScoreMax 最大目標スコア
	NertzTargetScoreMax = 500
	// NertzTargetScoreDefault 既定目標スコア
	NertzTargetScoreDefault = 100
)

// Nertz の CPU 1tick あたり手数の上限
const (
	// NertzCpuTickMovesMax CPU の 1tick あたり手数の絶対上限
	NertzCpuTickMovesMax = 20
)

// NertzConfig Nertz ゲーム設定
type NertzConfig struct {
	PlayerCount   int
	DrawCount     int // 1 or 3
	TargetScore   int
	CpuDifficulty NertzCpuDifficulty
	// CpuTickMoves は明示的な 1tick あたり CPU 手数。0 の場合は CpuDifficulty から自動算出。
	CpuTickMoves int
}

// DefaultNertzConfig デフォルト設定を返す
func DefaultNertzConfig() NertzConfig {
	return NertzConfig{
		PlayerCount:   NertzPlayerCntDefault,
		DrawCount:     3,
		TargetScore:   NertzTargetScoreDefault,
		CpuDifficulty: NertzCpuDifficultyNormal,
		CpuTickMoves:  0,
	}
}

// Validate 設定値の検証
func (c NertzConfig) Validate() error {
	if err := ValidateRange("player count", c.PlayerCount, NertzPlayerCntMin, NertzPlayerCntMax); err != nil {
		return err
	}
	if c.DrawCount != 1 && c.DrawCount != 3 {
		return fmt.Errorf("draw count must be 1 or 3, got %d", c.DrawCount)
	}
	if err := ValidateRange("target score", c.TargetScore, NertzTargetScoreMin, NertzTargetScoreMax); err != nil {
		return err
	}
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(NertzCpuDifficultyEasy), int(NertzCpuDifficultyHard)); err != nil {
		return err
	}
	return ValidateRange("CPU tick moves", c.CpuTickMoves, 0, NertzCpuTickMovesMax)
}

// ResolvedCpuTickMoves CpuTickMoves が 0 の場合に CpuDifficulty から自動算出した値を返す。
func (c NertzConfig) ResolvedCpuTickMoves() int {
	if c.CpuTickMoves > 0 {
		return c.CpuTickMoves
	}
	switch c.CpuDifficulty {
	case NertzCpuDifficultyEasy:
		return 1
	case NertzCpuDifficultyHard:
		return 5
	default:
		return 3
	}
}

// nertzConfigJSON is the JSON wire format for NertzConfig.
type nertzConfigJSON struct {
	PlayerCount   int                `json:"pc"`
	DrawCount     int                `json:"dc"`
	TargetScore   int                `json:"ts"`
	CpuDifficulty NertzCpuDifficulty `json:"cd"`
	CpuTickMoves  int                `json:"tm"`
}

// MarshalJSON implements json.Marshaler.
func (c NertzConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(nertzConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *NertzConfig) UnmarshalJSON(data []byte) error {
	var j nertzConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	c.PlayerCount = j.PlayerCount
	c.DrawCount = j.DrawCount
	c.TargetScore = j.TargetScore
	c.CpuDifficulty = j.CpuDifficulty
	c.CpuTickMoves = j.CpuTickMoves
	return nil
}
