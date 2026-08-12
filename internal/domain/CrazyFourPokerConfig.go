//go:build !js || !wasm || casino

package domain

import "encoding/json"

// 卓の構成。
const (
	// CrazyFourPokerHandSize は配る枚数 (プレイヤー・ディーラーとも 5 枚)。
	CrazyFourPokerHandSize = 5
	// CrazyFourPokerBestSize は勝負に使う枚数。**5 枚から最良の 4 枚**を選ぶ。
	CrazyFourPokerBestSize = 4
)

// プレイベットの倍率。
const (
	// CrazyFourPokerPlayMin はプレイベットの最低倍率 (アンティと同額)。
	CrazyFourPokerPlayMin = 1
	// CrazyFourPokerPlayNormalMax は通常の上限倍率。**同額しか置けない。**
	CrazyFourPokerPlayNormalMax = 1
	// CrazyFourPokerPlayAcesMax は**エースのペア以上**のときだけ許される上限倍率。
	//
	// この 3 倍がこのゲームの名前の由来で、独自ルールの本体。手が強いと分かった
	// 時点で乗せられる。issue は「2 倍は常に置ける」と読める書き方をしているが、
	// 素の規則では**倍率を動かせること自体が AA 以上の特典**なので、2 倍も
	// AA 以上でのみ許す。
	CrazyFourPokerPlayAcesMax = 3
)

// CrazyFourPokerMaxPlayMultiplier は手役に応じたプレイベットの上限倍率を返す。
//
// **規則をここ 1 か所に閉じ込める。** 上限をコントローラや画面で再計算すると必ずずれる。
func CrazyFourPokerMaxPlayMultiplier(hasAcesOrBetter bool) int {
	if hasAcesOrBetter {
		return CrazyFourPokerPlayAcesMax
	}
	return CrazyFourPokerPlayNormalMax
}

// ディーラーのクオリファイ条件。
//
// **キング以上のハイカードで成立する。** 成立しなければプレイベットはプッシュ、
// アンティだけが 1:1 で払われる。
const CrazyFourPokerDealerQualifyValue = 13

// チップとベットの範囲。
const (
	CrazyFourPokerChipsMin     = 100
	CrazyFourPokerChipsMax     = 100000
	CrazyFourPokerDefaultChips = 1000

	CrazyFourPokerAnteMin     = 10
	CrazyFourPokerAnteMax     = 500
	CrazyFourPokerDefaultAnte = 50
)

// CrazyFourPokerPayoutScale は配当倍率の分母。
//
// **1.5:1 を浮動小数で持たない。** 倍率は 10 倍した整数で持ち、アンティは 10 の
// 倍数に制限しているので端数が出ない (AndarBahar と同じ扱い)。
const CrazyFourPokerPayoutScale = 10

// CrazyFourPokerAnteUnit はアンティの刻み。配当が割り切れる保証になる。
const CrazyFourPokerAnteUnit = 10

// crazyFourPokerQueensUpPayouts は Queens Up サイドベットの配当 (X:1)。
//
// **出典の表を写さず、実際の分布から決めた。** 5 枚から作れる最良 4 枚役を
// C(52,5)=2,598,960 通り全列挙すると:
//
//	Four of a Kind      624  0.02401%
//	Straight Flush     2072  0.07972%
//	Three of a Kind   58656  2.25690%
//	Flush            114616  4.41007%
//	Straight         101808  3.91726%
//	Two Pair         123552  4.75390%
//	Pair of QQ+      242916  9.34666%
//
// **4 枚役ではフォーカードのほうがストレートフラッシュより希少**なので (624 対
// 2072)、配当も 4K > SF の順にしてある。5 枚役の感覚で SF を上に置くと、希少度と
// 配当の向きが食い違う。
//
// この表でのハウスエッジは**ちょうど 4.5203%** (期待値 -117,480 / 2,598,960)。
var crazyFourPokerQueensUpPayouts = map[int]int{
	FourCardHandFourOfAKind:   50,
	FourCardHandStraightFlush: 40,
	FourCardHandThreeOfAKind:  8,
	FourCardHandFlush:         4,
	FourCardHandStraight:      3,
	FourCardHandTwoPair:       2,
	FourCardHandPair:          1, // クイーンのペア以上のみ
}

// CrazyFourPokerQueensUpMinPair は Queens Up が成立する最低のペア (クイーン)。
const CrazyFourPokerQueensUpMinPair = 12

// crazyFourPokerSuperBonusPayouts は Super Bonus の配当 (10 倍した整数)。
//
// **こちらも希少度と同じ向きに並べてある。** 4 枚のエースだけを別枠にして、
// 残りのフォーカードをストレートフラッシュより上に置く。
var crazyFourPokerSuperBonusPayouts = map[int]int{
	FourCardHandFourOfAKind:   300, // 30:1 (エース 4 枚は別枠で 200:1)
	FourCardHandStraightFlush: 200, // 20:1
	FourCardHandThreeOfAKind:  20,  // 2:1
	FourCardHandFlush:         15,  // 1.5:1
	FourCardHandStraight:      10,  // 1:1
	FourCardHandTwoPair:       10,  // 1:1
	FourCardHandPair:          10,  // 1:1 (エースのペアのみ)
}

// CrazyFourPokerFourAcesPayout は 4 枚のエースへの Super Bonus 配当 (10 倍)。
const CrazyFourPokerFourAcesPayout = 2000 // 200:1

// CrazyFourPokerSuperBonusMinPair は Super Bonus が成立する最低のペア (エース)。
const CrazyFourPokerSuperBonusMinPair = 1

// CrazyFourPokerConfig はクレイジー 4 ポーカーのゲーム設定。
type CrazyFourPokerConfig struct {
	// InitialChips は初期チップ。
	InitialChips int
	// DefaultAnte はラウンド開始時に選んでおくアンティ額。
	DefaultAnte int
}

// DefaultCrazyFourPokerConfig はデフォルト設定を返す。
func DefaultCrazyFourPokerConfig() CrazyFourPokerConfig {
	return CrazyFourPokerConfig{
		InitialChips: CrazyFourPokerDefaultChips,
		DefaultAnte:  CrazyFourPokerDefaultAnte,
	}
}

// Validate は設定値の妥当性を検証する。
func (c CrazyFourPokerConfig) Validate() error {
	if err := ValidateRange("chips", c.InitialChips,
		CrazyFourPokerChipsMin, CrazyFourPokerChipsMax); err != nil {
		return err
	}
	return ValidateRange("ante", c.DefaultAnte, CrazyFourPokerAnteMin, CrazyFourPokerAnteMax)
}

// crazyFourPokerConfigJSON is the JSON wire format for CrazyFourPokerConfig.
type crazyFourPokerConfigJSON struct {
	InitialChips int `json:"ic"`
	DefaultAnte  int `json:"da"`
}

// MarshalJSON implements json.Marshaler.
func (c CrazyFourPokerConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(crazyFourPokerConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *CrazyFourPokerConfig) UnmarshalJSON(data []byte) error {
	var j crazyFourPokerConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	*c = CrazyFourPokerConfig(j)
	return c.Validate()
}

// CrazyFourPokerPhase は進行フェーズ。
type CrazyFourPokerPhase int

// フェーズ定数。
const (
	// CrazyFourPokerPhaseBet アンティと任意の Queens Up を賭ける
	CrazyFourPokerPhaseBet CrazyFourPokerPhase = iota
	// CrazyFourPokerPhaseDecide 手札を見てプレイベットかフォールドかを決める
	CrazyFourPokerPhaseDecide
	// CrazyFourPokerPhaseResult 決着後
	CrazyFourPokerPhaseResult
)

// CrazyFourPokerPhaseMax は最大のフェーズ値 (復元時の範囲検査に使う)。
const CrazyFourPokerPhaseMax = CrazyFourPokerPhaseResult

// CrazyFourPokerPhaseName はフェーズの識別子を返す (i18n キーの一部に使う)。
func CrazyFourPokerPhaseName(p CrazyFourPokerPhase) string {
	switch p {
	case CrazyFourPokerPhaseBet:
		return "bet"
	case CrazyFourPokerPhaseDecide:
		return "decide"
	default:
		return "result"
	}
}
