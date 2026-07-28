//go:build !js || !wasm || extra2

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// FaroWebInput はファロのWebインプット。
type FaroWebInput struct {
	BaseWebInput
	Rank   int   `json:"rank,omitempty"`
	Amount int   `json:"amount,omitempty"`
	Copper bool  `json:"copper,omitempty"`
	Order  []int `json:"order,omitempty"`
}

// FaroWebBet はレイアウト上の1ランクのベットを表す。
type FaroWebBet struct {
	Rank   int  `json:"rank"`
	Amount int  `json:"amount"`
	Copper bool `json:"copper"`
}

// FaroWebOutput はファロのWebアウトプット。
type FaroWebOutput struct {
	Phase       int              `json:"phase"`
	Chips       int              `json:"chips"`
	Bets        []*FaroWebBet    `json:"bets"`
	Soda        *WebOutputCard   `json:"soda,omitempty"`
	LosingCard  *WebOutputCard   `json:"losingCard,omitempty"`
	WinningCard *WebOutputCard   `json:"winningCard,omitempty"`
	Split       bool             `json:"split"`
	TurnsPlayed int              `json:"turnsPlayed"`
	TurnsTotal  int              `json:"turnsTotal"`
	Remaining   int              `json:"remaining"`
	CallCards   []*WebOutputCard `json:"callCards"`
	CallOrder   []int            `json:"callOrder"`
	CallWon     bool             `json:"callWon"`
	TotalPayout int              `json:"totalPayout"`
	GameEndFlag bool             `json:"gameEndFlag"`
	WebOutputBase
}

// FaroWebController はファロのWebコントローラー。
type FaroWebController = GameWebController[usecase.FaroInteractorIF, FaroWebInput, *FaroWebOutput]

// NewFaroWebController and NewFaroWebControllerWithProvider are the standard and
// provider-backed constructors for FaroWebController.
var NewFaroWebController, NewFaroWebControllerWithProvider = webControllerPair[usecase.FaroInteractorIF, FaroWebInput, *FaroWebOutput](
	newFaroDefaultOutput, faroDispatch,
)

func newFaroDefaultOutput(msg string) *FaroWebOutput {
	return &FaroWebOutput{
		Bets:          make([]*FaroWebBet, 0),
		CallCards:     make([]*WebOutputCard, 0),
		CallOrder:     make([]int, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func faroDispatch(bc *baseController, w http.ResponseWriter, fi usecase.FaroInteractorIF, param FaroWebInput, _ func(string) *FaroWebOutput) bool {
	switch param.Command {
	case "b", "bet":
		bc.writePresenterResponse(w, fi.PlaceBet(param.Rank, param.Amount, param.Copper))
	case "cb", "clearBet":
		bc.writePresenterResponse(w, fi.ClearBet(param.Rank))
	case "ca", "clearAll":
		bc.writePresenterResponse(w, fi.ClearAll())
	case "d", "deal":
		bc.writePresenterResponse(w, fi.DealTurn())
	case "call":
		bc.writePresenterResponse(w, fi.Call(param.Order))
	case "n", "next":
		bc.writePresenterResponse(w, fi.NextRound())
	default:
		return dispatchResetAndLog(param.Command, bc, w, fi.Reset, fi.ActionLog)
	}
	return true
}
