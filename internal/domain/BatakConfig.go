package domain

// BatakCpuDifficulty CPU の難易度レベル
type BatakCpuDifficulty int

// BatakのCPU難易度定数
const (
	// BatakCpuDifficultyEasy 低難易度
	BatakCpuDifficultyEasy BatakCpuDifficulty = iota
	// BatakCpuDifficultyNormal 中難易度
	BatakCpuDifficultyNormal
	// BatakCpuDifficultyHard 高難易度
	BatakCpuDifficultyHard
)

// BatakDefaultMaxRounds 固定ラウンド数 (Batak の伝統的なルール)
const BatakDefaultMaxRounds = 5

// BatakMaxAllowedRounds is the upper bound enforced by Validate. Tuned so
// that even at the cap the per-round ~80-entry ActionLog stays under
// batakMaxActionLogLen (= 5000) when restoring from JSON.
const BatakMaxAllowedRounds = 50

// BatakMinBid Batak の最小ビッド (Nil ビッドは無く、親の最低宣言値は 5)
const BatakMinBid = 5

// BatakMaxBid Batak の最大ビッド (13 トリック)
const BatakMaxBid = BatakHandSize

// BatakPassBid Batak のパスビッド (0)
const BatakPassBid = 0

// BatakConfig Batak ゲーム設定
//
// スコアは素の整数として保持する。
// 親 (declarer): tricks >= bid なら +bid、下回れば -bid
// 子 (non-declarer): +tricks (獲得トリック数がそのまま加点)
type BatakConfig struct {
	CpuDifficulty BatakCpuDifficulty `json:"cd"`
	MaxRounds     int                `json:"mr"` // ゲーム終了時のラウンド数 (デフォルト 5)
}

// DefaultBatakConfig デフォルト設定を返す
func DefaultBatakConfig() BatakConfig {
	return BatakConfig{
		CpuDifficulty: BatakCpuDifficultyNormal,
		MaxRounds:     BatakDefaultMaxRounds,
	}
}

// Validate 設定値のドメインバリデーション
func (c BatakConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(BatakCpuDifficultyEasy), int(BatakCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateRange("max rounds", c.MaxRounds, 1, BatakMaxAllowedRounds); err != nil {
		return err
	}
	return nil
}
