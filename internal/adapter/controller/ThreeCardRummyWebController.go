//go:build !js || !wasm || casino

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ThreeCardRummyWebInput スリーカード・ラミーWebインプット
type ThreeCardRummyWebInput struct {
	BaseWebInput
	Amount      int  `json:"amount,omitempty"`
	LowBonusBet *int `json:"lowBonusBet,omitempty"`
}

// ThreeCardRummyWebOutput スリーカード・ラミーWebアウトプット
type ThreeCardRummyWebOutput struct {
	PlayerHand      []*WebOutputCard `json:"playerHand"`
	DealerHand      []*WebOutputCard `json:"dealerHand"`
	Phase           int              `json:"phase"`
	Chips           int              `json:"chips"`
	AnteBet         int              `json:"anteBet"`
	LowBonusBet     int              `json:"lowBonusBet"`
	PlayBet         int              `json:"playBet"`
	Result          int              `json:"result"`
	AntePayout      int              `json:"antePayout"`
	PlayPayout      int              `json:"playPayout"`
	AnteBonusPayout int              `json:"anteBonusPayout"`
	LowBonusPayout  int              `json:"lowBonusPayout"`
	TotalPayout     int              `json:"totalPayout"`
	DealerQualified bool             `json:"dealerQualified"`
	// **点数であって役位ではない。低いほど強い。** 0 は「役」(同ランク3枚 /
	// 同スート連番3枚) で、このゲームの最強手 —— 「手が無い」でも「未計算」でも
	// ない。読む側はフェーズで伏せ札かどうかを見分けること。
	PlayerScore int `json:"playerScore"`
	DealerScore int `json:"dealerScore"`
	WebOutputBase
}

// ThreeCardRummyWebController スリーカード・ラミーWebコントローラークラス
type ThreeCardRummyWebController = GameWebController[usecase.ThreeCardRummyInteractorIF, ThreeCardRummyWebInput, *ThreeCardRummyWebOutput]

// NewThreeCardRummyWebController and NewThreeCardRummyWebControllerWithProvider are
// the standard and provider-backed constructors for ThreeCardRummyWebController.
var NewThreeCardRummyWebController, NewThreeCardRummyWebControllerWithProvider = webControllerPair[usecase.ThreeCardRummyInteractorIF, ThreeCardRummyWebInput, *ThreeCardRummyWebOutput](
	newThreeCardRummyDefaultOutput, threeCardRummyDispatch,
)

func newThreeCardRummyDefaultOutput(msg string) *ThreeCardRummyWebOutput {
	return &ThreeCardRummyWebOutput{
		PlayerHand:    make([]*WebOutputCard, 0),
		DealerHand:    make([]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func threeCardRummyDispatch(bc *baseController, w http.ResponseWriter, ti usecase.ThreeCardRummyInteractorIF, param ThreeCardRummyWebInput, _ func(string) *ThreeCardRummyWebOutput) bool {
	switch param.Command {
	case "b", "bet":
		lowBonus := deref(param.LowBonusBet)
		bc.writePresenterResponse(w, ti.Bet(param.Amount, lowBonus))
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
