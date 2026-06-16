//go:build !js || !wasm || casino

package domain

// CourtPieceCpuDifficulty CPU の難易度レベル
type CourtPieceCpuDifficulty int

// CourtPiece の CPU 難易度定数
const (
	// CourtPieceCpuDifficultyEasy 低難易度
	CourtPieceCpuDifficultyEasy CourtPieceCpuDifficulty = iota
	// CourtPieceCpuDifficultyNormal 中難易度
	CourtPieceCpuDifficultyNormal
	// CourtPieceCpuDifficultyHard 高難易度
	CourtPieceCpuDifficultyHard
)

// CourtPieceDefaultPointLimit 試合終了スコア (先にこの Sar 数に到達したチームが勝利)
const CourtPieceDefaultPointLimit = 7

// CourtPieceMaxPointLimit Validate で許容する PointLimit の上限
const CourtPieceMaxPointLimit = 50

// CourtPieceConfig Court Piece ゲーム設定
type CourtPieceConfig struct {
	CpuDifficulty CourtPieceCpuDifficulty `json:"cd"`
	PointLimit    int                     `json:"pl"` // 試合終了スコア (デフォルト 7)
}

// DefaultCourtPieceConfig デフォルト設定を返す
func DefaultCourtPieceConfig() CourtPieceConfig {
	return CourtPieceConfig{
		CpuDifficulty: CourtPieceCpuDifficultyNormal,
		PointLimit:    CourtPieceDefaultPointLimit,
	}
}

// Validate 設定値のドメインバリデーション
func (c CourtPieceConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(CourtPieceCpuDifficultyEasy), int(CourtPieceCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateRange("point limit", c.PointLimit, 1, CourtPieceMaxPointLimit); err != nil {
		return err
	}
	return nil
}
