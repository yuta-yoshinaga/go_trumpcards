//go:build !js || !wasm || solo

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// PyramidWebPresenter ピラミッドWebプレゼンタークラス
type PyramidWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (pr *PyramidWebPresenter) Output(p interfaces.PyramidGame, lastErr error) string {
	resObj := new(controller.PyramidWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, p, int(p.GetPhase()))
	resObj.StockCount = p.GetStockCount()

	// ウェイスト
	waste := p.GetWaste()
	if len(waste) > 0 {
		resObj.Waste = make([]*controller.WebOutputCard, len(waste))
		for i, c := range waste {
			resObj.Waste[i] = cardToOutput(c)
		}
	} else {
		resObj.Waste = make([]*controller.WebOutputCard, 0)
	}

	// ピラミッド
	pyramid := p.GetPyramid()
	resObj.Pyramid = make([][]*controller.PyramidWebOutputCard, domain.PyramidRowCnt)
	for row := range domain.PyramidRowCnt {
		rowCards := pyramid[row]
		resObj.Pyramid[row] = make([]*controller.PyramidWebOutputCard, len(rowCards))
		for col, pc := range rowCards {
			outPC := &controller.PyramidWebOutputCard{
				Removed: pc.Removed,
				Exposed: p.IsExposed(row, col),
			}
			if !pc.Removed {
				outPC.Card = cardToOutput(pc.Card)
			}
			resObj.Pyramid[row][col] = outPC
		}
	}

	// メッセージ
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	if p.GetPhase() == domain.PyramidPhasePlaying && !p.IsStalemate() {
		if hint := p.GetHint(); hint != nil {
			resObj.Hint = &controller.PyramidWebOutputHint{
				Type: hint.Type,
				Row1: hint.Row1,
				Col1: hint.Col1,
				Row2: hint.Row2,
				Col2: hint.Col2,
			}
		}
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		phase := p.GetPhase()
		switch phase {
		case domain.PyramidPhasePlaying:
			if p.IsStalemate() {
				resObj.MessageCode = "pyramid.stalemate"
			} else {
				resObj.MessageCode = "pyramid.playing"
			}
		case domain.PyramidPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", p.GetMoveCount())
			resObj.MessageCode = "pyramid.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", p.GetMoveCount())}
		case domain.PyramidPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "pyramid.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (pr *PyramidWebPresenter) HintOutput(p interfaces.PyramidGame) string {
	hint := p.GetHint()
	resObj := new(controller.PyramidWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, p, int(p.GetPhase()))
	resObj.StockCount = p.GetStockCount()
	resObj.Waste = make([]*controller.WebOutputCard, 0)
	resObj.Pyramid = make([][]*controller.PyramidWebOutputCard, 0)

	if hint != nil {
		resObj.Hint = &controller.PyramidWebOutputHint{
			Type: hint.Type,
			Row1: hint.Row1,
			Col1: hint.Col1,
			Row2: hint.Row2,
			Col2: hint.Col2,
		}
		resObj.MessageCode = "pyramid.hintAvailable"
	} else {
		resObj.MessageCode = "pyramid.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (pr *PyramidWebPresenter) ActionLogOutput(p interfaces.PyramidGame) string {
	return actionLogOutputJSON(p)
}
