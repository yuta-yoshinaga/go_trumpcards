//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"errors"
)

// 卓の範囲設定。
const (
	// TuSacMinSeats は最小席数。
	TuSacMinSeats = 2
	// TuSacMaxSeats は最大席数。
	TuSacMaxSeats = 4
	// TuSacDefaultSeats は既定の席数。
	TuSacDefaultSeats = 4

	// TuSacMinRounds は最小ラウンド数。
	TuSacMinRounds = 1
	// TuSacMaxRounds は最大ラウンド数。
	TuSacMaxRounds = 20
	// TuSacDefaultRounds は既定のラウンド数。
	TuSacDefaultRounds = 5
)

// 配札の形。
const (
	// TuSacHandSize は各席に配る枚数。
	//
	// **20 枚配る。** 四色牌は 112 枚と多く、メルドが 3 枚 (同色同種) や
	// 5 枚 (卒) と大きいので、10 枚程度では組み合わせが立たないまま山が
	// 尽きる。4 席 × 20 = 80 枚を配っても山に 32 枚残る。
	TuSacHandSize = 20
	// TuSacSetSize は同色同種のメルドの枚数。
	TuSacSetSize = 3
	// TuSacTrioSize は車馬砲のメルドの枚数。
	TuSacTrioSize = 3
	// TuSacSoldierSetSize は卒のメルドの枚数。
	TuSacSoldierSetSize = 5
)

// TuSacMeldKind はメルドの種類。
type TuSacMeldKind int

const (
	// TuSacMeldNone はメルドではない。
	TuSacMeldNone TuSacMeldKind = iota
	// TuSacMeldSameColorSet は同色・同種 3 枚。
	TuSacMeldSameColorSet
	// TuSacMeldChariotTrio は異色の 車・馬・砲 3 枚。
	TuSacMeldChariotTrio
	// TuSacMeldSoldierSet は卒 5 枚。
	TuSacMeldSoldierSet
)

// TuSacMeldKindMax は最大のメルド種別 (復元時の範囲検査に使う)。
const TuSacMeldKindMax = TuSacMeldSoldierSet

// TuSacMeldKindName はメルドの識別子を返す (i18n キーの一部に使う)。
func TuSacMeldKindName(k TuSacMeldKind) string {
	switch k {
	case TuSacMeldSameColorSet:
		return "sameColorSet"
	case TuSacMeldChariotTrio:
		return "chariotTrio"
	case TuSacMeldSoldierSet:
		return "soldierSet"
	default:
		return "none"
	}
}

// TuSacMeldPoints はメルド 1 つの得点を返す。
//
// **枚数が多いほど、そろえにくいほど高い。** 卒 5 枚は 112 枚から 5 枚を
// そろえる必要があり、同色同種 3 枚よりはるかに遠い ── 同じ点にすると
// 大きいメルドを狙う理由が無くなる。
func TuSacMeldPoints(k TuSacMeldKind) int {
	switch k {
	case TuSacMeldSoldierSet:
		return 5
	case TuSacMeldChariotTrio:
		return 3
	case TuSacMeldSameColorSet:
		return 2
	default:
		return 0
	}
}

// エラー値。設定検証で使う。
var (
	errTuSacSeatsRange  = errors.New("tusac: seats out of range")
	errTuSacRoundsRange = errors.New("tusac: rounds out of range")
	errTuSacDeckShort   = errors.New("tusac: the deck cannot serve this many seats")
)

// TuSacConfig は四色牌の卓設定。
type TuSacConfig struct {
	Seats  int
	Rounds int
}

// DefaultTuSacConfig は既定の設定を返す。
func DefaultTuSacConfig() TuSacConfig {
	return TuSacConfig{Seats: TuSacDefaultSeats, Rounds: TuSacDefaultRounds}
}

// Validate は設定が範囲内かを検査する。
//
// **山にも札を残す。** 配り切ってしまうと引くところが無くなり、手番が
// 「捨てるだけ」になる ── 引いて捨てるのがこのゲームの手番なので、
// 席数ぶん配ってなお山が残ることを確かめる。
func (c TuSacConfig) Validate() error {
	switch {
	case c.Seats < TuSacMinSeats || c.Seats > TuSacMaxSeats:
		return errTuSacSeatsRange
	case c.Rounds < TuSacMinRounds || c.Rounds > TuSacMaxRounds:
		return errTuSacRoundsRange
	}
	if c.Seats*TuSacHandSize >= TuSacDeckSize {
		return errTuSacDeckShort
	}
	return nil
}

// tuSacConfigJSON は TuSacConfig の JSON 表現。
type tuSacConfigJSON struct {
	Seats  int `json:"s"`
	Rounds int `json:"r"`
}

// MarshalJSON implements json.Marshaler.
func (c TuSacConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(tuSacConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *TuSacConfig) UnmarshalJSON(data []byte) error {
	var j tuSacConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	restored := TuSacConfig(j)
	if err := restored.Validate(); err != nil {
		return err
	}
	*c = restored
	return nil
}
