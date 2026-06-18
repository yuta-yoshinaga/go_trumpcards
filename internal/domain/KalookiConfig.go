//go:build !js || !wasm || solo

package domain

// KalookiCpuDifficulty Kalooki の CPU 難易度レベル
type KalookiCpuDifficulty int

// Kalooki の CPU 難易度定数
const (
	// KalookiCpuDifficultyEasy 低難易度
	KalookiCpuDifficultyEasy KalookiCpuDifficulty = iota
	// KalookiCpuDifficultyNormal 中難易度
	KalookiCpuDifficultyNormal
	// KalookiCpuDifficultyHard 高難易度
	KalookiCpuDifficultyHard
)

// Kalooki のプレイヤー数の下限・上限
const (
	// KalookiMinPlayers 最小プレイヤー数
	KalookiMinPlayers = 2
	// KalookiMaxPlayers 最大プレイヤー数
	KalookiMaxPlayers = 4
	// KalookiDefaultPlayers 既定プレイヤー数（人間 1 + CPU 3）
	KalookiDefaultPlayers = 4
)

// KalookiConfig Kalooki（カルーキ）の設定
type KalookiConfig struct {
	// CpuDifficulty CPU の難易度
	CpuDifficulty KalookiCpuDifficulty `json:"cd"`
	// PlayerCount このゲームのプレイヤー数（2〜4）
	PlayerCount int `json:"pc"`
	// OpeningThreshold 最初のメルド提出時に満たすべき合計点（オープニング要件）
	OpeningThreshold int `json:"ot"`
}

// DefaultKalookiConfig デフォルト設定を返す。
// 4 人（人間 1 + CPU 3）、CPU 難易度 Normal、オープニング要件 51 点。
func DefaultKalookiConfig() KalookiConfig {
	return KalookiConfig{
		CpuDifficulty:    KalookiCpuDifficultyNormal,
		PlayerCount:      KalookiDefaultPlayers,
		OpeningThreshold: 51,
	}
}

// Validate 設定値のドメインバリデーション
func (c KalookiConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(KalookiCpuDifficultyEasy), int(KalookiCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateRange("player count", c.PlayerCount, KalookiMinPlayers, KalookiMaxPlayers); err != nil {
		return err
	}
	if err := ValidateMin("opening threshold", c.OpeningThreshold, 0); err != nil {
		return err
	}
	return nil
}
