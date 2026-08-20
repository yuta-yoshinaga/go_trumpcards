//go:build !js || !wasm || casino

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ChinesePokerWebInput チャイニーズポーカーWebインプット
type ChinesePokerWebInput struct {
	BaseWebInput
	Amount        int   `json:"amount,omitempty"`
	FrontIndices  []int `json:"frontIndices,omitempty"`
	MiddleIndices []int `json:"middleIndices,omitempty"`
}

// ChinesePokerWebOutputArrangement は推奨する13枚の分け方。**インデックスは
// playerCards のもの**で、カードそのものではない (フロントが手札の並びを
// 変えても、同じ札を指し続ける)。
type ChinesePokerWebOutputArrangement struct {
	Front  []int `json:"front"`
	Middle []int `json:"middle"`
	Back   []int `json:"back"`
	// Foul はこの分け方がファウルになるか。ランク順に切るだけでは合法とは
	// 限らないので、勧めると同時に危険も伝える (#5615)。
	Foul bool `json:"foul"`
}

// ChinesePokerWebOutput チャイニーズポーカーWebアウトプット
type ChinesePokerWebOutput struct {
	PlayerCards      []*WebOutputCard `json:"playerCards"`
	DealerCards      []*WebOutputCard `json:"dealerCards"`
	PlayerFront      []*WebOutputCard `json:"playerFront"`
	PlayerMiddle     []*WebOutputCard `json:"playerMiddle"`
	PlayerBack       []*WebOutputCard `json:"playerBack"`
	DealerFront      []*WebOutputCard `json:"dealerFront"`
	DealerMiddle     []*WebOutputCard `json:"dealerMiddle"`
	DealerBack       []*WebOutputCard `json:"dealerBack"`
	Phase            int              `json:"phase"`
	Chips            int              `json:"chips"`
	Bet              int              `json:"bet"`
	Result           int              `json:"result"`
	FrontResult      int              `json:"frontResult"`
	MiddleResult     int              `json:"middleResult"`
	BackResult       int              `json:"backResult"`
	Payout           int              `json:"payout"`
	PlayerFrontRank  int              `json:"playerFrontRank"`
	PlayerMiddleRank int              `json:"playerMiddleRank"`
	PlayerBackRank   int              `json:"playerBackRank"`
	DealerFrontRank  int              `json:"dealerFrontRank"`
	DealerMiddleRank int              `json:"dealerMiddleRank"`
	DealerBackRank   int              `json:"dealerBackRank"`
	PlayerRoyalty    int              `json:"playerRoyalty"`
	DealerRoyalty    int              `json:"dealerRoyalty"`
	Scoop            bool             `json:"scoop"`
	// SuggestedArrangement はセットハンドで13枚そろっているときだけ入る。
	// 空配列ではなく省略するのは、「前列に置く札が無い」と読めてしまうため。
	SuggestedArrangement *ChinesePokerWebOutputArrangement `json:"suggestedArrangement,omitempty"`
	WebOutputBase
}

// ChinesePokerWebController チャイニーズポーカーWebコントローラークラス
type ChinesePokerWebController = GameWebController[usecase.ChinesePokerInteractorIF, ChinesePokerWebInput, *ChinesePokerWebOutput]

// NewChinesePokerWebController and NewChinesePokerWebControllerWithProvider are
// the standard and provider-backed constructors for ChinesePokerWebController.
var NewChinesePokerWebController, NewChinesePokerWebControllerWithProvider = webControllerPair[usecase.ChinesePokerInteractorIF, ChinesePokerWebInput, *ChinesePokerWebOutput](
	newChinesePokerDefaultOutput, chinesePokerDispatch,
)

func newChinesePokerDefaultOutput(msg string) *ChinesePokerWebOutput {
	return &ChinesePokerWebOutput{
		PlayerCards:   make([]*WebOutputCard, 0),
		DealerCards:   make([]*WebOutputCard, 0),
		PlayerFront:   make([]*WebOutputCard, 0),
		PlayerMiddle:  make([]*WebOutputCard, 0),
		PlayerBack:    make([]*WebOutputCard, 0),
		DealerFront:   make([]*WebOutputCard, 0),
		DealerMiddle:  make([]*WebOutputCard, 0),
		DealerBack:    make([]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func chinesePokerDispatch(bc *baseController, w http.ResponseWriter, ci usecase.ChinesePokerInteractorIF, param ChinesePokerWebInput, _ func(string) *ChinesePokerWebOutput) bool {
	switch param.Command {
	case "b", "bet":
		bc.writePresenterResponse(w, ci.Bet(param.Amount))
	case "s", "set":
		bc.writePresenterResponse(w, ci.SetHands(param.FrontIndices, param.MiddleIndices))
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, ci.Reset, ci.Hint, ci.ActionLog)
	}
	return true
}
