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
	resObj.Phase = int(s.GetPhase())
	resObj.MoveCount = s.GetMoveCount()
	resObj.StockCount = s.GetStockCount()
	resObj.CompletedSuits = s.GetCompletedSuits()
	resObj.CanUndo = s.CanUndo()
	resObj.Score = s.GetScore()
	resObj.Difficulty = int(s.GetDifficulty())
	resObj.IsStalemate = s.IsStalemate()
	resObj.UndoToEscape = s.UndoToEscape()

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

	// メッセージ
	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		phase := s.GetPhase()
		switch phase {
		case domain.SpiderPhasePlaying:
			if s.IsStalemate() {
				resObj.MessageCode = "spider.stalemate"
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
	resObj.Phase = int(s.GetPhase())
	resObj.MoveCount = s.GetMoveCount()
	resObj.StockCount = s.GetStockCount()
	resObj.CompletedSuits = s.GetCompletedSuits()
	resObj.CanUndo = s.CanUndo()
	resObj.Score = s.GetScore()
	resObj.Difficulty = int(s.GetDifficulty())
	resObj.IsStalemate = s.IsStalemate()
	resObj.UndoToEscape = s.UndoToEscape()
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
	phase := s.GetPhase()
	if phase == domain.SpiderPhasePlaying {
		return actionLogToJSON(nil)
	}
	return actionLogToJSON(s.GetActionLog())
}
