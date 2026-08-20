//go:build !js || !wasm || solo

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// PenguinWebPresenter ペンギンWebプレゼンタークラス
type PenguinWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *PenguinWebPresenter) Output(pg interfaces.PenguinGame, lastErr error) string {
	resObj := new(controller.PenguinWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, pg, int(pg.GetPhase()))

	// タブロー
	tableau := pg.GetTableau()
	resObj.Tableau = make([][]*controller.WebOutputCard, domain.PenguinTableauCnt)
	for i := 0; i < domain.PenguinTableauCnt; i++ {
		colCards := tableau[i]
		resObj.Tableau[i] = make([]*controller.WebOutputCard, len(colCards))
		for j, c := range colCards {
			resObj.Tableau[i][j] = cardToOutput(c)
		}
	}

	// フリーセル
	freeCells := pg.GetFreeCells()
	resObj.FreeCells = make([]*controller.WebOutputCard, domain.PenguinCellCnt)
	for i := 0; i < domain.PenguinCellCnt; i++ {
		resObj.FreeCells[i] = cardToOutput(freeCells[i])
	}

	// ファンデーション
	foundation := pg.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.PenguinFoundationCnt)
	for i := 0; i < domain.PenguinFoundationCnt; i++ {
		pile := foundation[i]
		resObj.Foundation[i] = make([]*controller.WebOutputCard, len(pile))
		for j, c := range pile {
			resObj.Foundation[i][j] = cardToOutput(c)
		}
	}

	// ベースランク
	resObj.BaseRank = pg.GetBaseRank()

	// 一度に動かせる枚数。空き列宛ては別枠 (#5614)。
	resObj.MaxMovableCards = pg.GetMaxMovableCards()
	resObj.MaxMovableCardsToEmptyColumn = pg.GetMaxMovableCardsToEmptyColumn()

	// メッセージ
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	if pg.GetPhase() == domain.PenguinPhasePlaying && !pg.IsStalemate() {
		if hint := pg.GetHint(); hint != nil {
			resObj.Hint = &controller.PenguinWebOutputHint{
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
		phase := pg.GetPhase()
		switch phase {
		case domain.PenguinPhasePlaying:
			if pg.IsStalemate() {
				if resObj.UndoToEscape > 0 {
					resObj.MessageCode = "penguin.stalemateWithEscape"
					resObj.MessageParams = map[string]string{"count": fmt.Sprintf("%d", resObj.UndoToEscape)}
				} else {
					resObj.MessageCode = "penguin.stalemate"
				}
			} else {
				resObj.MessageCode = "penguin.playing"
			}
		case domain.PenguinPhaseGameClear:
			resObj.MessageCode = "penguin.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", pg.GetMoveCount())}
		case domain.PenguinPhaseGameOver:
			resObj.MessageCode = "penguin.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *PenguinWebPresenter) HintOutput(pg interfaces.PenguinGame) string {
	hint := pg.GetHint()
	resObj := new(controller.PenguinWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, pg, int(pg.GetPhase()))
	resObj.FreeCells = make([]*controller.WebOutputCard, 0)
	resObj.Tableau = make([][]*controller.WebOutputCard, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)

	if hint != nil {
		resObj.Hint = &controller.PenguinWebOutputHint{
			FromZone:  hint.FromZone,
			FromCol:   hint.FromCol,
			CardIndex: hint.CardIndex,
			ToZone:    hint.ToZone,
			ToCol:     hint.ToCol,
		}
		resObj.MessageCode = "penguin.hintAvailable"
	} else {
		resObj.MessageCode = "penguin.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *PenguinWebPresenter) ActionLogOutput(pg interfaces.PenguinGame) string {
	return actionLogOutputJSON(pg)
}
