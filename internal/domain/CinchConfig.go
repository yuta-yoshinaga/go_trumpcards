//go:build !js || !wasm || extra

package domain

import "encoding/json"

// CinchCpuDifficulty は CPU の難易度レベル。
type CinchCpuDifficulty int

// CinchCpuDifficulty 定数
const (
	// CinchDifficultyEasy 簡単 (乱数を用いたビッド・プレイ)
	CinchDifficultyEasy CinchCpuDifficulty = 0
	// CinchDifficultyNormal 通常 (手札評価による貪欲な判断)
	CinchDifficultyNormal CinchCpuDifficulty = 1
	// CinchDifficultyHard 難しい (温存・セット狙いを含む最適化)
	CinchDifficultyHard CinchCpuDifficulty = 2
)

// CinchDifficultyNames 難易度名マップ
var CinchDifficultyNames = map[CinchCpuDifficulty]string{
	CinchDifficultyEasy:   "Easy",
	CinchDifficultyNormal: "Normal",
	CinchDifficultyHard:   "Hard",
}

// CinchConfig はチンチ (Cinch / Double Pedro / High Five) のローカルルール設定。
type CinchConfig struct {
	// CpuDifficulty CPU 難易度
	CpuDifficulty CinchCpuDifficulty `json:"cd"`
	// PointLimit ゲーム終了スコア (先に到達したプレイヤーが勝利, デフォルト 21)
	PointLimit int `json:"pl"`
}

// DefaultCinchConfig はデフォルトのローカルルール設定を返す。
//   - プレイヤー数: 4 (1 human + 3 CPU, 個人戦)
//   - デッキ: 52 枚
//   - 各プレイヤー 9 枚配り、1 ディールあたり計 14 ポイント
//   - CPU 難易度: 3 段階
//   - 目標得点: 21 (先取勝ち)
func DefaultCinchConfig() CinchConfig {
	return CinchConfig{
		CpuDifficulty: CinchDifficultyNormal,
		PointLimit:    21,
	}
}

// Validate は設定値のドメインバリデーションを行う。
func (c CinchConfig) Validate() error {
	if err := ValidateRange(
		"CPU difficulty",
		int(c.CpuDifficulty),
		int(CinchDifficultyEasy),
		int(CinchDifficultyHard),
	); err != nil {
		return err
	}
	return ValidateMin("point limit", c.PointLimit, 1)
}

// cinchConfigJSON is the JSON wire format for CinchConfig.
type cinchConfigJSON struct {
	CpuDifficulty CinchCpuDifficulty `json:"cd"`
	PointLimit    int                `json:"pl"`
}

// MarshalJSON implements json.Marshaler.
func (c CinchConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(cinchConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *CinchConfig) UnmarshalJSON(data []byte) error {
	var j cinchConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	*c = CinchConfig(j)
	return nil
}
