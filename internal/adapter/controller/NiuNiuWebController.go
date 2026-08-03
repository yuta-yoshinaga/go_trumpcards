//go:build !js || !wasm || extra3

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NiuNiuWebInput 闘牛 Web インプット
type NiuNiuWebInput struct {
	BaseWebInput
	// Amount is the stake for "bet".
	Amount *int `json:"amount,omitempty"`
}

// NiuNiuWebOutputHand 1 つの手の出力
type NiuNiuWebOutputHand struct {
	Cards []*WebOutputCard `json:"cards"`
	Bet   int              `json:"bet"`
	// ComboIdx は牛を作った 3 枚の位置。無牛なら空。画面でその 3 枚を目立たせる
	// ためのもので、クライアントに組み合わせ探索をやり直させないために送る。
	ComboIdx []int `json:"comboIdx"`
	// Rank は 0=無牛 / 1..9=牛1..牛9 / 10=牛牛。
	Rank      int    `json:"rank"`
	RankLabel string `json:"rankLabel"`
	// Multiplier はこの格で勝ったときの配当倍率。
	Multiplier int `json:"multiplier"`
	Payout     int `json:"payout"`
	// Hidden が真のとき、Cards の要素は null で格も伏せられる。枚数だけが残る。
	Hidden bool `json:"hidden"`
}

// NiuNiuWebOutputSeat 1 席の出力
type NiuNiuWebOutputSeat struct {
	Name  string               `json:"name"`
	IsCPU bool                 `json:"isCpu"`
	Hand  *NiuNiuWebOutputHand `json:"hand,omitempty"`
}

// NiuNiuWebOutput 闘牛 Web アウトプット
type NiuNiuWebOutput struct {
	Seats      []*NiuNiuWebOutputSeat `json:"seats"`
	BankerHand *NiuNiuWebOutputHand   `json:"bankerHand,omitempty"`
	BankerIdx  int                    `json:"bankerIdx"`
	Chips      int                    `json:"chips"`
	// MaxMultiplier は最大の配当倍率。**賭けられる上限は残高そのものではなく
	// 残高÷これ**になる。親が牛牛なら賭け金の 3 倍を取られるので、残高ちょうどを
	// 賭けると払えない。クライアントにこの割り算を再発明させないために送る。
	MaxMultiplier int    `json:"maxMultiplier"`
	LastResult    string `json:"lastResult"`
	Phase         int    `json:"phase"`
	WebOutputBase
}

// NiuNiuWebController 闘牛 Web コントローラークラス
type NiuNiuWebController = GameWebController[usecase.NiuNiuInteractorIF, NiuNiuWebInput, *NiuNiuWebOutput]

// NewNiuNiuWebController and NewNiuNiuWebControllerWithProvider are the
// standard and provider-backed constructors for NiuNiuWebController.
var NewNiuNiuWebController, NewNiuNiuWebControllerWithProvider = webControllerPair[usecase.NiuNiuInteractorIF, NiuNiuWebInput, *NiuNiuWebOutput](
	newNiuNiuDefaultOutput, niuNiuDispatch,
)

func newNiuNiuDefaultOutput(msg string) *NiuNiuWebOutput {
	return &NiuNiuWebOutput{
		Seats:         make([]*NiuNiuWebOutputSeat, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func niuNiuDispatch(bc *baseController, w http.ResponseWriter, ni usecase.NiuNiuInteractorIF, param NiuNiuWebInput, newDefault func(string) *NiuNiuWebOutput) bool {
	switch param.Command {
	case "b", "bet":
		if !requireParam(bc, w, newDefault, param.Amount == nil, "param error: amount is required.") {
			return true
		}
		bc.writePresenterResponse(w, ni.Bet(*param.Amount))
	case "log", "l":
		bc.writePresenterResponse(w, ni.ActionLog())
	case "r", "reset":
		bc.writePresenterResponse(w, ni.Reset())
	default:
		return false
	}
	return true
}
