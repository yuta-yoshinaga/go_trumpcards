//go:build !js || !wasm || extra3

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// CongressWebPresenter コングレス Web プレゼンタークラス
type CongressWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *CongressWebPresenter) Output(c interfaces.CongressGame, lastErr error) string {
	resObj := new(controller.CongressWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, c, int(c.GetPhase()))

	tableau := c.GetTableau()
	resObj.Tableau = make([][]*controller.WebOutputCard, domain.CongressTableauCnt)
	for i := range domain.CongressTableauCnt {
		pile := tableau[i]
		resObj.Tableau[i] = make([]*controller.WebOutputCard, len(pile))
		for j, card := range pile {
			resObj.Tableau[i][j] = cardToOutput(card)
		}
	}

	foundation := c.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.CongressFoundationCnt)
	for i := range domain.CongressFoundationCnt {
		pile := foundation[i]
		resObj.Foundation[i] = make([]*controller.WebOutputCard, len(pile))
		for j, card := range pile {
			resObj.Foundation[i][j] = cardToOutput(card)
		}
	}

	resObj.StockCount = c.GetStockCount()
	waste := c.GetWaste()
	resObj.Waste = make([]*controller.WebOutputCard, len(waste))
	for i, card := range waste {
		resObj.Waste[i] = cardToOutput(card)
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		switch c.GetPhase() {
		case domain.CongressPhasePlaying:
			if c.IsStalemate() {
				resObj.MessageCode = "congress.stalemate"
			} else {
				resObj.MessageCode = "congress.playing"
			}
		case domain.CongressPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", c.GetMoveCount())
			resObj.MessageCode = "congress.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", c.GetMoveCount())}
		case domain.CongressPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "congress.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *CongressWebPresenter) HintOutput(c interfaces.CongressGame) string {
	hint := c.GetHint()
	resObj := new(controller.CongressWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, c, int(c.GetPhase()))
	resObj.Tableau = make([][]*controller.WebOutputCard, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)
	resObj.Waste = make([]*controller.WebOutputCard, 0)

	if hint != nil {
		resObj.Hint = &controller.CongressWebOutputHint{
			FromZone: hint.FromZone,
			FromIdx:  hint.FromIdx,
			ToZone:   hint.ToZone,
			ToIdx:    hint.ToIdx,
		}
		resObj.MessageCode = "congress.hintAvailable"
	} else {
		resObj.MessageCode = "congress.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *CongressWebPresenter) ActionLogOutput(c interfaces.CongressGame) string {
	return actionLogOutputJSON(c)
}
