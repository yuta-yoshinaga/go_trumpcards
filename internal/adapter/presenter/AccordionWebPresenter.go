//go:build !js || !wasm || solo

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// AccordionWebPresenter アコーディオンWebプレゼンタークラス
type AccordionWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *AccordionWebPresenter) Output(a interfaces.AccordionGame, lastErr error) string {
	resObj := p.buildBase(a)

	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	if a.GetPhase() == domain.AccordionPhasePlaying && !a.IsStalemate() {
		if hint := a.GetHint(); hint != nil {
			resObj.Hint = &controller.AccordionWebOutputHint{
				FromIdx: hint.FromIdx,
				ToIdx:   hint.ToIdx,
			}
		}
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		switch a.GetPhase() {
		case domain.AccordionPhasePlaying:
			if a.IsStalemate() {
				resObj.MessageCode = "accordion.stalemate"
			} else {
				resObj.MessageCode = "accordion.playing"
			}
		case domain.AccordionPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", a.GetMoveCount())
			resObj.MessageCode = "accordion.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", a.GetMoveCount())}
		case domain.AccordionPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "accordion.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *AccordionWebPresenter) HintOutput(a interfaces.AccordionGame) string {
	resObj := p.buildBase(a)
	hint := a.GetHint()
	if hint != nil {
		resObj.Hint = &controller.AccordionWebOutputHint{
			FromIdx: hint.FromIdx,
			ToIdx:   hint.ToIdx,
		}
		resObj.MessageCode = "accordion.hintAvailable"
	} else {
		resObj.MessageCode = "accordion.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *AccordionWebPresenter) ActionLogOutput(a interfaces.AccordionGame) string {
	return actionLogOutputJSON(a)
}

// buildBase は共通フィールドを詰めたレスポンスオブジェクトを返す
func (p *AccordionWebPresenter) buildBase(a interfaces.AccordionGame) *controller.AccordionWebOutput {
	resObj := new(controller.AccordionWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, a, int(a.GetPhase()))
	resObj.PileCount = a.GetPileCount()

	piles := a.GetPiles()
	resObj.Piles = make([]*controller.AccordionWebOutputPile, len(piles))
	for i, pile := range piles {
		pileOut := &controller.AccordionWebOutputPile{
			Cards: make([]*controller.WebOutputCard, 0, 1),
			Size:  len(pile),
		}
		if len(pile) > 0 {
			top := pile[len(pile)-1]
			pileOut.Cards = append(pileOut.Cards, cardToOutput(top))
		}
		resObj.Piles[i] = pileOut
	}
	return resObj
}
