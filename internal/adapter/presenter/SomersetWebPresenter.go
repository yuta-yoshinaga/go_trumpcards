//go:build !js || !wasm || solo

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// SomersetWebPresenter Somerset Web プレゼンタークラス
type SomersetWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *SomersetWebPresenter) Output(bc interfaces.SomersetGame, lastErr error) string {
	resObj := new(controller.SomersetWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, bc, int(bc.GetPhase()))

	// タブロー — every card in Somerset is face-up by rule, but we
	// surface the domain's FaceUp field rather than hardcoding it so a future
	// hidden-deal variant would not silently leak state through this presenter.
	tableau := bc.GetTableau()
	resObj.Tableau = make([][]*controller.SomersetWebOutputTableauCard, domain.SomersetTableauCnt)
	for i := range domain.SomersetTableauCnt {
		colCards := tableau[i]
		resObj.Tableau[i] = make([]*controller.SomersetWebOutputTableauCard, len(colCards))
		for j, tc := range colCards {
			resObj.Tableau[i][j] = &controller.SomersetWebOutputTableauCard{
				Card:   cardToOutput(tc.Card),
				FaceUp: tc.FaceUp,
			}
		}
	}

	// ファンデーション
	foundation := bc.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.SomersetFoundationCnt)
	for i := range domain.SomersetFoundationCnt {
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
	if bc.GetPhase() == domain.SomersetPhasePlaying && !bc.IsStalemate() {
		if hint := bc.GetHint(); hint != nil {
			resObj.Hint = &controller.SomersetWebOutputHint{
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
		case domain.SomersetPhasePlaying:
			if bc.IsStalemate() {
				resObj.MessageCode = "somerset.stalemate"
			} else {
				resObj.MessageCode = "somerset.playing"
			}
		case domain.SomersetPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", bc.GetMoveCount())
			resObj.MessageCode = "somerset.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", bc.GetMoveCount())}
		case domain.SomersetPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "somerset.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *SomersetWebPresenter) HintOutput(bc interfaces.SomersetGame) string {
	hint := bc.GetHint()
	resObj := new(controller.SomersetWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, bc, int(bc.GetPhase()))
	resObj.Tableau = make([][]*controller.SomersetWebOutputTableauCard, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)

	if hint != nil {
		resObj.Hint = &controller.SomersetWebOutputHint{
			FromCol:   hint.FromCol,
			CardIndex: hint.CardIndex,
			ToZone:    hint.ToZone,
			ToCol:     hint.ToCol,
		}
		resObj.MessageCode = "somerset.hintAvailable"
	} else {
		resObj.MessageCode = "somerset.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *SomersetWebPresenter) ActionLogOutput(bc interfaces.SomersetGame) string {
	return actionLogOutputJSON(bc)
}
