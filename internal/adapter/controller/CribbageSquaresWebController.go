//go:build !js || !wasm || extra2

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CribbageSquaresWebInput はクリベッジ・スクエアズの Web 入力。
type CribbageSquaresWebInput struct {
	BaseWebInput
	Row *int `json:"row,omitempty"`
	Col *int `json:"col,omitempty"`
}

// CribbageSquaresWebOutputCard はボード上の 1 セル出力。
type CribbageSquaresWebOutputCard struct {
	Card *WebOutputCard `json:"card"`
}

// CribbageSquaresWebOutput はクリベッジ・スクエアズの Web 出力。
type CribbageSquaresWebOutput struct {
	Board       [][]*CribbageSquaresWebOutputCard `json:"board"`
	CurrentCard *WebOutputCard                    `json:"currentCard,omitempty"`
	PlacedCount int                               `json:"placedCount"`
	Phase       int                               `json:"phase"`
	CanUndo     bool                              `json:"canUndo"`
	RowScores   []int                             `json:"rowScores"`
	ColScores   []int                             `json:"colScores"`
	TotalScore  int                               `json:"totalScore"`
	// Starter は 17 枚目。16 マス埋まるまでは伏せたままなので null。
	Starter *WebOutputCard `json:"starter,omitempty"`
	// RowDetails / ColDetails は 8 手それぞれの得点内訳。どの手が何で
	// 稼いだのかを見せないと、合計だけでは上達のしようがない。
	RowDetails []*CribbageSquaresWebOutputScore `json:"rowDetails"`
	ColDetails []*CribbageSquaresWebOutputScore `json:"colDetails"`
	// RowPartialDetails / ColPartialDetails はスターター抜きで**既に確定して
	// いる**ぶん。RowDetails はスターターがめくれる 16 枚目まで必ず 0 なので、
	// これが無いと対局中の内訳を何も出せない (#6088)。スターターは点を足す
	// ことしかしないので、この値は最終点の下限になる。
	RowPartialDetails []*CribbageSquaresWebOutputScore `json:"rowPartialDetails"`
	ColPartialDetails []*CribbageSquaresWebOutputScore `json:"colPartialDetails"`
	// WinScore はクリア基準（61 点）。フロントで数値を持ち直さない。
	WinScore int                           `json:"winScore"`
	IsWin    bool                          `json:"isWin"`
	Hint     *CribbageSquaresWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
}

// CribbageSquaresWebOutputScore は 1 手ぶんのクリベッジ得点内訳。
type CribbageSquaresWebOutputScore struct {
	Fifteens int `json:"fifteens"`
	Pairs    int `json:"pairs"`
	Runs     int `json:"runs"`
	Flush    int `json:"flush"`
	Nobs     int `json:"nobs"`
	Total    int `json:"total"`
}

// CribbageSquaresWebOutputHint はサーバ側のシナジー考慮ヒント (#4790)。
type CribbageSquaresWebOutputHint struct {
	// Row は推奨するマスの行 (0-3)。
	Row int `json:"row"`
	// Col は推奨するマスの列 (0-3)。
	Col int `json:"col"`
	// Score はその配置が行と列に生む増分点。
	Score int `json:"score"`
	// Synergy はスコアが正（既存カードと噛み合う）かどうか。
	Synergy bool `json:"synergy"`
}

// CribbageSquaresWebController はクリベッジ・スクエアズ Web コントローラー。
type CribbageSquaresWebController = GameWebController[usecase.CribbageSquaresInteractorIF, CribbageSquaresWebInput, *CribbageSquaresWebOutput]

// NewCribbageSquaresWebController と NewCribbageSquaresWebControllerWithProvider は
// 標準コンストラクタおよびプロバイダ指定コンストラクタ。
var NewCribbageSquaresWebController, NewCribbageSquaresWebControllerWithProvider = webControllerPair[usecase.CribbageSquaresInteractorIF, CribbageSquaresWebInput, *CribbageSquaresWebOutput](
	newCribbageSquaresDefaultOutput, cribbageSquaresDispatch,
)

func newCribbageSquaresDefaultOutput(msg string) *CribbageSquaresWebOutput {
	return &CribbageSquaresWebOutput{
		Board:      make([][]*CribbageSquaresWebOutputCard, 0),
		RowScores:  make([]int, 0),
		ColScores:  make([]int, 0),
		RowDetails: make([]*CribbageSquaresWebOutputScore, 0),
		ColDetails: make([]*CribbageSquaresWebOutputScore, 0),
		// **配列は必ず空配列で返す。**null を返すと、フロントの map が落ちる。
		RowPartialDetails: make([]*CribbageSquaresWebOutputScore, 0),
		ColPartialDetails: make([]*CribbageSquaresWebOutputScore, 0),
		WinScore:          domain.CribbageSquaresWinScore,
		WebOutputBase:     WebOutputBase{Message: msg},
	}
}

func cribbageSquaresDispatch(bc *baseController, w http.ResponseWriter, pi usecase.CribbageSquaresInteractorIF, param CribbageSquaresWebInput, newDefault func(string) *CribbageSquaresWebOutput) bool {
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
