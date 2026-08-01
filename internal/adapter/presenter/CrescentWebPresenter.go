//go:build !js || !wasm || solo

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// CrescentWebPresenter クレセント・ソリティアの Web プレゼンター。
type CrescentWebPresenter struct{}

// Output ゲーム状態を JSON 出力する。
func (p *CrescentWebPresenter) Output(cr interfaces.CrescentGame, lastErr error) string {
	resObj := new(controller.CrescentWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, cr, int(cr.GetPhase()))
	resObj.RedealsRemaining = cr.GetRedealsRemaining()

	tableau := cr.GetTableau()
	resObj.Tableau = make([][]*controller.CrescentWebOutputTableauCard, domain.CrescentTableauCnt)
	for i := range domain.CrescentTableauCnt {
		colCards := tableau[i]
		resObj.Tableau[i] = make([]*controller.CrescentWebOutputTableauCard, len(colCards))
		for j, tc := range colCards {
			outTC := &controller.CrescentWebOutputTableauCard{FaceUp: tc.FaceUp}
			if tc.FaceUp {
				outTC.Card = cardToOutput(tc.Card)
			}
			resObj.Tableau[i][j] = outTC
		}
	}

	foundation := cr.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.CrescentFoundationCnt)
	for i := range domain.CrescentFoundationCnt {
		pile := foundation[i]
		resObj.Foundation[i] = make([]*controller.WebOutputCard, len(pile))
		for j, c := range pile {
			resObj.Foundation[i][j] = cardToOutput(c)
		}
	}

	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	if cr.GetPhase() == domain.CrescentPhasePlaying && !cr.IsStalemate() {
		if hint := cr.GetHint(); hint != nil {
			resObj.Hint = &controller.CrescentWebOutputHint{
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
		case domain.CrescentPhasePlaying:
			if cr.IsStalemate() {
				resObj.MessageCode = "crescent.stalemate"
			} else {
				resObj.MessageCode = "crescent.playing"
			}
		case domain.CrescentPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", cr.GetMoveCount())
			resObj.MessageCode = "crescent.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", cr.GetMoveCount())}
		case domain.CrescentPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "crescent.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントを JSON 出力する。
func (p *CrescentWebPresenter) HintOutput(cr interfaces.CrescentGame) string {
	hint := cr.GetHint()
	resObj := new(controller.CrescentWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, cr, int(cr.GetPhase()))
	resObj.RedealsRemaining = cr.GetRedealsRemaining()
	resObj.Tableau = make([][]*controller.CrescentWebOutputTableauCard, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)

	if hint != nil {
		resObj.Hint = &controller.CrescentWebOutputHint{
			FromCol: hint.FromCol,
			ToZone:  hint.ToZone,
			ToCol:   hint.ToCol,
			Redeal:  hint.Redeal,
		}
		resObj.MessageCode = "crescent.hintAvailable"
	} else {
		resObj.MessageCode = "crescent.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜を JSON 出力する。
func (p *CrescentWebPresenter) ActionLogOutput(cr interfaces.CrescentGame) string {
	return actionLogOutputJSON(cr)
}
