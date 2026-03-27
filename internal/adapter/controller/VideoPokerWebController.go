package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"

	"net/http"
)

// VideoPokerWebInput ビデオポーカーWebインプット
type VideoPokerWebInput struct {
	BaseWebInput
	Amount  int   `json:"amount,omitempty"`
	Indices []int `json:"indices,omitempty"`
}

// VideoPokerWebOutput ビデオポーカーWebアウトプット
type VideoPokerWebOutput struct {
	Hand        []*WebOutputCard `json:"hand"`
	Phase       int              `json:"phase"`
	Chips       int              `json:"chips"`
	BetAmount   int              `json:"betAmount"`
	Result      int              `json:"result"`
	Payout      int              `json:"payout"`
	HandRank    int              `json:"handRank"`
	HandName    string           `json:"handName"`
	HeldIndices [5]bool          `json:"heldIndices"`
	VariantName string           `json:"variantName"`
	WebOutputBase
}

// VideoPokerWebController ビデオポーカーWebコントローラークラス
type VideoPokerWebController = GameWebController[usecase.VideoPokerInteractorIF, VideoPokerWebInput, *VideoPokerWebOutput]

// NewVideoPokerWebController コンストラクタ
func NewVideoPokerWebController(factory func() usecase.VideoPokerInteractorIF) *VideoPokerWebController {
	return NewGameWebController(factory, newVideoPokerDefaultOutput, videoPokerDispatch)
}

func newVideoPokerDefaultOutput(msg string) *VideoPokerWebOutput {
	return &VideoPokerWebOutput{
		Hand:          make([]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func videoPokerDispatch(bc *baseController, w http.ResponseWriter, vi usecase.VideoPokerInteractorIF, param VideoPokerWebInput, _ func(string) *VideoPokerWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, vi.Reset())
	case "b", "bet":
		bc.writePresenterResponse(w, vi.Bet(param.Amount))
	case "h", "hold":
		indices := param.Indices
		if indices == nil {
			indices = []int{}
		}
		bc.writePresenterResponse(w, vi.Hold(indices))
	case "log", "l":
		bc.writePresenterResponse(w, vi.ActionLog())
	default:
		return false
	}
	return true
}
