//go:build !js || !wasm || solo

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
	Piles     []*AccordionWebOutputPile `json:"piles"`
	PileCount int                       `json:"pileCount"`
	Hint      *AccordionWebOutputHint   `json:"hint,omitempty"`
	SolitaireWebOutputBase
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
	case "ac", "autocomplete":
		// 独立 CUI には最初からある一括マージ。Web 側はページが自前でループを
		// 回していて、同じページの CLI モードからは呼べなかった (#5546)。
		bc.writePresenterResponse(w, ai.AutoComplete())
	case "undo_n":
		if !requireParam(bc, w, newDefault, param.N == nil, "param error: n is required.") {
			return true
		}
		bc.writePresenterResponse(w, ai.UndoN(*param.N))
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, ai.Reset, ai.Hint, ai.ActionLog)
	}
	return true
}

func accordionMoveDispatch(bc *baseController, w http.ResponseWriter, ai usecase.AccordionInteractorIF, param AccordionWebInput, newDefault func(string) *AccordionWebOutput) bool {
	if !requireParam(bc, w, newDefault, param.From == nil || param.To == nil, "param error: from and to are required.") {
		return true
	}
	if !requireParam(bc, w, newDefault, param.From.Zone != "pile" || param.To.Zone != "pile", "param error: invalid move zones. Only pile to pile is supported.") {
		return true
	}
	if !requireParam(bc, w, newDefault, param.From.Index == nil || param.To.Index == nil, "param error: from.index and to.index are required.") {
		return true
	}
	bc.writePresenterResponse(w, ai.Move(*param.From.Index, *param.To.Index))
	return true
}
