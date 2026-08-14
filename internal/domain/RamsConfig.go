//go:build !js || !wasm || extra2

package domain

// RamsPlayerCntMin プレイヤー数の下限
const RamsPlayerCntMin = 3

// RamsPlayerCntMax プレイヤー数の上限
const RamsPlayerCntMax = 5

// RamsPlayerCntDefault 既定のプレイヤー数
const RamsPlayerCntDefault = 4

// RamsRoundsMin ラウンド数の下限
const RamsRoundsMin = 1

// RamsRoundsMax ラウンド数の上限
const RamsRoundsMax = 12

// RamsRoundsDefault 既定のラウンド数
const RamsRoundsDefault = 4

// RamsConfig ラムス ゲーム設定
type RamsConfig struct {
	// PlayerCnt 参加人数 (3〜5)。**ラムスは可変人数が特徴**なので設定で変える。
	PlayerCnt int `json:"pc"`
	// Rounds 何ラウンドで打ち切るか
	Rounds int `json:"rd"`
}

// DefaultRamsConfig デフォルト設定を返す
func DefaultRamsConfig() RamsConfig {
	return RamsConfig{PlayerCnt: RamsPlayerCntDefault, Rounds: RamsRoundsDefault}
}

// Validate 設定値のドメインバリデーション
func (c RamsConfig) Validate() error {
	if err := ValidateRange("player count", c.PlayerCnt, RamsPlayerCntMin, RamsPlayerCntMax); err != nil {
		return err
	}
	return ValidateRange("rounds", c.Rounds, RamsRoundsMin, RamsRoundsMax)
}
