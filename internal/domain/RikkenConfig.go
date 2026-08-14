//go:build !js || !wasm || extra2

package domain

import (
	"encoding/json"
	"fmt"
)

// リッケンの卓の大きさ (4 人固定)。
const (
	// RikkenPlayerCnt は参加人数。
	RikkenPlayerCnt = 4
	// RikkenHandSize は 1 人あたりの手札枚数 (52 / 4)。
	RikkenHandSize = 13
	// RikkenTrickCnt は 1 ラウンドのトリック数。
	RikkenTrickCnt = RikkenHandSize
)

// 契約の種類。**数字がそのまま競りの強さ**で、上へしか積み増せません。
const (
	// RikkenContractNone まだ誰も宣言していない
	RikkenContractNone = 0
	// RikkenContractRik 呼んだ相方と組んで 8 トリック (切り札あり)
	RikkenContractRik = 1
	// RikkenContractMisere 1 トリックも取らない (単独・切り札なし)
	RikkenContractMisere = 2
	// RikkenContractSolo 単独で 6 トリック (切り札あり)
	RikkenContractSolo = 3
	// RikkenContractOpenMisere 手札を公開して 1 トリックも取らない (単独)
	RikkenContractOpenMisere = 4
)

// RikkenContractMax は最強の契約。
const RikkenContractMax = RikkenContractOpenMisere

// 契約ごとの必要トリック数。
const (
	// RikkenRikTarget は Rik の目標 (相方と合わせて)。
	RikkenRikTarget = 8
	// RikkenSoloTarget は Solo の目標 (単独で)。
	RikkenSoloTarget = 6
)

// 得点表。**すべてゼロサム**で、卓の合計は必ず 0 になります。
const (
	// RikkenRikBase は Rik を達成したときに宣言側の各人が受け取る基本点。
	RikkenRikBase = 1
	// RikkenSoloBase は Solo を達成したときの基本点 (守備側 1 人あたり)。
	RikkenSoloBase = 2
	// RikkenMiserePoints は Misere の点 (守備側 1 人あたり)。
	RikkenMiserePoints = 3
	// RikkenOpenMiserePoints は Open Misere の点 (守備側 1 人あたり)。
	RikkenOpenMiserePoints = 5
)

// RikkenRoundsMin / Max / Default は 1 ゲームのラウンド数の範囲。
const (
	RikkenRoundsMin     = 4
	RikkenRoundsMax     = 20
	RikkenDefaultRounds = 8
)

// RikkenConfig はリッケンのゲーム設定。
type RikkenConfig struct {
	// Rounds は 1 ゲームのラウンド数。
	Rounds int
}

// DefaultRikkenConfig はデフォルト設定を返す。
func DefaultRikkenConfig() RikkenConfig {
	return RikkenConfig{Rounds: RikkenDefaultRounds}
}

// Validate は設定値の妥当性を検証する。
//
// **ラウンド数は 4 の倍数でなくてよい。** 親が一巡しないと不公平ですが、
// リッケンは契約を競り落とした人が主役なので、席順の偏りは小さいためです。
func (c RikkenConfig) Validate() error {
	return ValidateRange("rounds", c.Rounds, RikkenRoundsMin, RikkenRoundsMax)
}

// rikkenConfigJSON is the JSON wire format for RikkenConfig.
type rikkenConfigJSON struct {
	Rounds int `json:"rd"`
}

// MarshalJSON implements json.Marshaler.
func (c RikkenConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(rikkenConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *RikkenConfig) UnmarshalJSON(data []byte) error {
	var j rikkenConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	c.Rounds = j.Rounds
	return c.Validate()
}

// RikkenContractName は契約の識別子を返す (i18n キーの一部に使います)。
func RikkenContractName(contract int) string {
	switch contract {
	case RikkenContractRik:
		return "rik"
	case RikkenContractMisere:
		return "misere"
	case RikkenContractSolo:
		return "solo"
	case RikkenContractOpenMisere:
		return "openMisere"
	default:
		return "none"
	}
}

// RikkenContractTarget は契約の目標トリック数を返す。
//
// **Misere 系は 0 トリックが目標**なので、ここは「取ってよい上限」ではなく
// 「達成に必要な数」として 0 を返します。判定側で向きを変えます。
func RikkenContractTarget(contract int) int {
	switch contract {
	case RikkenContractRik:
		return RikkenRikTarget
	case RikkenContractSolo:
		return RikkenSoloTarget
	default:
		return 0
	}
}

// RikkenIsMisere は契約が「1 トリックも取らない」系かを返す。
func RikkenIsMisere(contract int) bool {
	return contract == RikkenContractMisere || contract == RikkenContractOpenMisere
}

// RikkenNeedsTrump は契約に切り札が要るかを返す。
func RikkenNeedsTrump(contract int) bool {
	return contract == RikkenContractRik || contract == RikkenContractSolo
}

// RikkenHasPartner は契約に相方が付くかを返す。**Rik だけが 2 対 2** です。
func RikkenHasPartner(contract int) bool { return contract == RikkenContractRik }

// rikkenValidateContract は契約が範囲内かを検証する。
func rikkenValidateContract(contract int) error {
	if contract < RikkenContractNone || contract > RikkenContractMax {
		return fmt.Errorf("rikken: contract out of range: %d", contract)
	}
	return nil
}
