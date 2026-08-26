//go:build !js || !wasm || extra4

package domain

import "encoding/json"

// MightyCpuDifficulty CPU の難易度レベル
type MightyCpuDifficulty int

// マイティのCPU難易度定数
const (
	// MightyCpuDifficultyEasy 低難易度
	MightyCpuDifficultyEasy MightyCpuDifficulty = iota
	// MightyCpuDifficultyNormal 中難易度
	MightyCpuDifficultyNormal
	// MightyCpuDifficultyHard 高難易度
	MightyCpuDifficultyHard
)

// MightyConfig マイティゲーム設定
type MightyConfig struct {
	CpuDifficulty MightyCpuDifficulty
	MinBid        int // 最低ビッド値 (デフォルト13)
	NoTrumpExtra  int // ノートランプ宣言時のビッド加算値 (デフォルト2)
	PointLimit    int // ゲーム終了スコア (先に到達したチームが勝利)
}

// DefaultMightyConfig デフォルト設定を返す
func DefaultMightyConfig() MightyConfig {
	return MightyConfig{
		CpuDifficulty: MightyCpuDifficultyNormal,
		MinBid:        13,
		NoTrumpExtra:  2,
		PointLimit:    100,
	}
}

// Validate 設定値のドメインバリデーション
func (c MightyConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(MightyCpuDifficultyEasy), int(MightyCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateRange("min bid", c.MinBid, 1, MightyMaxPoints); err != nil {
		return err
	}
	if err := ValidateRange("no trump extra", c.NoTrumpExtra, 0, MightyMaxPoints-1); err != nil {
		return err
	}
	if err := ValidateMin("point limit", c.PointLimit, 1); err != nil {
		return err
	}
	return nil
}

// mightyConfigJSON is the JSON wire format for MightyConfig.
type mightyConfigJSON struct {
	CpuDifficulty MightyCpuDifficulty `json:"cd"`
	MinBid        int                 `json:"mb"`
	NoTrumpExtra  int                 `json:"nt"`
	PointLimit    int                 `json:"pl"`
}

// MarshalJSON implements json.Marshaler.
func (c MightyConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(mightyConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *MightyConfig) UnmarshalJSON(data []byte) error {
	var j mightyConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	c.CpuDifficulty = j.CpuDifficulty
	c.MinBid = j.MinBid
	c.NoTrumpExtra = j.NoTrumpExtra
	c.PointLimit = j.PointLimit
	return nil
}
