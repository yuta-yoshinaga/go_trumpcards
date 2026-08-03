//go:build !js || !wasm || extra2

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// BraidWebPresenter ブレイド Web プレゼンタークラス
type BraidWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *BraidWebPresenter) Output(b interfaces.BraidGame, lastErr error) string {
	resObj := new(controller.BraidWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, b, int(b.GetPhase()))

	braid := b.GetBraid()
	resObj.Braid = make([]*controller.WebOutputCard, len(braid))
	for i, card := range braid {
		resObj.Braid[i] = cardToOutput(card)
	}

	// 空き枠は null で送る。詰めてしまうとインデックスがずれ、ヒントの枠番号が
	// 画面と食い違う。
	fields := b.GetFields()
	resObj.Fields = make([]*controller.WebOutputCard, domain.BraidFieldCnt)
	for i := range domain.BraidFieldCnt {
		resObj.Fields[i] = cardToOutput(fields[i])
	}

	helpers := b.GetHelpers()
	resObj.Helpers = make([]*controller.WebOutputCard, domain.BraidHelperCnt)
	for i := range domain.BraidHelperCnt {
		resObj.Helpers[i] = cardToOutput(helpers[i])
	}

	foundation := b.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.BraidFoundationCnt)
	for i := range domain.BraidFoundationCnt {
		pile := foundation[i]
		resObj.Foundation[i] = make([]*controller.WebOutputCard, len(pile))
		for j, card := range pile {
			resObj.Foundation[i][j] = cardToOutput(card)
		}
	}

	resObj.StockCount = b.GetStockCount()
	waste := b.GetWaste()
	resObj.Waste = make([]*controller.WebOutputCard, len(waste))
	for i, card := range waste {
		resObj.Waste[i] = cardToOutput(card)
	}
	// direction と awaitingDirection を両方渡すのは、クライアントに「0 は未選択」
	// という約束を再実装させないため。
	resObj.BaseRank = b.GetBaseRank()
	resObj.Direction = int(b.GetDirection())
	resObj.AwaitingDirection = b.IsAwaitingDirection()
	resObj.RedealsLeft = domain.BraidMaxPasses - 1 - b.GetPassesUsed()
	resObj.CanRedeal = b.CanRedeal()

	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// 手が進められる局面だけで計算する。終局・向き待ち・手詰まりでは助言する手が
	// 無いので、探索を走らせるだけ無駄になる。
	if b.GetPhase() == domain.BraidPhasePlaying && !b.IsAwaitingDirection() && !b.IsStalemate() {
		if hint := b.GetHint(); hint != nil {
			resObj.Hint = &controller.BraidWebOutputHint{
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
		switch b.GetPhase() {
		case domain.BraidPhasePlaying:
			switch {
			case b.IsAwaitingDirection():
				resObj.MessageCode = "braid.chooseDirection"
			case b.IsStalemate():
				resObj.MessageCode = "braid.stalemate"
			default:
				resObj.MessageCode = "braid.playing"
			}
		case domain.BraidPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", b.GetMoveCount())
			resObj.MessageCode = "braid.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", b.GetMoveCount())}
		case domain.BraidPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "braid.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *BraidWebPresenter) HintOutput(b interfaces.BraidGame) string {
	hint := b.GetHint()
	resObj := new(controller.BraidWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, b, int(b.GetPhase()))
	resObj.Braid = make([]*controller.WebOutputCard, 0)
	resObj.Fields = make([]*controller.WebOutputCard, 0)
	resObj.Helpers = make([]*controller.WebOutputCard, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)
	resObj.Waste = make([]*controller.WebOutputCard, 0)

	if hint != nil {
		resObj.Hint = &controller.BraidWebOutputHint{
			FromZone: hint.FromZone,
			FromIdx:  hint.FromIdx,
			ToZone:   hint.ToZone,
			ToIdx:    hint.ToIdx,
		}
		resObj.MessageCode = "braid.hintAvailable"
	} else {
		resObj.MessageCode = "braid.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *BraidWebPresenter) ActionLogOutput(b interfaces.BraidGame) string {
	return actionLogOutputJSON(b)
}
