package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ClockSolitaireWebInput クロックソリティアWebインプット
type ClockSolitaireWebInput struct {
	BaseWebInput
}

// ClockSolitaireWebOutputCard パイルカード出力
type ClockSolitaireWebOutputCard struct {
	Card   *WebOutputCard `json:"card"`
	FaceUp bool           `json:"faceUp"`
}

// ClockSolitaireWebOutput クロックソリティアWebアウトプット
type ClockSolitaireWebOutput struct {
	Piles       [][]*ClockSolitaireWebOutputCard `json:"piles"`
	FaceUpCount []int                            `json:"faceUpCount"`
	Phase       int                              `json:"phase"`
	StepCount   int                              `json:"stepCount"`
	CurrentCard *WebOutputCard                   `json:"currentCard,omitempty"`
	WebOutputBase
}

// ClockSolitaireWebController クロックソリティアWebコントローラークラス
type ClockSolitaireWebController = GameWebController[usecase.ClockSolitaireInteractorIF, ClockSolitaireWebInput, *ClockSolitaireWebOutput]

// NewClockSolitaireWebController and NewClockSolitaireWebControllerWithProvider are
// the standard and provider-backed constructors for ClockSolitaireWebController.
var NewClockSolitaireWebController, NewClockSolitaireWebControllerWithProvider = webControllerPair[usecase.ClockSolitaireInteractorIF, ClockSolitaireWebInput, *ClockSolitaireWebOutput](
	newClockSolitaireDefaultOutput, clockSolitaireDispatch,
)

func newClockSolitaireDefaultOutput(msg string) *ClockSolitaireWebOutput {
	return &ClockSolitaireWebOutput{
		Piles:         make([][]*ClockSolitaireWebOutputCard, 0),
		FaceUpCount:   make([]int, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func clockSolitaireDispatch(bc *baseController, w http.ResponseWriter, ci usecase.ClockSolitaireInteractorIF, param ClockSolitaireWebInput, _ func(string) *ClockSolitaireWebOutput) bool {
	if param.Command == "a" || param.Command == "autoplay" {
		bc.writePresenterResponse(w, ci.AutoPlay())
		return true
	}
	return dispatchResetStepLog(param.Command, bc, w, ci)
}
