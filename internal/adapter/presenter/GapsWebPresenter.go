//go:build !js || !wasm || solo

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// GapsWebPresenter はGapsゲームのWebプレゼンター。
type GapsWebPresenter struct{}

// Output はゲーム状態をJSONで返す。
func (pr *GapsWebPresenter) Output(g interfaces.GapsGame, lastErr error) string {
	resObj := new(controller.GapsWebOutput)
	pr.populate(resObj, g)

	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	if g.GetPhase() == domain.GapsPhasePlaying && !g.IsStalemate() {
		if hint := g.GetHint(); hint != nil {
			resObj.Hint = &controller.GapsWebOutputHint{
				FromRow: hint.FromRow,
				FromCol: hint.FromCol,
				ToRow:   hint.ToRow,
				ToCol:   hint.ToCol,
			}
		}
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		switch g.GetPhase() {
		case domain.GapsPhasePlaying:
			if g.IsStalemate() {
				resObj.MessageCode = "gaps.stalemate"
			} else {
				resObj.MessageCode = "gaps.playing"
			}
		case domain.GapsPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", g.GetMoveCount())
			resObj.MessageCode = "gaps.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", g.GetMoveCount())}
		case domain.GapsPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "gaps.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput はヒントをJSONで返す。
func (pr *GapsWebPresenter) HintOutput(g interfaces.GapsGame) string {
	hint := g.GetHint()
	resObj := new(controller.GapsWebOutput)
	pr.populate(resObj, g)
	if hint != nil {
		resObj.Hint = &controller.GapsWebOutputHint{
			FromRow: hint.FromRow,
			FromCol: hint.FromCol,
			ToRow:   hint.ToRow,
			ToCol:   hint.ToCol,
		}
		resObj.MessageCode = "gaps.hintAvailable"
	} else {
		resObj.MessageCode = "gaps.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput は棋譜をJSONで返す。
func (pr *GapsWebPresenter) ActionLogOutput(g interfaces.GapsGame) string {
	return actionLogOutputJSON(g)
}

func (pr *GapsWebPresenter) populate(resObj *controller.GapsWebOutput, g interfaces.GapsGame) {
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, g, int(g.GetPhase()))
	resObj.RedealsUsed = g.GetRedealsUsed()
	resObj.RedealsRemaining = g.GetRedealsRemaining()
	grid := g.GetGrid()
	out := make([][]*controller.WebOutputCard, domain.GapsRowCnt)
	for r := 0; r < domain.GapsRowCnt; r++ {
		row := make([]*controller.WebOutputCard, domain.GapsColCnt)
		for c := 0; c < domain.GapsColCnt; c++ {
			if grid[r][c] != nil {
				row[c] = cardToOutput(grid[r][c])
			}
		}
		out[r] = row
	}
	resObj.Grid = out
}
