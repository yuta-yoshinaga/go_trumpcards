//go:build !js || !wasm || casino

package domain

// TarneebCpuDifficulty CPU の難易度レベル
type TarneebCpuDifficulty int

// TarneebのCPU難易度定数
const (
	// TarneebCpuDifficultyEasy 低難易度
	TarneebCpuDifficultyEasy TarneebCpuDifficulty = iota
	// TarneebCpuDifficultyNormal 中難易度
	TarneebCpuDifficultyNormal
	// TarneebCpuDifficultyHard 高難易度
	TarneebCpuDifficultyHard
)

// TarneebDefaultPointLimit ゲーム終了スコア (先に到達したチームが勝利)
const TarneebDefaultPointLimit = 31

// TarneebMaxPointLimit Validate で許容する PointLimit の上限
const TarneebMaxPointLimit = 200

// TarneebDefaultMinBid 最小ビッド (7 = 過半数 13/2 切り上げ)
const TarneebDefaultMinBid = 7

// TarneebPassBid パスを表す内部値
const TarneebPassBid = 0

// TarneebConfig Tarneeb ゲーム設定
type TarneebConfig struct {
	CpuDifficulty TarneebCpuDifficulty `json:"cd"`
	PointLimit    int                  `json:"pl"` // ゲーム終了スコア (デフォルト 31)
	MinBid        int                  `json:"mb"` // 最小ビッド (デフォルト 7)
}

// DefaultTarneebConfig デフォルト設定を返す
func DefaultTarneebConfig() TarneebConfig {
	return TarneebConfig{
		CpuDifficulty: TarneebCpuDifficultyNormal,
		PointLimit:    TarneebDefaultPointLimit,
		MinBid:        TarneebDefaultMinBid,
	}
}

// Validate 設定値のドメインバリデーション
func (c TarneebConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(TarneebCpuDifficultyEasy), int(TarneebCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateRange("point limit", c.PointLimit, 1, TarneebMaxPointLimit); err != nil {
		return err
	}
	if err := ValidateRange("min bid", c.MinBid, 1, 13); err != nil {
		return err
	}
	return nil
}
