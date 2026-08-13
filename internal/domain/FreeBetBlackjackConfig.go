//go:build !js || !wasm || casino

package domain

import "encoding/json"

// シューの構成。
const (
	// FreeBetDeckCount はシューに使うデッキ数 (標準 52 枚 × 6)。
	FreeBetDeckCount = 6
	// FreeBetMaxHands はスプリットで増やせる手札の上限。
	FreeBetMaxHands = 4
)

// 無料ダブルの条件。
//
// **ハードの 9 / 10 / 11 だけ。** ソフト 11 (A+10 相当) は含まない ── ソフトは
// 引き直しが利くので、無料で倍にできるとハウス側が持たない。
const (
	FreeBetDoubleMin = 9
	FreeBetDoubleMax = 11
)

// FreeBetNoSplitValue は無料スプリットできない札の値 (10 点札)。
//
// **10 のペアは割れない。** 20 は既に強い手なので、そこを無料で割れると
// 期待値がプレイヤー側に大きく傾く。
const FreeBetNoSplitValue = 10

// FreeBetDealerPushTotal はディーラーがこの数でバストしたとき引き分けになる合計。
//
// **22 だけが特別。** これが無料ダブル / 無料スプリットの対価で、ここを外すと
// ハウスエッジが消える。
const FreeBetDealerPushTotal = 22

// FreeBetDealerHitsSoft17 はディーラーがソフト 17 でヒットするか。
const FreeBetDealerHitsSoft17 = true

// ブラックジャックの配当 (3:2)。
const (
	FreeBetBlackjackPayoutNum = 3
	FreeBetBlackjackPayoutDen = 2
)

// チップとベットの範囲。
const (
	FreeBetChipsMin     = 100
	FreeBetChipsMax     = 100000
	FreeBetDefaultChips = 1000

	FreeBetAnteMin     = 10
	FreeBetAnteMax     = 500
	FreeBetDefaultAnte = 50

	// FreeBetAnteUnit はアンティの刻み。
	//
	// **3:2 の配当が割り切れることの保証。** 端数を丸めると控除率が動く。
	FreeBetAnteUnit = 10
)

// FreeBetBlackjackConfig はフリーベット・ブラックジャックのゲーム設定。
type FreeBetBlackjackConfig struct {
	// InitialChips は初期チップ。
	InitialChips int
	// DefaultAnte はラウンド開始時に選んでおくアンティ額。
	DefaultAnte int
}

// DefaultFreeBetBlackjackConfig はデフォルト設定を返す。
func DefaultFreeBetBlackjackConfig() FreeBetBlackjackConfig {
	return FreeBetBlackjackConfig{
		InitialChips: FreeBetDefaultChips,
		DefaultAnte:  FreeBetDefaultAnte,
	}
}

// Validate は設定値の妥当性を検証する。
func (c FreeBetBlackjackConfig) Validate() error {
	if err := ValidateRange("chips", c.InitialChips, FreeBetChipsMin, FreeBetChipsMax); err != nil {
		return err
	}
	return ValidateRange("ante", c.DefaultAnte, FreeBetAnteMin, FreeBetAnteMax)
}

// freeBetConfigJSON is the JSON wire format for FreeBetBlackjackConfig.
type freeBetConfigJSON struct {
	InitialChips int `json:"ic"`
	DefaultAnte  int `json:"da"`
}

// MarshalJSON implements json.Marshaler.
func (c FreeBetBlackjackConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(freeBetConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *FreeBetBlackjackConfig) UnmarshalJSON(data []byte) error {
	var j freeBetConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	*c = FreeBetBlackjackConfig(j)
	return c.Validate()
}

// FreeBetPhase は進行フェーズ。
type FreeBetPhase int

// フェーズ定数。
const (
	// FreeBetPhaseBet アンティを賭ける
	FreeBetPhaseBet FreeBetPhase = iota
	// FreeBetPhasePlay ヒット / スタンド / 無料ダブル / 無料スプリット
	FreeBetPhasePlay
	// FreeBetPhaseResult 決着後
	FreeBetPhaseResult
)

// FreeBetPhaseMax は最大のフェーズ値 (復元時の範囲検査に使う)。
const FreeBetPhaseMax = FreeBetPhaseResult

// FreeBetPhaseName はフェーズの識別子を返す (i18n キーの一部に使う)。
func FreeBetPhaseName(p FreeBetPhase) string {
	switch p {
	case FreeBetPhaseBet:
		return "bet"
	case FreeBetPhasePlay:
		return "play"
	default:
		return "result"
	}
}

// FreeBetCanFreeDouble は合計と手札の形から**無料ダブルできるか**を返す。
//
// **ハードの 9-11 で、まだ 2 枚のときだけ。**
func FreeBetCanFreeDouble(score int, soft bool, cardCount int) bool {
	if cardCount != 2 || soft {
		return false
	}
	return score >= FreeBetDoubleMin && score <= FreeBetDoubleMax
}

// FreeBetCanFreeSplit は 2 枚の値から**無料スプリットできるか**を返す。
//
// **同じ値のペアで、10 点札でないときだけ。**
func FreeBetCanFreeSplit(first, second int) bool {
	if first != second {
		return false
	}
	return freeBetCardPoints(first) != FreeBetNoSplitValue
}

// freeBetCardPoints は札の点数 (絵札は 10、A は 1) を返す。
func freeBetCardPoints(value int) int {
	if value > 10 {
		return 10
	}
	return value
}
