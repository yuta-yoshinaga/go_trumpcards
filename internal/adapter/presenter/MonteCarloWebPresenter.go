//go:build !js || !wasm || solo

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MonteCarloWebPresenter はモンテカルロ・ソリティアの Web プレゼンター。
type MonteCarloWebPresenter struct{}

// Output はゲーム状態を JSON で出力する。
func (pr *MonteCarloWebPresenter) Output(g interfaces.MonteCarloGame, lastErr error) string {
	resObj := pr.buildBase(g)

	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	if g.GetPhase() == domain.MonteCarloPhasePlaying && !g.IsStalemate() {
		if hint := g.Hint(); hint != nil {
			resObj.Hint = &controller.MonteCarloWebOutputHint{
				Action: hint.Action,
				FromR:  hint.FromR,
				FromC:  hint.FromC,
				ToR:    hint.ToR,
				ToC:    hint.ToC,
			}
		}
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		switch g.GetPhase() {
		case domain.MonteCarloPhasePlaying:
			if g.IsStalemate() {
				resObj.MessageCode = "montecarlo.stalemate"
			} else {
				resObj.MessageCode = "montecarlo.playing"
			}
		case domain.MonteCarloPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 補充回数: %d", g.GetDealCount())
			resObj.MessageCode = "montecarlo.gameClear"
			resObj.MessageParams = map[string]string{
				"dealCount":    fmt.Sprintf("%d", g.GetDealCount()),
				"removedCount": fmt.Sprintf("%d", g.GetRemovedCount()),
			}
		case domain.MonteCarloPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "montecarlo.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput はヒントを JSON で出力する。
func (pr *MonteCarloWebPresenter) HintOutput(g interfaces.MonteCarloGame) string {
	resObj := pr.buildBase(g)
	hint := g.Hint()
	if hint != nil {
		resObj.Hint = &controller.MonteCarloWebOutputHint{
			Action: hint.Action,
			FromR:  hint.FromR,
			FromC:  hint.FromC,
			ToR:    hint.ToR,
			ToC:    hint.ToC,
		}
		resObj.MessageCode = "montecarlo.hintAvailable"
	} else {
		resObj.MessageCode = "montecarlo.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput は棋譜を JSON で出力する。
func (pr *MonteCarloWebPresenter) ActionLogOutput(g interfaces.MonteCarloGame) string {
	return actionLogOutputJSON(g)
}

// buildBase は共通のレスポンスフィールドを詰めて返す。
func (pr *MonteCarloWebPresenter) buildBase(g interfaces.MonteCarloGame) *controller.MonteCarloWebOutput {
	resObj := new(controller.MonteCarloWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.StockCount = g.GetStockCount()
	resObj.RemovedCount = g.GetRemovedCount()
	resObj.DealCount = g.GetDealCount()
	resObj.CanUndo = g.CanUndo()
	resObj.IsStalemate = g.IsStalemate()

	board := g.GetBoard()
	resObj.Board = make([][]*controller.MonteCarloWebOutputCard, domain.MonteCarloGridSize)
	for r := range domain.MonteCarloGridSize {
		resObj.Board[r] = make([]*controller.MonteCarloWebOutputCard, domain.MonteCarloGridSize)
		for c := range domain.MonteCarloGridSize {
			cell := &controller.MonteCarloWebOutputCard{}
			if board[r][c] != nil {
				cell.Card = cardToOutput(board[r][c])
			}
			resObj.Board[r][c] = cell
		}
	}
	return resObj
}
