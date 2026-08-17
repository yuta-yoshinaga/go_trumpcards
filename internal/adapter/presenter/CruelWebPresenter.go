//go:build !js || !wasm || solo

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// CruelWebPresenter クルーエルWebプレゼンタークラス
type CruelWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *CruelWebPresenter) Output(c interfaces.CruelGame, lastErr error) string {
	resObj := new(controller.CruelWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, c, int(c.GetPhase()))

	// タブロー
	tableau := c.GetTableau()
	resObj.CanAutoComplete = c.CanAutoComplete()
	resObj.Tableau = make([][]*controller.KlondikeWebOutputTableauCard, domain.CruelTableauCnt)
	for i := range domain.CruelTableauCnt {
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

	// ファウンデーション
	foundation := c.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.CruelFoundationCnt)
	for i := range domain.CruelFoundationCnt {
		pile := foundation[i]
		resObj.Foundation[i] = make([]*controller.WebOutputCard, len(pile))
		for j, card := range pile {
			resObj.Foundation[i][j] = cardToOutput(card)
		}
	}

	// メッセージ
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	if c.GetPhase() == domain.CruelPhasePlaying && !c.IsStalemate() {
		if hint := c.GetHint(); hint != nil {
			resObj.Hint = &controller.CruelWebOutputHint{
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
		phase := c.GetPhase()
		switch phase {
		case domain.CruelPhasePlaying:
			if c.IsStalemate() {
				if resObj.UndoToEscape > 0 {
					resObj.MessageCode = "cruel.stalemateWithEscape"
					resObj.MessageParams = map[string]string{"count": fmt.Sprintf("%d", resObj.UndoToEscape)}
				} else {
					resObj.MessageCode = "cruel.stalemate"
				}
			} else {
				resObj.MessageCode = "cruel.playing"
			}
		case domain.CruelPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", c.GetMoveCount())
			resObj.MessageCode = "cruel.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", c.GetMoveCount())}
		case domain.CruelPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "cruel.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *CruelWebPresenter) HintOutput(c interfaces.CruelGame) string {
	hint := c.GetHint()
	resObj := new(controller.CruelWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, c, int(c.GetPhase()))
	resObj.Tableau = make([][]*controller.KlondikeWebOutputTableauCard, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)

	if hint != nil {
		resObj.Hint = &controller.CruelWebOutputHint{
			FromCol:   hint.FromCol,
			CardIndex: hint.CardIndex,
			ToZone:    hint.ToZone,
			ToCol:     hint.ToCol,
		}
		resObj.MessageCode = "cruel.hintAvailable"
	} else {
		resObj.MessageCode = "cruel.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *CruelWebPresenter) ActionLogOutput(c interfaces.CruelGame) string {
	return actionLogOutputJSON(c)
}
