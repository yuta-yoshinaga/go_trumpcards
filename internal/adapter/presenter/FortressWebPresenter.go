//go:build !js || !wasm || solo

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// FortressWebPresenter Fortress Web プレゼンタークラス
type FortressWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *FortressWebPresenter) Output(bc interfaces.FortressGame, lastErr error) string {
	resObj := new(controller.FortressWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, bc, int(bc.GetPhase()))

	// タブロー — every card in Fortress is face-up by rule, but we
	// surface the domain's FaceUp field rather than hardcoding it so a future
	// hidden-deal variant would not silently leak state through this presenter.
	tableau := bc.GetTableau()
	resObj.Tableau = make([][]*controller.FortressWebOutputTableauCard, domain.FortressTableauCnt)
	for i := range domain.FortressTableauCnt {
		colCards := tableau[i]
		resObj.Tableau[i] = make([]*controller.FortressWebOutputTableauCard, len(colCards))
		for j, tc := range colCards {
			resObj.Tableau[i][j] = &controller.FortressWebOutputTableauCard{
				Card:   cardToOutput(tc.Card),
				FaceUp: tc.FaceUp,
			}
		}
	}

	// ファンデーション
	foundation := bc.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.FortressFoundationCnt)
	for i := range domain.FortressFoundationCnt {
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
	if bc.GetPhase() == domain.FortressPhasePlaying && !bc.IsStalemate() {
		if hint := bc.GetHint(); hint != nil {
			resObj.Hint = &controller.FortressWebOutputHint{
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
		case domain.FortressPhasePlaying:
			if bc.IsStalemate() {
				resObj.MessageCode = "fortress.stalemate"
			} else {
				resObj.MessageCode = "fortress.playing"
			}
		case domain.FortressPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", bc.GetMoveCount())
			resObj.MessageCode = "fortress.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", bc.GetMoveCount())}
		case domain.FortressPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "fortress.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *FortressWebPresenter) HintOutput(bc interfaces.FortressGame) string {
	hint := bc.GetHint()
	resObj := new(controller.FortressWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, bc, int(bc.GetPhase()))
	resObj.Tableau = make([][]*controller.FortressWebOutputTableauCard, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)

	if hint != nil {
		resObj.Hint = &controller.FortressWebOutputHint{
			FromCol:   hint.FromCol,
			CardIndex: hint.CardIndex,
			ToZone:    hint.ToZone,
			ToCol:     hint.ToCol,
		}
		resObj.MessageCode = "fortress.hintAvailable"
	} else {
		resObj.MessageCode = "fortress.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *FortressWebPresenter) ActionLogOutput(bc interfaces.FortressGame) string {
	return actionLogOutputJSON(bc)
}
