//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"errors"
)

// 卓の範囲設定。
const (
	// MonteBankMinChips は最小の初期チップ。
	MonteBankMinChips = 100
	// MonteBankMaxChips は最大の初期チップ。
	MonteBankMaxChips = 100000
	// MonteBankDefaultChips は既定の初期チップ。
	MonteBankDefaultChips = 1000

	// MonteBankMinBet は最小の賭け金。
	MonteBankMinBet = 10
	// MonteBankMaxBet は最大の賭け金。
	MonteBankMaxBet = 500
	// MonteBankDefaultBet は既定の賭け金。
	MonteBankDefaultBet = 50
	// MonteBankBetUnit は賭け金の刻み。**3:1 が整数で割り切れるように固定する。**
	MonteBankBetUnit = 10
)

// 場札と山の規則。
const (
	// MonteBankDeckSize はラテン 40 枚デッキの枚数 (8・9・10 を抜いた形)。
	MonteBankDeckSize = 40
	// MonteBankSuitSize は 1 スートあたりの枚数。
	MonteBankSuitSize = 10
	// MonteBankLayoutSize は場札の枚数 (山の上下から 2 枚ずつ)。
	MonteBankLayoutSize = 4
	// MonteBankLayoutHalf は上または下から取る枚数。
	MonteBankLayoutHalf = MonteBankLayoutSize / 2
	// MonteBankCardsPerRound は 1 ラウンドに要る枚数 (場札 4 + ゲート 1)。
	MonteBankCardsPerRound = MonteBankLayoutSize + 1
)

// MonteBankPayout は的中したときの配当倍率 (3:1)。
//
// **この値は写したものではなく、数えた結果である。**
//
// 場札 4 枚が見えている時点で、次にめくるゲートは残り 36 枚の一様な 1 枚。
// 賭けた札のスートが場札に何枚出ているかで勝率が決まる:
//
//	1 枚だけ  → 9/36 = 25.00%   → 3:1 でちょうど互角 (控除率 0%)
//	2 枚      → 8/36 = 22.22%   → 控除率 11.11%
//	3 枚      → 7/36 = 19.44%   → 控除率 22.22%
//	4 枚      → 6/36 = 16.67%   → 控除率 33.33%
//
// **つまり胴元の取り分は、すべてプレイヤーの選択から出る。** 場札に 1 枚しか
// 出ていないスートを選び続ければ互角で、重複したスートを選ぶたびに損をする。
// 選択の無いゲームにしないために、この配当でなければならない。
//
// 何も見ずに賭けた場合の総合勝率は 9/39 = 3/13 = 23.08% で、3:1 なら控除率
// 7.69%。issue が併記していた 4:1 は**プレイヤー側に 15.4% の優位**が付くので
// 使えない。
const MonteBankPayout = 3

// エラー値。設定検証で使う。
var (
	errMonteBankChipsRange  = errors.New("montebank: initial chips out of range")
	errMonteBankBetRangeCfg = errors.New("montebank: default bet out of range")
	errMonteBankBetUnitCfg  = errors.New("montebank: default bet must be a multiple of the unit")
)

// MonteBankConfig はモンテバンクの卓設定。
type MonteBankConfig struct {
	InitialChips int
	DefaultBet   int
}

// DefaultMonteBankConfig は既定の設定を返す。
func DefaultMonteBankConfig() MonteBankConfig {
	return MonteBankConfig{
		InitialChips: MonteBankDefaultChips,
		DefaultBet:   MonteBankDefaultBet,
	}
}

// Validate は設定が範囲内かを検査する。
func (c MonteBankConfig) Validate() error {
	switch {
	case c.InitialChips < MonteBankMinChips || c.InitialChips > MonteBankMaxChips:
		return errMonteBankChipsRange
	case c.DefaultBet < MonteBankMinBet || c.DefaultBet > MonteBankMaxBet:
		return errMonteBankBetRangeCfg
	case c.DefaultBet%MonteBankBetUnit != 0:
		return errMonteBankBetUnitCfg
	}
	return nil
}

// monteBankConfigJSON は MonteBankConfig の JSON 表現。
type monteBankConfigJSON struct {
	InitialChips int `json:"c"`
	DefaultBet   int `json:"b"`
}

// MarshalJSON implements json.Marshaler.
func (c MonteBankConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(monteBankConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **復元した設定も Validate に通す。** 保存を書き換えれば刻み外れの賭け金や
// 範囲外のチップを送り込めるので、検査を入口だけに置くと素通りする。
func (c *MonteBankConfig) UnmarshalJSON(data []byte) error {
	var j monteBankConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	restored := MonteBankConfig(j)
	if err := restored.Validate(); err != nil {
		return err
	}
	*c = restored
	return nil
}
