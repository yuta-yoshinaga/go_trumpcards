//go:build !js || !wasm || extra

package domain

import "fmt"

// SergeantMajorRoundsMin / SergeantMajorRoundsMax は許容するラウンド数の範囲。
//
// **役割が 3 つあるので 3 の倍数が自然。** 3 ラウンドで全員が 8・5・3 を
// 一度ずつ引き受けます。
const (
	SergeantMajorRoundsMin = 3
	SergeantMajorRoundsMax = 30
)

// SergeantMajorConfig はサージェントメジャーのゲーム設定。
type SergeantMajorConfig struct {
	// Rounds は打つラウンド数。
	Rounds int `json:"r"`
}

// DefaultSergeantMajorConfig はデフォルト設定を返す。
func DefaultSergeantMajorConfig() SergeantMajorConfig {
	return SergeantMajorConfig{Rounds: SergeantMajorDefaultRounds}
}

// Validate は設定値のドメインバリデーション。
//
// **3 の倍数でなければ不公平になる（レビュー指摘 PR #5311）。** 親は毎ラウンド
// 1 つずつ回るだけなので、たとえば 4 ラウンドだと 1 人だけノルマ 8 を 2 回
// 引き受け、他の 2 人は 1 回で済む。ノルマ 8 は届きにくく得点が下がりやすい
// ので、これは有利不利がそのまま偏るということ。
func (c SergeantMajorConfig) Validate() error {
	if err := ValidateRange("rounds", c.Rounds, SergeantMajorRoundsMin, SergeantMajorRoundsMax); err != nil {
		return err
	}
	if c.Rounds%SergeantMajorPlayerCnt != 0 {
		return fmt.Errorf("rounds must be a multiple of %d so every seat takes each target equally, got %d",
			SergeantMajorPlayerCnt, c.Rounds)
	}
	return nil
}
