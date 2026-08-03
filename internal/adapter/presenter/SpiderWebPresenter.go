//go:build !js || !wasm || solo

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// SpiderWebPresenter スパイダーソリティアWebプレゼンタークラス
type SpiderWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *SpiderWebPresenter) Output(s interfaces.SpiderGame, lastErr error) string {
	resObj := new(controller.SpiderWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, s, int(s.GetPhase()))
	resObj.StockCount = s.GetStockCount()
	resObj.CompletedSuits = s.GetCompletedSuits()
	resObj.Score = s.GetScore()
	resObj.Difficulty = int(s.GetDifficulty())

	// タブロー
	tableau := s.GetTableau()
	resObj.Tableau = make([][]*controller.SpiderWebOutputTableauCard, domain.SpiderTableauCnt)
	for i := range domain.SpiderTableauCnt {
		colCards := tableau[i]
		resObj.Tableau[i] = make([]*controller.SpiderWebOutputTableauCard, len(colCards))
		for j, tc := range colCards {
			outTC := &controller.SpiderWebOutputTableauCard{FaceUp: tc.FaceUp}
			if tc.FaceUp {
				outTC.Card = cardToOutput(tc.Card)
			}
			resObj.Tableau[i][j] = outTC
		}
	}

	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	if s.GetPhase() == domain.SpiderPhasePlaying && !s.IsStalemate() {
		if hint := s.GetHint(); hint != nil {
			resObj.Hint = &controller.SpiderWebOutputHint{
				FromCol:   hint.FromCol,
				CardIndex: hint.CardIndex,
				ToCol:     hint.ToCol,
			}
		}
	}

	// メッセージ
	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		phase := s.GetPhase()
		switch phase {
		case domain.SpiderPhasePlaying:
			if s.IsStalemate() {
				if resObj.UndoToEscape > 0 {
					resObj.MessageCode = "spider.stalemateWithEscape"
					resObj.MessageParams = map[string]string{"count": fmt.Sprintf("%d", resObj.UndoToEscape)}
				} else {
					resObj.MessageCode = "spider.stalemate"
				}
			} else {
				resObj.MessageCode = "spider.playing"
			}
		case domain.SpiderPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d スコア: %d", s.GetMoveCount(), s.GetScore())
			resObj.MessageCode = "spider.gameClear"
			resObj.MessageParams = map[string]string{
				"moveCount": fmt.Sprintf("%d", s.GetMoveCount()),
				"score":     fmt.Sprintf("%d", s.GetScore()),
			}
		case domain.SpiderPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "spider.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *SpiderWebPresenter) HintOutput(s interfaces.SpiderGame) string {
	hint := s.GetHint()
	resObj := new(controller.SpiderWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, s, int(s.GetPhase()))
	resObj.StockCount = s.GetStockCount()
	resObj.CompletedSuits = s.GetCompletedSuits()
	resObj.Score = s.GetScore()
	resObj.Difficulty = int(s.GetDifficulty())
	resObj.Tableau = make([][]*controller.SpiderWebOutputTableauCard, 0)

	if hint != nil {
		resObj.Hint = &controller.SpiderWebOutputHint{
			FromCol:   hint.FromCol,
			CardIndex: hint.CardIndex,
			ToCol:     hint.ToCol,
		}
		resObj.MessageCode = "spider.hintAvailable"
	} else {
		resObj.MessageCode = "spider.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *SpiderWebPresenter) ActionLogOutput(s interfaces.SpiderGame) string {
	return actionLogOutputJSON(s)
}
