//go:build !js || !wasm || classic

package domain

// HokmTargetMin 勝利に必要なハンド勝ち点の下限
const HokmTargetMin = 1

// HokmTargetMax 勝利に必要なハンド勝ち点の上限
const HokmTargetMax = 13

// HokmTargetDefault 既定の勝利ハンド数（本場のホクムは 7 ハンド先取）
const HokmTargetDefault = 7

// HokmConfig ホクム ゲーム設定
type HokmConfig struct {
	// Target 勝利に必要なハンド勝ち点
	Target int `json:"tg"`
}

// DefaultHokmConfig デフォルト設定を返す
func DefaultHokmConfig() HokmConfig {
	return HokmConfig{Target: HokmTargetDefault}
}

// Validate 設定値のドメインバリデーション
func (c HokmConfig) Validate() error {
	return ValidateRange("target", c.Target, HokmTargetMin, HokmTargetMax)
}
