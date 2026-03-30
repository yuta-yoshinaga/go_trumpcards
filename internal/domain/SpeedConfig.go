package domain

import "encoding/json"

// SpeedCpuDifficulty CPUの難易度
type SpeedCpuDifficulty int

// Speedの難易度定数
const (
	// SpeedCpuDifficultyEasy ランダムに1枚だけ出す
	SpeedCpuDifficultyEasy SpeedCpuDifficulty = iota
	// SpeedCpuDifficultyNormal 貪欲に複数枚出す
	SpeedCpuDifficultyNormal
	// SpeedCpuDifficultyHard 貪欲＋戦略的判断
	SpeedCpuDifficultyHard
)

// SpeedConfig スピードゲーム設定
type SpeedConfig struct {
	CpuDifficulty SpeedCpuDifficulty
}

// DefaultSpeedConfig デフォルト設定を返す
func DefaultSpeedConfig() SpeedConfig {
	return SpeedConfig{CpuDifficulty: SpeedCpuDifficultyNormal}
}

// Validate 設定値の検証
func (c SpeedConfig) Validate() error {
	return ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(SpeedCpuDifficultyEasy), int(SpeedCpuDifficultyHard))
}

// speedConfigJSON is the JSON wire format for SpeedConfig.
type speedConfigJSON struct {
	CpuDifficulty SpeedCpuDifficulty `json:"cd"`
}

// MarshalJSON implements json.Marshaler.
func (c SpeedConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(speedConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *SpeedConfig) UnmarshalJSON(data []byte) error {
	var j speedConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	c.CpuDifficulty = j.CpuDifficulty
	return nil
}
