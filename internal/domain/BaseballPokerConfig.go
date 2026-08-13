//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"errors"
)

// 卓の範囲設定。
const (
	// BaseballMinSeats は最小席数。
	BaseballMinSeats = 2
	// BaseballMaxSeats は最大席数。
	BaseballMaxSeats = 6
	// BaseballDefaultSeats は既定の席数。
	BaseballDefaultSeats = 4

	// BaseballMinChips は最小の初期チップ。
	BaseballMinChips = 100
	// BaseballMaxChips は最大の初期チップ。
	BaseballMaxChips = 100000
	// BaseballDefaultChips は既定の初期チップ。
	BaseballDefaultChips = 1000

	// BaseballMinAnte は最小のアンティ。
	BaseballMinAnte = 5
	// BaseballMaxAnte は最大のアンティ。
	BaseballMaxAnte = 100
	// BaseballDefaultAnte は既定のアンティ。
	BaseballDefaultAnte = 10
)

// 配札の形。**セブンカードスタッドと同じ 2 伏せ + 4 表 + 1 伏せ。**
const (
	// BaseballDownCards は最初に伏せて配る枚数。
	BaseballDownCards = 2
	// BaseballUpCards は表向きに配る枚数 (3rd〜6th ストリート)。
	BaseballUpCards = 4
	// BaseballLastDownCard は 7th ストリートで伏せて配る枚数。
	BaseballLastDownCard = 1
	// BaseballBaseCards は追加なしで配られる総枚数。
	BaseballBaseCards = BaseballDownCards + BaseballUpCards + BaseballLastDownCard
	// BaseballHandSize は役を作る枚数。
	BaseballHandSize = 5
)

// **ワイルドとイベントの札。** 野球にちなんだこのゲームの名物ルール。
const (
	// BaseballWildThree は 3 (ワイルド、かつ表で出るとポット買い増しを迫られる)。
	BaseballWildThree = 3
	// BaseballWildNine は 9 (ワイルド、9 イニングにちなむ)。
	BaseballWildNine = 9
	// BaseballBonusFour は 4 (表で出ると伏せ札を 1 枚もらえる)。
	BaseballBonusFour = 4
)

// BaseballMaxBonusCards は 1 席が受け取れるボーナス札の上限。
//
// **表向きに配るのは 4 枚なので、4 が 4 回出れば 4 枚増える。** 山の残りを
// 使い切らないよう、席数の検算はこの上限で行う。
const BaseballMaxBonusCards = BaseballUpCards

// BaseballIsWild は 3 と 9 をワイルドとして扱う。
//
// **evalWildHand にそのまま渡せる形にしておく。** 判定を各所に散らすと、
// 役の評価と画面の印が食い違って「なぜ勝ったのか分からない」盤面になる。
func BaseballIsWild(c *Card) bool {
	if c == nil {
		return false
	}
	v := c.GetValue()
	return v == BaseballWildThree || v == BaseballWildNine
}

// エラー値。設定検証で使う。
var (
	errBaseballSeatsRange = errors.New("baseballpoker: seats out of range")
	errBaseballChipsRange = errors.New("baseballpoker: initial chips out of range")
	errBaseballAnteRange  = errors.New("baseballpoker: ante out of range")
	errBaseballDeckShort  = errors.New("baseballpoker: the deck cannot serve this many seats")
)

// BaseballPokerConfig はベースボールポーカーの卓設定。
type BaseballPokerConfig struct {
	Seats        int
	InitialChips int
	Ante         int
}

// DefaultBaseballPokerConfig は既定の設定を返す。
func DefaultBaseballPokerConfig() BaseballPokerConfig {
	return BaseballPokerConfig{
		Seats:        BaseballDefaultSeats,
		InitialChips: BaseballDefaultChips,
		Ante:         BaseballDefaultAnte,
	}
}

// Validate は設定が範囲内かを検査する。
//
// **席数は山の枚数でも縛る。** 1 人 7 枚に加えて、表の 4 でボーナス札が
// 配られる。山に 4 は 4 枚しかないので追加は卓全体で最大 4 枚 ── 7 席だと
// 53 枚要って配り切れない。ここで弾かないと配札の途中で山が尽き、手札に
// nil が入る。
func (c BaseballPokerConfig) Validate() error {
	switch {
	case c.Seats < BaseballMinSeats || c.Seats > BaseballMaxSeats:
		return errBaseballSeatsRange
	case c.InitialChips < BaseballMinChips || c.InitialChips > BaseballMaxChips:
		return errBaseballChipsRange
	case c.Ante < BaseballMinAnte || c.Ante > BaseballMaxAnte:
		return errBaseballAnteRange
	}
	if c.Seats*BaseballBaseCards+BaseballMaxBonusCards > 52 {
		return errBaseballDeckShort
	}
	return nil
}

// baseballConfigJSON は BaseballPokerConfig の JSON 表現。
type baseballConfigJSON struct {
	Seats        int `json:"s"`
	InitialChips int `json:"c"`
	Ante         int `json:"a"`
}

// MarshalJSON implements json.Marshaler.
func (c BaseballPokerConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(baseballConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **範囲だけでなく山の枚数も通す。** 復元した設定をそのまま `Validate` に
// かけるので、保存を書き換えて 7 席にしても配札の前に落ちる。
func (c *BaseballPokerConfig) UnmarshalJSON(data []byte) error {
	var j baseballConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	restored := BaseballPokerConfig(j)
	if err := restored.Validate(); err != nil {
		return err
	}
	*c = restored
	return nil
}
