//go:build !js || !wasm || classic

package domain

import "encoding/json"

// ScopaCpuDifficulty CPU難易度レベル
type ScopaCpuDifficulty int

// ScopaCpuDifficulty定数
const (
	// ScopaDifficultyEasy 簡単 (ランダムな合法手)
	ScopaDifficultyEasy ScopaCpuDifficulty = 0
	// ScopaDifficultyNormal 通常 (貪欲: 得点価値が最大の手)
	ScopaDifficultyNormal ScopaCpuDifficulty = 1
	// ScopaDifficultyHard 難しい (相手への得点漏れを回避)
	ScopaDifficultyHard ScopaCpuDifficulty = 2
)

// ScopaDifficultyNames 難易度名マップ
var ScopaDifficultyNames = map[ScopaCpuDifficulty]string{
	ScopaDifficultyEasy:   "Easy",
	ScopaDifficultyNormal: "Normal",
	ScopaDifficultyHard:   "Hard",
}

// Scopa スコア定数 (いずれもラウンドあたりの加点)
const (
	// ScopaDefaultTargetScore デフォルトの目標点 (11点)
	ScopaDefaultTargetScore = 11

	// ScopaMinTargetScore 目標点の下限
	ScopaMinTargetScore = 1
	// ScopaMaxTargetScore 目標点の上限
	ScopaMaxTargetScore = 999
	// ScopaScoreMostCards 最多カード獲得 (carte) ボーナス
	ScopaScoreMostCards = 1
	// ScopaScoreMostDiamonds 最多ダイヤ獲得 (denari) ボーナス
	ScopaScoreMostDiamonds = 1
	// ScopaScoreSetteBello 7♦ (セッテ・ベッロ) ボーナス
	ScopaScoreSetteBello = 1
	// ScopaScoreMostSevens 最多 7 獲得 (簡易プリミエラ) ボーナス
	ScopaScoreMostSevens = 1
	// ScopaScoreScopa スコパ (場全取り) 1回あたりボーナス
	ScopaScoreScopa = 1
)

// ScopaConfig スコパのローカルルール設定
type ScopaConfig struct {
	// TargetScore ゲーム終了のための目標点 (デフォルト 11)
	TargetScore int
	// CpuDifficulty CPU難易度
	CpuDifficulty ScopaCpuDifficulty
}

// DefaultScopaConfig デフォルトのローカルルール設定
// 問題 #1990 のユーザー回答により:
//   - プレイヤー数: 2 (1 human + 1 CPU, ヘッズアップ)
//   - デッキ: 40 枚 (8,9,10 を除外)
//   - 目標点: 11
//   - プリミエラ: 簡易版 (最多 7 で +1)
//   - CPU 難易度: 3 段階
//   - 言語: ja/en
func DefaultScopaConfig() ScopaConfig {
	return ScopaConfig{
		TargetScore:   ScopaDefaultTargetScore,
		CpuDifficulty: ScopaDifficultyNormal,
	}
}

// Validate 設定値のドメインバリデーション
func (c ScopaConfig) Validate() error {
	if err := ValidateRange(
		"CPU difficulty",
		int(c.CpuDifficulty),
		int(ScopaDifficultyEasy),
		int(ScopaDifficultyHard),
	); err != nil {
		return err
	}
	return ValidateRange("target score", c.TargetScore, ScopaMinTargetScore, ScopaMaxTargetScore)
}

// scopaConfigJSON is the JSON wire format for ScopaConfig.
type scopaConfigJSON struct {
	TargetScore   int                `json:"ts"`
	CpuDifficulty ScopaCpuDifficulty `json:"di"`
}

// MarshalJSON implements json.Marshaler.
func (c ScopaConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(scopaConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *ScopaConfig) UnmarshalJSON(data []byte) error {
	var j scopaConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	*c = ScopaConfig(j)
	return nil
}
