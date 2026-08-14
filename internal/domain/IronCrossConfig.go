//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"errors"
)

// 卓の範囲設定。
const (
	// IronCrossMinSeats は最小席数。
	IronCrossMinSeats = 3
	// IronCrossMaxSeats は最大席数。
	IronCrossMaxSeats = 7
	// IronCrossDefaultSeats は既定の席数。
	IronCrossDefaultSeats = 4

	// IronCrossMinChips は最小の初期チップ。
	IronCrossMinChips = 100
	// IronCrossMaxChips は最大の初期チップ。
	IronCrossMaxChips = 100000
	// IronCrossDefaultChips は既定の初期チップ。
	IronCrossDefaultChips = 1000

	// IronCrossMinAnte は最小のアンティ。
	IronCrossMinAnte = 5
	// IronCrossMaxAnte は最大のアンティ。
	IronCrossMaxAnte = 100
	// IronCrossDefaultAnte は既定のアンティ。
	IronCrossDefaultAnte = 10
)

// 十字の配置。**中央は縦と横で共有する。**
const (
	// IronCrossHoleCards は各プレイヤーの手札枚数。
	IronCrossHoleCards = 4
	// IronCrossCommunityCards は十字に並べるコミュニティの枚数。
	IronCrossCommunityCards = 5
	// IronCrossLineCards は 1 本の列の枚数 (端 + 中央 + 端)。
	IronCrossLineCards = 3
	// IronCrossHandSize は役を作る枚数。
	IronCrossHandSize = 5
	// IronCrossPoolSize は役を選ぶ母数 (手札 4 + 選んだ列 3)。
	//
	// **選ばなかった列は使えない。** ここが Holdem との違いで、7 枚から
	// 最良の 5 枚を選ぶこと自体は同じでも、**7 枚の中身が人によって変わる**。
	IronCrossPoolSize = IronCrossHoleCards + IronCrossLineCards
)

// 十字の位置。**並びが配置そのもの。**
//
//	    [1]
//	[3] [0] [4]
//	    [2]
const (
	// IronCrossCenter は中央 (縦にも横にも入る)。
	IronCrossCenter = 0
	// IronCrossTop は上。
	IronCrossTop = 1
	// IronCrossBottom は下。
	IronCrossBottom = 2
	// IronCrossLeft は左。
	IronCrossLeft = 3
	// IronCrossRight は右。
	IronCrossRight = 4
)

// IronCrossLine はプレイヤーが選ぶ列。
type IronCrossLine int

const (
	// IronCrossLineNone はまだ選んでいない。
	IronCrossLineNone IronCrossLine = iota
	// IronCrossLineVertical は縦 (上 - 中央 - 下)。
	IronCrossLineVertical
	// IronCrossLineHorizontal は横 (左 - 中央 - 右)。
	IronCrossLineHorizontal
)

// IronCrossLineMax は最大の列値 (復元時の範囲検査に使う)。
const IronCrossLineMax = IronCrossLineHorizontal

// IronCrossLineName は列の識別子を返す (i18n キーの一部に使う)。
func IronCrossLineName(l IronCrossLine) string {
	switch l {
	case IronCrossLineVertical:
		return "vertical"
	case IronCrossLineHorizontal:
		return "horizontal"
	default:
		return "none"
	}
}

// IronCrossLineIndexes は列に含まれる十字の位置を返す。
//
// **中央は必ず入る。** 縦と横は中央 1 枚だけを共有していて、そこが
// 「どちらを選んでも半分は同じ」という選択の妙になっている。
func IronCrossLineIndexes(l IronCrossLine) []int {
	switch l {
	case IronCrossLineVertical:
		return []int{IronCrossTop, IronCrossCenter, IronCrossBottom}
	case IronCrossLineHorizontal:
		return []int{IronCrossLeft, IronCrossCenter, IronCrossRight}
	default:
		return nil
	}
}

// エラー値。設定検証で使う。
var (
	errIronCrossSeatsRange = errors.New("ironcross: seats out of range")
	errIronCrossChipsRange = errors.New("ironcross: initial chips out of range")
	errIronCrossAnteRange  = errors.New("ironcross: ante out of range")
	errIronCrossDeckShort  = errors.New("ironcross: the deck cannot serve this many seats")
)

// IronCrossConfig はアイアンクロスの卓設定。
type IronCrossConfig struct {
	Seats        int
	InitialChips int
	Ante         int
}

// DefaultIronCrossConfig は既定の設定を返す。
func DefaultIronCrossConfig() IronCrossConfig {
	return IronCrossConfig{
		Seats:        IronCrossDefaultSeats,
		InitialChips: IronCrossDefaultChips,
		Ante:         IronCrossDefaultAnte,
	}
}

// Validate は設定が範囲内かを検査する。
//
// **席数は山の枚数でも縛る。** 1 人 4 枚 + 十字 5 枚なので、7 人なら 33 枚要る。
func (c IronCrossConfig) Validate() error {
	switch {
	case c.Seats < IronCrossMinSeats || c.Seats > IronCrossMaxSeats:
		return errIronCrossSeatsRange
	case c.InitialChips < IronCrossMinChips || c.InitialChips > IronCrossMaxChips:
		return errIronCrossChipsRange
	case c.Ante < IronCrossMinAnte || c.Ante > IronCrossMaxAnte:
		return errIronCrossAnteRange
	}
	if c.Seats*IronCrossHoleCards+IronCrossCommunityCards > 52 {
		return errIronCrossDeckShort
	}
	return nil
}

// ironCrossConfigJSON は IronCrossConfig の JSON 表現。
type ironCrossConfigJSON struct {
	Seats        int `json:"s"`
	InitialChips int `json:"c"`
	Ante         int `json:"a"`
}

// MarshalJSON implements json.Marshaler.
func (c IronCrossConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(ironCrossConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *IronCrossConfig) UnmarshalJSON(data []byte) error {
	var j ironCrossConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	restored := IronCrossConfig(j)
	if err := restored.Validate(); err != nil {
		return err
	}
	*c = restored
	return nil
}
