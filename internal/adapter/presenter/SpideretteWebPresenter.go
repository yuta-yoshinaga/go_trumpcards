//go:build !js || !wasm || solo

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// SpideretteWebPresenter スパイダレットWebプレゼンタークラス
type SpideretteWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *SpideretteWebPresenter) Output(s interfaces.SpideretteGame, lastErr error) string {
	resObj := new(controller.SpideretteWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, s, int(s.GetPhase()))
	resObj.StockCount = s.GetStockCount()
	resObj.CompletedSuits = s.GetCompletedSuits()
	resObj.Score = s.GetScore()

	tableau := s.GetTableau()
	resObj.Tableau = make([][]*controller.SpideretteWebOutputTableauCard, domain.SpideretteTableauCnt)
	for i := range domain.SpideretteTableauCnt {
		colCards := tableau[i]
		resObj.Tableau[i] = make([]*controller.SpideretteWebOutputTableauCard, len(colCards))
		for j, tc := range colCards {
			outTC := &controller.SpideretteWebOutputTableauCard{FaceUp: tc.FaceUp}
			if tc.FaceUp {
				outTC.Card = cardToOutput(tc.Card)
			}
			resObj.Tableau[i][j] = outTC
		}
	}

	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	if s.GetPhase() == domain.SpiderettePhasePlaying && !s.IsStalemate() {
		if hint := s.GetHint(); hint != nil {
			resObj.Hint = &controller.SpideretteWebOutputHint{
				FromCol:   hint.FromCol,
				CardIndex: hint.CardIndex,
				ToCol:     hint.ToCol,
			}
		}
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		phase := s.GetPhase()
		switch phase {
		case domain.SpiderettePhasePlaying:
			if s.IsStalemate() {
				if resObj.UndoToEscape > 0 {
					resObj.MessageCode = "spiderette.stalemateWithEscape"
					resObj.MessageParams = map[string]string{"count": fmt.Sprintf("%d", resObj.UndoToEscape)}
				} else {
					resObj.MessageCode = "spiderette.stalemate"
				}
			} else {
				resObj.MessageCode = "spiderette.playing"
			}
		case domain.SpiderettePhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d スコア: %d", s.GetMoveCount(), s.GetScore())
			resObj.MessageCode = "spiderette.gameClear"
			resObj.MessageParams = map[string]string{
				"moveCount": fmt.Sprintf("%d", s.GetMoveCount()),
				"score":     fmt.Sprintf("%d", s.GetScore()),
			}
		case domain.SpiderettePhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "spiderette.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *SpideretteWebPresenter) HintOutput(s interfaces.SpideretteGame) string {
	hint := s.GetHint()
	resObj := new(controller.SpideretteWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, s, int(s.GetPhase()))
	resObj.StockCount = s.GetStockCount()
	resObj.CompletedSuits = s.GetCompletedSuits()
	resObj.Score = s.GetScore()
	resObj.Tableau = make([][]*controller.SpideretteWebOutputTableauCard, 0)

	if hint != nil {
		resObj.Hint = &controller.SpideretteWebOutputHint{
			FromCol:   hint.FromCol,
			CardIndex: hint.CardIndex,
			ToCol:     hint.ToCol,
		}
		resObj.MessageCode = "spiderette.hintAvailable"
	} else {
		resObj.MessageCode = "spiderette.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *SpideretteWebPresenter) ActionLogOutput(s interfaces.SpideretteGame) string {
	return actionLogOutputJSON(s)
}
