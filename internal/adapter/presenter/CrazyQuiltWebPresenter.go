//go:build !js || !wasm || solo

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// CrazyQuiltWebPresenter クレイジーキルト Web プレゼンタークラス
type CrazyQuiltWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *CrazyQuiltWebPresenter) Output(c interfaces.CrazyQuiltGame, lastErr error) string {
	resObj := new(controller.CrazyQuiltWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, c, int(c.GetPhase()))

	// キルトはマス番号順のカード配列。取り除いたマスは null。
	quilt := c.GetQuilt()
	resObj.Quilt = make([]*controller.WebOutputCard, domain.CrazyQuiltCells)
	// **取れるかどうかはサーバが計算して送る。**短辺の露出は向きに依存する
	// ので、フロントで再実装すると規則が 2 か所に散る。
	resObj.Available = make([]bool, domain.CrazyQuiltCells)
	for i := range domain.CrazyQuiltCells {
		resObj.Quilt[i] = cardToOutput(quilt[i])
		resObj.Available[i] = c.IsAvailable(i)
	}
	resObj.RedealsLeft = c.GetRedealsLeft()
	resObj.FoundationAscending = make([]bool, domain.CrazyQuiltFoundationCnt)
	for i := range domain.CrazyQuiltFoundationCnt {
		resObj.FoundationAscending[i] = c.IsAscendingFoundation(i)
	}

	foundation := c.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.CrazyQuiltFoundationCnt)
	for i := range domain.CrazyQuiltFoundationCnt {
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
	if c.GetPhase() == domain.CrazyQuiltPhasePlaying {
		if hint := c.GetHint(); hint != nil {
			resObj.Hint = &controller.CrazyQuiltWebOutputHint{
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
		case domain.CrazyQuiltPhasePlaying:
			if c.IsStalemate() {
				resObj.MessageCode = "crazyquilt.stalemate"
			} else {
				resObj.MessageCode = "crazyquilt.playing"
			}
		case domain.CrazyQuiltPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", c.GetMoveCount())
			resObj.MessageCode = "crazyquilt.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", c.GetMoveCount())}
		case domain.CrazyQuiltPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "crazyquilt.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *CrazyQuiltWebPresenter) HintOutput(c interfaces.CrazyQuiltGame) string {
	hint := c.GetHint()
	resObj := new(controller.CrazyQuiltWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, c, int(c.GetPhase()))
	resObj.Quilt = make([]*controller.WebOutputCard, 0)
	resObj.Available = make([]bool, 0)
	resObj.FoundationAscending = make([]bool, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)
	resObj.Waste = make([]*controller.WebOutputCard, 0)

	if hint != nil {
		resObj.Hint = &controller.CrazyQuiltWebOutputHint{
			FromZone: hint.FromZone,
			FromIdx:  hint.FromIdx,
			ToZone:   hint.ToZone,
			ToIdx:    hint.ToIdx,
		}
		resObj.MessageCode = "crazyquilt.hintAvailable"
	} else {
		resObj.MessageCode = "crazyquilt.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *CrazyQuiltWebPresenter) ActionLogOutput(c interfaces.CrazyQuiltGame) string {
	return actionLogOutputJSON(c)
}
