//go:build !js || !wasm || solo

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// EightOffWebPresenter エイトオフWebプレゼンタークラス
type EightOffWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *EightOffWebPresenter) Output(e interfaces.EightOffGame, lastErr error) string {
	resObj := new(controller.EightOffWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, e, int(e.GetPhase()))

	// タブロー
	tableau := e.GetTableau()
	resObj.Tableau = make([][]*controller.WebOutputCard, domain.EightOffTableauCnt)
	for i := 0; i < domain.EightOffTableauCnt; i++ {
		colCards := tableau[i]
		resObj.Tableau[i] = make([]*controller.WebOutputCard, len(colCards))
		for j, c := range colCards {
			resObj.Tableau[i][j] = cardToOutput(c)
		}
	}

	// フリーセル
	freeCells := e.GetFreeCells()
	resObj.FreeCells = make([]*controller.WebOutputCard, domain.EightOffCellCnt)
	for i := 0; i < domain.EightOffCellCnt; i++ {
		resObj.FreeCells[i] = cardToOutput(freeCells[i])
	}

	// ファンデーション
	foundation := e.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.EightOffFoundationCnt)
	for i := 0; i < domain.EightOffFoundationCnt; i++ {
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
	if e.GetPhase() == domain.EightOffPhasePlaying && !e.IsStalemate() {
		if hint := e.GetHint(); hint != nil {
			resObj.Hint = &controller.EightOffWebOutputHint{
				FromZone:  hint.FromZone,
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
		phase := e.GetPhase()
		switch phase {
		case domain.EightOffPhasePlaying:
			if e.IsStalemate() {
				if resObj.UndoToEscape > 0 {
					resObj.MessageCode = "eightoff.stalemateWithEscape"
					resObj.MessageParams = map[string]string{"count": fmt.Sprintf("%d", resObj.UndoToEscape)}
				} else {
					resObj.MessageCode = "eightoff.stalemate"
				}
			} else {
				resObj.MessageCode = "eightoff.playing"
			}
		case domain.EightOffPhaseGameClear:
			resObj.MessageCode = "eightoff.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", e.GetMoveCount())}
		case domain.EightOffPhaseGameOver:
			resObj.MessageCode = "eightoff.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *EightOffWebPresenter) HintOutput(e interfaces.EightOffGame) string {
	hint := e.GetHint()
	resObj := new(controller.EightOffWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, e, int(e.GetPhase()))
	resObj.FreeCells = make([]*controller.WebOutputCard, 0)
	resObj.Tableau = make([][]*controller.WebOutputCard, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)

	if hint != nil {
		resObj.Hint = &controller.EightOffWebOutputHint{
			FromZone:  hint.FromZone,
			FromCol:   hint.FromCol,
			CardIndex: hint.CardIndex,
			ToZone:    hint.ToZone,
			ToCol:     hint.ToCol,
		}
		resObj.MessageCode = "eightoff.hintAvailable"
	} else {
		resObj.MessageCode = "eightoff.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *EightOffWebPresenter) ActionLogOutput(e interfaces.EightOffGame) string {
	return actionLogOutputJSON(e)
}
