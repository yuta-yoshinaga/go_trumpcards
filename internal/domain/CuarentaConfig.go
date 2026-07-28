//go:build !js || !wasm || extra2

package domain

import "encoding/json"

// CuarentaCpuDifficulty CPU難易度レベル。
type CuarentaCpuDifficulty int

// CuarentaCpuDifficulty 定数。
const (
	// CuarentaDifficultyEasy 簡単 (ランダムな合法手)。
	CuarentaDifficultyEasy CuarentaCpuDifficulty = 0
	// CuarentaDifficultyNormal 通常 (捕獲を優先する貪欲法)。
	CuarentaDifficultyNormal CuarentaCpuDifficulty = 1
	// CuarentaDifficultyHard 難しい (caída/limpia を狙い、相手への漏れを避ける)。
	CuarentaDifficultyHard CuarentaCpuDifficulty = 2
)

// CuarentaDifficultyNames 難易度名マップ。
var CuarentaDifficultyNames = map[CuarentaCpuDifficulty]string{
	CuarentaDifficultyEasy:   "Easy",
	CuarentaDifficultyNormal: "Normal",
	CuarentaDifficultyHard:   "Hard",
}

// クアレンタの得点定数 (いずれも加点)。
const (
	// CuarentaDefaultTargetScore デフォルトの目標点 (40点)。
	CuarentaDefaultTargetScore = 40
	// CuarentaScoreCaida カイーダ (相手の置いたカードを即捕獲) ボーナス。
	CuarentaScoreCaida = 2
	// CuarentaScoreLimpia リンピア (場を全て掃く) ボーナス。
	CuarentaScoreLimpia = 1
	// CuarentaScoreRondaPerExtra ロンダ (同ランク 3 枚以上の連続捕獲) で
	// 3 枚目以降 1 枚あたりに加点する点数。
	CuarentaScoreRondaPerExtra = 1
	// CuarentaMostCardsThreshold このカード枚数を超えて捕獲したチーム
	// (40 枚の過半) に最多取りボーナスを与える。
	CuarentaMostCardsThreshold = 20
	// CuarentaScoreMostCards 最多取りチームへのラウンド終了ボーナス。
	CuarentaScoreMostCards = 6
)

// CuarentaConfig クアレンタのローカルルール設定。
type CuarentaConfig struct {
	// TargetScore ゲーム終了のための目標点 (デフォルト 40)。
	TargetScore int
	// CpuDifficulty CPU 難易度。
	CpuDifficulty CuarentaCpuDifficulty
}

// DefaultCuarentaConfig デフォルトのローカルルール設定。
//
// 問題 #2330 のルール:
//   - プレイヤー数: 4 (1 human + 3 CPU)、2 チーム ({0,2} vs {1,3})
//   - デッキ: 40 枚 (8,9,10 を除外)
//   - 目標点: 40
//   - caída +2 / limpia +1 / ronda 連続加点 / 最多取り +6
//   - CPU 難易度: 3 段階
func DefaultCuarentaConfig() CuarentaConfig {
	return CuarentaConfig{
		TargetScore:   CuarentaDefaultTargetScore,
		CpuDifficulty: CuarentaDifficultyNormal,
	}
}

// Validate 設定値のドメインバリデーション。
func (c CuarentaConfig) Validate() error {
	if err := ValidateRange(
		"CPU difficulty",
		int(c.CpuDifficulty),
		int(CuarentaDifficultyEasy),
		int(CuarentaDifficultyHard),
	); err != nil {
		return err
	}
	return ValidateRange("target score", c.TargetScore, 1, 999)
}

// cuarentaConfigJSON is the JSON wire format for CuarentaConfig.
type cuarentaConfigJSON struct {
	TargetScore   int                   `json:"ts"`
	CpuDifficulty CuarentaCpuDifficulty `json:"di"`
}

// MarshalJSON implements json.Marshaler.
func (c CuarentaConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(cuarentaConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *CuarentaConfig) UnmarshalJSON(data []byte) error {
	var j cuarentaConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	*c = CuarentaConfig(j)
	return nil
}
