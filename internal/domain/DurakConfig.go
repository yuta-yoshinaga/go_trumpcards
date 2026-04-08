package domain

import (
	"encoding/json"
	"fmt"
)

// DurakCpuDifficulty CPU難易度レベル
type DurakCpuDifficulty int

// DurakCpuDifficulty定数
const (
	// DurakDifficultyNormal 通常難易度 (デフォルト)
	DurakDifficultyNormal DurakCpuDifficulty = 0
	// DurakDifficultyEasy 簡単
	DurakDifficultyEasy DurakCpuDifficulty = 1
	// DurakDifficultyHard 難しい
	DurakDifficultyHard DurakCpuDifficulty = 2
)

// DurakPlayerCntMin 最小プレイヤー数
const DurakPlayerCntMin = 2

// DurakPlayerCntMax 最大プレイヤー数
const DurakPlayerCntMax = 6

// DurakPlayerCntDefault デフォルトプレイヤー数
const DurakPlayerCntDefault = 4

// DurakHandSize 各プレイヤーの初期手札枚数
const DurakHandSize = 6

// DurakConfig ドゥラーク設定
type DurakConfig struct {
	PlayerCount     int                `json:"-"` // プレイヤー数 (2-6)
	CpuDifficulty   DurakCpuDifficulty `json:"-"` // CPU難易度
	TransferEnabled bool               `json:"-"` // ペレヴォード (防御者が同ランクカードで次へ転送)
}

// DefaultDurakConfig デフォルト設定
func DefaultDurakConfig() DurakConfig {
	return DurakConfig{
		PlayerCount:     DurakPlayerCntDefault,
		CpuDifficulty:   DurakDifficultyNormal,
		TransferEnabled: false,
	}
}

// Validate 設定値のバリデーション
func (c DurakConfig) Validate() error {
	if err := ValidateRange("player count", c.PlayerCount, DurakPlayerCntMin, DurakPlayerCntMax); err != nil {
		return err
	}
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(DurakDifficultyNormal), int(DurakDifficultyHard)); err != nil {
		return err
	}
	// 36枚デッキ / プレイヤー数 で配りきれるか検証
	if c.PlayerCount*DurakHandSize > 36 {
		return fmt.Errorf("too many players for 36-card deck: %d players × %d cards = %d > 36", c.PlayerCount, DurakHandSize, c.PlayerCount*DurakHandSize)
	}
	return nil
}

// durakConfigJSON is the JSON wire format for DurakConfig.
type durakConfigJSON struct {
	PlayerCount     int                `json:"pc"`
	CpuDifficulty   DurakCpuDifficulty `json:"di"`
	TransferEnabled bool               `json:"te"`
}

// MarshalJSON implements json.Marshaler.
func (c DurakConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(durakConfigJSON{
		PlayerCount:     c.PlayerCount,
		CpuDifficulty:   c.CpuDifficulty,
		TransferEnabled: c.TransferEnabled,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *DurakConfig) UnmarshalJSON(data []byte) error {
	var j durakConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	c.PlayerCount = j.PlayerCount
	c.CpuDifficulty = j.CpuDifficulty
	c.TransferEnabled = j.TransferEnabled
	return nil
}
