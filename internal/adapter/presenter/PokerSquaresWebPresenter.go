//go:build !js || !wasm || solo

package presenter

import (
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
	return pr.Output(p, nil)
}

// ActionLogOutput は棋譜を JSON で出力する。
func (pr *PokerSquaresWebPresenter) ActionLogOutput(p interfaces.PokerSquaresGame) string {
	return actionLogOutputJSON(p)
}
