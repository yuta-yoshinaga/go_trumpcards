//go:build !js || !wasm || extra9

package presenter

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// StalactitesWebPresenter フリーセルWebプレゼンタークラス
type StalactitesWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *StalactitesWebPresenter) Output(f interfaces.StalactitesGame, lastErr error) string {
	resObj := new(controller.StalactitesWebOutput)
	resObj.LegalTargets = p.buildLegalTargets(f)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, f, int(f.GetPhase()))

	// タブロー
	tableau := f.GetTableau()
	resObj.Tableau = make([][]*controller.WebOutputCard, domain.StalactitesTableauCnt)
	for i := 0; i < domain.StalactitesTableauCnt; i++ {
		colCards := tableau[i]
		resObj.Tableau[i] = make([]*controller.WebOutputCard, len(colCards))
		for j, c := range colCards {
			resObj.Tableau[i][j] = cardToOutput(c)
		}
	}

	// フリーセル
	cells := f.GetCells()
	resObj.Cells = make([]*controller.WebOutputCard, domain.StalactitesCellCnt)
	for i := 0; i < domain.StalactitesCellCnt; i++ {
		resObj.Cells[i] = cardToOutput(cells[i])
	}

	// ファンデーション
	foundation := f.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.StalactitesFoundationCnt)
	for i := 0; i < domain.StalactitesFoundationCnt; i++ {
		pile := foundation[i]
		resObj.Foundation[i] = make([]*controller.WebOutputCard, len(pile))
		for j, c := range pile {
			resObj.Foundation[i][j] = cardToOutput(c)
		}
	}

	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	if f.GetPhase() == domain.StalactitesPhasePlaying && !f.IsStalemate() {
		if hint := f.GetHint(); hint != nil {
			resObj.Hint = &controller.StalactitesWebOutputHint{
				FromZone:  hint.FromZone,
				FromCol:   hint.FromCol,
				CardIndex: hint.CardIndex,
				ToZone:    hint.ToZone,
				ToCol:     hint.ToCol,
			}
		}
	}

	// 一度に動かせる枚数。空き列宛ては別枠 (#5975)。
	resObj.BaseRank = f.GetBaseRank()
	resObj.MaxMovableCards = f.GetMaxMovableCards()
	resObj.MaxMovableCardsToEmptyColumn = f.GetMaxMovableCardsToEmptyColumn()

	// メッセージ
	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		phase := f.GetPhase()
		switch phase {
		case domain.StalactitesPhasePlaying:
			if f.IsStalemate() {
				if resObj.UndoToEscape > 0 {
					resObj.MessageCode = "stalactites.stalemateWithEscape"
					resObj.MessageParams = map[string]string{"count": fmt.Sprintf("%d", resObj.UndoToEscape)}
				} else {
					resObj.MessageCode = "stalactites.stalemate"
				}
			} else {
				resObj.MessageCode = "stalactites.playing"
			}
		case domain.StalactitesPhaseGameClear:
			resObj.MessageCode = "stalactites.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", f.GetMoveCount())}
		case domain.StalactitesPhaseGameOver:
			resObj.MessageCode = "stalactites.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *StalactitesWebPresenter) HintOutput(f interfaces.StalactitesGame) string {
	hint := f.GetHint()
	resObj := new(controller.StalactitesWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, f, int(f.GetPhase()))
	resObj.Cells = make([]*controller.WebOutputCard, 0)
	resObj.Tableau = make([][]*controller.WebOutputCard, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)
	resObj.LegalTargets = p.buildLegalTargets(f)

	if hint != nil {
		resObj.Hint = &controller.StalactitesWebOutputHint{
			FromZone:  hint.FromZone,
			FromCol:   hint.FromCol,
			CardIndex: hint.CardIndex,
			ToZone:    hint.ToZone,
			ToCol:     hint.ToCol,
		}
		resObj.MessageCode = "stalactites.hintAvailable"
	} else {
		resObj.MessageCode = "stalactites.noHint"
	}

	return marshalOrError(resObj)
}

func (p *StalactitesWebPresenter) buildLegalTargets(f interfaces.StalactitesGame) map[string]controller.StalactitesWebOutputTargets {
	result := make(map[string]controller.StalactitesWebOutputTargets)
	for col := 0; col < domain.StalactitesTableauCnt; col++ {
		tab, foundation := f.LegalTargetsFrom(col)
		if len(tab)+len(foundation) == 0 {
			continue
		}
		if tab == nil {
			tab = []int{}
		}
		if foundation == nil {
			foundation = []int{}
		}
		result["tableau-"+strconv.Itoa(col)] = controller.StalactitesWebOutputTargets{Tableau: tab, Foundation: foundation}
	}
	for cell := 0; cell < domain.StalactitesCellCnt; cell++ {
		tab, foundation := f.CellLegalTargets(cell)
		if len(tab)+len(foundation) == 0 {
			continue
		}
		if tab == nil {
			tab = []int{}
		}
		if foundation == nil {
			foundation = []int{}
		}
		result["stalactites-"+strconv.Itoa(cell)] = controller.StalactitesWebOutputTargets{Tableau: tab, Foundation: foundation}
	}
	return result
}

// ActionLogOutput 棋譜をJSON出力
func (p *StalactitesWebPresenter) ActionLogOutput(f interfaces.StalactitesGame) string {
	return actionLogOutputJSON(f)
}
