//go:build !js || !wasm || classic

package domain

// EscobaCpuDifficulty CPU の難易度レベル
type EscobaCpuDifficulty int

// Escoba の CPU 難易度定数
const (
	// EscobaCpuDifficultyEasy 低難易度
	EscobaCpuDifficultyEasy EscobaCpuDifficulty = iota
	// EscobaCpuDifficultyNormal 中難易度
	EscobaCpuDifficultyNormal
	// EscobaCpuDifficultyHard 高難易度
	EscobaCpuDifficultyHard
)

// EscobaDefaultTargetScore 試合終了スコア (先に到達したプレイヤーが勝利)
const EscobaDefaultTargetScore = 10

// EscobaMaxTargetScore Validate で許容する TargetScore の上限
const EscobaMaxTargetScore = 100

// EscobaConfig Escoba ゲーム設定
type EscobaConfig struct {
	CpuDifficulty EscobaCpuDifficulty `json:"cd"`
	TargetScore   int                 `json:"ts"` // 試合終了スコア (デフォルト 10)
}

// DefaultEscobaConfig デフォルト設定を返す
func DefaultEscobaConfig() EscobaConfig {
	return EscobaConfig{
		CpuDifficulty: EscobaCpuDifficultyNormal,
		TargetScore:   EscobaDefaultTargetScore,
	}
}

// Validate 設定値のドメインバリデーション
func (c EscobaConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(EscobaCpuDifficultyEasy), int(EscobaCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateRange("target score", c.TargetScore, 1, EscobaMaxTargetScore); err != nil {
		return err
	}
	return nil
}
