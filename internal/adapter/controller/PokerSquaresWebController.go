//go:build !js || !wasm || solo

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
	Hint        *PokerSquaresWebOutputHint     `json:"hint,omitempty"`
	WebOutputBase
}

// PokerSquaresWebOutputHint はサーバ側のシナジー考慮ヒント (#4790)。
type PokerSquaresWebOutputHint struct {
	// Row は推奨するマスの行 (0-4)。
	Row int `json:"row"`
	// Col は推奨するマスの列 (0-4)。
	Col int `json:"col"`
	// Score はその配置が生む行・列の相乗効果スコア。
	Score int `json:"score"`
	// Synergy はスコアが正 (既存カードと相乗効果あり) かどうか。
	Synergy bool `json:"synergy"`
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
	case "p", "place":
		if !requireParam(bc, w, newDefault, param.Row == nil || param.Col == nil, "param error: row and col are required.") {
			return true
		}
		bc.writePresenterResponse(w, pi.Place(*param.Row, *param.Col))
	case "u", "undo":
		bc.writePresenterResponse(w, pi.Undo())
	case "g", "giveup":
		bc.writePresenterResponse(w, pi.GiveUp())
	case "h", "hint":
		bc.writePresenterResponse(w, pi.Hint())
	default:
		return dispatchResetAndLog(param.Command, bc, w, pi.Reset, pi.ActionLog)
	}
	return true
}
