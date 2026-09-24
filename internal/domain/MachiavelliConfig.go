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

// MachiavelliConfig マキャヴェッリの設定
type MachiavelliConfig struct {
	// PlayerCount 参加プレイヤー数（人間 1 + CPU）。2〜5。
	PlayerCount int `json:"pc"`
	// TargetRounds ゲーム終了までのラウンド数（この回数を消化した時点で累計最少が勝利）
	TargetRounds int `json:"tr"`
}

// DefaultMachiavelliConfig デフォルト設定を返す
func DefaultMachiavelliConfig() MachiavelliConfig {
	return MachiavelliConfig{
		PlayerCount:  MachiavelliDefaultPlayerCount,
		TargetRounds: MachiavelliDefaultTargetRounds,
	}
}

// Validate 設定値のドメインバリデーション
func (c MachiavelliConfig) Validate() error {
	if err := ValidateRange("player count", c.PlayerCount, MachiavelliPlayerCountMin, MachiavelliPlayerCountMax); err != nil {
		return err
	}
	if err := ValidateMin("target rounds", c.TargetRounds, 1); err != nil {
		return err
	}
	return nil
}
