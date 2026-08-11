//go:build !js || !wasm || extra3

package domain

// HasenpfefferTargetMin / HasenpfefferTargetMax は許容する目標点の範囲。
const (
	HasenpfefferTargetMin = 5
	HasenpfefferTargetMax = 50
)

// HasenpfefferConfig はハーゼンプフェファーのゲーム設定。
type HasenpfefferConfig struct {
	// Target は勝利に必要な点数。
	Target int `json:"t"`
}

// DefaultHasenpfefferConfig はデフォルト設定を返す。
func DefaultHasenpfefferConfig() HasenpfefferConfig {
	return HasenpfefferConfig{Target: HasenpfefferDefaultTarget}
}

// Validate は設定値のドメインバリデーション。
func (c HasenpfefferConfig) Validate() error {
	return ValidateRange("target", c.Target, HasenpfefferTargetMin, HasenpfefferTargetMax)
}
