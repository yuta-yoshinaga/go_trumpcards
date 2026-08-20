//go:build !js || !wasm || extra4

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// AlaskaWebPresenter アラスカWebプレゼンタークラス
type AlaskaWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *AlaskaWebPresenter) Output(r interfaces.AlaskaGame, lastErr error) string {
	resObj := new(controller.AlaskaWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, r, int(r.GetPhase()))

	// タブロー
	tableau := r.GetTableau()
	resObj.Tableau = make([][]*controller.AlaskaWebOutputTableauCard, domain.AlaskaTableauCnt)
	for i := range domain.AlaskaTableauCnt {
		colCards := tableau[i]
		resObj.Tableau[i] = make([]*controller.AlaskaWebOutputTableauCard, len(colCards))
		for j, tc := range colCards {
			outTC := &controller.AlaskaWebOutputTableauCard{FaceUp: tc.FaceUp}
			if tc.FaceUp {
				outTC.Card = cardToOutput(tc.Card)
			}
			resObj.Tableau[i][j] = outTC
		}
	}

	// ファンデーション
	foundation := r.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.AlaskaFoundationCnt)
	for i := range domain.AlaskaFoundationCnt {
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
	if r.GetPhase() == domain.AlaskaPhasePlaying && !r.IsStalemate() {
		if hint := r.GetHint(); hint != nil {
			resObj.Hint = &controller.AlaskaWebOutputHint{
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
		phase := r.GetPhase()
		switch phase {
		case domain.AlaskaPhasePlaying:
			if r.IsStalemate() {
				if resObj.UndoToEscape > 0 {
					resObj.MessageCode = "alaska.stalemateWithEscape"
					resObj.MessageParams = map[string]string{"count": fmt.Sprintf("%d", resObj.UndoToEscape)}
				} else {
					resObj.MessageCode = "alaska.stalemate"
				}
			} else {
				resObj.MessageCode = "alaska.playing"
			}
		case domain.AlaskaPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", r.GetMoveCount())
			resObj.MessageCode = "alaska.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", r.GetMoveCount())}
		case domain.AlaskaPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "alaska.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *AlaskaWebPresenter) HintOutput(r interfaces.AlaskaGame) string {
	hint := r.GetHint()
	resObj := new(controller.AlaskaWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, r, int(r.GetPhase()))
	resObj.Tableau = make([][]*controller.AlaskaWebOutputTableauCard, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)

	if hint != nil {
		resObj.Hint = &controller.AlaskaWebOutputHint{
			FromCol:   hint.FromCol,
			CardIndex: hint.CardIndex,
			ToZone:    hint.ToZone,
			ToCol:     hint.ToCol,
		}
		resObj.MessageCode = "alaska.hintAvailable"
	} else {
		resObj.MessageCode = "alaska.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *AlaskaWebPresenter) ActionLogOutput(r interfaces.AlaskaGame) string {
	return actionLogOutputJSON(r)
}
