//go:build !js || !wasm || extra4

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// FourteenOutWebPresenter はフォーティーンアウト・ソリティアの Web プレゼンター。
type FourteenOutWebPresenter struct{}

// Output はゲーム状態を JSON で出力する。
func (pr *FourteenOutWebPresenter) Output(g interfaces.FourteenOutGame, lastErr error) string {
	resObj := pr.buildBase(g)

	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	if g.GetPhase() == domain.FourteenOutPhasePlaying && !g.IsStalemate() {
		if hint := g.Hint(); hint != nil {
			resObj.Hint = &controller.FourteenOutWebOutputHint{
				Action:  hint.Action,
				FromCol: hint.FromCol,
				ToCol:   hint.ToCol,
			}
		}
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		switch g.GetPhase() {
		case domain.FourteenOutPhasePlaying:
			if g.IsStalemate() {
				resObj.MessageCode = "fourteenout.stalemate"
			} else {
				resObj.MessageCode = "fourteenout.playing"
			}
		case domain.FourteenOutPhaseGameClear:
			resObj.Message = "ゲームクリア！"
			resObj.MessageCode = "fourteenout.gameClear"
			resObj.MessageParams = map[string]string{
				"removedCount": fmt.Sprintf("%d", g.GetRemovedCount()),
			}
		case domain.FourteenOutPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "fourteenout.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput はヒントを JSON で出力する。
func (pr *FourteenOutWebPresenter) HintOutput(g interfaces.FourteenOutGame) string {
	resObj := pr.buildBase(g)
	hint := g.Hint()
	if hint != nil {
		resObj.Hint = &controller.FourteenOutWebOutputHint{
			Action:  hint.Action,
			FromCol: hint.FromCol,
			ToCol:   hint.ToCol,
		}
		resObj.MessageCode = "fourteenout.hintAvailable"
	} else {
		resObj.MessageCode = "fourteenout.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput は棋譜を JSON で出力する。
func (pr *FourteenOutWebPresenter) ActionLogOutput(g interfaces.FourteenOutGame) string {
	return actionLogOutputJSON(g)
}

// buildBase は共通のレスポンスフィールドを詰めて返す。
func (pr *FourteenOutWebPresenter) buildBase(g interfaces.FourteenOutGame) *controller.FourteenOutWebOutput {
	resObj := new(controller.FourteenOutWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RemovedCount = g.GetRemovedCount()
	resObj.RemovablePairs = g.CountRemovablePairs()
	resObj.CanUndo = g.CanUndo()
	resObj.IsStalemate = g.IsStalemate()

	cols := g.GetColumns()
	resObj.Columns = make([][]*controller.FourteenOutWebOutputCard, len(cols))
	for c, col := range cols {
		resObj.Columns[c] = make([]*controller.FourteenOutWebOutputCard, len(col))
		for i, card := range col {
			cell := &controller.FourteenOutWebOutputCard{}
			if card != nil {
				cell.Card = cardToOutput(card)
			}
			resObj.Columns[c][i] = cell
		}
	}
	return resObj
}
