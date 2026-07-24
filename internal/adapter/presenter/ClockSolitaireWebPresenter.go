//go:build !js || !wasm || solo

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// ClockSolitaireWebPresenter クロックソリティアWebプレゼンタークラス
type ClockSolitaireWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (pr *ClockSolitaireWebPresenter) Output(g interfaces.ClockSolitaireGame, lastErr error) string {
	resObj := new(controller.ClockSolitaireWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.StepCount = g.GetStepCount()
	resObj.CanUndo = g.CanUndo()

	// パイル
	piles := g.GetPiles()
	resObj.Piles = make([][]*controller.ClockSolitaireWebOutputCard, domain.ClockSolitairePileCount)
	for i := range domain.ClockSolitairePileCount {
		pile := piles[i]
		resObj.Piles[i] = make([]*controller.ClockSolitaireWebOutputCard, len(pile))
		for j, pc := range pile {
			outCard := &controller.ClockSolitaireWebOutputCard{
				FaceUp: pc.FaceUp,
			}
			if pc.FaceUp {
				outCard.Card = cardToOutput(pc.Card)
			}
			resObj.Piles[i][j] = outCard
		}
	}

	// 表向き枚数
	fuc := g.GetFaceUpCount()
	resObj.FaceUpCount = make([]int, domain.ClockSolitairePileCount)
	copy(resObj.FaceUpCount, fuc[:])

	// 現在のカード
	if cc := g.GetCurrentCard(); cc != nil {
		resObj.CurrentCard = cardToOutput(cc)
	}

	// メッセージ
	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		phase := g.GetPhase()
		switch phase {
		case domain.ClockSolitairePhasePlaying:
			resObj.MessageCode = "clocksolitaire.playing"
		case domain.ClockSolitairePhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ ステップ数: %d", g.GetStepCount())
			resObj.MessageCode = "clocksolitaire.gameClear"
			resObj.MessageParams = map[string]string{"stepCount": fmt.Sprintf("%d", g.GetStepCount())}
		case domain.ClockSolitairePhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "clocksolitaire.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (pr *ClockSolitaireWebPresenter) ActionLogOutput(g interfaces.ClockSolitaireGame) string {
	return actionLogOutputJSON(g)
}
