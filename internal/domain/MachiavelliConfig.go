//go:build !js || !wasm || extra

package domain

// Machiavelli プレイヤー数の下限・上限・既定値
const (
	// MachiavelliPlayerCountMin プレイヤー数の下限
	MachiavelliPlayerCountMin = 2
	// MachiavelliPlayerCountMax プレイヤー数の上限
	MachiavelliPlayerCountMax = 5
	// MachiavelliDefaultPlayerCount 既定プレイヤー数（人間 1 + CPU 3）
	MachiavelliDefaultPlayerCount = 4
)

// MachiavelliCpuDifficulty CPU の難易度レベル
type MachiavelliCpuDifficulty int

// Machiavelli の CPU 難易度定数
const (
	// MachiavelliCpuDifficultyEasy 低難易度
	MachiavelliCpuDifficultyEasy MachiavelliCpuDifficulty = iota
	// MachiavelliCpuDifficultyNormal 中難易度
	MachiavelliCpuDifficultyNormal
	// MachiavelliCpuDifficultyHard 高難易度
	MachiavelliCpuDifficultyHard
)

// MachiavelliConfig マキャヴェッリの設定
type MachiavelliConfig struct {
	// PlayerCount 参加プレイヤー数（人間 1 + CPU）。2〜5。
	PlayerCount int `json:"pc"`
	// CpuDifficulty CPU 難易度
	CpuDifficulty MachiavelliCpuDifficulty `json:"cd"`
	// TargetRounds ゲーム終了までのラウンド数（この回数を消化した時点で累計最少が勝利）
	TargetRounds int `json:"tr"`
}

// DefaultMachiavelliConfig デフォルト設定を返す
func DefaultMachiavelliConfig() MachiavelliConfig {
	return MachiavelliConfig{
		PlayerCount:   MachiavelliDefaultPlayerCount,
		CpuDifficulty: MachiavelliCpuDifficultyNormal,
		TargetRounds:  MachiavelliDefaultTargetRounds,
	}
}

// Validate 設定値のドメインバリデーション
func (c MachiavelliConfig) Validate() error {
	if err := ValidateRange("player count", c.PlayerCount, MachiavelliPlayerCountMin, MachiavelliPlayerCountMax); err != nil {
		return err
	}
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(MachiavelliCpuDifficultyEasy), int(MachiavelliCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("target rounds", c.TargetRounds, 1); err != nil {
		return err
	}
	return nil
}
