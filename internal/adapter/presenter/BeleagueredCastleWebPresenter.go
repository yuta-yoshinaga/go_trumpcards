//go:build !js || !wasm || solo

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// BeleagueredCastleWebPresenter Beleaguered Castle Web プレゼンタークラス
type BeleagueredCastleWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *BeleagueredCastleWebPresenter) Output(bc interfaces.BeleagueredCastleGame, lastErr error) string {
	resObj := new(controller.BeleagueredCastleWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, bc, int(bc.GetPhase()))

	// タブロー — every card in Beleaguered Castle is face-up by rule, but we
	// surface the domain's FaceUp field rather than hardcoding it so a future
	// hidden-deal variant would not silently leak state through this presenter.
	tableau := bc.GetTableau()
	resObj.Tableau = make([][]*controller.BeleagueredCastleWebOutputTableauCard, domain.BeleagueredCastleTableauCnt)
	for i := range domain.BeleagueredCastleTableauCnt {
		colCards := tableau[i]
		resObj.Tableau[i] = make([]*controller.BeleagueredCastleWebOutputTableauCard, len(colCards))
		for j, tc := range colCards {
			resObj.Tableau[i][j] = &controller.BeleagueredCastleWebOutputTableauCard{
				Card:   cardToOutput(tc.Card),
				FaceUp: tc.FaceUp,
			}
		}
	}

	// ファンデーション
	foundation := bc.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.BeleagueredCastleFoundationCnt)
	for i := range domain.BeleagueredCastleFoundationCnt {
		pile := foundation[i]
		resObj.Foundation[i] = make([]*controller.WebOutputCard, len(pile))
		for j, c := range pile {
			resObj.Foundation[i][j] = cardToOutput(c)
		}
	}

	// メッセージ
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	if bc.GetPhase() == domain.BeleagueredCastlePhasePlaying && !bc.IsStalemate() {
		if hint := bc.GetHint(); hint != nil {
			resObj.Hint = &controller.BeleagueredCastleWebOutputHint{
				FromCol:   hint.FromCol,
				CardIndex: hint.CardIndex,
				ToZone:    hint.ToZone,
				ToCol:     hint.ToCol,
			}
		}
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		phase := bc.GetPhase()
		switch phase {
		case domain.BeleagueredCastlePhasePlaying:
			if bc.IsStalemate() {
				resObj.MessageCode = "beleagueredcastle.stalemate"
			} else {
				resObj.MessageCode = "beleagueredcastle.playing"
			}
		case domain.BeleagueredCastlePhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", bc.GetMoveCount())
			resObj.MessageCode = "beleagueredcastle.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", bc.GetMoveCount())}
		case domain.BeleagueredCastlePhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "beleagueredcastle.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *BeleagueredCastleWebPresenter) HintOutput(bc interfaces.BeleagueredCastleGame) string {
	hint := bc.GetHint()
	resObj := new(controller.BeleagueredCastleWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, bc, int(bc.GetPhase()))
	resObj.Tableau = make([][]*controller.BeleagueredCastleWebOutputTableauCard, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)

	if hint != nil {
		resObj.Hint = &controller.BeleagueredCastleWebOutputHint{
			FromCol:   hint.FromCol,
			CardIndex: hint.CardIndex,
			ToZone:    hint.ToZone,
			ToCol:     hint.ToCol,
		}
		resObj.MessageCode = "beleagueredcastle.hintAvailable"
	} else {
		resObj.MessageCode = "beleagueredcastle.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *BeleagueredCastleWebPresenter) ActionLogOutput(bc interfaces.BeleagueredCastleGame) string {
	return actionLogOutputJSON(bc)
}
