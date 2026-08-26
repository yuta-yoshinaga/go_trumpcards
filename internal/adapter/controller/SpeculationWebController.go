//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SpeculationWebInput スペキュレーションWebインプット
type SpeculationWebInput struct {
	BaseWebInput
	// Amount は競りで上乗せする額。`bid` でのみ読む。
	Amount *int `json:"amount,omitempty"`
}

// SpeculationWebOutCfg はスペキュレーションの卓設定
type SpeculationWebOutCfg struct {
	Players      int `json:"players"`
	InitialChips int `json:"initialChips"`
	Stake        int `json:"stake"`
	Rounds       int `json:"rounds"`
}

// SpeculationWebOutputSeat は 1 席分の公開情報。
//
// **伏せ札の中身は出さない。** 枚数だけを返す —— 中身を送ると、競りで
// いくら出すべきかが盤面から丸見えになり、賭けが賭けでなくなる。
type SpeculationWebOutputSeat struct {
	Name  string `json:"name"`
	Chips int    `json:"chips"`
	// HiddenCount はまだめくっていない伏せ札の枚数。
	HiddenCount int `json:"hiddenCount"`
	// Best はこの席が持つ最高切り札。持っていなければ null。
	Best *WebOutputCard `json:"best,omitempty"`
}

// SpeculationWebOutput スペキュレーションWebアウトプット
type SpeculationWebOutput struct {
	Phase int                         `json:"phase"`
	Seats []*SpeculationWebOutputSeat `json:"seats"`
	// TrumpSuit はこのラウンドの切り札スート。
	TrumpSuit int `json:"trumpSuit"`
	// TrumpCard は切り札を決めた札。
	TrumpCard *WebOutputCard `json:"trumpCard,omitempty"`
	Pot       int            `json:"pot"`
	// TurnSeat は次にめくる席。
	TurnSeat int `json:"turnSeat"`
	// BestSeat は最高切り札を持つ席。誰も持っていなければ -1。
	BestSeat int `json:"bestSeat"`
	// OfferFrom / OfferTo / OfferAmount は競りの申し出。無ければ -1 / -1 / 0。
	OfferFrom   int `json:"offerFrom"`
	OfferTo     int `json:"offerTo"`
	OfferAmount int `json:"offerAmount"`
	RoundNo     int `json:"roundNo"`
	// WinnerSeat は直前のラウンドの勝者席。決着前・流局なら -1。
	WinnerSeat  int  `json:"winnerSeat"`
	GameEndFlag bool `json:"gameEndFlag"`

	Config *SpeculationWebOutCfg `json:"config,omitempty"`
	WebOutputBase
}

// SpeculationWebController スペキュレーションWebコントローラークラス
type SpeculationWebController = GameWebController[usecase.SpeculationInteractorIF, SpeculationWebInput, *SpeculationWebOutput]

// NewSpeculationWebController and NewSpeculationWebControllerWithProvider
// are the standard and provider-backed constructors.
var NewSpeculationWebController, NewSpeculationWebControllerWithProvider = webControllerPair[usecase.SpeculationInteractorIF, SpeculationWebInput, *SpeculationWebOutput](
	newSpeculationDefaultOutput, speculationDispatch,
)

func newSpeculationDefaultOutput(msg string) *SpeculationWebOutput {
	return &SpeculationWebOutput{
		Seats: make([]*SpeculationWebOutputSeat, 0),
		// **「無し」は -1。** 0 は正当な席番号なので、0 を既定にすると
		// 「誰も持っていない」が「座席 0 が持っている」に化ける。
		TrumpSuit:     -1,
		BestSeat:      -1,
		OfferFrom:     -1,
		OfferTo:       -1,
		WinnerSeat:    -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func speculationDispatch(bc *baseController, w http.ResponseWriter, ci usecase.SpeculationInteractorIF, param SpeculationWebInput, newOut func(string) *SpeculationWebOutput) bool {
	switch param.Command {
	case "f", "flip":
		bc.writePresenterResponse(w, ci.Flip())
	case "a", "accept":
		bc.writePresenterResponse(w, ci.Accept())
	case "d", "decline":
		bc.writePresenterResponse(w, ci.Decline())
	case "bid":
		// **上乗せ額は必須。** 省略を 0 と読むと、断ったのか 0 で買おうと
		// したのか区別が付かない。
		if !requireParam(bc, w, newOut, param.Amount == nil, "param error: amount is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.Bid(*param.Amount))
	case "next":
		bc.writePresenterResponse(w, ci.NextRound())
	case "hint":
		bc.writePresenterResponse(w, ci.Hint())
	default:
		return dispatchResetAndLog(param.Command, bc, w, ci.Reset, ci.ActionLog)
	}
	return true
}
