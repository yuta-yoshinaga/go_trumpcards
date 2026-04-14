package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PokerSquaresWebInput はポーカー・スクエアズの Web 入力。
type PokerSquaresWebInput struct {
	BaseWebInput
	Row *int `json:"row,omitempty"`
	Col *int `json:"col,omitempty"`
}

// PokerSquaresWebOutputCard はボード上の 1 セル出力。
type PokerSquaresWebOutputCard struct {
	Card *WebOutputCard `json:"card"`
}

// PokerSquaresWebOutput はポーカー・スクエアズの Web 出力。
type PokerSquaresWebOutput struct {
	Board       [][]*PokerSquaresWebOutputCard `json:"board"`
	CurrentCard *WebOutputCard                 `json:"currentCard,omitempty"`
	PlacedCount int                            `json:"placedCount"`
	Phase       int                            `json:"phase"`
	CanUndo     bool                           `json:"canUndo"`
	RowScores   []int                          `json:"rowScores"`
	ColScores   []int                          `json:"colScores"`
	TotalScore  int                            `json:"totalScore"`
	WebOutputBase
}

// PokerSquaresWebController はポーカー・スクエアズ Web コントローラー。
type PokerSquaresWebController = GameWebController[usecase.PokerSquaresInteractorIF, PokerSquaresWebInput, *PokerSquaresWebOutput]

// NewPokerSquaresWebController と NewPokerSquaresWebControllerWithProvider は
// 標準コンストラクタおよびプロバイダ指定コンストラクタ。
var NewPokerSquaresWebController, NewPokerSquaresWebControllerWithProvider = webControllerPair[usecase.PokerSquaresInteractorIF, PokerSquaresWebInput, *PokerSquaresWebOutput](
	newPokerSquaresDefaultOutput, pokerSquaresDispatch,
)

func newPokerSquaresDefaultOutput(msg string) *PokerSquaresWebOutput {
	return &PokerSquaresWebOutput{
		Board:         make([][]*PokerSquaresWebOutputCard, 0),
		RowScores:     make([]int, 0),
		ColScores:     make([]int, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func pokerSquaresDispatch(bc *baseController, w http.ResponseWriter, pi usecase.PokerSquaresInteractorIF, param PokerSquaresWebInput, newDefault func(string) *PokerSquaresWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, pi.Reset())
	case "p", "place":
		if param.Row == nil || param.Col == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: row and col are required."))
			return true
		}
		bc.writePresenterResponse(w, pi.Place(*param.Row, *param.Col))
	case "u", "undo":
		bc.writePresenterResponse(w, pi.Undo())
	case "g", "giveup":
		bc.writePresenterResponse(w, pi.GiveUp())
	default:
		return dispatchLog(param.Command, bc, w, pi.ActionLog)
	}
	return true
}
