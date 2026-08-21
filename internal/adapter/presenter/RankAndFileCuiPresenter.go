//go:build !js || !wasm || extra4

package presenter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// rankAndFileColumnStr returns the display string for a RankAndFile tableau column.
func rankAndFileColumnStr(colCards []*domain.RankAndFileTableauCard) string {
	parts := make([]string, len(colCards))
	for j, tc := range colCards {
		// **伏せ札は中身を出さない。**各列 4 枚のうち 3 枚が伏せで配られるので、
		// FaceUp を無視するとこの CUI が隠し札を全部見せてしまう（クローン元の
		// Forty Thieves は全部表向きなので無害だった）。
		if !tc.FaceUp {
			parts[j] = fmt.Sprintf(" [%d]??", j)
			continue
		}
		parts[j] = fmt.Sprintf(" [%d]%s", j, cuiCardStr(tc.Card))
	}
	return strings.Join(parts, " ")
}

// RankAndFileCuiPresenter renders the Rank and File Solitaire CUI view.
type RankAndFileCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *RankAndFileCuiPresenter) Output(ft interfaces.RankAndFileGame, lastErr error) string {
	return buildCuiOutput(i18n.T("rankandfile.helpTitle"), func(b *strings.Builder) {
		// Foundation
		b.WriteString(i18n.T("rankandfile.foundationHeader"))
		foundation := ft.GetFoundation()
		for i := range domain.RankAndFileFoundationCnt {
			if i != 0 {
				b.WriteString(" | ")
			}
			pile := foundation[i]
			if len(pile) == 0 {
				b.WriteString(i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(cuiCardStr(pile[len(pile)-1]))
			}
		}
		b.WriteString("\n")

		// Stock + waste
		b.WriteString(i18n.Tf("rankandfile.stockLine",
			"count", strconv.Itoa(ft.GetStockCount())))
		waste := ft.GetWaste()
		if len(waste) > 0 {
			b.WriteString(i18n.Tf("rankandfile.wasteCard",
				"card", cuiCardStr(waste[len(waste)-1])))
		} else {
			b.WriteString(i18n.T("rankandfile.wasteEmpty"))
		}
		b.WriteString("\n")

		b.WriteString("----------\n")

		// Tableau
		tableau := ft.GetTableau()
		for col := range domain.RankAndFileTableauCnt {
			colCards := tableau[col]
			b.WriteString(i18n.Tf("rankandfile.columnLabel", "col", strconv.Itoa(col)))
			if len(colCards) == 0 {
				b.WriteString(" " + i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(rankAndFileColumnStr(colCards))
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch ft.GetPhase() {
		case domain.RankAndFilePhasePlaying:
			if ft.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				// Tell the player how many undos escape the dead end, matching the
				// web StalemateEscapeButton.
				if n := ft.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("cuiSolitaireUndoToEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
			}
			b.WriteString(i18n.T("rankandfile.cuiCommandHint") + "\n")
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(ft.GetMoveCount())) + "\n")
		case domain.RankAndFilePhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(ft.GetMoveCount())) + "\n")
		case domain.RankAndFilePhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current Rank and File hint.
func (p *RankAndFileCuiPresenter) HintOutput(ft interfaces.RankAndFileGame) string {
	hint := ft.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	// 盤上に手が無くストックだけ残っている局面。移動の体裁 (「A → B」) に
	// 落とすと列 -1 が漏れるので、専用の文言で言う (#5525)。
	if hint.FromZone == "stock" {
		return i18n.T("rankandfile.hintDraw") + "\n"
	}
	var from string
	if hint.FromZone == "tableau" {
		from = i18n.Tf("rankandfile.hintFromTableau",
			"col", strconv.Itoa(hint.FromCol),
			"idx", strconv.Itoa(hint.CardIndex))
	} else {
		from = i18n.T("rankandfile.hintFromWaste")
	}
	var to string
	if hint.ToZone == "foundation" {
		to = i18n.T("rankandfile.hintToFoundation")
	} else {
		to = i18n.Tf("rankandfile.hintToTableau", "col", strconv.Itoa(hint.ToCol))
	}
	return i18n.Tf("rankandfile.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *RankAndFileCuiPresenter) ActionLogOutput(ft interfaces.RankAndFileGame) string {
	if ft.GetPhase() == domain.RankAndFilePhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(ft.GetActionLog())
}
