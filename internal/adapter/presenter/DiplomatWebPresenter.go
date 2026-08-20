//go:build !js || !wasm || extra

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// DiplomatWebPresenter ディプロマット Web プレゼンタークラス
type DiplomatWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *DiplomatWebPresenter) Output(c interfaces.DiplomatGame, lastErr error) string {
	resObj := new(controller.DiplomatWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, c, int(c.GetPhase()))

	tableau := c.GetTableau()
	resObj.Tableau = make([][]*controller.WebOutputCard, domain.DiplomatTableauCnt)
	resObj.TableauDeadEnd = make([]bool, domain.DiplomatTableauCnt)
	for i := range domain.DiplomatTableauCnt {
		pile := tableau[i]
		resObj.Tableau[i] = make([]*controller.WebOutputCard, len(pile))
		for j, card := range pile {
			resObj.Tableau[i][j] = cardToOutput(card)
		}
		if len(pile) > 0 {
			resObj.TableauDeadEnd[i] = domain.DiplomatIsDeadEndTop(pile[len(pile)-1])
		}
	}

	foundation := c.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.DiplomatFoundationCnt)
	for i := range domain.DiplomatFoundationCnt {
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

	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	if c.GetPhase() == domain.DiplomatPhasePlaying {
		if hint := c.GetHint(); hint != nil {
			resObj.Hint = &controller.DiplomatWebOutputHint{
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
		case domain.DiplomatPhasePlaying:
			if c.IsStalemate() {
				resObj.MessageCode = "diplomat.stalemate"
			} else {
				resObj.MessageCode = "diplomat.playing"
			}
		case domain.DiplomatPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", c.GetMoveCount())
			resObj.MessageCode = "diplomat.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", c.GetMoveCount())}
		case domain.DiplomatPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "diplomat.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *DiplomatWebPresenter) HintOutput(c interfaces.DiplomatGame) string {
	hint := c.GetHint()
	resObj := new(controller.DiplomatWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, c, int(c.GetPhase()))
	resObj.Tableau = make([][]*controller.WebOutputCard, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)
	resObj.Waste = make([]*controller.WebOutputCard, 0)

	if hint != nil {
		resObj.Hint = &controller.DiplomatWebOutputHint{
			FromZone: hint.FromZone,
			FromIdx:  hint.FromIdx,
			ToZone:   hint.ToZone,
			ToIdx:    hint.ToIdx,
		}
		resObj.MessageCode = "diplomat.hintAvailable"
	} else {
		resObj.MessageCode = "diplomat.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *DiplomatWebPresenter) ActionLogOutput(c interfaces.DiplomatGame) string {
	return actionLogOutputJSON(c)
}
