//go:build !js || !wasm || extra

package domain

import (
	"encoding/json"
	"errors"
)

// 卓の範囲設定。
const (
	// KingoMinSeats は最小席数 (親 1 + 子 1)。
	KingoMinSeats = 2
	// KingoMaxSeats は最大席数。
	KingoMaxSeats = 5
	// KingoDefaultSeats は既定の席数。
	KingoDefaultSeats = 4

	// KingoMinChips は最小の初期チップ。
	KingoMinChips = 100
	// KingoMaxChips は最大の初期チップ。
	KingoMaxChips = 100000
	// KingoDefaultChips は既定の初期チップ。
	KingoDefaultChips = 1000

	// KingoMinBet は最小の張り額。
	KingoMinBet = 10
	// KingoMaxBet は最大の張り額。
	KingoMaxBet = 500
	// KingoBetStep は張り額の刻み。
	KingoBetStep = 10

	// KingoMinRounds は最小ラウンド数。
	KingoMinRounds = 2
	// KingoMaxRounds は最大ラウンド数。
	KingoMaxRounds = 50
	// KingoDefaultRounds は既定のラウンド数。
	KingoDefaultRounds = 10
)

// 配札の形。**株札 40 枚 (1〜10 が 4 枚ずつ) を使う。**
//
// デッキ自体は `buildOichoKabuDeck` をそのまま使う ── 同じ株札なので、
// 枚数の定義を 2 か所に置くと必ず片方だけずれる。
const (
	// KingoHandSize は 1 席に配る枚数。
	//
	// **3 枚配る。** おいちょかぶの 2〜3 枚と違い、キンゴは「同じ数字を
	// 何枚そろえたか」で competing するので、3 枚ないと「そろえた」と
	// 「そろわなかった」の差がほとんど出ない。
	KingoHandSize = 3
	// KingoDeckSize は株札の総枚数。
	KingoDeckSize = OichoKabuDeckSize
	// KingoValueMax は株札の最大数字。
	KingoValueMax = OichoKabuValueMax
)

// KingoRank は手役の強さ。**大きいほど強い。**
type KingoRank int

const (
	// KingoRankNone は役なし (同じ数字が 1 枚もそろわない)。
	KingoRankNone KingoRank = iota
	// KingoRankPair は 2 枚そろい。
	KingoRankPair
	// KingoRankArashi は 3 枚そろい (嵐)。
	KingoRankArashi
)

// KingoRankMax は最大のランク値 (復元時の範囲検査に使う)。
const KingoRankMax = KingoRankArashi

// KingoRankName は役の識別子を返す (i18n キーの一部に使う)。
func KingoRankName(r KingoRank) string {
	switch r {
	case KingoRankPair:
		return "pair"
	case KingoRankArashi:
		return "arashi"
	default:
		return "none"
	}
}

// KingoPayout は役ごとの配当倍率を返す。
//
// **嵐は 3 倍、2 枚そろいは等倍。** 3 枚そろう確率は 2 枚そろいよりずっと
// 低いので、同じ配当にすると「そろえに行く意味」が消える。株札 40 枚から
// 3 枚を引いたときの実数は C(40,3) = 9880 通りに対し
//
//	嵐 (3 枚そろい)     10 × C(4,3) =    40 通り =  0.40%
//	2 枚そろい          10 × C(4,2) × 36 = 2160 通り = 21.86%
//	役なし                              7680 通り = 77.73%
//
// で、嵐は 2 枚そろいの 54 分の 1 しか出ない。
func KingoPayout(r KingoRank) int {
	switch r {
	case KingoRankArashi:
		return 3
	case KingoRankPair:
		return 1
	default:
		return 1
	}
}

// エラー値。設定検証で使う。
var (
	errKingoSeatsRange  = errors.New("kingo: seats out of range")
	errKingoChipsRange  = errors.New("kingo: initial chips out of range")
	errKingoBetRange    = errors.New("kingo: bet out of range")
	errKingoRoundsRange = errors.New("kingo: rounds out of range")
	errKingoDeckShort   = errors.New("kingo: the deck cannot serve this many seats")
)

// KingoConfig はキンゴの卓設定。
type KingoConfig struct {
	Seats        int
	InitialChips int
	MinBet       int
	Rounds       int
}

// DefaultKingoConfig は既定の設定を返す。
func DefaultKingoConfig() KingoConfig {
	return KingoConfig{
		Seats:        KingoDefaultSeats,
		InitialChips: KingoDefaultChips,
		MinBet:       KingoMinBet,
		Rounds:       KingoDefaultRounds,
	}
}

// Validate は設定が範囲内かを検査する。
//
// **ラウンド数は席数の倍数でなくてよいが、席数以上でなければならない。**
// 親が順に回るので、席数より少ないラウンドで終えると**親を一度もやらない席**が
// 出る ── 親は総取りの側なので、そこが回らないと有利不利が席順で固定される。
func (c KingoConfig) Validate() error {
	switch {
	case c.Seats < KingoMinSeats || c.Seats > KingoMaxSeats:
		return errKingoSeatsRange
	case c.InitialChips < KingoMinChips || c.InitialChips > KingoMaxChips:
		return errKingoChipsRange
	case c.MinBet < KingoMinBet || c.MinBet > KingoMaxBet || c.MinBet%KingoBetStep != 0:
		return errKingoBetRange
	case c.Rounds < KingoMinRounds || c.Rounds > KingoMaxRounds:
		return errKingoRoundsRange
	case c.Rounds < c.Seats:
		return errKingoRoundsRange
	}
	if c.Seats*KingoHandSize > KingoDeckSize {
		return errKingoDeckShort
	}
	return nil
}

// kingoConfigJSON は KingoConfig の JSON 表現。
type kingoConfigJSON struct {
	Seats        int `json:"s"`
	InitialChips int `json:"c"`
	MinBet       int `json:"b"`
	Rounds       int `json:"r"`
}

// MarshalJSON implements json.Marshaler.
func (c KingoConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(kingoConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **復元した設定も `Validate` に通す。** 範囲だけでなく「ラウンド数 >= 席数」も
// 見るので、書き換えた保存で親が回らない卓を作れない。
func (c *KingoConfig) UnmarshalJSON(data []byte) error {
	var j kingoConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	restored := KingoConfig(j)
	if err := restored.Validate(); err != nil {
		return err
	}
	*c = restored
	return nil
}
