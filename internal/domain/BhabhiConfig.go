//go:build !js || !wasm || extra

package domain

// BhabhiMinPlayers / BhabhiMaxPlayers は許容するプレイヤー数の範囲。
//
// **3 人未満だと成立しない。** 2 人ではフォローできなかった側が場札を引き取り、
// 引き取った側が必ずリードに戻るので、手札が減る経路が実質消える。
const (
	BhabhiMinPlayers = 3
	BhabhiMaxPlayers = 7
)

// BhabhiConfig はバービーのゲーム設定。
type BhabhiConfig struct {
	// PlayerCnt は参加人数（人間 1 人 + CPU）。
	PlayerCnt int `json:"pc"`
}

// DefaultBhabhiConfig はデフォルト設定を返す。
func DefaultBhabhiConfig() BhabhiConfig {
	return BhabhiConfig{PlayerCnt: BhabhiDefaultPlayers}
}

// Validate は設定値のドメインバリデーション。
func (c BhabhiConfig) Validate() error {
	return ValidateRange("player count", c.PlayerCnt, BhabhiMinPlayers, BhabhiMaxPlayers)
}
