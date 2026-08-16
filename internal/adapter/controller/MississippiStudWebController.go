//go:build !js || !wasm || casino

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// MississippiStudWebInput ミシシッピ・スタッドWebインプット
type MississippiStudWebInput struct {
	BaseWebInput
	Amount     int `json:"amount,omitempty"`
	Multiplier int `json:"multiplier,omitempty"`
}

// MississippiStudWebOutput ミシシッピ・スタッドWebアウトプット
type MississippiStudWebOutput struct {
	PlayerHand        []*WebOutputCard `json:"playerHand"`
	CommunityCards    []*WebOutputCard `json:"communityCards"`
	CommunityRevealed []bool           `json:"communityRevealed"`
	Phase             int              `json:"phase"`
	Chips             int              `json:"chips"`
	AnteAmount        int              `json:"anteAmount"`
	StreetMultipliers []int            `json:"streetMultipliers"`
	Folded            bool             `json:"folded"`
	TotalBet          int              `json:"totalBet"`
	Result            int              `json:"result"`
	HandRank          int              `json:"handRank"`
	PayoutMultiplier  int              `json:"payoutMultiplier"`
	AntePayout        int              `json:"antePayout"`
	StreetPayouts     []int            `json:"streetPayouts"`
	TotalPayout       int              `json:"totalPayout"`
	WebOutputBase
}

// MississippiStudWebController ミシシッピ・スタッドWebコントローラー
type MississippiStudWebController = GameWebController[usecase.MississippiStudInteractorIF, MississippiStudWebInput, *MississippiStudWebOutput]

// NewMississippiStudWebController and NewMississippiStudWebControllerWithProvider are
// the standard and provider-backed constructors for MississippiStudWebController.
var NewMississippiStudWebController, NewMississippiStudWebControllerWithProvider = webControllerPair[usecase.MississippiStudInteractorIF, MississippiStudWebInput, *MississippiStudWebOutput](
	newMississippiStudDefaultOutput, mississippiStudDispatch,
)

func newMississippiStudDefaultOutput(msg string) *MississippiStudWebOutput {
	return &MississippiStudWebOutput{
		PlayerHand:        make([]*WebOutputCard, 0),
		CommunityCards:    make([]*WebOutputCard, 0),
		CommunityRevealed: make([]bool, 0),
		StreetMultipliers: make([]int, 0),
		StreetPayouts:     make([]int, 0),
		WebOutputBase:     WebOutputBase{Message: msg},
	}
}

func mississippiStudDispatch(bc *baseController, w http.ResponseWriter, mi usecase.MississippiStudInteractorIF, param MississippiStudWebInput, _ func(string) *MississippiStudWebOutput) bool {
	switch param.Command {
	case "b", "bet":
		bc.writePresenterResponse(w, mi.Bet(param.Amount))
	case "p", "play":
		bc.writePresenterResponse(w, mi.Play(param.Multiplier))
	case "f", "fold":
		bc.writePresenterResponse(w, mi.Fold())
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, mi.Reset, mi.Hint, mi.ActionLog)
	}
	return true
}
