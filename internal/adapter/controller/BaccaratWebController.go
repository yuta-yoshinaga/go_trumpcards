//go:build !js || !wasm || casino

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"

	"net/http"
)

// BaccaratWebInput バカラWebインプット
type BaccaratWebInput struct {
	BaseWebInput
	Amount        int  `json:"amount,omitempty"`
	BetType       *int `json:"betType,omitempty"`
	PlayerPairBet *int `json:"playerPairBet,omitempty"`
	BankerPairBet *int `json:"bankerPairBet,omitempty"`
}

// BaccaratWebOutputSideBetResult サイドベット結果
type BaccaratWebOutputSideBetResult struct {
	BetType    int    `json:"betType"`
	ResultType int    `json:"resultType"`
	ResultName string `json:"resultName"`
	BetAmount  int    `json:"betAmount"`
	Payout     int    `json:"payout"`
}

// BaccaratWebOutput バカラWebアウトプット
type BaccaratWebOutput struct {
	PlayerHand      []*WebOutputCard                  `json:"playerHand"`
	BankerHand      []*WebOutputCard                  `json:"bankerHand"`
	PlayerHandValue int                               `json:"playerHandValue"`
	BankerHandValue int                               `json:"bankerHandValue"`
	Phase           int                               `json:"phase"`
	Chips           int                               `json:"chips"`
	BetAmount       int                               `json:"betAmount"`
	BetType         int                               `json:"betType"`
	Result          int                               `json:"result"`
	Payout          int                               `json:"payout"`
	History         []int                             `json:"history"`
	PlayerPairBet   int                               `json:"playerPairBet"`
	BankerPairBet   int                               `json:"bankerPairBet"`
	SideBetResults  []*BaccaratWebOutputSideBetResult `json:"sideBetResults"`
	WebOutputBase
}

// BaccaratWebController バカラWebコントローラークラス
type BaccaratWebController = GameWebController[usecase.BaccaratInteractorIF, BaccaratWebInput, *BaccaratWebOutput]

// NewBaccaratWebController and NewBaccaratWebControllerWithProvider are
// the standard and provider-backed constructors for BaccaratWebController.
var NewBaccaratWebController, NewBaccaratWebControllerWithProvider = webControllerPair[usecase.BaccaratInteractorIF, BaccaratWebInput, *BaccaratWebOutput](
	newBaccaratDefaultOutput, baccaratDispatch,
)

func newBaccaratDefaultOutput(msg string) *BaccaratWebOutput {
	return &BaccaratWebOutput{
		PlayerHand:     make([]*WebOutputCard, 0),
		BankerHand:     make([]*WebOutputCard, 0),
		History:        make([]int, 0),
		SideBetResults: make([]*BaccaratWebOutputSideBetResult, 0),
		WebOutputBase:  WebOutputBase{Message: msg},
	}
}

func baccaratDispatch(bc *baseController, w http.ResponseWriter, bi usecase.BaccaratInteractorIF, param BaccaratWebInput, _ func(string) *BaccaratWebOutput) bool {
	switch param.Command {
	case "b", "bet":
		bt := deref(param.BetType)
		ppBet := deref(param.PlayerPairBet)
		bpBet := deref(param.BankerPairBet)
		bc.writePresenterResponse(w, bi.Bet(param.Amount, bt, ppBet, bpBet))
	case "ch", "clearhistory":
		bc.writePresenterResponse(w, bi.ClearHistory())
	default:
		return dispatchResetAndLog(param.Command, bc, w, bi.Reset, bi.ActionLog)
	}
	return true
}
