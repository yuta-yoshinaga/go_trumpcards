//go:build !js || !wasm || casino

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CaribbeanDrawWebInput カリビアン・ドロー・ポーカーWebインプット
type CaribbeanDrawWebInput struct {
	BaseWebInput
	Amount     int  `json:"amount,omitempty"`
	JackpotBet *int `json:"jackpotBet,omitempty"`
	// Indices は交換する札の**0 始まり**の添字。`draw` コマンドでのみ読む。
	// 省略/空は「交換しない」を意味し、`null` と `[]` は同じ扱い。
	Indices []int `json:"indices,omitempty"`
}

// CaribbeanDrawWebOutput カリビアン・ドロー・ポーカーWebアウトプット
type CaribbeanDrawWebOutput struct {
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
	// DrawCost はこのラウンドで払った交換手数料。引かなければ 0。
	DrawCost       int `json:"drawCost"`
	PlayerHandRank int `json:"playerHandRank"`
	DealerHandRank int `json:"dealerHandRank"`
	WebOutputBase
}

// CaribbeanDrawWebController カリビアン・ドロー・ポーカーWebコントローラークラス
type CaribbeanDrawWebController = GameWebController[usecase.CaribbeanDrawInteractorIF, CaribbeanDrawWebInput, *CaribbeanDrawWebOutput]

// NewCaribbeanDrawWebController and NewCaribbeanDrawWebControllerWithProvider are
// the standard and provider-backed constructors for CaribbeanDrawWebController.
var NewCaribbeanDrawWebController, NewCaribbeanDrawWebControllerWithProvider = webControllerPair[usecase.CaribbeanDrawInteractorIF, CaribbeanDrawWebInput, *CaribbeanDrawWebOutput](
	newCaribbeanDrawDefaultOutput, caribbeanDrawDispatch,
)

func newCaribbeanDrawDefaultOutput(msg string) *CaribbeanDrawWebOutput {
	return &CaribbeanDrawWebOutput{
		PlayerHand:    make([]*WebOutputCard, 0),
		DealerHand:    make([]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func caribbeanDrawDispatch(bc *baseController, w http.ResponseWriter, ci usecase.CaribbeanDrawInteractorIF, param CaribbeanDrawWebInput, _ func(string) *CaribbeanDrawWebOutput) bool {
	switch param.Command {
	case "b", "bet":
		jackpot := deref(param.JackpotBet)
		bc.writePresenterResponse(w, ci.Bet(param.Amount, jackpot))
	case "d", "draw":
		bc.writePresenterResponse(w, ci.Draw(param.Indices))
	case "p", "play":
		bc.writePresenterResponse(w, ci.Play())
	case "f", "fold":
		bc.writePresenterResponse(w, ci.Fold())
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, ci.Reset, ci.Hint, ci.ActionLog)
	}
	return true
}
