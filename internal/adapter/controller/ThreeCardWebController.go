//go:build !js || !wasm || casino

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ThreeCardWebInput スリーカードポーカーWebインプット
type ThreeCardWebInput struct {
	BaseWebInput
	Amount      int  `json:"amount,omitempty"`
	PairPlusBet *int `json:"pairPlusBet,omitempty"`
}

// ThreeCardWebOutput スリーカードポーカーWebアウトプット
type ThreeCardWebOutput struct {
	PlayerHand      []*WebOutputCard `json:"playerHand"`
	DealerHand      []*WebOutputCard `json:"dealerHand"`
	Phase           int              `json:"phase"`
	Chips           int              `json:"chips"`
	AnteBet         int              `json:"anteBet"`
	PairPlusBet     int              `json:"pairPlusBet"`
	PlayBet         int              `json:"playBet"`
	Result          int              `json:"result"`
	AntePayout      int              `json:"antePayout"`
	PlayPayout      int              `json:"playPayout"`
	AnteBonusPayout int              `json:"anteBonusPayout"`
	PairPlusPayout  int              `json:"pairPlusPayout"`
	TotalPayout     int              `json:"totalPayout"`
	DealerQualified bool             `json:"dealerQualified"`
	PlayerHandRank  int              `json:"playerHandRank"`
	DealerHandRank  int              `json:"dealerHandRank"`
	WebOutputBase
}

// ThreeCardWebController スリーカードポーカーWebコントローラークラス
type ThreeCardWebController = GameWebController[usecase.ThreeCardInteractorIF, ThreeCardWebInput, *ThreeCardWebOutput]

// NewThreeCardWebController and NewThreeCardWebControllerWithProvider are
// the standard and provider-backed constructors for ThreeCardWebController.
var NewThreeCardWebController, NewThreeCardWebControllerWithProvider = webControllerPair[usecase.ThreeCardInteractorIF, ThreeCardWebInput, *ThreeCardWebOutput](
	newThreeCardDefaultOutput, threeCardDispatch,
)

func newThreeCardDefaultOutput(msg string) *ThreeCardWebOutput {
	return &ThreeCardWebOutput{
		PlayerHand:    make([]*WebOutputCard, 0),
		DealerHand:    make([]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func threeCardDispatch(bc *baseController, w http.ResponseWriter, ti usecase.ThreeCardInteractorIF, param ThreeCardWebInput, _ func(string) *ThreeCardWebOutput) bool {
	switch param.Command {
	case "b", "bet":
		ppBet := deref(param.PairPlusBet)
		bc.writePresenterResponse(w, ti.Bet(param.Amount, ppBet))
	case "rb", "rebet":
		// 直前と同じ額で賭け直す。金額はサーバが覚えているので、CLI は額を
		// 送らなくてよい (#5513)。
		bc.writePresenterResponse(w, ti.Rebet())
	case "p", "play":
		bc.writePresenterResponse(w, ti.Play())
	case "f", "fold":
		bc.writePresenterResponse(w, ti.Fold())
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, ti.Reset, ti.Hint, ti.ActionLog)
	}
	return true
}
