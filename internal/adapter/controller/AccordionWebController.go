package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// AccordionWebInput アコーディオンWebインプット
type AccordionWebInput struct {
	BaseWebInput
	From *AccordionWebZone `json:"from,omitempty"`
	To   *AccordionWebZone `json:"to,omitempty"`
}

// AccordionWebZone ゾーン指定
type AccordionWebZone struct {
	Zone  string `json:"zone"`
	Index *int   `json:"index,omitempty"`
}

// AccordionWebOutputPile パイル出力
type AccordionWebOutputPile struct {
	Cards []*WebOutputCard `json:"cards"`
	// Size は埋め込みカードの枚数。Cards は常に top カードのみを含むが、
	// 合流した山の厚みを UI が表現できるように size を別フィールドで持つ。
	Size int `json:"size"`
}

// AccordionWebOutputHint ヒント出力
type AccordionWebOutputHint struct {
	FromIdx int `json:"fromIdx"`
	ToIdx   int `json:"toIdx"`
}

// AccordionWebOutput アコーディオンWebアウトプット
type AccordionWebOutput struct {
	Piles        []*AccordionWebOutputPile `json:"piles"`
	PileCount    int                       `json:"pileCount"`
	Phase        int                       `json:"phase"`
	MoveCount    int                       `json:"moveCount"`
	CanUndo      bool                      `json:"canUndo"`
	IsStalemate  bool                      `json:"isStalemate"`
	UndoToEscape int                       `json:"undoToEscape"`
	Hint         *AccordionWebOutputHint   `json:"hint,omitempty"`
	WebOutputBase
}

// AccordionWebController アコーディオンWebコントローラークラス
type AccordionWebController = GameWebController[usecase.AccordionInteractorIF, AccordionWebInput, *AccordionWebOutput]

// NewAccordionWebController and NewAccordionWebControllerWithProvider are
// the standard and provider-backed constructors for AccordionWebController.
var NewAccordionWebController, NewAccordionWebControllerWithProvider = webControllerPair[usecase.AccordionInteractorIF, AccordionWebInput, *AccordionWebOutput](
	newAccordionDefaultOutput, accordionDispatch,
)

func newAccordionDefaultOutput(msg string) *AccordionWebOutput {
	return &AccordionWebOutput{
		Piles:         make([]*AccordionWebOutputPile, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func accordionDispatch(bc *baseController, w http.ResponseWriter, ai usecase.AccordionInteractorIF, param AccordionWebInput, newDefault func(string) *AccordionWebOutput) bool {
	switch param.Command {
	case "m", "move":
		return accordionMoveDispatch(bc, w, ai, param, newDefault)
	case "g", "giveup":
		bc.writePresenterResponse(w, ai.GiveUp())
	case "u", "undo":
		bc.writePresenterResponse(w, ai.Undo())
	case "undo_n":
		if param.N == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: n is required."))
			return true
		}
		bc.writePresenterResponse(w, ai.UndoN(*param.N))
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, ai.Reset, ai.Hint, ai.ActionLog)
	}
	return true
}

func accordionMoveDispatch(bc *baseController, w http.ResponseWriter, ai usecase.AccordionInteractorIF, param AccordionWebInput, newDefault func(string) *AccordionWebOutput) bool {
	if param.From == nil || param.To == nil {
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: from and to are required."))
		return true
	}
	if param.From.Zone != "pile" || param.To.Zone != "pile" {
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid move zones. Only pile to pile is supported."))
		return true
	}
	if param.From.Index == nil || param.To.Index == nil {
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: from.index and to.index are required."))
		return true
	}
	bc.writePresenterResponse(w, ai.Move(*param.From.Index, *param.To.Index))
	return true
}
