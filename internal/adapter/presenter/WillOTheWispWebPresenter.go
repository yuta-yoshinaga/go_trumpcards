//go:build !js || !wasm || solo

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// WillOTheWispWebPresenter ウィル・オ・ザ・ウィスプWebプレゼンタークラス
type WillOTheWispWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *WillOTheWispWebPresenter) Output(s interfaces.WillOTheWispGame, lastErr error) string {
	resObj := new(controller.WillOTheWispWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, s, int(s.GetPhase()))
	resObj.StockCount = s.GetStockCount()
	resObj.CompletedSuits = s.GetCompletedSuits()
	resObj.Score = s.GetScore()
	resObj.Scoring = controller.WillOTheWispWebOutputScoring{
		Start:       domain.WillOTheWispStartScore,
		MovePenalty: domain.WillOTheWispMovePenalty,
		SuitBonus:   domain.WillOTheWispSuitBonus,
	}

	tableau := s.GetTableau()
	resObj.Tableau = make([][]*controller.WillOTheWispWebOutputTableauCard, domain.WillOTheWispTableauCnt)
	for i := range domain.WillOTheWispTableauCnt {
		colCards := tableau[i]
		resObj.Tableau[i] = make([]*controller.WillOTheWispWebOutputTableauCard, len(colCards))
		for j, tc := range colCards {
			outTC := &controller.WillOTheWispWebOutputTableauCard{FaceUp: tc.FaceUp}
			if tc.FaceUp {
				outTC.Card = cardToOutput(tc.Card)
			}
			resObj.Tableau[i][j] = outTC
		}
	}

	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	if s.GetPhase() == domain.WillOTheWispPhasePlaying && !s.IsStalemate() {
		if hint := s.GetHint(); hint != nil {
			resObj.Hint = &controller.WillOTheWispWebOutputHint{
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
		case domain.WillOTheWispPhasePlaying:
			if s.IsStalemate() {
				if resObj.UndoToEscape > 0 {
					resObj.MessageCode = "willothewisp.stalemateWithEscape"
					resObj.MessageParams = map[string]string{"count": fmt.Sprintf("%d", resObj.UndoToEscape)}
				} else {
					resObj.MessageCode = "willothewisp.stalemate"
				}
			} else {
				resObj.MessageCode = "willothewisp.playing"
			}
		case domain.WillOTheWispPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d スコア: %d", s.GetMoveCount(), s.GetScore())
			resObj.MessageCode = "willothewisp.gameClear"
			resObj.MessageParams = map[string]string{
				"moveCount": fmt.Sprintf("%d", s.GetMoveCount()),
				"score":     fmt.Sprintf("%d", s.GetScore()),
			}
		case domain.WillOTheWispPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "willothewisp.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *WillOTheWispWebPresenter) HintOutput(s interfaces.WillOTheWispGame) string {
	hint := s.GetHint()
	resObj := new(controller.WillOTheWispWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, s, int(s.GetPhase()))
	resObj.StockCount = s.GetStockCount()
	resObj.CompletedSuits = s.GetCompletedSuits()
	resObj.Score = s.GetScore()
	resObj.Scoring = controller.WillOTheWispWebOutputScoring{
		Start:       domain.WillOTheWispStartScore,
		MovePenalty: domain.WillOTheWispMovePenalty,
		SuitBonus:   domain.WillOTheWispSuitBonus,
	}
	resObj.Tableau = make([][]*controller.WillOTheWispWebOutputTableauCard, 0)

	if hint != nil {
		resObj.Hint = &controller.WillOTheWispWebOutputHint{
			FromCol:   hint.FromCol,
			CardIndex: hint.CardIndex,
			ToCol:     hint.ToCol,
		}
		resObj.MessageCode = "willothewisp.hintAvailable"
	} else {
		resObj.MessageCode = "willothewisp.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *WillOTheWispWebPresenter) ActionLogOutput(s interfaces.WillOTheWispGame) string {
	return actionLogOutputJSON(s)
}
