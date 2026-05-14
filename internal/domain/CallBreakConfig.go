package domain

// CallBreakCpuDifficulty CPU の難易度レベル
type CallBreakCpuDifficulty int

// CallBreakのCPU難易度定数
const (
	// CallBreakCpuDifficultyEasy 低難易度
	CallBreakCpuDifficultyEasy CallBreakCpuDifficulty = iota
	// CallBreakCpuDifficultyNormal 中難易度
	CallBreakCpuDifficultyNormal
	// CallBreakCpuDifficultyHard 高難易度
	CallBreakCpuDifficultyHard
)

// CallBreakDefaultMaxRounds 固定ラウンド数 (Call Break の伝統的なルール)
const CallBreakDefaultMaxRounds = 5

// CallBreakMaxAllowedRounds is the upper bound enforced by Validate. Tuned so
// that even at the cap the per-round ~80-entry ActionLog stays under
// callBreakMaxActionLogLen (= 5000) when restoring from JSON.
const CallBreakMaxAllowedRounds = 50

// CallBreakMinBid Call Break の最小ビッド (Nil ビッドは無い)
const CallBreakMinBid = 1

// CallBreakConfig Call Break ゲーム設定
//
// Score は内部的に「×10 された整数」として保持する。
// 例: bid=4 で 5 トリック獲得 → roundScore = 41 (表示は 4.1)、
//
//	bid=4 で 3 トリック獲得 → roundScore = -40 (表示は -4.0)。
//
// これにより浮動小数点誤差なくテストでき、JSON シリアライズも整数のままになる。
type CallBreakConfig struct {
	CpuDifficulty CallBreakCpuDifficulty `json:"cd"`
	MaxRounds     int                    `json:"mr"` // ゲーム終了時のラウンド数 (デフォルト 5)
}

// DefaultCallBreakConfig デフォルト設定を返す
func DefaultCallBreakConfig() CallBreakConfig {
	return CallBreakConfig{
		CpuDifficulty: CallBreakCpuDifficultyNormal,
		MaxRounds:     CallBreakDefaultMaxRounds,
	}
}

// Validate 設定値のドメインバリデーション
func (c CallBreakConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(CallBreakCpuDifficultyEasy), int(CallBreakCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateRange("max rounds", c.MaxRounds, 1, CallBreakMaxAllowedRounds); err != nil {
		return err
	}
	return nil
}
