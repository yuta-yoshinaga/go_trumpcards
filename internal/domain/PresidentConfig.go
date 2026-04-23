package domain

import "encoding/json"

// PresidentCpuDifficulty CPU難易度レベル
type PresidentCpuDifficulty int

// PresidentCpuDifficulty定数
const (
	// PresidentDifficultyEasy 簡単 (常に最弱を出す)
	PresidentDifficultyEasy PresidentCpuDifficulty = 0
	// PresidentDifficultyNormal 通常 (デフォルト)
	PresidentDifficultyNormal PresidentCpuDifficulty = 1
	// PresidentDifficultyHard 難しい (ヒューリスティック)
	PresidentDifficultyHard PresidentCpuDifficulty = 2
)

// PresidentDifficultyNames 難易度名マップ
var PresidentDifficultyNames = map[PresidentCpuDifficulty]string{
	PresidentDifficultyEasy:   "Easy",
	PresidentDifficultyNormal: "Normal",
	PresidentDifficultyHard:   "Hard",
}

// PresidentConfig プレジデントのローカルルール設定
type PresidentConfig struct {
	// RevolutionEnabled 革命: 4枚出しで強さ順が反転
	RevolutionEnabled bool
	// CardExchangeEnabled カード交換: 前ラウンドのランクに基づくカード交換
	CardExchangeEnabled bool
	// PassFieldFlushEnabled パス即場流れ: プレイヤーがパスすると場がクリアされる
	// false の場合は大富豪式 (全員パスで場流れ)
	PassFieldFlushEnabled bool
	// CpuDifficulty CPU難易度
	CpuDifficulty PresidentCpuDifficulty
}

// DefaultPresidentConfig デフォルトのローカルルール設定
// 問題 #1468 のユーザー回答により:
//   - 革命: 有効
//   - カード交換: 有効 (2 cards President↔Scum, 1 card Vice President↔Vice Scum)
//   - パス即場流れ: 有効 (default on)
func DefaultPresidentConfig() PresidentConfig {
	return PresidentConfig{
		RevolutionEnabled:     true,
		CardExchangeEnabled:   true,
		PassFieldFlushEnabled: true,
		CpuDifficulty:         PresidentDifficultyNormal,
	}
}

// Validate 設定値のドメインバリデーション
func (c PresidentConfig) Validate() error {
	return ValidateRange(
		"CPU difficulty",
		int(c.CpuDifficulty),
		int(PresidentDifficultyEasy),
		int(PresidentDifficultyHard),
	)
}

// presidentConfigJSON is the JSON wire format for PresidentConfig.
type presidentConfigJSON struct {
	RevolutionEnabled     bool                   `json:"re"`
	CardExchangeEnabled   bool                   `json:"ce"`
	PassFieldFlushEnabled bool                   `json:"pf"`
	CpuDifficulty         PresidentCpuDifficulty `json:"di"`
}

// MarshalJSON implements json.Marshaler.
func (c PresidentConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(presidentConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *PresidentConfig) UnmarshalJSON(data []byte) error {
	var j presidentConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	*c = PresidentConfig(j)
	return nil
}
