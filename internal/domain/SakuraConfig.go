//go:build !js || !wasm || extra3

package domain

import (
	"encoding/json"
	"errors"
)

// 卓の範囲設定。
const (
	// SakuraMinSeats は最小席数。
	SakuraMinSeats = 2
	// SakuraMaxSeats は最大席数。
	SakuraMaxSeats = 4
	// SakuraDefaultSeats は既定の席数。
	SakuraDefaultSeats = 3

	// SakuraMinRounds は最小ラウンド数。
	SakuraMinRounds = 1
	// SakuraMaxRounds は最大ラウンド数。
	SakuraMaxRounds = 12
	// SakuraDefaultRounds は既定のラウンド数。
	SakuraDefaultRounds = 3
)

// 配札の形。**花札 48 枚をそのまま使う。**
//
// デッキは `buildKoiKoiDeck` を流用する ── 同じ花札なので、月とインデックスの
// 対応を 2 か所に置くと必ず片方だけずれる。札の正体 (光/タネ/短冊/カス) も
// `KoiKoiCardCategory` が持っているので、こちらでは持たない。
const (
	// SakuraDeckSize は花札の総枚数。
	SakuraDeckSize = KoiKoiMonthCnt * KoiKoiCardsPerMonth
	// SakuraHandSize は各席に配る枚数。
	SakuraHandSize = 7
	// SakuraFieldSize は場に並べる枚数。
	SakuraFieldSize = 6
)

// 札の点数。**さくらは役ではなく点数の合計で競う。**
//
// こいこいや八八が「役ができたか」で勝負するのに対し、さくらは獲得した札を
// 種類ごとに数え上げて合計する ── 勝敗の決め方そのものが違うので、役の判定を
// 持ち込むと別のゲームになる。
const (
	// SakuraBrightPoints は光札の点数。
	SakuraBrightPoints = 20
	// SakuraAnimalPoints はタネ札の点数。
	SakuraAnimalPoints = 10
	// SakuraRibbonPoints は短冊札の点数。
	SakuraRibbonPoints = 5
	// SakuraChaffPoints はカス札の点数。
	SakuraChaffPoints = 1
)

// SakuraCardPoints は 1 枚の点数を返す。
//
// **札の正体は KoiKoi の分類をそのまま使う。** 同じ花札なので、光/タネ/短冊/
// カスの対応をここで作り直すと、同じ札が 2 つの正体を持つことになる。
func SakuraCardPoints(c *Card) int {
	if c == nil {
		return 0
	}
	switch KoiKoiCardCategory(c) {
	case KoiKoiBright:
		return SakuraBrightPoints
	case KoiKoiAnimal:
		return SakuraAnimalPoints
	case KoiKoiRibbon:
		return SakuraRibbonPoints
	default:
		return SakuraChaffPoints
	}
}

// 追加役に使う札の座標。**KoiKoi のデッキと同じ (月, インデックス)。**
//
// 花札の札は月とインデックスで一意に決まるので、名前ではなく座標で指す ──
// 「桜に幕」は 3 月の 1 番、「芒に月」は 8 月の 1 番。
const (
	// SakuraCurtainMonth は「桜に幕」の月 (3 月)。
	SakuraCurtainMonth = 3
	// SakuraCurtainIndex は「桜に幕」のインデックス。
	SakuraCurtainIndex = 1
	// SakuraMoonMonth は「芒に月」の月 (8 月)。
	SakuraMoonMonth = 8
	// SakuraMoonIndex は「芒に月」のインデックス。
	SakuraMoonIndex = 1
	// SakuraAllBrightsCount は光札の総枚数 (5 枚)。
	SakuraAllBrightsCount = 5
)

// SakuraBonus は独自の追加役。
type SakuraBonus int

const (
	// SakuraBonusNone は追加役なし。
	SakuraBonusNone SakuraBonus = iota
	// SakuraBonusSakuraSake は「桜に幕」と「芒に月」を両方取った役。
	SakuraBonusSakuraSake
	// SakuraBonusAllBrights は光札 5 枚すべてを取った役。
	SakuraBonusAllBrights
)

// SakuraBonusMax は最大の追加役 (復元時の範囲検査に使う)。
const SakuraBonusMax = SakuraBonusAllBrights

// SakuraBonusName は追加役の識別子を返す (i18n キーの一部に使う)。
func SakuraBonusName(b SakuraBonus) string {
	switch b {
	case SakuraBonusSakuraSake:
		return "sakuraSake"
	case SakuraBonusAllBrights:
		return "allBrights"
	default:
		return "none"
	}
}

// SakuraBonusPoints は追加役の点数を返す。
func SakuraBonusPoints(b SakuraBonus) int {
	switch b {
	case SakuraBonusAllBrights:
		return 100
	case SakuraBonusSakuraSake:
		return 30
	default:
		return 0
	}
}

// エラー値。設定検証で使う。
var (
	errSakuraSeatsRange  = errors.New("sakura: seats out of range")
	errSakuraRoundsRange = errors.New("sakura: rounds out of range")
	errSakuraDeckShort   = errors.New("sakura: the deck cannot serve this many seats")
)

// SakuraConfig はさくらの卓設定。
type SakuraConfig struct {
	Seats  int
	Rounds int
}

// DefaultSakuraConfig は既定の設定を返す。
func DefaultSakuraConfig() SakuraConfig {
	return SakuraConfig{Seats: SakuraDefaultSeats, Rounds: SakuraDefaultRounds}
}

// Validate は設定が範囲内かを検査する。
//
// **山にも札を残す。** 手札と場札を配ってなお山が残らないと、めくる手が
// 成り立たない ── さくらの手番は「手札を合わせて、山をめくって合わせる」の
// 2 段なので、山が空だと後半が消える。
func (c SakuraConfig) Validate() error {
	switch {
	case c.Seats < SakuraMinSeats || c.Seats > SakuraMaxSeats:
		return errSakuraSeatsRange
	case c.Rounds < SakuraMinRounds || c.Rounds > SakuraMaxRounds:
		return errSakuraRoundsRange
	}
	if c.Seats*SakuraHandSize+SakuraFieldSize >= SakuraDeckSize {
		return errSakuraDeckShort
	}
	return nil
}

// sakuraConfigJSON は SakuraConfig の JSON 表現。
type sakuraConfigJSON struct {
	Seats  int `json:"s"`
	Rounds int `json:"r"`
}

// MarshalJSON implements json.Marshaler.
func (c SakuraConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(sakuraConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *SakuraConfig) UnmarshalJSON(data []byte) error {
	var j sakuraConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	restored := SakuraConfig(j)
	if err := restored.Validate(); err != nil {
		return err
	}
	*c = restored
	return nil
}
