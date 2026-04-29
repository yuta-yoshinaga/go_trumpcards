package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TexasHoldemBonusWebInput テキサスホールデムボーナスポーカーWebインプット
type TexasHoldemBonusWebInput struct {
	BaseWebInput
	Amount   int  `json:"amount,omitempty"`
	BonusBet *int `json:"bonusBet,omitempty"`
}

// TexasHoldemBonusWebOutput テキサスホールデムボーナスポーカーWebアウトプット
type TexasHoldemBonusWebOutput struct {
	PlayerHand     []*WebOutputCard `json:"playerHand"`
	DealerHand     []*WebOutputCard `json:"dealerHand"`
	Community      []*WebOutputCard `json:"community"`
	Phase          int              `json:"phase"`
	Chips          int              `json:"chips"`
	AnteBet        int              `json:"anteBet"`
	BonusBet       int              `json:"bonusBet"`
	FlopBet        int              `json:"flopBet"`
	TurnBet        int              `json:"turnBet"`
	RiverBet       int              `json:"riverBet"`
	TotalPlayBet   int              `json:"totalPlayBet"`
	Result         int              `json:"result"`
	AntePayout     int              `json:"antePayout"`
	PlayPayout     int              `json:"playPayout"`
	BonusPayout    int              `json:"bonusPayout"`
	TotalPayout    int              `json:"totalPayout"`
	PlayerHandRank int              `json:"playerHandRank"`
	DealerHandRank int              `json:"dealerHandRank"`
	WebOutputBase
}

// TexasHoldemBonusWebController テキサスホールデムボーナスポーカーWebコントローラークラス
type TexasHoldemBonusWebController = GameWebController[usecase.TexasHoldemBonusInteractorIF, TexasHoldemBonusWebInput, *TexasHoldemBonusWebOutput]

// NewTexasHoldemBonusWebController and NewTexasHoldemBonusWebControllerWithProvider are
// the standard and provider-backed constructors for TexasHoldemBonusWebController.
var NewTexasHoldemBonusWebController, NewTexasHoldemBonusWebControllerWithProvider = webControllerPair[usecase.TexasHoldemBonusInteractorIF, TexasHoldemBonusWebInput, *TexasHoldemBonusWebOutput](
	newTexasHoldemBonusDefaultOutput, texasHoldemBonusDispatch,
)

func newTexasHoldemBonusDefaultOutput(msg string) *TexasHoldemBonusWebOutput {
	return &TexasHoldemBonusWebOutput{
		PlayerHand:    make([]*WebOutputCard, 0),
		DealerHand:    make([]*WebOutputCard, 0),
		Community:     make([]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func texasHoldemBonusDispatch(bc *baseController, w http.ResponseWriter, ti usecase.TexasHoldemBonusInteractorIF, param TexasHoldemBonusWebInput, _ func(string) *TexasHoldemBonusWebOutput) bool {
	switch param.Command {
	case "b", "bet":
		bonus := deref(param.BonusBet)
		bc.writePresenterResponse(w, ti.Bet(param.Amount, bonus))
	case "p", "play":
		bc.writePresenterResponse(w, ti.Play())
	case "f", "fold":
		bc.writePresenterResponse(w, ti.Fold())
	case "c", "check":
		bc.writePresenterResponse(w, ti.Check())
	case "ra", "raise":
		bc.writePresenterResponse(w, ti.Raise())
	default:
		return dispatchResetAndLog(param.Command, bc, w, ti.Reset, ti.ActionLog)
	}
	return true
}
