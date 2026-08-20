//go:build !js || !wasm || extra4

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// LetItRideWebInput レット・イット・ライドWebインプット
type LetItRideWebInput struct {
	BaseWebInput
	Amount int `json:"amount,omitempty"`
}

// LetItRideWebOutput レット・イット・ライドWebアウトプット
type LetItRideWebOutput struct {
	PlayerHand     []*WebOutputCard `json:"playerHand"`
	CommunityCards []*WebOutputCard `json:"communityCards"`
	Phase          int              `json:"phase"`
	Chips          int              `json:"chips"`
	BetAmount      int              `json:"betAmount"`
	Bet1Active     bool             `json:"bet1Active"`
	Bet2Active     bool             `json:"bet2Active"`
	Bet3Active     bool             `json:"bet3Active"`
	Result         int              `json:"result"`
	HandRank       int              `json:"handRank"`
	Bet1Payout     int              `json:"bet1Payout"`
	Bet2Payout     int              `json:"bet2Payout"`
	Bet3Payout     int              `json:"bet3Payout"`
	TotalPayout    int              `json:"totalPayout"`
	WebOutputBase
}

// LetItRideWebController レット・イット・ライドWebコントローラークラス
type LetItRideWebController = GameWebController[usecase.LetItRideInteractorIF, LetItRideWebInput, *LetItRideWebOutput]

// NewLetItRideWebController and NewLetItRideWebControllerWithProvider are
// the standard and provider-backed constructors for LetItRideWebController.
var NewLetItRideWebController, NewLetItRideWebControllerWithProvider = webControllerPair[usecase.LetItRideInteractorIF, LetItRideWebInput, *LetItRideWebOutput](
	newLetItRideDefaultOutput, letItRideDispatch,
)

func newLetItRideDefaultOutput(msg string) *LetItRideWebOutput {
	return &LetItRideWebOutput{
		PlayerHand:     make([]*WebOutputCard, 0),
		CommunityCards: make([]*WebOutputCard, 0),
		WebOutputBase:  WebOutputBase{Message: msg},
	}
}

func letItRideDispatch(bc *baseController, w http.ResponseWriter, li usecase.LetItRideInteractorIF, param LetItRideWebInput, _ func(string) *LetItRideWebOutput) bool {
	switch param.Command {
	case "b", "bet":
		bc.writePresenterResponse(w, li.Bet(param.Amount))
	case "p", "pull":
		bc.writePresenterResponse(w, li.Pull())
	case "l", "letitride":
		bc.writePresenterResponse(w, li.LetItRide())
	default:
		return dispatchResetAndLog(param.Command, bc, w, li.Reset, li.ActionLog)
	}
	return true
}
