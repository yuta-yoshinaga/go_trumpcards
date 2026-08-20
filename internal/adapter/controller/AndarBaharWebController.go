//go:build !js || !wasm || extra4

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// AndarBaharWebInput アンダーバハールWebインプット
type AndarBaharWebInput struct {
	BaseWebInput
	Amount     int `json:"amount,omitempty"`
	Target     int `json:"target,omitempty"`
	SideAmount int `json:"sideAmount,omitempty"`
	SideBand   int `json:"sideBand,omitempty"`
}

// AndarBaharWebOutput アンダーバハールWebアウトプット
type AndarBaharWebOutput struct {
	Joker *WebOutputCard `json:"joker,omitempty"`
	// AndarCards / BaharCards は **常に配列** で返す (null にしない)。
	AndarCards  []*WebOutputCard `json:"andarCards"`
	BaharCards  []*WebOutputCard `json:"baharCards"`
	FirstColumn int              `json:"firstColumn"`
	DealtCount  int              `json:"dealtCount"`
	Phase       int              `json:"phase"`
	Chips       int              `json:"chips"`
	BetAmount   int              `json:"betAmount"`
	BetTarget   int              `json:"betTarget"`
	SideAmount  int              `json:"sideAmount"`
	SideBand    int              `json:"sideBand"`
	Winner      int              `json:"winner"`
	Result      int              `json:"result"`
	Payout      int              `json:"payout"`
	// MainPayout / SidePayout は Payout の内訳。
	//
	// **サイドベットは別の賭け。** 合計だけでは、外したのがメインなのかサイドなのか
	// 画面から読めません。合計は常に両者の和です。
	MainPayout int   `json:"mainPayout"`
	SidePayout int   `json:"sidePayout"`
	History    []int `json:"history"`
	WebOutputBase
}

// AndarBaharWebController アンダーバハールWebコントローラークラス
type AndarBaharWebController = GameWebController[usecase.AndarBaharInteractorIF, AndarBaharWebInput, *AndarBaharWebOutput]

// NewAndarBaharWebController and NewAndarBaharWebControllerWithProvider are
// the standard and provider-backed constructors for AndarBaharWebController.
var NewAndarBaharWebController, NewAndarBaharWebControllerWithProvider = webControllerPair[usecase.AndarBaharInteractorIF, AndarBaharWebInput, *AndarBaharWebOutput](
	newAndarBaharDefaultOutput, andarBaharDispatch,
)

func newAndarBaharDefaultOutput(msg string) *AndarBaharWebOutput {
	return &AndarBaharWebOutput{
		AndarCards:    make([]*WebOutputCard, 0),
		BaharCards:    make([]*WebOutputCard, 0),
		History:       make([]int, 0),
		Winner:        -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func andarBaharDispatch(bc *baseController, w http.ResponseWriter, ai usecase.AndarBaharInteractorIF, param AndarBaharWebInput, _ func(string) *AndarBaharWebOutput) bool {
	switch param.Command {
	case "b", "bet":
		bc.writePresenterResponse(w,
			ai.Bet(param.Amount, param.Target, param.SideAmount, andarBaharSideBandOf(param)))
	case "clear":
		bc.writePresenterResponse(w, ai.ClearHistory())
	case "h", "hint":
		bc.writePresenterResponse(w, ai.Hint())
	default:
		return dispatchResetAndLog(param.Command, bc, w, ai.Reset, ai.ActionLog)
	}
	return true
}

// andarBaharSideBandOf はリクエストからサイドベットの帯を決める。
//
// **賭けていなければ帯は「無し」。** JSON で `sideBand` を省くと 0 になりますが、
// 0 は「1 枚目の帯」という有効な値なので、そのまま渡すと**サイドベットを
// していない普通のベットが「賭け金 0 のサイドベット」として弾かれます**。
// 金額が入っているときだけ帯を見ます。
func andarBaharSideBandOf(param AndarBaharWebInput) int {
	if param.SideAmount <= 0 {
		return domain.AndarBaharSideNone
	}
	return param.SideBand
}
