//go:build !js || !wasm || extra4

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PerseveranceWebInput パーシビアランスWebインプット
type PerseveranceWebInput struct {
	BaseWebInput
	From *PerseveranceWebZone `json:"from,omitempty"`
	To   *PerseveranceWebZone `json:"to,omitempty"`
}

// PerseveranceWebZone ゾーン指定
type PerseveranceWebZone struct {
	Zone      string `json:"zone"`
	Col       *int   `json:"col,omitempty"`
	CardIndex *int   `json:"cardIndex,omitempty"`
}

// PerseveranceWebOutputTableauCard タブローカード出力
type PerseveranceWebOutputTableauCard struct {
	Card   *WebOutputCard `json:"card"`
	FaceUp bool           `json:"faceUp"`
}

// PerseveranceWebOutputHint ヒント出力
type PerseveranceWebOutputHint struct {
	FromCol   int    `json:"fromCol"`
	CardIndex int    `json:"cardIndex"`
	ToZone    string `json:"toZone"`
	ToCol     int    `json:"toCol"`
}

// PerseveranceWebOutput パーシビアランスWebアウトプット
type PerseveranceWebOutput struct {
	Tableau    [][]*PerseveranceWebOutputTableauCard `json:"tableau"`
	Foundation [][]*WebOutputCard                    `json:"foundation"`
	Hint       *PerseveranceWebOutputHint            `json:"hint,omitempty"`
	// RedealsLeft 残りの再配り回数 (0-2)。0 になると再配りボタンは押せない。
	RedealsLeft int `json:"redealsLeft"`
	SolitaireWebOutputBase
	WebOutputBase
}

// PerseveranceWebController パーシビアランスWebコントローラークラス
type PerseveranceWebController = GameWebController[usecase.PerseveranceInteractorIF, PerseveranceWebInput, *PerseveranceWebOutput]

// NewPerseveranceWebController and NewPerseveranceWebControllerWithProvider are
// the standard and provider-backed constructors for PerseveranceWebController.
var NewPerseveranceWebController, NewPerseveranceWebControllerWithProvider = webControllerPair[usecase.PerseveranceInteractorIF, PerseveranceWebInput, *PerseveranceWebOutput](
	newPerseveranceDefaultOutput, perseveranceDispatch,
)

func newPerseveranceDefaultOutput(msg string) *PerseveranceWebOutput {
	return &PerseveranceWebOutput{
		Tableau:       make([][]*PerseveranceWebOutputTableauCard, 0),
		Foundation:    make([][]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func perseveranceDispatch(bc *baseController, w http.ResponseWriter, bi usecase.PerseveranceInteractorIF, param PerseveranceWebInput, newDefault func(string) *PerseveranceWebOutput) bool {
	switch param.Command {
	case "m", "move":
		return perseveranceMoveDispatch(bc, w, bi, param, newDefault)
	case "rd", "redeal":
		bc.writePresenterResponse(w, bi.Redeal())
	case "g", "giveup":
		bc.writePresenterResponse(w, bi.GiveUp())
	case "ac", "autocomplete":
		bc.writePresenterResponse(w, bi.AutoComplete())
	case "u", "undo":
		bc.writePresenterResponse(w, bi.Undo())
	case "undo_n":
		if !requireParam(bc, w, newDefault, param.N == nil, "param error: n is required.") {
			return true
		}
		bc.writePresenterResponse(w, bi.UndoN(*param.N))
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, bi.Reset, bi.Hint, bi.ActionLog)
	}
	return true
}

func perseveranceMoveDispatch(bc *baseController, w http.ResponseWriter, bi usecase.PerseveranceInteractorIF, param PerseveranceWebInput, newDefault func(string) *PerseveranceWebOutput) bool {
	mv := topCardMove{haveFrom: param.From != nil, haveTo: param.To != nil}
	if param.From != nil {
		mv.fromZone, mv.fromCol = param.From.Zone, param.From.Col
	}
	if param.To != nil {
		mv.toZone, mv.toCol = param.To.Zone, param.To.Col
	}
	return dispatchTopCardMove(bc, w, mv, topCardMoveFns{
		tableauToTableau:    bi.MoveTableauToTableau,
		tableauToFoundation: bi.MoveTableauToFoundation,
	}, newDefault)
}
