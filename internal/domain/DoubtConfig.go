package domain

import "encoding/json"

// DoubtMemoryLevel CPU の記憶力レベル
type DoubtMemoryLevel int

// Doubtの記憶レベル定数
const (
	// DoubtMemoryLevelEasy 低記憶力 (約30%の確率で記憶)
	DoubtMemoryLevelEasy DoubtMemoryLevel = iota
	// DoubtMemoryLevelNormal 中記憶力 (約70%の確率で記憶)
	DoubtMemoryLevelNormal
	// DoubtMemoryLevelHard 高記憶力 (100%記憶)
	DoubtMemoryLevelHard
)

// DoubtConfig ダウトゲーム設定
type DoubtConfig struct {
	DoubtWindowSec       int
	CpuMemoryLevel       DoubtMemoryLevel
	PenaltyDrawLimit     int  // 0 = unlimited; >0 = loser draws at most N cards
	CpuHesitationEnabled bool // CPU迷い時間ディレイ
	CpuMetaAI            bool // メタAI: セッション内学習
}

// DefaultDoubtConfig デフォルト設定を返す
func DefaultDoubtConfig() DoubtConfig {
	return DoubtConfig{DoubtWindowSec: 10, CpuMemoryLevel: DoubtMemoryLevelNormal}
}

// doubtConfigJSON is the JSON wire format for DoubtConfig.
type doubtConfigJSON struct {
	DoubtWindowSec       int              `json:"dw"`
	CpuMemoryLevel       DoubtMemoryLevel `json:"ml"`
	PenaltyDrawLimit     int              `json:"pd"`
	CpuHesitationEnabled bool             `json:"ch"`
	CpuMetaAI            bool             `json:"ca"`
}

// MarshalJSON implements json.Marshaler.
func (c DoubtConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(doubtConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *DoubtConfig) UnmarshalJSON(data []byte) error {
	var j doubtConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	c.DoubtWindowSec = j.DoubtWindowSec
	c.CpuMemoryLevel = j.CpuMemoryLevel
	c.PenaltyDrawLimit = j.PenaltyDrawLimit
	c.CpuHesitationEnabled = j.CpuHesitationEnabled
	c.CpuMetaAI = j.CpuMetaAI
	return nil
}
