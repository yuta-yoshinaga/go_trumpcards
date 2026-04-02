package domain

import "encoding/json"

// GoFishCpuDifficulty CPU の難易度レベル
type GoFishCpuDifficulty int

// GoFishのCPU難易度定数
const (
	// GoFishCpuDifficultyEasy 低難易度: ランダムに要求
	GoFishCpuDifficultyEasy GoFishCpuDifficulty = iota
	// GoFishCpuDifficultyNormal 中難易度: 手札のランクのみ要求、直近の記憶に依存
	GoFishCpuDifficultyNormal
	// GoFishCpuDifficultyHard 高難易度: 全要求履歴から最適な相手とランクを推測
	GoFishCpuDifficultyHard
)

// GoFishConfig Go Fishゲーム設定
type GoFishConfig struct {
	CpuDifficulty GoFishCpuDifficulty
	CpuMetaAI     bool // メタAI: セッション内学習
}

// DefaultGoFishConfig デフォルト設定を返す
func DefaultGoFishConfig() GoFishConfig {
	return GoFishConfig{CpuDifficulty: GoFishCpuDifficultyNormal}
}

// Validate 設定値のドメインバリデーション
func (c GoFishConfig) Validate() error {
	return ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(GoFishCpuDifficultyEasy), int(GoFishCpuDifficultyHard))
}

// goFishConfigJSON is the JSON wire format for GoFishConfig.
type goFishConfigJSON struct {
	CpuDifficulty GoFishCpuDifficulty `json:"cd"`
	CpuMetaAI     bool                `json:"ma"`
}

// MarshalJSON implements json.Marshaler.
func (c GoFishConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(goFishConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *GoFishConfig) UnmarshalJSON(data []byte) error {
	var j goFishConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	c.CpuDifficulty = j.CpuDifficulty
	c.CpuMetaAI = j.CpuMetaAI
	return nil
}
