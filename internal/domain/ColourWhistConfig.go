//go:build !js || !wasm || extra4

package domain

import (
	"encoding/json"
	"fmt"
)

// カラーホイストの卓の大きさ (4 人固定)。
const (
	// ColourWhistPlayerCnt は参加人数。
	ColourWhistPlayerCnt = 4
	// ColourWhistHandSize は 1 人あたりの手札枚数 (52 / 4)。
	ColourWhistHandSize = 13
	// ColourWhistTrickCnt は 1 ラウンドのトリック数。
	ColourWhistTrickCnt = ColourWhistHandSize
)

// 契約の種類。**数字がそのまま競りの強さ**で、上へしか積み増せません。
const (
	// ColourWhistContractNone まだ誰も宣言していない
	ColourWhistContractNone = 0
	// ColourWhistContractSamen 相方を呼んで 2 対 2 で 8 トリック
	ColourWhistContractSamen = 1
	// ColourWhistContractAlleen 単独で 8 トリック
	ColourWhistContractAlleen = 2
	// ColourWhistContractMiserie 1 トリックも取らない (単独)
	ColourWhistContractMiserie = 3
	// ColourWhistContractTroel エース 3 枚で強制的に成立する 2 対 2 の契約
	//
	// **競りでは選べません。** 配った時点で誰かがエースを 3 枚持っていれば
	// 自動的にこれになります。
	ColourWhistContractTroel = 4
)

// ColourWhistContractMax は契約値の上限。
const ColourWhistContractMax = ColourWhistContractTroel

// ColourWhistBidMax は**競りで宣言できる**上限。Troel は競りに出てきません。
const ColourWhistBidMax = ColourWhistContractMiserie

// 契約の目標トリック数。
const (
	// ColourWhistTrickTarget は 8 トリック系契約の目標。
	ColourWhistTrickTarget = 8
	// ColourWhistTroelAces は Troel が発生するエースの枚数。
	ColourWhistTroelAces = 3
)

// 得点表。**すべてゼロサム**で、卓の合計は必ず 0 になります。
const (
	// ColourWhistSamenBase は Samen / Troel の基本点 (守備側 1 人あたり)。
	ColourWhistSamenBase = 1
	// ColourWhistAlleenBase は Alleen の基本点 (守備側 1 人あたり)。
	ColourWhistAlleenBase = 3
	// ColourWhistMiseriePoints は Miserie の点 (守備側 1 人あたり)。
	ColourWhistMiseriePoints = 4
)

// ColourWhistRoundsMin / Max / Default は 1 ゲームのラウンド数の範囲。
const (
	ColourWhistRoundsMin     = 4
	ColourWhistRoundsMax     = 20
	ColourWhistDefaultRounds = 8
)

// ColourWhistConfig はカラーホイストのゲーム設定。
type ColourWhistConfig struct {
	// Rounds は 1 ゲームのラウンド数。
	Rounds int
}

// DefaultColourWhistConfig はデフォルト設定を返す。
func DefaultColourWhistConfig() ColourWhistConfig {
	return ColourWhistConfig{Rounds: ColourWhistDefaultRounds}
}

// Validate は設定値の妥当性を検証する。
func (c ColourWhistConfig) Validate() error {
	return ValidateRange("rounds", c.Rounds, ColourWhistRoundsMin, ColourWhistRoundsMax)
}

// colourWhistConfigJSON is the JSON wire format for ColourWhistConfig.
type colourWhistConfigJSON struct {
	Rounds int `json:"rd"`
}

// MarshalJSON implements json.Marshaler.
func (c ColourWhistConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(colourWhistConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *ColourWhistConfig) UnmarshalJSON(data []byte) error {
	var j colourWhistConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	c.Rounds = j.Rounds
	return c.Validate()
}

// ColourWhistContractName は契約の識別子を返す (i18n キーの一部に使います)。
func ColourWhistContractName(contract int) string {
	switch contract {
	case ColourWhistContractSamen:
		return "samen"
	case ColourWhistContractAlleen:
		return "alleen"
	case ColourWhistContractMiserie:
		return "miserie"
	case ColourWhistContractTroel:
		return "troel"
	default:
		return "none"
	}
}

// ColourWhistIsMiserie は契約が「1 トリックも取らない」系かを返す。
func ColourWhistIsMiserie(contract int) bool { return contract == ColourWhistContractMiserie }

// ColourWhistNeedsTrump は契約に切り札が要るかを返す。
//
// **Miserie だけが切り札なし**です。
func ColourWhistNeedsTrump(contract int) bool {
	return contract == ColourWhistContractSamen ||
		contract == ColourWhistContractAlleen ||
		contract == ColourWhistContractTroel
}

// ColourWhistHasPartner は契約に相方が付くかを返す。
//
// **Samen と Troel が 2 対 2** です。Troel は競りではなく配りで決まります。
func ColourWhistHasPartner(contract int) bool {
	return contract == ColourWhistContractSamen || contract == ColourWhistContractTroel
}

// ColourWhistContractTarget は契約の目標トリック数を返す。
func ColourWhistContractTarget(contract int) int {
	if ColourWhistIsMiserie(contract) || contract == ColourWhistContractNone {
		return 0
	}
	return ColourWhistTrickTarget
}

// colourWhistValidateContract は契約が範囲内かを検証する。
func colourWhistValidateContract(contract int) error {
	if contract < ColourWhistContractNone || contract > ColourWhistContractMax {
		return fmt.Errorf("colourwhist: contract out of range: %d", contract)
	}
	return nil
}
