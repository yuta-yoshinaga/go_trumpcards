package domain

import "encoding/json"

// PigsTailMinPlayers ぶたのしっぽ最小プレイヤー数 (人間1 + CPU1)
const PigsTailMinPlayers = 2

// PigsTailMaxPlayers ぶたのしっぽ最大プレイヤー数
const PigsTailMaxPlayers = 6

// PigsTailConfig ぶたのしっぽ設定
type PigsTailConfig struct {
	CpuHesitationEnabled bool // CPU迷い時間ディレイ
	PlayerCount          int  // 参加人数 (人間1 + CPU) — PigsTailMinPlayers..PigsTailMaxPlayers
}

// DefaultPigsTailConfig デフォルト設定を返す (4人: 人間1 + CPU3)
func DefaultPigsTailConfig() PigsTailConfig {
	return PigsTailConfig{PlayerCount: PigsTailPlayerCnt}
}

// Validate 設定値のドメインバリデーション
func (c PigsTailConfig) Validate() error {
	return ValidateRange("player count", c.PlayerCount, PigsTailMinPlayers, PigsTailMaxPlayers)
}

// pigsTailConfigJSON is the JSON wire format for PigsTailConfig.
type pigsTailConfigJSON struct {
	CpuHesitationEnabled bool `json:"ch"`
	PlayerCount          int  `json:"pc"`
}

// MarshalJSON implements json.Marshaler.
func (c PigsTailConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(pigsTailConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *PigsTailConfig) UnmarshalJSON(data []byte) error {
	var j pigsTailConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	c.CpuHesitationEnabled = j.CpuHesitationEnabled
	// 旧スナップショット (pc 未設定) との後方互換: 0 は既定値へ丸める。
	if j.PlayerCount == 0 {
		j.PlayerCount = PigsTailPlayerCnt
	}
	c.PlayerCount = j.PlayerCount
	return nil
}
