//go:build !js || !wasm || extra4

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// FourteenOutWebInput はフォーティーンアウト・ソリティアの Web 入力。
type FourteenOutWebInput struct {
	BaseWebInput
	// **列番号 2 つで足りる。**動かせるのは各列の末尾だけなので、クローン元の
	// Monte Carlo が要る (行,列) x2 は Fourteen Out では過剰な自由度になる。
	FromCol *int `json:"fromCol,omitempty"`
	ToCol   *int `json:"toCol,omitempty"`
}

// FourteenOutWebOutputCard はボード上の 1 セル出力。
type FourteenOutWebOutputCard struct {
	Card *WebOutputCard `json:"card"`
}

// FourteenOutWebOutputHint はヒント情報。Action は "remove" または "deal"。
type FourteenOutWebOutputHint struct {
	Action  string `json:"action"`
	FromCol int    `json:"fromCol"`
	ToCol   int    `json:"toCol"`
}

// FourteenOutWebOutput はフォーティーンアウト・ソリティアの Web 出力。
type FourteenOutWebOutput struct {
	// Columns は 12 列。各列の末尾だけが露出している札。
	Columns      [][]*FourteenOutWebOutputCard `json:"columns"`
	Phase        int                           `json:"phase"`
	RemovedCount int                           `json:"removedCount"`
	// RemovablePairs は残っている取り除ける組の数。手が見えているかの判断材料。
	RemovablePairs int                       `json:"removablePairs"`
	CanUndo        bool                      `json:"canUndo"`
	IsStalemate    bool                      `json:"isStalemate"`
	Hint           *FourteenOutWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
}

// FourteenOutWebController はフォーティーンアウト・ソリティア Web コントローラー。
type FourteenOutWebController = GameWebController[usecase.FourteenOutInteractorIF, FourteenOutWebInput, *FourteenOutWebOutput]

// NewFourteenOutWebController と NewFourteenOutWebControllerWithProvider は
// 標準コンストラクタおよびプロバイダ指定コンストラクタ。
var NewFourteenOutWebController, NewFourteenOutWebControllerWithProvider = webControllerPair[usecase.FourteenOutInteractorIF, FourteenOutWebInput, *FourteenOutWebOutput](
	newFourteenOutDefaultOutput, fourteenOutDispatch,
)

func newFourteenOutDefaultOutput(msg string) *FourteenOutWebOutput {
	return &FourteenOutWebOutput{
		Columns:       make([][]*FourteenOutWebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func fourteenOutDispatch(bc *baseController, w http.ResponseWriter, mi usecase.FourteenOutInteractorIF, param FourteenOutWebInput, newDefault func(string) *FourteenOutWebOutput) bool {
	switch param.Command {
	case "m", "move", "remove":
		if !requireParam(bc, w, newDefault, param.FromCol == nil || param.ToCol == nil, "param error: fromCol and toCol are required.") {
			return true
		}
		bc.writePresenterResponse(w, mi.Remove(*param.FromCol, *param.ToCol))
	case "u", "undo":
		bc.writePresenterResponse(w, mi.Undo())
	case "g", "giveup":
		bc.writePresenterResponse(w, mi.GiveUp())
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, mi.Reset, mi.Hint, mi.ActionLog)
	}
	return true
}
