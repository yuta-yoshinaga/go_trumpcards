//go:build !js || !wasm || extra2

package domain

// JulepePlayerCntMin プレイヤー数の下限
const JulepePlayerCntMin = 3

// JulepePlayerCntMax プレイヤー数の上限
const JulepePlayerCntMax = 5

// JulepePlayerCntDefault 既定のプレイヤー数
const JulepePlayerCntDefault = 4

// JulepeRoundsMin ラウンド数の下限
const JulepeRoundsMin = 1

// JulepeRoundsMax ラウンド数の上限
const JulepeRoundsMax = 12

// JulepeRoundsDefault 既定のラウンド数
const JulepeRoundsDefault = 4

// JulepeConfig フレペ ゲーム設定
type JulepeConfig struct {
	// PlayerCnt 参加人数 (3〜5)。**フレペは可変人数が特徴**なので設定で変える。
	PlayerCnt int `json:"pc"`
	// Rounds 何ラウンドで打ち切るか
	Rounds int `json:"rd"`
}

// DefaultJulepeConfig デフォルト設定を返す
func DefaultJulepeConfig() JulepeConfig {
	return JulepeConfig{PlayerCnt: JulepePlayerCntDefault, Rounds: JulepeRoundsDefault}
}

// Validate 設定値のドメインバリデーション
func (c JulepeConfig) Validate() error {
	if err := ValidateRange("player count", c.PlayerCnt, JulepePlayerCntMin, JulepePlayerCntMax); err != nil {
		return err
	}
	return ValidateRange("rounds", c.Rounds, JulepeRoundsMin, JulepeRoundsMax)
}
