//go:build !js || !wasm || extra2

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// SlyFoxWebPresenter スライ・フォックス Web プレゼンタークラス
type SlyFoxWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *SlyFoxWebPresenter) Output(c interfaces.SlyFoxGame, lastErr error) string {
	resObj := new(controller.SlyFoxWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, c, int(c.GetPhase()))

	tableau := c.GetTableau()
	resObj.Tableau = make([][]*controller.WebOutputCard, domain.SlyFoxTableauCnt)
	for i := range domain.SlyFoxTableauCnt {
		pile := tableau[i]
		resObj.Tableau[i] = make([]*controller.WebOutputCard, len(pile))
		for j, card := range pile {
			resObj.Tableau[i][j] = cardToOutput(card)
		}
	}

	foundation := c.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.SlyFoxFoundationCnt)
	resObj.FoundationAscending = make([]bool, domain.SlyFoxFoundationCnt)
	for i := range domain.SlyFoxFoundationCnt {
		pile := foundation[i]
		resObj.Foundation[i] = make([]*controller.WebOutputCard, len(pile))
		for j, card := range pile {
			resObj.Foundation[i][j] = cardToOutput(card)
		}
		resObj.FoundationAscending[i] = c.IsAscendingFoundation(i)
	}

	resObj.StockCount = c.GetStockCount()
	resObj.DealtThisCycle = c.DealtThisCycle()
	resObj.DealCycle = domain.SlyFoxDealCycle
	resObj.ReserveLocked = c.ReserveIsLocked()

	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	if c.GetPhase() == domain.SlyFoxPhasePlaying {
		if hint := c.GetHint(); hint != nil {
			resObj.Hint = &controller.SlyFoxWebOutputHint{
				FromZone: hint.FromZone,
				FromIdx:  hint.FromIdx,
				ToZone:   hint.ToZone,
				ToIdx:    hint.ToIdx,
			}
		}
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		switch c.GetPhase() {
		case domain.SlyFoxPhasePlaying:
			if c.IsStalemate() {
				resObj.MessageCode = "slyfox.stalemate"
			} else {
				resObj.MessageCode = "slyfox.playing"
			}
		case domain.SlyFoxPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", c.GetMoveCount())
			resObj.MessageCode = "slyfox.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", c.GetMoveCount())}
		case domain.SlyFoxPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "slyfox.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *SlyFoxWebPresenter) HintOutput(c interfaces.SlyFoxGame) string {
	hint := c.GetHint()
	resObj := new(controller.SlyFoxWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, c, int(c.GetPhase()))
	resObj.Tableau = make([][]*controller.WebOutputCard, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)
	resObj.FoundationAscending = make([]bool, 0)
	resObj.DealCycle = domain.SlyFoxDealCycle

	if hint != nil {
		resObj.Hint = &controller.SlyFoxWebOutputHint{
			FromZone: hint.FromZone,
			FromIdx:  hint.FromIdx,
			ToZone:   hint.ToZone,
			ToIdx:    hint.ToIdx,
		}
		resObj.MessageCode = "slyfox.hintAvailable"
	} else {
		resObj.MessageCode = "slyfox.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *SlyFoxWebPresenter) ActionLogOutput(c interfaces.SlyFoxGame) string {
	return actionLogOutputJSON(c)
}
