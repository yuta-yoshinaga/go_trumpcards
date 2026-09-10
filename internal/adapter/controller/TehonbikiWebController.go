//go:build !js || !wasm || extra2

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TehonbikiWebInput is the wager request. Numbers are the covered 1–6 cards.
type TehonbikiWebInput struct {
	BaseWebInput
	Numbers []int                   `json:"numbers,omitempty"`
	BetType domain.TehonbikiBetType `json:"betType,omitempty"`
	Bet     *int                    `json:"bet,omitempty"`
}
type TehonbikiWebOutCfg struct {
	InitialChips int `json:"initialChips"`
	DefaultBet   int `json:"defaultBet"`
}

// TehonbikiWebOutput is the public game state.
type TehonbikiWebOutput struct {
	Phase          int                     `json:"phase"`
	ParentCard     *int                    `json:"parentCard,omitempty"`
	Numbers        []int                   `json:"numbers"`
	BetType        domain.TehonbikiBetType `json:"betType"`
	Bet            int                     `json:"bet"`
	Result         int                     `json:"result"`
	Payout         int                     `json:"payout"`
	Chips          int                     `json:"chips"`
	RoundNumber    int                     `json:"roundNumber"`
	RemainingCards int                     `json:"remainingCards"`
	GameEndFlag    bool                    `json:"gameEndFlag"`
	PayoutNum      int                     `json:"payoutNum"`
	PayoutDen      int                     `json:"payoutDen"`
	Config         *TehonbikiWebOutCfg     `json:"config,omitempty"`
	WebOutputBase
}
type TehonbikiWebController = GameWebController[usecase.TehonbikiInteractorIF, TehonbikiWebInput, *TehonbikiWebOutput]

var NewTehonbikiWebController, NewTehonbikiWebControllerWithProvider = webControllerPair[usecase.TehonbikiInteractorIF, TehonbikiWebInput, *TehonbikiWebOutput](newTehonbikiDefaultOutput, tehonbikiDispatch)

func newTehonbikiDefaultOutput(msg string) *TehonbikiWebOutput {
	return &TehonbikiWebOutput{Numbers: []int{}, WebOutputBase: WebOutputBase{Message: msg}}
}
func tehonbikiDispatch(bc *baseController, w http.ResponseWriter, ci usecase.TehonbikiInteractorIF, p TehonbikiWebInput, newOut func(string) *TehonbikiWebOutput) bool {
	switch p.Command {
	case "bet", "b":
		if !requireParam(bc, w, newOut, p.Bet == nil, "param error: bet is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.PlaceBet(p.Numbers, p.BetType, *p.Bet))
	case "next":
		bc.writePresenterResponse(w, ci.NextRound())
	case "hint":
		bc.writePresenterResponse(w, ci.Hint())
	default:
		return dispatchResetAndLog(p.Command, bc, w, ci.Reset, ci.ActionLog)
	}
	return true
}
