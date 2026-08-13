//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"errors"
)

// HorseDiscipline は H.O.R.S.E. の 1 種目。
//
// **並びが頭文字そのもの。** H-O-R-S-E の順に回るので、値の順序を入れ替えると
// ゲームの名前と実際のローテーションが食い違う。
type HorseDiscipline int

const (
	// HorseHoldem は Texas Hold'em (H)。
	HorseHoldem HorseDiscipline = iota
	// HorseOmahaHiLo は Omaha Hi-Lo (O)。
	HorseOmahaHiLo
	// HorseRazz は Razz (R)。
	HorseRazz
	// HorseStud は Seven Card Stud (S)。
	HorseStud
	// HorseStudHiLo は Seven Card Stud Hi-Lo Eight or Better (E)。
	HorseStudHiLo
)

// HorseDisciplineCount は種目数。
const HorseDisciplineCount = 5

// HorseDisciplineMax は最大の種目値 (復元時の範囲検査に使う)。
const HorseDisciplineMax = HorseStudHiLo

// HorseDisciplineName は種目の識別子を返す (i18n キーの一部に使う)。
func HorseDisciplineName(d HorseDiscipline) string {
	switch d {
	case HorseHoldem:
		return "holdem"
	case HorseOmahaHiLo:
		return "omahaHiLo"
	case HorseRazz:
		return "razz"
	case HorseStud:
		return "stud"
	case HorseStudHiLo:
		return "studHiLo"
	default:
		return "unknown"
	}
}

// HorseDisciplineLetter は種目の頭文字を返す (H.O.R.S.E. の由来)。
func HorseDisciplineLetter(d HorseDiscipline) string {
	switch d {
	case HorseHoldem:
		return "H"
	case HorseOmahaHiLo:
		return "O"
	case HorseRazz:
		return "R"
	case HorseStud:
		return "S"
	case HorseStudHiLo:
		return "E"
	default:
		return "?"
	}
}

// 卓の範囲設定。
const (
	// HorseMinSeats は最小席数。
	HorseMinSeats = 2
	// HorseMaxSeats は最大席数。
	HorseMaxSeats = 6
	// HorseDefaultSeats は既定の席数。
	HorseDefaultSeats = 6

	// HorseMinChips は最小の初期チップ。
	HorseMinChips = 100
	// HorseMaxChips は最大の初期チップ。
	HorseMaxChips = 100000
	// HorseDefaultChips は既定の初期チップ。
	HorseDefaultChips = 1000

	// HorseMinHandsPerDiscipline は 1 種目あたりの最小ハンド数。
	HorseMinHandsPerDiscipline = 1
	// HorseMaxHandsPerDiscipline は 1 種目あたりの最大ハンド数。
	HorseMaxHandsPerDiscipline = 10
	// HorseDefaultHandsPerDiscipline は既定のハンド数。
	HorseDefaultHandsPerDiscipline = 2
)

// エラー値。設定検証で使う。
var (
	errHorseSeatsRange = errors.New("horse: seats out of range")
	errHorseChipsRange = errors.New("horse: initial chips out of range")
	errHorseHandsRange = errors.New("horse: hands per discipline out of range")
)

// HorseConfig は H.O.R.S.E. の卓設定。
type HorseConfig struct {
	Seats              int
	InitialChips       int
	HandsPerDiscipline int
}

// DefaultHorseConfig は既定の設定を返す。
func DefaultHorseConfig() HorseConfig {
	return HorseConfig{
		Seats:              HorseDefaultSeats,
		InitialChips:       HorseDefaultChips,
		HandsPerDiscipline: HorseDefaultHandsPerDiscipline,
	}
}

// Validate は設定が範囲内かを検査する。
func (c HorseConfig) Validate() error {
	switch {
	case c.Seats < HorseMinSeats || c.Seats > HorseMaxSeats:
		return errHorseSeatsRange
	case c.InitialChips < HorseMinChips || c.InitialChips > HorseMaxChips:
		return errHorseChipsRange
	case c.HandsPerDiscipline < HorseMinHandsPerDiscipline || c.HandsPerDiscipline > HorseMaxHandsPerDiscipline:
		return errHorseHandsRange
	}
	return nil
}

// horseConfigJSON は HorseConfig の JSON 表現。
type horseConfigJSON struct {
	Seats              int `json:"s"`
	InitialChips       int `json:"c"`
	HandsPerDiscipline int `json:"h"`
}

// MarshalJSON implements json.Marshaler.
func (c HorseConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(horseConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **復元した設定も Validate に通す。** 保存を書き換えれば席数 0 の卓を作れる。
func (c *HorseConfig) UnmarshalJSON(data []byte) error {
	var j horseConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	restored := HorseConfig(j)
	if err := restored.Validate(); err != nil {
		return err
	}
	*c = restored
	return nil
}
