//go:build !js || !wasm || solo

package domain

// BidWhistCpuDifficulty CPU の難易度レベル
type BidWhistCpuDifficulty int

// Bid Whist の CPU 難易度定数
const (
	// BidWhistCpuDifficultyEasy 低難易度
	BidWhistCpuDifficultyEasy BidWhistCpuDifficulty = iota
	// BidWhistCpuDifficultyNormal 中難易度
	BidWhistCpuDifficultyNormal
	// BidWhistCpuDifficultyHard 高難易度
	BidWhistCpuDifficultyHard
)

// BidWhistDefaultTargetScore ゲーム終了スコア (先に到達したチームが勝利)。
// Bid Whist の標準である「先取7点」を採用する。
const BidWhistDefaultTargetScore = 7

// BidWhistConfig Bid Whist ゲーム設定
type BidWhistConfig struct {
	CpuDifficulty BidWhistCpuDifficulty `json:"cd"`
	TargetScore   int                   `json:"ts"` // 勝利スコア (デフォルト7)
}

// DefaultBidWhistConfig デフォルト設定を返す
func DefaultBidWhistConfig() BidWhistConfig {
	return BidWhistConfig{
		CpuDifficulty: BidWhistCpuDifficultyNormal,
		TargetScore:   BidWhistDefaultTargetScore,
	}
}

// Validate 設定値のドメインバリデーション
func (c BidWhistConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(BidWhistCpuDifficultyEasy), int(BidWhistCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("target score", c.TargetScore, 1); err != nil {
		return err
	}
	return nil
}
