package domain

import "encoding/json"

// SlapjackCpuDifficulty CPU の難易度
type SlapjackCpuDifficulty int

// Slapjack CPU 難易度定数
const (
	// SlapjackCpuEasy 反応がゆっくりした初心者向け
	SlapjackCpuEasy SlapjackCpuDifficulty = 0
	// SlapjackCpuNormal 標準的な反応速度
	SlapjackCpuNormal SlapjackCpuDifficulty = 1
	// SlapjackCpuHard 高速な反応速度 (人間が勝てる範囲)
	SlapjackCpuHard SlapjackCpuDifficulty = 2
)

// SlapjackJackValue J のカード値
const SlapjackJackValue = 11

// SlapjackPenaltyCount 誤スラップ時のペナルティ枚数
const SlapjackPenaltyCount = 1

// SlapjackMinReactionMs CPU 反応時間の下限 (人間に勝ち目を残すための床値)
const SlapjackMinReactionMs = 80

// SlapjackConfig スラップジャック設定
type SlapjackConfig struct {
	// CpuDifficulty CPU の難易度
	CpuDifficulty SlapjackCpuDifficulty
}

// DefaultSlapjackConfig デフォルト設定を返す
func DefaultSlapjackConfig() SlapjackConfig {
	return SlapjackConfig{CpuDifficulty: SlapjackCpuNormal}
}

// Validate 設定値の検証
func (c SlapjackConfig) Validate() error {
	return ValidateRange("cpu difficulty", int(c.CpuDifficulty),
		int(SlapjackCpuEasy), int(SlapjackCpuHard))
}

// ReactionMeanMs 難易度に対応する反応時間の平均 (ms)
func (c SlapjackConfig) ReactionMeanMs() int {
	switch c.CpuDifficulty {
	case SlapjackCpuEasy:
		return 1100
	case SlapjackCpuHard:
		return 300
	default:
		return 600
	}
}

// ReactionStdDevMs 難易度に対応する反応時間のばらつき (ms)
func (c SlapjackConfig) ReactionStdDevMs() int {
	switch c.CpuDifficulty {
	case SlapjackCpuEasy:
		return 300
	case SlapjackCpuHard:
		return 120
	default:
		return 200
	}
}

// slapjackConfigJSON is the JSON wire format for SlapjackConfig.
type slapjackConfigJSON struct {
	CpuDifficulty int `json:"cd"`
}

// MarshalJSON implements json.Marshaler.
func (c SlapjackConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(slapjackConfigJSON{CpuDifficulty: int(c.CpuDifficulty)})
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *SlapjackConfig) UnmarshalJSON(data []byte) error {
	var j slapjackConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	c.CpuDifficulty = SlapjackCpuDifficulty(j.CpuDifficulty)
	return nil
}
