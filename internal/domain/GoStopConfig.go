//go:build !js || !wasm || extra

package domain

import "encoding/json"

// GoStopCpuDifficulty は CPU の難易度レベル。
type GoStopCpuDifficulty int

// Go-Stop の CPU 難易度定数
const (
	// GoStopCpuDifficultyEasy 低難易度 (合法手からランダム、決断は常にストップ)
	GoStopCpuDifficultyEasy GoStopCpuDifficulty = iota
	// GoStopCpuDifficultyNormal 中難易度 (捕獲価値を優先)
	GoStopCpuDifficultyNormal
	// GoStopCpuDifficultyHard 高難易度 (得点を狙い、ゴーで積極的に続行)
	GoStopCpuDifficultyHard
)

// GoStopDifficultyNames 難易度名マップ
var GoStopDifficultyNames = map[GoStopCpuDifficulty]string{
	GoStopCpuDifficultyEasy:   "Easy",
	GoStopCpuDifficultyNormal: "Normal",
	GoStopCpuDifficultyHard:   "Hard",
}

// GoStopTargetScoreMin / Max は目標得点の許容範囲。
const (
	GoStopTargetScoreMin = 1
	GoStopTargetScoreMax = 200
)

// GoStopConfig はゴーストップ (Go-Stop) のローカルルール設定。
type GoStopConfig struct {
	// CpuDifficulty CPU 難易度
	CpuDifficulty GoStopCpuDifficulty `json:"cd"`
	// TargetScore この累計得点に到達したプレイヤーが出た時点でゲーム終了。
	// 到達者がいなくても GoStopMaxRounds ラウンドで打ち切る。
	TargetScore int `json:"ts"`
}

// DefaultGoStopConfig はデフォルトのローカルルール設定を返す。
//   - プレイヤー数: 2 (1 human + 1 CPU)
//   - デッキ: 花札 48 枚 (12 か月 × 4)
//   - 手札 10 枚 / 場 8 枚
//   - 目標得点: 7 (これに到達で終局、最長 GoStopMaxRounds ラウンド)
//   - CPU 難易度: 3 段階
func DefaultGoStopConfig() GoStopConfig {
	return GoStopConfig{
		CpuDifficulty: GoStopCpuDifficultyNormal,
		TargetScore:   7,
	}
}

// Validate は設定値のドメインバリデーションを行う。
func (c GoStopConfig) Validate() error {
	if err := ValidateRange(
		"CPU difficulty",
		int(c.CpuDifficulty),
		int(GoStopCpuDifficultyEasy),
		int(GoStopCpuDifficultyHard),
	); err != nil {
		return err
	}
	return ValidateRange("target score", c.TargetScore, GoStopTargetScoreMin, GoStopTargetScoreMax)
}

// gostopConfigJSON is the JSON wire format for GoStopConfig.
type gostopConfigJSON struct {
	CpuDifficulty GoStopCpuDifficulty `json:"cd"`
	TargetScore   int                 `json:"ts"`
}

// MarshalJSON implements json.Marshaler.
func (c GoStopConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(gostopConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *GoStopConfig) UnmarshalJSON(data []byte) error {
	var j gostopConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	*c = GoStopConfig(j)
	return nil
}
