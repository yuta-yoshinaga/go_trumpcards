//go:build !js || !wasm || extra

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// StHelenaWebPresenter セント・ヘレナ・ソリティアの Web プレゼンター。
type StHelenaWebPresenter struct{}

// Output ゲーム状態を JSON 出力する。
func (p *StHelenaWebPresenter) Output(cr interfaces.StHelenaGame, lastErr error) string {
	resObj := new(controller.StHelenaWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, cr, int(cr.GetPhase()))
	resObj.RedealsRemaining = cr.GetRedealsRemaining()
	resObj.RestrictionsActive = cr.RestrictionsActive()

	tableau := cr.GetTableau()
	resObj.Tableau = make([][]*controller.StHelenaWebOutputTableauCard, domain.StHelenaTableauCnt)
	for i := range domain.StHelenaTableauCnt {
		colCards := tableau[i]
		resObj.Tableau[i] = make([]*controller.StHelenaWebOutputTableauCard, len(colCards))
		for j, tc := range colCards {
			outTC := &controller.StHelenaWebOutputTableauCard{FaceUp: tc.FaceUp}
			if tc.FaceUp {
				outTC.Card = cardToOutput(tc.Card)
			}
			resObj.Tableau[i][j] = outTC
		}
	}

	foundation := cr.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.StHelenaFoundationCnt)
	for i := range domain.StHelenaFoundationCnt {
		pile := foundation[i]
		resObj.Foundation[i] = make([]*controller.WebOutputCard, len(pile))
		for j, c := range pile {
			resObj.Foundation[i][j] = cardToOutput(c)
		}
	}

	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	if cr.GetPhase() == domain.StHelenaPhasePlaying && !cr.IsStalemate() {
		if hint := cr.GetHint(); hint != nil {
			resObj.Hint = &controller.StHelenaWebOutputHint{
				FromCol: hint.FromCol,
				ToZone:  hint.ToZone,
				ToCol:   hint.ToCol,
				Redeal:  hint.Redeal,
			}
		}
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		switch cr.GetPhase() {
		case domain.StHelenaPhasePlaying:
			if cr.IsStalemate() {
				resObj.MessageCode = "sthelena.stalemate"
			} else {
				resObj.MessageCode = "sthelena.playing"
			}
		case domain.StHelenaPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", cr.GetMoveCount())
			resObj.MessageCode = "sthelena.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", cr.GetMoveCount())}
		case domain.StHelenaPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "sthelena.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントを JSON 出力する。
func (p *StHelenaWebPresenter) HintOutput(cr interfaces.StHelenaGame) string {
	hint := cr.GetHint()
	resObj := new(controller.StHelenaWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, cr, int(cr.GetPhase()))
	resObj.RedealsRemaining = cr.GetRedealsRemaining()
	resObj.RestrictionsActive = cr.RestrictionsActive()
	resObj.Tableau = make([][]*controller.StHelenaWebOutputTableauCard, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)

	if hint != nil {
		resObj.Hint = &controller.StHelenaWebOutputHint{
			FromCol: hint.FromCol,
			ToZone:  hint.ToZone,
			ToCol:   hint.ToCol,
			Redeal:  hint.Redeal,
		}
		resObj.MessageCode = "sthelena.hintAvailable"
	} else {
		resObj.MessageCode = "sthelena.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜を JSON 出力する。
func (p *StHelenaWebPresenter) ActionLogOutput(cr interfaces.StHelenaGame) string {
	return actionLogOutputJSON(cr)
}
