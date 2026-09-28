//go:build !js || !wasm || extra3

package domain

// ToepenConfig はトゥーペンのゲーム設定。
type ToepenConfig struct {
	// PlayerCnt は参加人数。pagat は 3〜8 人としている。
	PlayerCnt int `json:"pc"`
}

// DefaultToepenConfig はデフォルト設定を返す。
func DefaultToepenConfig() ToepenConfig {
	return ToepenConfig{PlayerCnt: ToepenPlayerCnt}
}

// Validate は設定値のドメインバリデーション。
func (c ToepenConfig) Validate() error {
	return ValidateRange("player count", c.PlayerCnt, ToepenMinPlayers, ToepenMaxPlayers)
}
