package domain

import "encoding/json"

// ShitheadCpuDifficulty CPU難易度レベル
type ShitheadCpuDifficulty int

// ShitheadCpuDifficulty定数
const (
	// ShitheadDifficultyEasy 簡単 (常に最弱を出す)
	ShitheadDifficultyEasy ShitheadCpuDifficulty = 0
	// ShitheadDifficultyNormal 通常 (デフォルト)
	ShitheadDifficultyNormal ShitheadCpuDifficulty = 1
	// ShitheadDifficultyHard 難しい (マジックカードを温存)
	ShitheadDifficultyHard ShitheadCpuDifficulty = 2
)

// ShitheadDifficultyNames 難易度名マップ
var ShitheadDifficultyNames = map[ShitheadCpuDifficulty]string{
	ShitheadDifficultyEasy:   "Easy",
	ShitheadDifficultyNormal: "Normal",
	ShitheadDifficultyHard:   "Hard",
}

// ShitheadConfig シットヘッドのローカルルール設定
type ShitheadConfig struct {
	// MagicTwo 2をリセットカードとして扱う (次は何でも出せる)
	MagicTwo bool
	// MagicSeven 7を「次は7以下しか出せない」カードとして扱う
	MagicSeven bool
	// MagicEight 8を次プレイヤースキップカードとして扱う
	MagicEight bool
	// MagicTen 10を場札焼却カードとして扱う (場札をゲームから除外し、同じ人が続ける)
	MagicTen bool
	// FourOfAKindBurn 同じ値が4枚揃ったら場札を焼却する
	FourOfAKindBurn bool
	// CpuDifficulty CPU難易度
	CpuDifficulty ShitheadCpuDifficulty
}

// DefaultShitheadConfig デフォルトのローカルルール設定
// 標準的なShithead/Karmaのルール: 2/10/4枚焼却を有効、7/8 (地域差) を有効
func DefaultShitheadConfig() ShitheadConfig {
	return ShitheadConfig{
		MagicTwo:        true,
		MagicSeven:      true,
		MagicEight:      true,
		MagicTen:        true,
		FourOfAKindBurn: true,
		CpuDifficulty:   ShitheadDifficultyNormal,
	}
}

// Validate 設定値のドメインバリデーション
func (c ShitheadConfig) Validate() error {
	return ValidateRange(
		"CPU difficulty",
		int(c.CpuDifficulty),
		int(ShitheadDifficultyEasy),
		int(ShitheadDifficultyHard),
	)
}

// shitheadConfigJSON is the JSON wire format for ShitheadConfig.
type shitheadConfigJSON struct {
	MagicTwo        bool                  `json:"m2"`
	MagicSeven      bool                  `json:"m7"`
	MagicEight      bool                  `json:"m8"`
	MagicTen        bool                  `json:"mT"`
	FourOfAKindBurn bool                  `json:"f4"`
	CpuDifficulty   ShitheadCpuDifficulty `json:"di"`
}

// MarshalJSON implements json.Marshaler.
func (c ShitheadConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(shitheadConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *ShitheadConfig) UnmarshalJSON(data []byte) error {
	var j shitheadConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	*c = ShitheadConfig(j)
	return nil
}
