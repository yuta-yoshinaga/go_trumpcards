//go:build !js || !wasm || extra2

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// AuldLangSyneWebPresenter オールド・ラング・サインWebプレゼンタークラス
type AuldLangSyneWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *AuldLangSyneWebPresenter) Output(g interfaces.AuldLangSyneGame, lastErr error) string {
	resObj := p.buildBase(g)

	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	if g.GetPhase() == domain.AuldLangSynePhasePlaying && !g.IsStalemate() {
		if hint := g.GetHint(); hint != nil {
			resObj.Hint = &controller.AuldLangSyneWebOutputHint{
				WasteIdx:      hint.WasteIdx,
				FoundationIdx: hint.FoundationIdx,
			}
		}
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		switch g.GetPhase() {
		case domain.AuldLangSynePhasePlaying:
			if g.IsStalemate() {
				resObj.MessageCode = "auldlangsyne.stalemate"
			} else {
				resObj.MessageCode = "auldlangsyne.playing"
			}
		case domain.AuldLangSynePhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", g.GetMoveCount())
			resObj.MessageCode = "auldlangsyne.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", g.GetMoveCount())}
		case domain.AuldLangSynePhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "auldlangsyne.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *AuldLangSyneWebPresenter) HintOutput(g interfaces.AuldLangSyneGame) string {
	resObj := p.buildBase(g)
	hint := g.GetHint()
	if hint != nil {
		resObj.Hint = &controller.AuldLangSyneWebOutputHint{
			WasteIdx:      hint.WasteIdx,
			FoundationIdx: hint.FoundationIdx,
		}
		resObj.MessageCode = "auldlangsyne.hintAvailable"
	} else {
		resObj.MessageCode = "auldlangsyne.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *AuldLangSyneWebPresenter) ActionLogOutput(g interfaces.AuldLangSyneGame) string {
	return actionLogOutputJSON(g)
}

// buildBase は共通フィールドを詰めたレスポンスオブジェクトを返す
func (p *AuldLangSyneWebPresenter) buildBase(g interfaces.AuldLangSyneGame) *controller.AuldLangSyneWebOutput {
	resObj := new(controller.AuldLangSyneWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, g, int(g.GetPhase()))
	resObj.StockCount = g.GetStockCount()

	foundations := g.GetFoundations()
	resObj.Foundations = make([][]*controller.WebOutputCard, domain.AuldLangSyneFoundationCnt)
	for i := range domain.AuldLangSyneFoundationCnt {
		pile := foundations[i]
		out := make([]*controller.WebOutputCard, len(pile))
		for j, card := range pile {
			out[j] = cardToOutput(card)
		}
		resObj.Foundations[i] = out
	}

	wastes := g.GetWastes()
	resObj.Wastes = make([][]*controller.WebOutputCard, domain.AuldLangSyneWasteCnt)
	for i := range domain.AuldLangSyneWasteCnt {
		pile := wastes[i]
		out := make([]*controller.WebOutputCard, len(pile))
		for j, card := range pile {
			out[j] = cardToOutput(card)
		}
		resObj.Wastes[i] = out
	}
	return resObj
}
