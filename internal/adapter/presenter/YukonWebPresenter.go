//go:build !js || !wasm || solo

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// YukonWebPresenter ユーコンWebプレゼンタークラス
type YukonWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *YukonWebPresenter) Output(y interfaces.YukonGame, lastErr error) string {
	resObj := new(controller.YukonWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, y, int(y.GetPhase()))

	// タブロー
	tableau := y.GetTableau()
	resObj.Tableau = make([][]*controller.KlondikeWebOutputTableauCard, domain.YukonTableauCnt)
	for i := range domain.YukonTableauCnt {
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
	foundation := y.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.YukonFoundationCnt)
	for i := range domain.YukonFoundationCnt {
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
	if y.GetPhase() == domain.YukonPhasePlaying && !y.IsStalemate() {
		if hint := y.GetHint(); hint != nil {
			resObj.Hint = &controller.YukonWebOutputHint{
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
		phase := y.GetPhase()
		switch phase {
		case domain.YukonPhasePlaying:
			if y.IsStalemate() {
				if resObj.UndoToEscape > 0 {
					resObj.MessageCode = "yukon.stalemateWithEscape"
					resObj.MessageParams = map[string]string{"count": fmt.Sprintf("%d", resObj.UndoToEscape)}
				} else {
					resObj.MessageCode = "yukon.stalemate"
				}
			} else {
				resObj.MessageCode = "yukon.playing"
			}
		case domain.YukonPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d", y.GetMoveCount())
			resObj.MessageCode = "yukon.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", y.GetMoveCount())}
		case domain.YukonPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "yukon.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *YukonWebPresenter) HintOutput(y interfaces.YukonGame) string {
	hint := y.GetHint()
	resObj := new(controller.YukonWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, y, int(y.GetPhase()))
	resObj.Tableau = make([][]*controller.KlondikeWebOutputTableauCard, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)

	if hint != nil {
		resObj.Hint = &controller.YukonWebOutputHint{
			FromCol:   hint.FromCol,
			CardIndex: hint.CardIndex,
			ToZone:    hint.ToZone,
			ToCol:     hint.ToCol,
		}
		resObj.MessageCode = "yukon.hintAvailable"
	} else {
		resObj.MessageCode = "yukon.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *YukonWebPresenter) ActionLogOutput(y interfaces.YukonGame) string {
	return actionLogOutputJSON(y)
}
