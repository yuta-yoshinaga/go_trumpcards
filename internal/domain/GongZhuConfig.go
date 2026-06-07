package domain

// GongZhuCpuDifficulty CPU の難易度レベル
type GongZhuCpuDifficulty int

// Gong ZhuのCPU難易度定数
const (
	// GongZhuCpuDifficultyEasy 低難易度
	GongZhuCpuDifficultyEasy GongZhuCpuDifficulty = iota
	// GongZhuCpuDifficultyNormal 中難易度
	GongZhuCpuDifficultyNormal
	// GongZhuCpuDifficultyHard 高難易度
	GongZhuCpuDifficultyHard
)

// GongZhuConfig 拱猪（Gong Zhu）ゲーム設定
type GongZhuConfig struct {
	CpuDifficulty GongZhuCpuDifficulty `json:"cd"`
	// PointLimit ゲーム終了しきい値。いずれかのプレイヤーの累積スコアが -PointLimit 以下に
	// 達したらゲーム終了し、最高スコアのプレイヤーが勝者となる。
	PointLimit int `json:"pl"`
}

// DefaultGongZhuConfig デフォルト設定を返す
func DefaultGongZhuConfig() GongZhuConfig {
	return GongZhuConfig{CpuDifficulty: GongZhuCpuDifficultyNormal, PointLimit: 1000}
}

// Validate 設定値のドメインバリデーション
func (c GongZhuConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(GongZhuCpuDifficultyEasy), int(GongZhuCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("point limit", c.PointLimit, 1); err != nil {
		return err
	}
	return nil
}
