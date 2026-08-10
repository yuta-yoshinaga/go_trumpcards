//go:build !js || !wasm || classic

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// ColoradoWebPresenter コロラド Web プレゼンタークラス
type ColoradoWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *ColoradoWebPresenter) Output(c interfaces.ColoradoGame, lastErr error) string {
	resObj := new(controller.ColoradoWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, c, int(c.GetPhase()))

	tableau := c.GetTableau()
	resObj.Tableau = make([][]*controller.WebOutputCard, domain.ColoradoTableauCnt)
	for i := range domain.ColoradoTableauCnt {
		pile := tableau[i]
		resObj.Tableau[i] = make([]*controller.WebOutputCard, len(pile))
		for j, card := range pile {
			resObj.Tableau[i][j] = cardToOutput(card)
		}
	}

	foundation := c.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.ColoradoFoundationCnt)
	resObj.FoundationAscending = make([]bool, domain.ColoradoFoundationCnt)
	for i := range domain.ColoradoFoundationCnt {
		pile := foundation[i]
		resObj.Foundation[i] = make([]*controller.WebOutputCard, len(pile))
		for j, card := range pile {
			resObj.Foundation[i][j] = cardToOutput(card)
		}
		resObj.FoundationAscending[i] = c.IsAscendingFoundation(i)
	}

	resObj.StockCount = c.GetStockCount()
	waste := c.GetWaste()
	resObj.Waste = make([]*controller.WebOutputCard, len(waste))
	for i, card := range waste {
		resObj.Waste[i] = cardToOutput(card)
	}

	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	if c.GetPhase() == domain.ColoradoPhasePlaying {
		if hint := c.GetHint(); hint != nil {
			resObj.Hint = &controller.ColoradoWebOutputHint{
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
		case domain.ColoradoPhasePlaying:
			if c.IsStalemate() {
				resObj.MessageCode = "colorado.stalemate"
			} else {
				resObj.MessageCode = "colorado.playing"
			}
		case domain.ColoradoPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", c.GetMoveCount())
			resObj.MessageCode = "colorado.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", c.GetMoveCount())}
		case domain.ColoradoPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "colorado.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *ColoradoWebPresenter) HintOutput(c interfaces.ColoradoGame) string {
	hint := c.GetHint()
	resObj := new(controller.ColoradoWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, c, int(c.GetPhase()))
	resObj.Tableau = make([][]*controller.WebOutputCard, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)
	resObj.FoundationAscending = make([]bool, 0)
	resObj.Waste = make([]*controller.WebOutputCard, 0)

	if hint != nil {
		resObj.Hint = &controller.ColoradoWebOutputHint{
			FromZone: hint.FromZone,
			FromIdx:  hint.FromIdx,
			ToZone:   hint.ToZone,
			ToIdx:    hint.ToIdx,
		}
		resObj.MessageCode = "colorado.hintAvailable"
	} else {
		resObj.MessageCode = "colorado.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *ColoradoWebPresenter) ActionLogOutput(c interfaces.ColoradoGame) string {
	return actionLogOutputJSON(c)
}
