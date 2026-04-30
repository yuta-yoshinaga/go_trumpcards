package domain

import "encoding/json"

// EgyptianRatscrewCpuDifficulty CPU の難易度
type EgyptianRatscrewCpuDifficulty int

// Egyptian Ratscrew CPU 難易度定数
const (
	// EgyptianRatscrewCpuEasy 反応がゆっくりした初心者向け
	EgyptianRatscrewCpuEasy EgyptianRatscrewCpuDifficulty = 0
	// EgyptianRatscrewCpuNormal 標準的な反応速度
	EgyptianRatscrewCpuNormal EgyptianRatscrewCpuDifficulty = 1
	// EgyptianRatscrewCpuHard 高速な反応速度 (人間が勝てる範囲)
	EgyptianRatscrewCpuHard EgyptianRatscrewCpuDifficulty = 2
)

// 絵札カード値 (Slapjack の SlapjackJackValue と同じく 11=J, 12=Q, 13=K, 14=A)
const (
	// EgyptianRatscrewJackValue J のカード値
	EgyptianRatscrewJackValue = 11
	// EgyptianRatscrewQueenValue Q のカード値
	EgyptianRatscrewQueenValue = 12
	// EgyptianRatscrewKingValue K のカード値
	EgyptianRatscrewKingValue = 13
	// EgyptianRatscrewAceValue A のカード値
	EgyptianRatscrewAceValue = 14
)

// EgyptianRatscrewPenaltyCount 誤スラップ時のペナルティ枚数
const EgyptianRatscrewPenaltyCount = 1

// EgyptianRatscrewMinReactionMs CPU 反応時間の下限 (人間に勝ち目を残すための床値)
const EgyptianRatscrewMinReactionMs = 80

// FaceCardChances 絵札ごとに与えられるチャンス回数。
// J=1, Q=2, K=3, A=4 (古典的な Egyptian Ratscrew の規則)。
func FaceCardChances(cardValue int) int {
	switch cardValue {
	case EgyptianRatscrewJackValue:
		return 1
	case EgyptianRatscrewQueenValue:
		return 2
	case EgyptianRatscrewKingValue:
		return 3
	case EgyptianRatscrewAceValue:
		return 4
	default:
		return 0
	}
}

// IsFaceCard 値が絵札 (J/Q/K/A) なら true。
func IsFaceCard(cardValue int) bool {
	return FaceCardChances(cardValue) > 0
}

// EgyptianRatscrewConfig エジプシャン・ラットスクリュー設定
type EgyptianRatscrewConfig struct {
	// CpuDifficulty CPU の難易度
	CpuDifficulty EgyptianRatscrewCpuDifficulty
}

// DefaultEgyptianRatscrewConfig デフォルト設定を返す
func DefaultEgyptianRatscrewConfig() EgyptianRatscrewConfig {
	return EgyptianRatscrewConfig{CpuDifficulty: EgyptianRatscrewCpuNormal}
}

// Validate 設定値の検証
func (c EgyptianRatscrewConfig) Validate() error {
	return ValidateRange("cpu difficulty", int(c.CpuDifficulty),
		int(EgyptianRatscrewCpuEasy), int(EgyptianRatscrewCpuHard))
}

// ReactionMeanMs 難易度に対応する反応時間の平均 (ms)
func (c EgyptianRatscrewConfig) ReactionMeanMs() int {
	switch c.CpuDifficulty {
	case EgyptianRatscrewCpuEasy:
		return 1100
	case EgyptianRatscrewCpuHard:
		return 300
	default:
		return 600
	}
}

// ReactionStdDevMs 難易度に対応する反応時間のばらつき (ms)
func (c EgyptianRatscrewConfig) ReactionStdDevMs() int {
	switch c.CpuDifficulty {
	case EgyptianRatscrewCpuEasy:
		return 300
	case EgyptianRatscrewCpuHard:
		return 120
	default:
		return 200
	}
}

// egyptianRatscrewConfigJSON is the JSON wire format for EgyptianRatscrewConfig.
type egyptianRatscrewConfigJSON struct {
	CpuDifficulty int `json:"cd"`
}

// MarshalJSON implements json.Marshaler.
func (c EgyptianRatscrewConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(egyptianRatscrewConfigJSON{CpuDifficulty: int(c.CpuDifficulty)})
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *EgyptianRatscrewConfig) UnmarshalJSON(data []byte) error {
	var j egyptianRatscrewConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	c.CpuDifficulty = EgyptianRatscrewCpuDifficulty(j.CpuDifficulty)
	return nil
}
