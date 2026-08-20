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

// AlaskaCuiPresenter renders the Russian Solitaire CUI view.
type AlaskaCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
// alaskaColumnStr は 1 列を "[i]札" の並びに整形する。裏向きは "??"。
// klondikeColumnStr と同じ整形だが、あちらは KlondikeTableauCard を取り、
// Alaska は別バケット (extra4) なので独自の AlaskaTableauCard を使う。
func alaskaColumnStr(colCards []*domain.AlaskaTableauCard) string {
	parts := make([]string, len(colCards))
	for j, tc := range colCards {
		if tc.FaceUp {
			parts[j] = fmt.Sprintf(" [%d]%s", j, cuiCardStr(tc.Card))
		} else {
			parts[j] = fmt.Sprintf(" [%d]??", j)
		}
	}
	return strings.Join(parts, " ")
}

func (p *AlaskaCuiPresenter) Output(r interfaces.AlaskaGame, lastErr error) string {
	return buildCuiOutput(i18n.T("alaska.helpTitle"), func(b *strings.Builder) {
		// Foundation
		b.WriteString(i18n.T("alaska.foundationHeader"))
		foundation := r.GetFoundation()
		for i := range domain.AlaskaFoundationCnt {
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

		b.WriteString("----------\n")

		// Tableau
		tableau := r.GetTableau()
		for col := range domain.AlaskaTableauCnt {
			colCards := tableau[col]
			b.WriteString(i18n.Tf("alaska.columnLabel", "col", strconv.Itoa(col)))
			if len(colCards) == 0 {
				b.WriteString(" " + i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(alaskaColumnStr(colCards))
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch r.GetPhase() {
		case domain.AlaskaPhasePlaying:
			if r.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				// Tell the player how many undos escape the dead end, matching the
				// web StalemateEscapeButton.
				if n := r.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("cuiSolitaireUndoToEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(r.GetMoveCount())) + "\n")
		case domain.AlaskaPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(r.GetMoveCount())) + "\n")
		case domain.AlaskaPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current Russian Solitaire hint.
func (p *AlaskaCuiPresenter) HintOutput(r interfaces.AlaskaGame) string {
	hint := r.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	from := i18n.Tf("alaska.hintFrom",
		"col", strconv.Itoa(hint.FromCol),
		"idx", strconv.Itoa(hint.CardIndex))
	var to string
	if hint.ToZone == "foundation" {
		to = i18n.T("alaska.hintToFoundation")
	} else {
		to = i18n.Tf("alaska.hintToTableau", "col", strconv.Itoa(hint.ToCol))
	}
	return i18n.Tf("alaska.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *AlaskaCuiPresenter) ActionLogOutput(r interfaces.AlaskaGame) string {
	if r.GetPhase() == domain.AlaskaPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(r.GetActionLog())
}
