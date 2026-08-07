//go:build !js || !wasm || solo

package presenter

import (
	"encoding/json"
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// PokerSquaresWebPresenter はポーカー・スクエアズの Web プレゼンター。
type PokerSquaresWebPresenter struct{}

// Output はゲーム状態を JSON で出力する。
func (pr *PokerSquaresWebPresenter) Output(p interfaces.PokerSquaresGame, lastErr error) string {
	resObj := new(controller.PokerSquaresWebOutput)
	resObj.Phase = int(p.GetPhase())
	resObj.PlacedCount = p.GetPlacedCount()
	resObj.CanUndo = p.CanUndo()
	resObj.CurrentCard = cardToOutput(p.GetCurrentCard())

	board := p.GetBoard()
	resObj.Board = make([][]*controller.PokerSquaresWebOutputCard, domain.PokerSquaresGridSize)
	for r := range domain.PokerSquaresGridSize {
		resObj.Board[r] = make([]*controller.PokerSquaresWebOutputCard, domain.PokerSquaresGridSize)
		for c := range domain.PokerSquaresGridSize {
			cell := &controller.PokerSquaresWebOutputCard{}
			if board[r][c] != nil {
				cell.Card = cardToOutput(board[r][c])
			}
			resObj.Board[r][c] = cell
		}
	}

	resObj.RowScores = make([]int, domain.PokerSquaresGridSize)
	resObj.ColScores = make([]int, domain.PokerSquaresGridSize)
	for i := range domain.PokerSquaresGridSize {
		resObj.RowScores[i] = p.RowScore(i)
		resObj.ColScores[i] = p.ColScore(i)
		resObj.TotalScore += resObj.RowScores[i] + resObj.ColScores[i]
	}

	// **受動ヒントは Output() でも埋める。**command:"hint" 専用のレスポンスは
	// ページの state にマージされないので、ここで埋めないと state.hint が
	// 永久に undefined になる (#4483)。
	resObj.Hint = pokerSquaresWebHint(p)

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		switch p.GetPhase() {
		case domain.PokerSquaresPhasePlaying:
			resObj.MessageCode = "pokersquares.playing"
		case domain.PokerSquaresPhaseComplete:
			resObj.Message = fmt.Sprintf("ゲーム終了 合計得点: %d", resObj.TotalScore)
			resObj.MessageCode = "pokersquares.complete"
			resObj.MessageParams = map[string]string{"totalScore": fmt.Sprintf("%d", resObj.TotalScore)}
		}
	}

	return marshalOrError(resObj)
}

// HintOutput はヒントを JSON で出力する。Web GUI はクライアント側の
// ヒューリスティックヒントを使うため、ここでは JSON 化した状態を返して
// PokerSquaresPresenter インタフェースを満たす (CUI のみが利用する)。
func (pr *PokerSquaresWebPresenter) HintOutput(p interfaces.PokerSquaresGame) string {
	out := pr.Output(p, nil)
	// **押したときだけ出す印を付ける。**ページは isRequestedHint で
	// 「今まさに要求されたヒント」かを見分ける。Output() が常に載せる hint と
	// 区別が付かないと、ボタンを押していないのに助言が出続ける。
	var resObj controller.PokerSquaresWebOutput
	if err := json.Unmarshal([]byte(out), &resObj); err != nil {
		return out
	}
	if resObj.Hint != nil {
		resObj.MessageCode = "pokersquares.hintAvailable"
	} else {
		resObj.MessageCode = "pokersquares.noHint"
	}
	return marshalOrError(&resObj)
}

// pokerSquaresWebHint はドメインのヒントを JSON 用の形に移す。
func pokerSquaresWebHint(p interfaces.PokerSquaresGame) *controller.PokerSquaresWebOutputHint {
	hint := p.GetHint()
	if hint == nil {
		return nil
	}
	return &controller.PokerSquaresWebOutputHint{
		Row: hint.Row, Col: hint.Col, Score: hint.Score, Synergy: hint.Synergy,
	}
}

// ActionLogOutput は棋譜を JSON で出力する。
func (pr *PokerSquaresWebPresenter) ActionLogOutput(p interfaces.PokerSquaresGame) string {
	return actionLogOutputJSON(p)
}
