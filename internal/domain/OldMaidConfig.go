package domain

import (
	"encoding/json"
	"fmt"
)

// OldMaidMode ババ抜きモード
type OldMaidMode int

// OldMaidのゲームモード定数
const (
	OldMaidModeNormal   OldMaidMode = iota // ババ抜き: ジョーカーが奇数カード
	OldMaidModeJijiNuki                    // ジジ抜き: ランダム1枚除外
)

// OldMaidConfig ババ抜き設定
type OldMaidConfig struct {
	Mode                 OldMaidMode
	CpuPlacementStrategy bool // CPU心理戦: 奇数カードを端に配置
	CpuMemoryAI          bool // CPU記憶AI: 引いた位置を記憶して戦略的に選択
	CpuHesitationEnabled bool // CPU迷い時間ディレイ
	CpuMetaAI            bool // メタAI: セッション内学習
}

// DefaultOldMaidConfig デフォルト設定を返す
func DefaultOldMaidConfig() OldMaidConfig {
	return OldMaidConfig{Mode: OldMaidModeNormal, CpuPlacementStrategy: false}
}

// Validate 設定値のドメインバリデーション
func (c OldMaidConfig) Validate() error {
	if c.Mode != OldMaidModeNormal && c.Mode != OldMaidModeJijiNuki {
		return fmt.Errorf("invalid game mode %d, must be %d (Normal) or %d (JijiNuki)", int(c.Mode), int(OldMaidModeNormal), int(OldMaidModeJijiNuki))
	}
	return nil
}

// oldMaidConfigJSON is the JSON wire format for OldMaidConfig.
type oldMaidConfigJSON struct {
	Mode                 OldMaidMode `json:"md"`
	CpuPlacementStrategy bool        `json:"cp"`
	CpuMemoryAI          bool        `json:"cm"`
	CpuHesitationEnabled bool        `json:"ch"`
	CpuMetaAI            bool        `json:"ca"`
}

// MarshalJSON implements json.Marshaler.
func (c OldMaidConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(oldMaidConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *OldMaidConfig) UnmarshalJSON(data []byte) error {
	var j oldMaidConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	c.Mode = j.Mode
	c.CpuPlacementStrategy = j.CpuPlacementStrategy
	c.CpuMemoryAI = j.CpuMemoryAI
	c.CpuHesitationEnabled = j.CpuHesitationEnabled
	c.CpuMetaAI = j.CpuMetaAI
	return nil
}
