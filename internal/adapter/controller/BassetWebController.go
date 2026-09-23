//go:build !js || !wasm || extra4

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BassetWebInput はバセットのWebインプット。
type BassetWebInput struct {
	BaseWebInput
	Rank   int `json:"rank,omitempty"`
	Amount int `json:"amount,omitempty"`
}

// BassetWebBet はレイアウト上の1ランクのベットを表す。
type BassetWebBet struct {
	Rank   int `json:"rank"`
	Amount int `json:"amount"`
	Stage  int `json:"stage"`
}

// BassetWebOutput はバセットのWebアウトプット。
type BassetWebOutput struct {
	Phase       int            `json:"phase"`
	Chips       int            `json:"chips"`
	Bet         *BassetWebBet  `json:"bet,omitempty"`
	BankerCard  *WebOutputCard `json:"bankerCard,omitempty"`
	PlayerCard  *WebOutputCard `json:"playerCard,omitempty"`
	Hit         bool           `json:"hit"`
	TurnsPlayed int            `json:"turnsPlayed"`
	TurnsTotal  int            `json:"turnsTotal"`
	Remaining   int            `json:"remaining"`
	// RemainingByRank は各ランクの残り枚数 (index 1..13 が A..K、0 は未使用)。
	// **未配の山札から直接数えた値**で、クライアントが公開札を蓄えて再構成する
	// 必要はない ── 蓄える形はリロードで消えて「全ランク満数」と嘘をつく (#6471)。
	RemainingByRank []int `json:"remainingByRank"`
	TotalPayout     int   `json:"totalPayout"`
	GameEndFlag     bool  `json:"gameEndFlag"`
	WebOutputBase
}

// BassetWebController はバセットのWebコントローラー。
type BassetWebController = GameWebController[usecase.BassetInteractorIF, BassetWebInput, *BassetWebOutput]

// NewBassetWebController and NewBassetWebControllerWithProvider are the standard and
// provider-backed constructors for BassetWebController.
var NewBassetWebController, NewBassetWebControllerWithProvider = webControllerPair[usecase.BassetInteractorIF, BassetWebInput, *BassetWebOutput](
	newBassetDefaultOutput, bassetDispatch,
)

func newBassetDefaultOutput(msg string) *BassetWebOutput {
	return &BassetWebOutput{
		Bet: nil,
		// エラー応答でも `null` を返さない ── クライアントは添字で引くので、
		// `null` だと画面が落ちる。
		RemainingByRank: make([]int, 0),
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func bassetDispatch(bc *baseController, w http.ResponseWriter, fi usecase.BassetInteractorIF, param BassetWebInput, _ func(string) *BassetWebOutput) bool {
	switch param.Command {
	case "b", "bet":
		bc.writePresenterResponse(w, fi.PlaceBet(param.Rank, param.Amount))
	case "d", "deal":
		bc.writePresenterResponse(w, fi.DealTurn())
	case "take", "takeWinnings":
		bc.writePresenterResponse(w, fi.TakeWinnings())
	case "paroli", "pressParoli":
		bc.writePresenterResponse(w, fi.PressParoli())
	case "n", "next":
		bc.writePresenterResponse(w, fi.NextRound())
	default:
		return dispatchResetAndLog(param.Command, bc, w, fi.Reset, fi.ActionLog)
	}
	return true
}
