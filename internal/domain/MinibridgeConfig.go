//go:build !js || !wasm || extra3

package domain

import "fmt"

const (
	// MinibridgeRoundsMin は最小ラウンド数。
	MinibridgeRoundsMin = MinibridgePlayerCnt
	// MinibridgeRoundsMax は最大ラウンド数。
	MinibridgeRoundsMax = 20
	// MinibridgeDefaultRounds は既定のラウンド数（全員が 1 回ずつ親）。
	MinibridgeDefaultRounds = MinibridgePlayerCnt
)

// MinibridgeConfig はミニブリッジの設定。
type MinibridgeConfig struct {
	// Rounds は打つディール数。
	Rounds int `json:"r"`
}

// DefaultMinibridgeConfig は既定設定を返す。
func DefaultMinibridgeConfig() MinibridgeConfig {
	return MinibridgeConfig{Rounds: MinibridgeDefaultRounds}
}

// Validate は設定を検証する。
//
// **4 の倍数でなければ不公平になる。** 親は毎ディール 1 つ回るだけなので、
// 6 ラウンドだと 2 席だけが親を 2 回引き受ける。親の側はペア同点（20-20、
// 実測 8.1%）のときに宣言側を取るので、その回数が偏るとそのまま有利不利になる。
func (c MinibridgeConfig) Validate() error {
	if c.Rounds < MinibridgeRoundsMin || c.Rounds > MinibridgeRoundsMax {
		return fmt.Errorf("rounds must be between %d and %d, got %d",
			MinibridgeRoundsMin, MinibridgeRoundsMax, c.Rounds)
	}
	if c.Rounds%MinibridgePlayerCnt != 0 {
		return fmt.Errorf("rounds must be a multiple of %d so every seat deals equally, got %d",
			MinibridgePlayerCnt, c.Rounds)
	}
	return nil
}
