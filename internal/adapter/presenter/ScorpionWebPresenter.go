//go:build !js || !wasm || solo

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// ScorpionWebPresenter スコーピオンWebプレゼンタークラス
type ScorpionWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *ScorpionWebPresenter) Output(s interfaces.ScorpionGame, lastErr error) string {
	resObj := p.buildBase(s)

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		switch s.GetPhase() {
		case domain.ScorpionPhasePlaying:
			if s.IsStalemate() {
				resObj.MessageCode = "scorpion.stalemate"
			} else {
				resObj.MessageCode = "scorpion.playing"
			}
		case domain.ScorpionPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", s.GetMoveCount())
			resObj.MessageCode = "scorpion.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", s.GetMoveCount())}
		case domain.ScorpionPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "scorpion.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *ScorpionWebPresenter) HintOutput(s interfaces.ScorpionGame) string {
	resObj := p.buildBase(s)
	hint := s.GetHint()
	if hint != nil {
		resObj.Hint = &controller.ScorpionWebOutputHint{
			FromCol:   hint.FromCol,
			CardIndex: hint.CardIndex,
			ToCol:     hint.ToCol,
		}
		resObj.MessageCode = "scorpion.hintAvailable"
	} else {
		resObj.MessageCode = "scorpion.noHint"
	}
	return marshalOrError(resObj)
}

// LegalMovesOutput は合法な移動先を返す。Web版は ScorpionPage が
// scorpionLegalTargets でクライアント側に移動先を算出・ハイライトするため、
// 通常の状態JSONへ委譲する。
func (p *ScorpionWebPresenter) LegalMovesOutput(s interfaces.ScorpionGame, _ int) string {
	return p.Output(s, nil)
}

// ActionLogOutput 棋譜をJSON出力
func (p *ScorpionWebPresenter) ActionLogOutput(s interfaces.ScorpionGame) string {
	return actionLogOutputJSON(s)
}

// buildBase は共通フィールドを詰めたレスポンスオブジェクトを返す
func (p *ScorpionWebPresenter) buildBase(s interfaces.ScorpionGame) *controller.ScorpionWebOutput {
	resObj := new(controller.ScorpionWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, s, int(s.GetPhase()))
	resObj.StockCount = s.GetStockCount()
	resObj.CompletedSuits = s.GetCompletedSuits()

	tableau := s.GetTableau()
	resObj.Tableau = make([][]*controller.KlondikeWebOutputTableauCard, domain.ScorpionTableauCnt)
	for i := range domain.ScorpionTableauCnt {
		colCards := tableau[i]
		resObj.Tableau[i] = make([]*controller.KlondikeWebOutputTableauCard, len(colCards))
		for j, tc := range colCards {
			outTC := &controller.KlondikeWebOutputTableauCard{FaceUp: tc.FaceUp}
			if tc.FaceUp {
				outTC.Card = cardToOutput(tc.Card)
			}
			resObj.Tableau[i][j] = outTC
		}
	}
	return resObj
}
