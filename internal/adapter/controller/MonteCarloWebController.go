package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// MonteCarloWebInput はモンテカルロ・ソリティアの Web 入力。
type MonteCarloWebInput struct {
	BaseWebInput
	FromR *int `json:"fromR,omitempty"`
	FromC *int `json:"fromC,omitempty"`
	ToR   *int `json:"toR,omitempty"`
	ToC   *int `json:"toC,omitempty"`
}

// MonteCarloWebOutputCard はボード上の 1 セル出力。
type MonteCarloWebOutputCard struct {
	Card *WebOutputCard `json:"card"`
}

// MonteCarloWebOutputHint はヒント情報。Action は "remove" または "deal"。
type MonteCarloWebOutputHint struct {
	Action string `json:"action"`
	FromR  int    `json:"fromR,omitempty"`
	FromC  int    `json:"fromC,omitempty"`
	ToR    int    `json:"toR,omitempty"`
	ToC    int    `json:"toC,omitempty"`
}

// MonteCarloWebOutput はモンテカルロ・ソリティアの Web 出力。
type MonteCarloWebOutput struct {
	Board        [][]*MonteCarloWebOutputCard `json:"board"`
	Phase        int                          `json:"phase"`
	StockCount   int                          `json:"stockCount"`
	RemovedCount int                          `json:"removedCount"`
	DealCount    int                          `json:"dealCount"`
	CanUndo      bool                         `json:"canUndo"`
	IsStalemate  bool                         `json:"isStalemate"`
	Hint         *MonteCarloWebOutputHint     `json:"hint,omitempty"`
	WebOutputBase
}

// MonteCarloWebController はモンテカルロ・ソリティア Web コントローラー。
type MonteCarloWebController = GameWebController[usecase.MonteCarloInteractorIF, MonteCarloWebInput, *MonteCarloWebOutput]

// NewMonteCarloWebController と NewMonteCarloWebControllerWithProvider は
// 標準コンストラクタおよびプロバイダ指定コンストラクタ。
var NewMonteCarloWebController, NewMonteCarloWebControllerWithProvider = webControllerPair[usecase.MonteCarloInteractorIF, MonteCarloWebInput, *MonteCarloWebOutput](
	newMonteCarloDefaultOutput, monteCarloDispatch,
)

func newMonteCarloDefaultOutput(msg string) *MonteCarloWebOutput {
	return &MonteCarloWebOutput{
		Board:         make([][]*MonteCarloWebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func monteCarloDispatch(bc *baseController, w http.ResponseWriter, mi usecase.MonteCarloInteractorIF, param MonteCarloWebInput, newDefault func(string) *MonteCarloWebOutput) bool {
	switch param.Command {
	case "m", "move", "remove":
		if param.FromR == nil || param.FromC == nil || param.ToR == nil || param.ToC == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: fromR, fromC, toR, toC are required."))
			return true
		}
		bc.writePresenterResponse(w, mi.Remove(*param.FromR, *param.FromC, *param.ToR, *param.ToC))
	case "d", "deal":
		bc.writePresenterResponse(w, mi.Deal())
	case "u", "undo":
		bc.writePresenterResponse(w, mi.Undo())
	case "g", "giveup":
		bc.writePresenterResponse(w, mi.GiveUp())
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, mi.Reset, mi.Hint, mi.ActionLog)
	}
	return true
}
