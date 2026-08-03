//go:build !js || !wasm || extra2

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// SirTommyWebPresenter サー・トミーWebプレゼンタークラス
type SirTommyWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *SirTommyWebPresenter) Output(g interfaces.SirTommyGame, lastErr error) string {
	resObj := p.buildBase(g)

	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	if g.GetPhase() == domain.SirTommyPhasePlaying && !g.IsStalemate() {
		if hint := g.GetHint(); hint != nil {
			resObj.Hint = &controller.SirTommyWebOutputHint{
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
		case domain.SirTommyPhasePlaying:
			if g.IsStalemate() {
				resObj.MessageCode = "sirtommy.stalemate"
			} else {
				resObj.MessageCode = "sirtommy.playing"
			}
		case domain.SirTommyPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", g.GetMoveCount())
			resObj.MessageCode = "sirtommy.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", g.GetMoveCount())}
		case domain.SirTommyPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "sirtommy.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *SirTommyWebPresenter) HintOutput(g interfaces.SirTommyGame) string {
	resObj := p.buildBase(g)
	hint := g.GetHint()
	if hint != nil {
		resObj.Hint = &controller.SirTommyWebOutputHint{
			FromZone:      hint.FromZone,
			WasteIdx:      hint.WasteIdx,
			FoundationIdx: hint.FoundationIdx,
		}
		resObj.MessageCode = "sirtommy.hintAvailable"
	} else {
		resObj.MessageCode = "sirtommy.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *SirTommyWebPresenter) ActionLogOutput(g interfaces.SirTommyGame) string {
	return actionLogOutputJSON(g)
}

// buildBase は共通フィールドを詰めたレスポンスオブジェクトを返す
func (p *SirTommyWebPresenter) buildBase(g interfaces.SirTommyGame) *controller.SirTommyWebOutput {
	resObj := new(controller.SirTommyWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, g, int(g.GetPhase()))
	resObj.StockCount = g.GetStockCount()
	if top := g.GetStockTop(); top != nil {
		resObj.StockTop = cardToOutput(top)
	}

	foundations := g.GetFoundations()
	resObj.Foundations = make([][]*controller.WebOutputCard, domain.SirTommyFoundationCnt)
	for i := range domain.SirTommyFoundationCnt {
		pile := foundations[i]
		out := make([]*controller.WebOutputCard, len(pile))
		for j, card := range pile {
			out[j] = cardToOutput(card)
		}
		resObj.Foundations[i] = out
	}

	wastes := g.GetWastes()
	resObj.Wastes = make([][]*controller.WebOutputCard, domain.SirTommyWasteCnt)
	for i := range domain.SirTommyWasteCnt {
		pile := wastes[i]
		out := make([]*controller.WebOutputCard, len(pile))
		for j, card := range pile {
			out[j] = cardToOutput(card)
		}
		resObj.Wastes[i] = out
	}
	return resObj
}
