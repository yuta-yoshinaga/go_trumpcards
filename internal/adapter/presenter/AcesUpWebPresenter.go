//go:build !js || !wasm || solo

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// AcesUpWebPresenter エースアップWebプレゼンタークラス
type AcesUpWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (pr *AcesUpWebPresenter) Output(g interfaces.AcesUpGame, lastErr error) string {
	resObj := new(controller.AcesUpWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, g, int(g.GetPhase()))
	resObj.StockCount = g.GetStockCount()
	resObj.DiscardCount = g.GetDiscardCount()
	resObj.DiscardTop = cardToOutput(g.GetDiscardTop())
	resObj.Columns = acesUpColumns(g)

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		switch g.GetPhase() {
		case domain.AcesUpPhasePlaying:
			if g.IsStalemate() {
				resObj.MessageCode = "acesup.stalemate"
			} else {
				resObj.MessageCode = "acesup.playing"
			}
		case domain.AcesUpPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", g.GetMoveCount())
			resObj.MessageCode = "acesup.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", g.GetMoveCount())}
		case domain.AcesUpPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "acesup.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// acesUpColumns はドメインの列表現をWeb出力に変換する。
func acesUpColumns(g interfaces.AcesUpGame) [][]*controller.AcesUpWebOutputCard {
	cols := g.GetColumns()
	out := make([][]*controller.AcesUpWebOutputCard, domain.AcesUpColCnt)
	for c := range domain.AcesUpColCnt {
		col := cols[c]
		out[c] = make([]*controller.AcesUpWebOutputCard, len(col))
		for i, card := range col {
			top := i == len(col)-1
			outCard := &controller.AcesUpWebOutputCard{
				Card: cardToOutput(card),
				Top:  top,
			}
			if top {
				outCard.Removable = g.CanRemove(c)
				outCard.Movable = g.CanMove(c)
			}
			out[c][i] = outCard
		}
	}
	return out
}

// HintOutput ヒントをJSON出力
func (pr *AcesUpWebPresenter) HintOutput(g interfaces.AcesUpGame) string {
	hint := g.GetHint()
	resObj := new(controller.AcesUpWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, g, int(g.GetPhase()))
	resObj.StockCount = g.GetStockCount()
	resObj.DiscardCount = g.GetDiscardCount()
	resObj.Columns = make([][]*controller.AcesUpWebOutputCard, 0)

	if hint != nil {
		resObj.Hint = &controller.AcesUpWebOutputHint{
			Type: hint.Type,
			Col:  hint.Col,
		}
		resObj.MessageCode = "acesup.hintAvailable"
	} else {
		resObj.MessageCode = "acesup.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (pr *AcesUpWebPresenter) ActionLogOutput(g interfaces.AcesUpGame) string {
	return actionLogOutputJSON(g)
}
