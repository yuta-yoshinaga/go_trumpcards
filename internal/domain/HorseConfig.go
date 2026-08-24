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
	// HorseNLHoldem は No-Limit Texas Hold'em。**Eight-Game Mix だけが回す。**
	HorseNLHoldem
	// HorsePLOmaha は Pot-Limit Omaha (ハイのみ)。**Eight-Game Mix だけが回す。**
	HorsePLOmaha
	// HorseTripleDraw は 2-7 Triple Draw Lowball。**Eight-Game Mix だけが回す。**
	HorseTripleDraw
)

// HorseDisciplineCount は H.O.R.S.E. の種目数。
//
// **種目の総数ではない。** 定義されている種目は Eight-Game Mix ぶんを含めて
// 8 つあり、どこまで回すかは**ローテーションが決める** (`HorseRotation`)。
const HorseDisciplineCount = 5

// HorseDisciplineMax は最大の種目値 (復元時の範囲検査に使う)。
const HorseDisciplineMax = HorseTripleDraw

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
	case HorseNLHoldem:
		return "nlHoldem"
	case HorsePLOmaha:
		return "plOmaha"
	case HorseTripleDraw:
		return "tripleDraw"
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
	// **8 種目のほうに頭字語は無い。** H.O.R.S.E. の 5 文字は名前そのものだが、
	// Eight-Game Mix は「8 種目」としか呼ばれない。残り 3 種目の記号は
	// 画面で種目を見分けるための短縮で、ゲーム名の由来ではない。
	case HorseNLHoldem:
		return "NLH"
	case HorsePLOmaha:
		return "PLO"
	case HorseTripleDraw:
		return "2-7"
	default:
		return "?"
	}
}

// HorseVariant はどのローテーションを回す卓かを表す。
//
// **種目の並びだけが違い、進行も精算も同じ。** H.O.R.S.E. と Eight-Game Mix は
// 同じオーケストレータで、回す種目の一覧が違うだけである。
type HorseVariant int

const (
	// HorseVariantHorse は H.O.R.S.E. (5 種目)。
	HorseVariantHorse HorseVariant = iota
	// HorseVariantEightGame は Eight-Game Mix (8 種目)。
	HorseVariantEightGame
)

// HorseVariantMax は最大のバリアント値 (復元時の範囲検査に使う)。
const HorseVariantMax = HorseVariantEightGame

// horseRotations はバリアントごとの種目の並び。
//
// **並びが競技そのもの。** H.O.R.S.E. は頭文字の順、Eight-Game Mix は
// 「H.O.R.S.E. の 5 種目 + NLH / PLO / 2-7」という定義そのままの順で回す。
var horseRotations = [][]HorseDiscipline{
	HorseVariantHorse: {
		HorseHoldem, HorseOmahaHiLo, HorseRazz, HorseStud, HorseStudHiLo,
	},
	HorseVariantEightGame: {
		HorseHoldem, HorseOmahaHiLo, HorseRazz, HorseStud, HorseStudHiLo,
		HorseNLHoldem, HorsePLOmaha, HorseTripleDraw,
	},
}

// HorseRotation はバリアントが回す種目の並びを返す。
//
// **返すのは複製。** 呼び出し側が並び替えても卓の進行順は変わらない。
func HorseRotation(v HorseVariant) []HorseDiscipline {
	if v < 0 || int(v) >= len(horseRotations) {
		return nil
	}
	return append([]HorseDiscipline(nil), horseRotations[v]...)
}

// HorseRotationIndex は種目がローテーションの何番目かを返す。無ければ -1。
func HorseRotationIndex(v HorseVariant, d HorseDiscipline) int {
	if v < 0 || int(v) >= len(horseRotations) {
		return -1
	}
	for i, cand := range horseRotations[v] {
		if cand == d {
			return i
		}
	}
	return -1
}

// 卓の範囲設定。
const (
	// 席数は **種目側が受け付ける卓サイズしか選べない**。
	//
	// Holdem 系の卓は 4 / 6 / 9 人のいずれかで、それ以外を渡すと
	// `NewOmahaPlayersForTable` などが**黙って 4 人に落とす** ── こちらが 3 人ぶんの
	// 残高を配っても 4 人が打つことになり、回収が別のプレイヤーを読んで人間の残高が
	// まったく動かない (実測: 卓 4 人 / 正本 3 席、総量が ±数十ずれた)。
	//
	// HorseSeatSizes に無い数は Validate で弾く。

	// HorseMinSeats は最小席数。
	HorseMinSeats = 4
	// HorseMaxSeats は最大席数。
	HorseMaxSeats = 9
	// HorseDefaultSeats は既定の席数。
	HorseDefaultSeats = 4

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

// HorseSeatSizes は選べる席数。**種目側の卓サイズと同じものしか選べない。**
var HorseSeatSizes = []int{4, 6, 9}

// HorseEightGameSeatSizes は Eight-Game Mix で選べる席数。
//
// **4 人卓しか作れない。** 追加の 3 種目のうち 2-7 Triple Draw は
// `DeuceToSevenConfig.CpuCount` が 1..3 ── つまり**卓は最大 4 席**で、6 人や
// 9 人を渡すと `Validate` が落ちる。落ちた種目は卓を作れず `startHand` が
// マッチをそのまま畳むので、**6 人卓を選べるようにすると 6 ハンド目で理由も
// 出さずにゲームが終わる**。
//
// 席数だけの問題でもない。トリプルドローは 1 人 5 枚配ったうえで 3 回まで
// 引き直すので、9 人卓なら配った時点で山が 7 枚しか残らず、引き直しが
// 成立しない。52 枚のデッキで 8 種目を回すなら 4 人卓が上限である。
var HorseEightGameSeatSizes = []int{4}

// HorseSeatSizesFor はバリアントで選べる席数を返す。
func HorseSeatSizesFor(v HorseVariant) []int {
	if v == HorseVariantEightGame {
		return HorseEightGameSeatSizes
	}
	return HorseSeatSizes
}

// HorseValidSeats は席数が H.O.R.S.E. の受け付ける卓サイズかを返す。
func HorseValidSeats(n int) bool {
	return HorseValidSeatsFor(HorseVariantHorse, n)
}

// HorseValidSeatsFor はバリアントごとの席数検査。
//
// **バリアントを渡さない検査は H.O.R.S.E. の答えしか返さない。** Eight-Game Mix は
// 4 人卓だけなので、6 を「有効」と読むと 6 種目目でマッチが終わる卓が作れてしまう。
func HorseValidSeatsFor(v HorseVariant, n int) bool {
	for _, size := range HorseSeatSizesFor(v) {
		if size == n {
			return true
		}
	}
	return false
}

// エラー値。設定検証で使う。
var (
	errHorseSeatsRange   = errors.New("horse: seats out of range")
	errHorseChipsRange   = errors.New("horse: initial chips out of range")
	errHorseHandsRange   = errors.New("horse: hands per discipline out of range")
	errHorseVariantRange = errors.New("horse: variant out of range")
)

// HorseConfig はミックスゲームの卓設定。
type HorseConfig struct {
	Seats              int
	InitialChips       int
	HandsPerDiscipline int
	// Variant はどのローテーションを回すか。既定は H.O.R.S.E.。
	//
	// **席数の上限がこれで変わる。** Eight-Game Mix は 4 人卓しか作れない
	// (`HorseEightGameSeatSizes`)。
	Variant HorseVariant
}

// DefaultHorseConfig は既定の設定を返す。
func DefaultHorseConfig() HorseConfig {
	return HorseConfig{
		Seats:              HorseDefaultSeats,
		InitialChips:       HorseDefaultChips,
		HandsPerDiscipline: HorseDefaultHandsPerDiscipline,
	}
}

// DefaultEightGameConfig は Eight-Game Mix の既定の設定を返す。
//
// **席数は 4 で固定。** 8 種目のうち 2-7 Triple Draw が 4 席までしか作れない
// (`HorseEightGameSeatSizes` の理由を参照)。
func DefaultEightGameConfig() HorseConfig {
	c := DefaultHorseConfig()
	c.Variant = HorseVariantEightGame
	c.Seats = HorseEightGameSeatSizes[0]
	return c
}

// Validate は設定が範囲内かを検査する。
func (c HorseConfig) Validate() error {
	switch {
	case c.Variant < 0 || c.Variant > HorseVariantMax:
		return errHorseVariantRange
	case !HorseValidSeatsFor(c.Variant, c.Seats):
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
	Seats              int          `json:"s"`
	InitialChips       int          `json:"c"`
	HandsPerDiscipline int          `json:"h"`
	Variant            HorseVariant `json:"v"`
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
