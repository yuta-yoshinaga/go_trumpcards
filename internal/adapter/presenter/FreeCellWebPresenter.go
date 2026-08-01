//go:build !js || !wasm || solo

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// FreeCellWebPresenter フリーセルWebプレゼンタークラス
type FreeCellWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *FreeCellWebPresenter) Output(f interfaces.FreeCellGame, lastErr error) string {
	resObj := new(controller.FreeCellWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, f, int(f.GetPhase()))

	// タブロー
	tableau := f.GetTableau()
	resObj.Tableau = make([][]*controller.WebOutputCard, domain.FreeCellTableauCnt)
	for i := 0; i < domain.FreeCellTableauCnt; i++ {
		colCards := tableau[i]
		resObj.Tableau[i] = make([]*controller.WebOutputCard, len(colCards))
		for j, c := range colCards {
			resObj.Tableau[i][j] = cardToOutput(c)
		}
	}

	// フリーセル
	freeCells := f.GetFreeCells()
	resObj.FreeCells = make([]*controller.WebOutputCard, domain.FreeCellCellCnt)
	for i := 0; i < domain.FreeCellCellCnt; i++ {
		resObj.FreeCells[i] = cardToOutput(freeCells[i])
	}

	// ファンデーション
	foundation := f.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.FreeCellFoundationCnt)
	for i := 0; i < domain.FreeCellFoundationCnt; i++ {
		pile := foundation[i]
		resObj.Foundation[i] = make([]*controller.WebOutputCard, len(pile))
		for j, c := range pile {
			resObj.Foundation[i][j] = cardToOutput(c)
		}
	}

	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	if f.GetPhase() == domain.FreeCellPhasePlaying && !f.IsStalemate() {
		if hint := f.GetHint(); hint != nil {
			resObj.Hint = &controller.FreeCellWebOutputHint{
				FromZone:  hint.FromZone,
				FromCol:   hint.FromCol,
				CardIndex: hint.CardIndex,
				ToZone:    hint.ToZone,
				ToCol:     hint.ToCol,
			}
		}
	}

	// メッセージ
	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		phase := f.GetPhase()
		switch phase {
		case domain.FreeCellPhasePlaying:
			if f.IsStalemate() {
				if resObj.UndoToEscape > 0 {
					resObj.MessageCode = "freecell.stalemateWithEscape"
					resObj.MessageParams = map[string]string{"count": fmt.Sprintf("%d", resObj.UndoToEscape)}
				} else {
					resObj.MessageCode = "freecell.stalemate"
				}
			} else {
				resObj.MessageCode = "freecell.playing"
			}
		case domain.FreeCellPhaseGameClear:
			resObj.MessageCode = "freecell.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", f.GetMoveCount())}
		case domain.FreeCellPhaseGameOver:
			resObj.MessageCode = "freecell.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *FreeCellWebPresenter) HintOutput(f interfaces.FreeCellGame) string {
	hint := f.GetHint()
	resObj := new(controller.FreeCellWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, f, int(f.GetPhase()))
	resObj.FreeCells = make([]*controller.WebOutputCard, 0)
	resObj.Tableau = make([][]*controller.WebOutputCard, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)

	if hint != nil {
		resObj.Hint = &controller.FreeCellWebOutputHint{
			FromZone:  hint.FromZone,
			FromCol:   hint.FromCol,
			CardIndex: hint.CardIndex,
			ToZone:    hint.ToZone,
			ToCol:     hint.ToCol,
		}
		resObj.MessageCode = "freecell.hintAvailable"
	} else {
		resObj.MessageCode = "freecell.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *FreeCellWebPresenter) ActionLogOutput(f interfaces.FreeCellGame) string {
	return actionLogOutputJSON(f)
}
