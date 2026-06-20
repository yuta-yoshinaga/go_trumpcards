//go:build !js || !wasm || casino

package domain

// OpenFaceChineseCpuDifficulty CPU の難易度レベル
type OpenFaceChineseCpuDifficulty int

// OFC の CPU 難易度定数
const (
	// OpenFaceChineseCpuDifficultyEasy 低難易度（ランダム配置）
	OpenFaceChineseCpuDifficultyEasy OpenFaceChineseCpuDifficulty = iota
	// OpenFaceChineseCpuDifficultyNormal 中難易度（ヒューリスティック配置）
	OpenFaceChineseCpuDifficultyNormal
	// OpenFaceChineseCpuDifficultyHard 高難易度（ヒューリスティック配置）
	OpenFaceChineseCpuDifficultyHard
)

// OpenFaceChinesePlayerMin 最小プレイヤー数
const OpenFaceChinesePlayerMin = 2

// OpenFaceChinesePlayerMax 最大プレイヤー数
const OpenFaceChinesePlayerMax = 4

// OpenFaceChineseConfig オープンフェイス・チャイニーズポーカーのゲーム設定
type OpenFaceChineseConfig struct {
	// CpuDifficulty CPU の難易度。
	CpuDifficulty OpenFaceChineseCpuDifficulty `json:"cd"`
	// PlayerCount プレイヤー数（人間 1 + CPU、2〜4）。
	PlayerCount int `json:"pn"`
	// TargetRounds マッチを構成するラウンド数。これだけプレイしたら最高得点者が勝者。
	TargetRounds int `json:"tr"`
}

// DefaultOpenFaceChineseConfig デフォルト設定を返す（人間 1 + CPU 1 のヘッズアップ、4 ラウンドマッチ）。
func DefaultOpenFaceChineseConfig() OpenFaceChineseConfig {
	return OpenFaceChineseConfig{
		CpuDifficulty: OpenFaceChineseCpuDifficultyNormal,
		PlayerCount:   2,
		TargetRounds:  4,
	}
}

// Validate 設定値のドメインバリデーション
func (c OpenFaceChineseConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(OpenFaceChineseCpuDifficultyEasy), int(OpenFaceChineseCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateRange("player count", c.PlayerCount, OpenFaceChinesePlayerMin, OpenFaceChinesePlayerMax); err != nil {
		return err
	}
	if err := ValidateMin("target rounds", c.TargetRounds, 1); err != nil {
		return err
	}
	return nil
}
