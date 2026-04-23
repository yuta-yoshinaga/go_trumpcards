package domain

import "encoding/json"

// CassinoCpuDifficulty CPU難易度レベル
type CassinoCpuDifficulty int

// CassinoCpuDifficulty定数
const (
	// CassinoDifficultyEasy 簡単 (ランダムな合法手)
	CassinoDifficultyEasy CassinoCpuDifficulty = 0
	// CassinoDifficultyNormal 通常 (貪欲: 得点スコアが最大の手)
	CassinoDifficultyNormal CassinoCpuDifficulty = 1
	// CassinoDifficultyHard 難しい (相手への得点漏れを回避)
	CassinoDifficultyHard CassinoCpuDifficulty = 2
)

// CassinoDifficultyNames 難易度名マップ
var CassinoDifficultyNames = map[CassinoCpuDifficulty]string{
	CassinoDifficultyEasy:   "Easy",
	CassinoDifficultyNormal: "Normal",
	CassinoDifficultyHard:   "Hard",
}

// Cassino スコア定数
const (
	// CassinoDefaultTargetScore デフォルトの目標点 (21点)
	CassinoDefaultTargetScore = 21
	// CassinoScoreMostCards 最多カード獲得ボーナス
	CassinoScoreMostCards = 3
	// CassinoScoreMostSpades 最多スペード獲得ボーナス
	CassinoScoreMostSpades = 1
	// CassinoScoreBigCasino 10♦ (ビッグカシノ) ボーナス
	CassinoScoreBigCasino = 2
	// CassinoScoreLittleCasino 2♠ (リトルカシノ) ボーナス
	CassinoScoreLittleCasino = 1
	// CassinoScoreAce A (エース) 1枚あたりボーナス
	CassinoScoreAce = 1
	// CassinoScoreSweep スイープ (場全取り) 1回あたりボーナス
	CassinoScoreSweep = 1
)

// CassinoConfig カシノのローカルルール設定
type CassinoConfig struct {
	// TargetScore ゲーム終了のための目標点 (デフォルト 21)
	TargetScore int
	// MultiBuildEnabled 複合ビルド (ペアビルド) を許可するか
	MultiBuildEnabled bool
	// SweepBonusEnabled スイープ (場全取り) にボーナスを与えるか
	SweepBonusEnabled bool
	// CpuDifficulty CPU難易度
	CpuDifficulty CassinoCpuDifficulty
}

// DefaultCassinoConfig デフォルトのローカルルール設定
// 問題 #1469 のユーザー回答により:
//   - プレイヤー数: 4 (1 human + 3 CPU)
//   - デッキ: 標準 52 枚
//   - 目標点: 21
//   - スイープ: 有効
//   - ビルド: 複合ビルド有効
//   - CPU 難易度: 3 段階
//   - 言語: ja/en
func DefaultCassinoConfig() CassinoConfig {
	return CassinoConfig{
		TargetScore:       CassinoDefaultTargetScore,
		MultiBuildEnabled: true,
		SweepBonusEnabled: true,
		CpuDifficulty:     CassinoDifficultyNormal,
	}
}

// Validate 設定値のドメインバリデーション
func (c CassinoConfig) Validate() error {
	if err := ValidateRange(
		"CPU difficulty",
		int(c.CpuDifficulty),
		int(CassinoDifficultyEasy),
		int(CassinoDifficultyHard),
	); err != nil {
		return err
	}
	return ValidateRange("target score", c.TargetScore, 1, 999)
}

// cassinoConfigJSON is the JSON wire format for CassinoConfig.
type cassinoConfigJSON struct {
	TargetScore       int                  `json:"ts"`
	MultiBuildEnabled bool                 `json:"mb"`
	SweepBonusEnabled bool                 `json:"sb"`
	CpuDifficulty     CassinoCpuDifficulty `json:"di"`
}

// MarshalJSON implements json.Marshaler.
func (c CassinoConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(cassinoConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *CassinoConfig) UnmarshalJSON(data []byte) error {
	var j cassinoConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	*c = CassinoConfig(j)
	return nil
}
