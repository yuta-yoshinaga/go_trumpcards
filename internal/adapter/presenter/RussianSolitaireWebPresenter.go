//go:build !js || !wasm || solo

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// RussianSolitaireWebPresenter ロシアンソリティアWebプレゼンタークラス
type RussianSolitaireWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *RussianSolitaireWebPresenter) Output(r interfaces.RussianSolitaireGame, lastErr error) string {
	resObj := new(controller.RussianSolitaireWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, r, int(r.GetPhase()))

	// タブロー
	tableau := r.GetTableau()
	resObj.Tableau = make([][]*controller.KlondikeWebOutputTableauCard, domain.RussianSolitaireTableauCnt)
	for i := range domain.RussianSolitaireTableauCnt {
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

	// ファンデーション
	foundation := r.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.RussianSolitaireFoundationCnt)
	for i := range domain.RussianSolitaireFoundationCnt {
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
	if r.GetPhase() == domain.RussianSolitairePhasePlaying && !r.IsStalemate() {
		if hint := r.GetHint(); hint != nil {
			resObj.Hint = &controller.RussianSolitaireWebOutputHint{
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
		case domain.RussianSolitairePhasePlaying:
			if r.IsStalemate() {
				if resObj.UndoToEscape > 0 {
					resObj.MessageCode = "russiansolitaire.stalemateWithEscape"
					resObj.MessageParams = map[string]string{"count": fmt.Sprintf("%d", resObj.UndoToEscape)}
				} else {
					resObj.MessageCode = "russiansolitaire.stalemate"
				}
			} else {
				resObj.MessageCode = "russiansolitaire.playing"
			}
		case domain.RussianSolitairePhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", r.GetMoveCount())
			resObj.MessageCode = "russiansolitaire.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", r.GetMoveCount())}
		case domain.RussianSolitairePhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "russiansolitaire.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *RussianSolitaireWebPresenter) HintOutput(r interfaces.RussianSolitaireGame) string {
	hint := r.GetHint()
	resObj := new(controller.RussianSolitaireWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, r, int(r.GetPhase()))
	resObj.Tableau = make([][]*controller.KlondikeWebOutputTableauCard, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)

	if hint != nil {
		resObj.Hint = &controller.RussianSolitaireWebOutputHint{
			FromCol:   hint.FromCol,
			CardIndex: hint.CardIndex,
			ToZone:    hint.ToZone,
			ToCol:     hint.ToCol,
		}
		resObj.MessageCode = "russiansolitaire.hintAvailable"
	} else {
		resObj.MessageCode = "russiansolitaire.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *RussianSolitaireWebPresenter) ActionLogOutput(r interfaces.RussianSolitaireGame) string {
	return actionLogOutputJSON(r)
}
