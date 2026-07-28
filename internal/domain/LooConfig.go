//go:build !js || !wasm || extra3

package domain

import "encoding/json"

// LooCpuDifficulty は CPU の難易度レベル。
type LooCpuDifficulty int

// Loo の CPU 難易度定数
const (
	// LooCpuDifficultyEasy 低難易度 (ランダムな play/pass とプレイ)
	LooCpuDifficultyEasy LooCpuDifficulty = iota
	// LooCpuDifficultyNormal 中難易度 (手札強度による判断)
	LooCpuDifficultyNormal
	// LooCpuDifficultyHard 高難易度 (より厳しい参加基準と戦略プレイ)
	LooCpuDifficultyHard
)

// LooDifficultyNames 難易度名マップ
var LooDifficultyNames = map[LooCpuDifficulty]string{
	LooCpuDifficultyEasy:   "Easy",
	LooCpuDifficultyNormal: "Normal",
	LooCpuDifficultyHard:   "Hard",
}

// LooConfig はルー (Loo / Lanterloo) のローカルルール設定。
type LooConfig struct {
	// CpuDifficulty CPU 難易度
	CpuDifficulty LooCpuDifficulty `json:"cd"`
	// Ante 各ディールで各プレイヤーがポットへ支払うチップ数 (既定 3)
	Ante int `json:"an"`
}

// DefaultLooConfig はデフォルトのローカルルール設定を返す。
//   - プレイヤー数: 4 (1 human + 3 CPU, 個人戦)
//   - デッキ: 52 枚
//   - 各プレイヤー 5 枚配り (Five-card Loo)、5 トリック
//   - CPU 難易度: 3 段階
//   - アンティ: 3 (1 トリック = ポット/5 を獲得)
func DefaultLooConfig() LooConfig {
	return LooConfig{
		CpuDifficulty: LooCpuDifficultyNormal,
		Ante:          3,
	}
}

// Validate は設定値のドメインバリデーションを行う。
func (c LooConfig) Validate() error {
	if err := ValidateRange(
		"CPU difficulty",
		int(c.CpuDifficulty),
		int(LooCpuDifficultyEasy),
		int(LooCpuDifficultyHard),
	); err != nil {
		return err
	}
	return ValidateMin("ante", c.Ante, 1)
}

// looConfigJSON is the JSON wire format for LooConfig.
type looConfigJSON struct {
	CpuDifficulty LooCpuDifficulty `json:"cd"`
	Ante          int              `json:"an"`
}

// MarshalJSON implements json.Marshaler.
func (c LooConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(looConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *LooConfig) UnmarshalJSON(data []byte) error {
	var j looConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	*c = LooConfig(j)
	return nil
}
