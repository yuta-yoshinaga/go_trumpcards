//go:build !js || !wasm || extra

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MatrimonyWebPresenter マトリモニー Web プレゼンタークラス
type MatrimonyWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *MatrimonyWebPresenter) Output(c interfaces.MatrimonyGame, lastErr error) string {
	resObj := new(controller.MatrimonyWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, c, int(c.GetPhase()))

	// タブローは 1 枠 1 枚。空き枠は null で送る。
	tableau := c.GetTableau()
	resObj.Tableau = make([]*controller.WebOutputCard, domain.MatrimonyTableauCnt)
	for i := range domain.MatrimonyTableauCnt {
		resObj.Tableau[i] = cardToOutput(tableau[i])
	}

	foundation := c.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.MatrimonyFoundationCnt)
	for i := range domain.MatrimonyFoundationCnt {
		pile := foundation[i]
		resObj.Foundation[i] = make([]*controller.WebOutputCard, len(pile))
		for j, card := range pile {
			resObj.Foundation[i][j] = cardToOutput(card)
		}
	}

	resObj.StockCount = c.GetStockCount()
	resObj.RedealCount = c.GetRedealCount()
	waste := c.GetWaste()
	resObj.Waste = make([]*controller.WebOutputCard, len(waste))
	for i, card := range waste {
		resObj.Waste[i] = cardToOutput(card)
	}

	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	if c.GetPhase() == domain.MatrimonyPhasePlaying {
		if hint := c.GetHint(); hint != nil {
			resObj.Hint = &controller.MatrimonyWebOutputHint{
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
		case domain.MatrimonyPhasePlaying:
			if c.IsStalemate() {
				resObj.MessageCode = "matrimony.stalemate"
			} else {
				resObj.MessageCode = "matrimony.playing"
			}
		case domain.MatrimonyPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", c.GetMoveCount())
			resObj.MessageCode = "matrimony.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", c.GetMoveCount())}
		case domain.MatrimonyPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "matrimony.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *MatrimonyWebPresenter) HintOutput(c interfaces.MatrimonyGame) string {
	hint := c.GetHint()
	resObj := new(controller.MatrimonyWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, c, int(c.GetPhase()))
	resObj.Tableau = make([]*controller.WebOutputCard, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)
	resObj.Waste = make([]*controller.WebOutputCard, 0)

	if hint != nil {
		resObj.Hint = &controller.MatrimonyWebOutputHint{
			FromZone: hint.FromZone,
			FromIdx:  hint.FromIdx,
			ToZone:   hint.ToZone,
			ToIdx:    hint.ToIdx,
		}
		resObj.MessageCode = "matrimony.hintAvailable"
	} else {
		resObj.MessageCode = "matrimony.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *MatrimonyWebPresenter) ActionLogOutput(c interfaces.MatrimonyGame) string {
	return actionLogOutputJSON(c)
}
