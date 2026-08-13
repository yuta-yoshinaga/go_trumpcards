//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"errors"
)

// 席数・チップ・ラウンド数の範囲。
const (
	// BanLuckMinSeats は最小席数 (親 1 + 子 1)。
	BanLuckMinSeats = 2
	// BanLuckMaxSeats は最大席数。
	BanLuckMaxSeats = 6
	// BanLuckDefaultSeats は既定の席数。
	BanLuckDefaultSeats = 4

	// BanLuckMinChips は最小の初期チップ。
	BanLuckMinChips = 100
	// BanLuckMaxChips は最大の初期チップ。
	BanLuckMaxChips = 100000
	// BanLuckDefaultChips は既定の初期チップ。
	BanLuckDefaultChips = 1000

	// BanLuckMinRounds は最小ラウンド数。
	BanLuckMinRounds = 2
	// BanLuckMaxRounds は最大ラウンド数。
	BanLuckMaxRounds = 50
	// BanLuckDefaultRounds は既定のラウンド数。
	BanLuckDefaultRounds = 10

	// BanLuckMinBet は 1 席あたりの最小賭け金。
	BanLuckMinBet = 10
	// BanLuckMaxBet は 1 席あたりの最大賭け金。
	BanLuckMaxBet = 500
	// BanLuckDefaultBet は既定の賭け金。
	BanLuckDefaultBet = 50
	// BanLuckBetUnit は賭け金の刻み。**配当が整数で割り切れるように固定する。**
	BanLuckBetUnit = 10
)

// 役と手札の規則。
const (
	// BanLuckTarget は目標値。これを超えるとバスト。
	BanLuckTarget = 21
	// BanLuckBankerMustHitUnder は**親の義務ヒットの境目**。
	//
	// 親はこの値未満の間、引くのを選べない。子は自由なのに親だけが縛られる
	// ── これが「親をやると有利」を打ち消している仕掛けで、ここを子と同じ
	// 自由裁量にすると親の期待値が跳ね上がる。
	BanLuckBankerMustHitUnder = 15
	// BanLuckFiveDragonCards は Five Dragon に必要な枚数。
	BanLuckFiveDragonCards = 5
	// BanLuckMaxHandCards は 1 手札の上限枚数 (= Five Dragon で打ち止め)。
	BanLuckMaxHandCards = BanLuckFiveDragonCards
)

// 役ごとの配当倍率。**勝った側がこの倍率で受け取る。**
const (
	// BanLuckPayoutNormal は通常の勝ちの倍率。
	BanLuckPayoutNormal = 1
	// BanLuckPayoutFiveDragon は Five Dragon の倍率。
	BanLuckPayoutFiveDragon = 2
	// BanLuckPayoutBanLuck は Ban Luck (2 枚 21) の倍率。
	BanLuckPayoutBanLuck = 2
	// BanLuckPayoutBanBan は Ban Ban (A+A) の倍率。**最強の役。**
	BanLuckPayoutBanBan = 3
)

// エラー値。設定検証で使う。
var (
	errBanLuckSeatsRange  = errors.New("banluck: seats out of range")
	errBanLuckChipsRange  = errors.New("banluck: initial chips out of range")
	errBanLuckRoundsRange = errors.New("banluck: rounds out of range")
	errBanLuckBetRange    = errors.New("banluck: default bet out of range")
	errBanLuckBetUnit     = errors.New("banluck: default bet must be a multiple of the unit")
)

// BanLuckConfig はバンラックの卓設定。
type BanLuckConfig struct {
	Seats        int
	InitialChips int
	Rounds       int
	DefaultBet   int
}

// DefaultBanLuckConfig は既定の設定を返す。
func DefaultBanLuckConfig() BanLuckConfig {
	return BanLuckConfig{
		Seats:        BanLuckDefaultSeats,
		InitialChips: BanLuckDefaultChips,
		Rounds:       BanLuckDefaultRounds,
		DefaultBet:   BanLuckDefaultBet,
	}
}

// Validate は設定が範囲内かを検査する。
//
// **範囲だけでなく刻みも見る。** Ban Ban の 3 倍配当は整数で割り切れるが、
// 刻みを外した賭け金を許すとチップの総量が端数で崩れる経路が生まれる。
func (c BanLuckConfig) Validate() error {
	switch {
	case c.Seats < BanLuckMinSeats || c.Seats > BanLuckMaxSeats:
		return errBanLuckSeatsRange
	case c.InitialChips < BanLuckMinChips || c.InitialChips > BanLuckMaxChips:
		return errBanLuckChipsRange
	case c.Rounds < BanLuckMinRounds || c.Rounds > BanLuckMaxRounds:
		return errBanLuckRoundsRange
	case c.DefaultBet < BanLuckMinBet || c.DefaultBet > BanLuckMaxBet:
		return errBanLuckBetRange
	case c.DefaultBet%BanLuckBetUnit != 0:
		return errBanLuckBetUnit
	}
	return nil
}

// banLuckConfigJSON は BanLuckConfig の JSON 表現。
type banLuckConfigJSON struct {
	Seats        int `json:"s"`
	InitialChips int `json:"c"`
	Rounds       int `json:"r"`
	DefaultBet   int `json:"b"`
}

// MarshalJSON implements json.Marshaler.
func (c BanLuckConfig) MarshalJSON() ([]byte, error) {
	// **変換で書く。** 項目が 1 対 1 なので、片方に項目を足してもう片方に
	// 足し忘れたらコンパイルが落ちる ── 手書きの代入だと黙って落ちる。
	return json.Marshal(banLuckConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **復元した設定も Validate に通す。** 保存を書き換えれば席数 0 や刻み外れの
// 賭け金を送り込めるので、範囲検査を入口だけに置くと素通りする。
func (c *BanLuckConfig) UnmarshalJSON(data []byte) error {
	var j banLuckConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	restored := BanLuckConfig(j)
	if err := restored.Validate(); err != nil {
		return err
	}
	*c = restored
	return nil
}
