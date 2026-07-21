//go:build !js || !wasm || solo

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// WaspWebPresenter ワスプWebプレゼンタークラス
type WaspWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *WaspWebPresenter) Output(s interfaces.WaspGame, lastErr error) string {
	resObj := p.buildBase(s)

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		switch s.GetPhase() {
		case domain.WaspPhasePlaying:
			if s.IsStalemate() {
				resObj.MessageCode = "wasp.stalemate"
			} else {
				resObj.MessageCode = "wasp.playing"
			}
		case domain.WaspPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", s.GetMoveCount())
			resObj.MessageCode = "wasp.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", s.GetMoveCount())}
		case domain.WaspPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "wasp.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *WaspWebPresenter) HintOutput(s interfaces.WaspGame) string {
	resObj := p.buildBase(s)
	hint := s.GetHint()
	if hint != nil {
		resObj.Hint = &controller.WaspWebOutputHint{
			FromCol:   hint.FromCol,
			CardIndex: hint.CardIndex,
			ToCol:     hint.ToCol,
		}
		resObj.MessageCode = "wasp.hintAvailable"
	} else {
		resObj.MessageCode = "wasp.noHint"
	}
	return marshalOrError(resObj)
}

// LegalMovesOutput は合法な移動先を返す。Web版は WaspPage の waspLegalTargets が
// クライアント側で移動先を算出・ハイライトするため、通常の状態JSONへ委譲する。
func (p *WaspWebPresenter) LegalMovesOutput(s interfaces.WaspGame, _ int) string {
	return p.Output(s, nil)
}

// ActionLogOutput 棋譜をJSON出力
func (p *WaspWebPresenter) ActionLogOutput(s interfaces.WaspGame) string {
	return actionLogOutputJSON(s)
}

// buildBase は共通フィールドを詰めたレスポンスオブジェクトを返す
func (p *WaspWebPresenter) buildBase(s interfaces.WaspGame) *controller.WaspWebOutput {
	resObj := new(controller.WaspWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, s, int(s.GetPhase()))
	resObj.StockCount = s.GetStockCount()
	resObj.CompletedSuits = s.GetCompletedSuits()

	tableau := s.GetTableau()
	resObj.Tableau = make([][]*controller.KlondikeWebOutputTableauCard, domain.WaspTableauCnt)
	for i := range domain.WaspTableauCnt {
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
