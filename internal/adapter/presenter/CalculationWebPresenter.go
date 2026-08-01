//go:build !js || !wasm || solo

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// CalculationWebPresenter カルキュレーションWebプレゼンタークラス
type CalculationWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *CalculationWebPresenter) Output(g interfaces.CalculationGame, lastErr error) string {
	resObj := p.buildBase(g)

	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	if g.GetPhase() == domain.CalculationPhasePlaying && !g.IsStalemate() {
		if hint := g.GetHint(); hint != nil {
			resObj.Hint = &controller.CalculationWebOutputHint{
				FromZone:      hint.FromZone,
				WasteIdx:      hint.WasteIdx,
				FoundationIdx: hint.FoundationIdx,
			}
		}
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		switch g.GetPhase() {
		case domain.CalculationPhasePlaying:
			if g.IsStalemate() {
				resObj.MessageCode = "calculation.stalemate"
			} else {
				resObj.MessageCode = "calculation.playing"
			}
		case domain.CalculationPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", g.GetMoveCount())
			resObj.MessageCode = "calculation.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", g.GetMoveCount())}
		case domain.CalculationPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "calculation.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *CalculationWebPresenter) HintOutput(g interfaces.CalculationGame) string {
	resObj := p.buildBase(g)
	hint := g.GetHint()
	if hint != nil {
		resObj.Hint = &controller.CalculationWebOutputHint{
			FromZone:      hint.FromZone,
			WasteIdx:      hint.WasteIdx,
			FoundationIdx: hint.FoundationIdx,
		}
		resObj.MessageCode = "calculation.hintAvailable"
	} else {
		resObj.MessageCode = "calculation.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *CalculationWebPresenter) ActionLogOutput(g interfaces.CalculationGame) string {
	return actionLogOutputJSON(g)
}

// buildBase は共通フィールドを詰めたレスポンスオブジェクトを返す
func (p *CalculationWebPresenter) buildBase(g interfaces.CalculationGame) *controller.CalculationWebOutput {
	resObj := new(controller.CalculationWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, g, int(g.GetPhase()))
	resObj.StockCount = g.GetStockCount()
	if top := g.GetStockTop(); top != nil {
		resObj.StockTop = cardToOutput(top)
	}

	foundations := g.GetFoundations()
	resObj.Foundations = make([][]*controller.WebOutputCard, domain.CalculationFoundationCnt)
	for i := range domain.CalculationFoundationCnt {
		pile := foundations[i]
		out := make([]*controller.WebOutputCard, len(pile))
		for j, card := range pile {
			out[j] = cardToOutput(card)
		}
		resObj.Foundations[i] = out
	}

	wastes := g.GetWastes()
	resObj.Wastes = make([][]*controller.WebOutputCard, domain.CalculationWasteCnt)
	for i := range domain.CalculationWasteCnt {
		pile := wastes[i]
		out := make([]*controller.WebOutputCard, len(pile))
		for j, card := range pile {
			out[j] = cardToOutput(card)
		}
		resObj.Wastes[i] = out
	}
	return resObj
}
