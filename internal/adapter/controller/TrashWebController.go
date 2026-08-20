package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TrashWebInput トラッシュWebインプット
type TrashWebInput struct {
	BaseWebInput
	Position *int `json:"position,omitempty"`
}

// TrashWebSlot 1スロット出力
type TrashWebSlot struct {
	Card   *WebOutputCard `json:"card,omitempty"`
	FaceUp bool           `json:"faceUp"`
}

// TrashWebPlayer 1プレイヤー出力
type TrashWebPlayer struct {
	Slots [domain.TrashSlotCnt]TrashWebSlot `json:"slots"`
	IsCpu bool                              `json:"isCpu"`
}

// TrashWebOutput トラッシュWebアウトプット
type TrashWebOutput struct {
	Phase       int                                   `json:"phase"`
	Current     int                                   `json:"current"`
	Players     [domain.TrashPlayerCnt]TrashWebPlayer `json:"players"`
	StockSize   int                                   `json:"stockSize"`
	DiscardSize int                                   `json:"discardSize"`
	DiscardTop  *WebOutputCard                        `json:"discardTop,omitempty"`
	Pending     *WebOutputCard                        `json:"pending,omitempty"`
	MoveCount   int                                   `json:"moveCount"`
	Winner      int                                   `json:"winner"`
	WebOutputBase
}

// TrashWebController トラッシュWebコントローラークラス
type TrashWebController = GameWebController[usecase.TrashInteractorIF, TrashWebInput, *TrashWebOutput]

// NewTrashWebController and NewTrashWebControllerWithProvider are the standard
// and provider-backed constructors for TrashWebController.
var NewTrashWebController, NewTrashWebControllerWithProvider = webControllerPair[usecase.TrashInteractorIF, TrashWebInput, *TrashWebOutput](
	newTrashDefaultOutput, trashDispatch,
)

func newTrashDefaultOutput(msg string) *TrashWebOutput {
	return &TrashWebOutput{
		Winner:        -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func trashDispatch(bc *baseController, w http.ResponseWriter, ti usecase.TrashInteractorIF, param TrashWebInput, newDefault func(string) *TrashWebOutput) bool {
	switch param.Command {
	case "d", "draw":
		bc.writePresenterResponse(w, ti.Draw())
	case "p", "place", "placeWild":
		if !requireParam(bc, w, newDefault, param.Position == nil, "param error: position is required.") {
			return true
		}
		bc.writePresenterResponse(w, ti.PlaceWild(*param.Position))
	case "cpu":
		bc.writePresenterResponse(w, ti.CpuStep())
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, ti.Reset, ti.Hint, ti.ActionLog)
	}
	return true
}
