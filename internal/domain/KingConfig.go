//go:build !js || !wasm || extra

package domain

import "encoding/json"

// KingCpuDifficulty は CPU の難易度レベル。
type KingCpuDifficulty int

// KingCpuDifficulty 定数
const (
	// KingDifficultyEasy 簡単 (ランダムな合法手・コントラクトを最小 id 選択)
	KingDifficultyEasy KingCpuDifficulty = 0
	// KingDifficultyNormal 通常 (貪欲: 失点を避ける / 得点を狙う)
	KingDifficultyNormal KingCpuDifficulty = 1
	// KingDifficultyHard 難しい (手札評価でコントラクト・プレイを最適化)
	KingDifficultyHard KingCpuDifficulty = 2
)

// KingDifficultyNames 難易度名マップ
var KingDifficultyNames = map[KingCpuDifficulty]string{
	KingDifficultyEasy:   "Easy",
	KingDifficultyNormal: "Normal",
	KingDifficultyHard:   "Hard",
}

// KingConfig はキングのローカルルール設定。
type KingConfig struct {
	// CpuDifficulty CPU 難易度
	CpuDifficulty KingCpuDifficulty
}

// DefaultKingConfig はデフォルトのローカルルール設定を返す。
//   - プレイヤー数: 4
//   - デッキ: 52 枚
//   - 計 7 ディール (全 7 コントラクトを 1 巡)
//   - CPU 難易度: 3 段階
func DefaultKingConfig() KingConfig {
	return KingConfig{
		CpuDifficulty: KingDifficultyNormal,
	}
}

// Validate は設定値のドメインバリデーションを行う。
func (c KingConfig) Validate() error {
	return ValidateRange(
		"CPU difficulty",
		int(c.CpuDifficulty),
		int(KingDifficultyEasy),
		int(KingDifficultyHard),
	)
}

// kingConfigJSON is the JSON wire format for KingConfig.
type kingConfigJSON struct {
	CpuDifficulty KingCpuDifficulty `json:"di"`
}

// MarshalJSON implements json.Marshaler.
func (c KingConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(kingConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *KingConfig) UnmarshalJSON(data []byte) error {
	var j kingConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	*c = KingConfig(j)
	return nil
}
