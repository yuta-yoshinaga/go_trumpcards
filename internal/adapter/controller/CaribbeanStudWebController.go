//go:build !js || !wasm || casino

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CaribbeanStudWebInput カリビアンスタッドポーカーWebインプット
type CaribbeanStudWebInput struct {
	BaseWebInput
	Amount     int  `json:"amount,omitempty"`
	JackpotBet *int `json:"jackpotBet,omitempty"`
}

// CaribbeanStudWebOutput カリビアンスタッドポーカーWebアウトプット
type CaribbeanStudWebOutput struct {
	PlayerHand      []*WebOutputCard `json:"playerHand"`
	DealerHand      []*WebOutputCard `json:"dealerHand"`
	Phase           int              `json:"phase"`
	Chips           int              `json:"chips"`
	AnteBet         int              `json:"anteBet"`
	JackpotBet      int              `json:"jackpotBet"`
	PlayBet         int              `json:"playBet"`
	Result          int              `json:"result"`
	AntePayout      int              `json:"antePayout"`
	PlayPayout      int              `json:"playPayout"`
	JackpotPayout   int              `json:"jackpotPayout"`
	TotalPayout     int              `json:"totalPayout"`
	DealerQualified bool             `json:"dealerQualified"`
	PlayerHandRank  int              `json:"playerHandRank"`
	DealerHandRank  int              `json:"dealerHandRank"`
	WebOutputBase
}

// CaribbeanStudWebController カリビアンスタッドポーカーWebコントローラークラス
type CaribbeanStudWebController = GameWebController[usecase.CaribbeanStudInteractorIF, CaribbeanStudWebInput, *CaribbeanStudWebOutput]

// NewCaribbeanStudWebController and NewCaribbeanStudWebControllerWithProvider are
// the standard and provider-backed constructors for CaribbeanStudWebController.
var NewCaribbeanStudWebController, NewCaribbeanStudWebControllerWithProvider = webControllerPair[usecase.CaribbeanStudInteractorIF, CaribbeanStudWebInput, *CaribbeanStudWebOutput](
	newCaribbeanStudDefaultOutput, caribbeanStudDispatch,
)

func newCaribbeanStudDefaultOutput(msg string) *CaribbeanStudWebOutput {
	return &CaribbeanStudWebOutput{
		PlayerHand:    make([]*WebOutputCard, 0),
		DealerHand:    make([]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func caribbeanStudDispatch(bc *baseController, w http.ResponseWriter, ci usecase.CaribbeanStudInteractorIF, param CaribbeanStudWebInput, _ func(string) *CaribbeanStudWebOutput) bool {
	switch param.Command {
	case "b", "bet":
		jackpot := deref(param.JackpotBet)
		bc.writePresenterResponse(w, ci.Bet(param.Amount, jackpot))
	case "p", "play":
		bc.writePresenterResponse(w, ci.Play())
	case "f", "fold":
		bc.writePresenterResponse(w, ci.Fold())
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, ci.Reset, ci.Hint, ci.ActionLog)
	}
	return true
}
