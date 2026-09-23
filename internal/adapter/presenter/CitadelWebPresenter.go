//go:build !js || !wasm || solo

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// CitadelWebPresenter Citadel Web プレゼンタークラス
type CitadelWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *CitadelWebPresenter) Output(bc interfaces.CitadelGame, lastErr error) string {
	resObj := new(controller.CitadelWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, bc, int(bc.GetPhase()))

	// タブロー — every card in Citadel is face-up by rule, but we
	// surface the domain's FaceUp field rather than hardcoding it so a future
	// hidden-deal variant would not silently leak state through this presenter.
	tableau := bc.GetTableau()
	resObj.Tableau = make([][]*controller.CitadelWebOutputTableauCard, domain.CitadelTableauCnt)
	for i := range domain.CitadelTableauCnt {
		colCards := tableau[i]
		resObj.Tableau[i] = make([]*controller.CitadelWebOutputTableauCard, len(colCards))
		for j, tc := range colCards {
			resObj.Tableau[i][j] = &controller.CitadelWebOutputTableauCard{
				Card:   cardToOutput(tc.Card),
				FaceUp: tc.FaceUp,
			}
		}
	}

	// ファンデーション
	foundation := bc.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.CitadelFoundationCnt)
	for i := range domain.CitadelFoundationCnt {
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
	if bc.GetPhase() == domain.CitadelPhasePlaying && !bc.IsStalemate() {
		if hint := bc.GetHint(); hint != nil {
			resObj.Hint = &controller.CitadelWebOutputHint{
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
		case domain.CitadelPhasePlaying:
			if bc.IsStalemate() {
				resObj.MessageCode = "citadel.stalemate"
			} else {
				resObj.MessageCode = "citadel.playing"
			}
		case domain.CitadelPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", bc.GetMoveCount())
			resObj.MessageCode = "citadel.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", bc.GetMoveCount())}
		case domain.CitadelPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "citadel.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *CitadelWebPresenter) HintOutput(bc interfaces.CitadelGame) string {
	hint := bc.GetHint()
	resObj := new(controller.CitadelWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, bc, int(bc.GetPhase()))
	resObj.Tableau = make([][]*controller.CitadelWebOutputTableauCard, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)

	if hint != nil {
		resObj.Hint = &controller.CitadelWebOutputHint{
			FromCol:   hint.FromCol,
			CardIndex: hint.CardIndex,
			ToZone:    hint.ToZone,
			ToCol:     hint.ToCol,
		}
		resObj.MessageCode = "citadel.hintAvailable"
	} else {
		resObj.MessageCode = "citadel.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *CitadelWebPresenter) ActionLogOutput(bc interfaces.CitadelGame) string {
	return actionLogOutputJSON(bc)
}
