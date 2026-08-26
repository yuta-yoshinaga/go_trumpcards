//go:build !js || !wasm || extra4

package domain

import "encoding/json"

// NapoleonCpuDifficulty CPU の難易度レベル
type NapoleonCpuDifficulty int

// ナポレオンのCPU難易度定数
const (
	// NapoleonCpuDifficultyEasy 低難易度
	NapoleonCpuDifficultyEasy NapoleonCpuDifficulty = iota
	// NapoleonCpuDifficultyNormal 中難易度
	NapoleonCpuDifficultyNormal
	// NapoleonCpuDifficultyHard 高難易度
	NapoleonCpuDifficultyHard
)

// NapoleonConfig ナポレオンゲーム設定
type NapoleonConfig struct {
	CpuDifficulty NapoleonCpuDifficulty
	MinBid        int // 最低ビッド値 (デフォルト12)
	PointLimit    int // ゲーム終了スコア (先に到達したプレイヤーが勝利)
}

// DefaultNapoleonConfig デフォルト設定を返す
func DefaultNapoleonConfig() NapoleonConfig {
	return NapoleonConfig{
		CpuDifficulty: NapoleonCpuDifficultyNormal,
		MinBid:        12,
		PointLimit:    100,
	}
}

// Validate 設定値のドメインバリデーション
func (c NapoleonConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(NapoleonCpuDifficultyEasy), int(NapoleonCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateRange("min bid", c.MinBid, 1, NapoleonMaxPictureCards); err != nil {
		return err
	}
	if err := ValidateMin("point limit", c.PointLimit, 1); err != nil {
		return err
	}
	return nil
}

// napoleonConfigJSON is the JSON wire format for NapoleonConfig.
type napoleonConfigJSON struct {
	CpuDifficulty NapoleonCpuDifficulty `json:"cd"`
	MinBid        int                   `json:"mb"`
	PointLimit    int                   `json:"pl"`
}

// MarshalJSON implements json.Marshaler.
func (c NapoleonConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(napoleonConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *NapoleonConfig) UnmarshalJSON(data []byte) error {
	var j napoleonConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	c.CpuDifficulty = j.CpuDifficulty
	c.MinBid = j.MinBid
	c.PointLimit = j.PointLimit
	return nil
}
