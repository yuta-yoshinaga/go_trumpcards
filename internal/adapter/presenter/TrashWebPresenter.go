package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// TrashWebPresenter トラッシュWebプレゼンタークラス
type TrashWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *TrashWebPresenter) Output(t interfaces.TrashGame, lastErr error) string {
	resObj := p.buildBase(t)

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		switch t.GetPhase() {
		case domain.TrashPhasePlayerTurn:
			if t.IsCpuTurn() {
				resObj.MessageCode = "trash.cpuTurn"
			} else {
				resObj.MessageCode = "trash.playerTurn"
			}
		case domain.TrashPhaseAwaitWild:
			if t.IsCpuTurn() {
				resObj.MessageCode = "trash.cpuAwaitWild"
			} else {
				resObj.MessageCode = "trash.awaitWild"
			}
		case domain.TrashPhaseGameOver:
			if t.GetWinner() == domain.TrashHumanIdx {
				resObj.MessageCode = "trash.gameOverWin"
			} else {
				resObj.MessageCode = "trash.gameOverLose"
			}
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", t.GetMoveCount())}
		}
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *TrashWebPresenter) ActionLogOutput(t interfaces.TrashGame) string {
	if t.GetPhase() != domain.TrashPhaseGameOver {
		return actionLogToJSON(nil)
	}
	return actionLogToJSON(t.GetActionLog())
}

// HintOutput ヒントを出力する。Web ではヒントはクライアント側 (useGameHint) で
// 算出するため、通常の状態出力を返す。TrashPresenter インタフェースを満たすための実装。
func (p *TrashWebPresenter) HintOutput(t interfaces.TrashGame) string {
	return p.Output(t, nil)
}

// buildBase は共通フィールドを詰めたレスポンスオブジェクトを返す
func (p *TrashWebPresenter) buildBase(t interfaces.TrashGame) *controller.TrashWebOutput {
	resObj := new(controller.TrashWebOutput)
	resObj.Phase = int(t.GetPhase())
	resObj.Current = t.GetCurrent()
	resObj.StockSize = t.GetStockSize()
	resObj.DiscardSize = t.GetDiscardSize()
	resObj.MoveCount = t.GetMoveCount()
	resObj.Winner = t.GetWinner()

	if top := t.GetDiscardTop(); top != nil {
		resObj.DiscardTop = cardToOutput(top)
	}
	if pending := t.GetPending(); pending != nil {
		resObj.Pending = cardToOutput(pending)
	}

	for i := range resObj.Players {
		slots := t.GetPlayerSlots(i)
		resObj.Players[i].IsCpu = t.IsCpuPlayer(i)
		for j := range resObj.Players[i].Slots {
			if j >= len(slots) {
				break
			}
			s := slots[j]
			resObj.Players[i].Slots[j].FaceUp = s.FaceUp
			// Hide face-down cards from the opponent AND from the active player — true to the real
			// game (cards stay face-down until flipped). The human sees only their own face-up slots.
			if s.FaceUp && s.Card != nil {
				resObj.Players[i].Slots[j].Card = cardToOutput(s.Card)
			}
		}
	}
	return resObj
}
