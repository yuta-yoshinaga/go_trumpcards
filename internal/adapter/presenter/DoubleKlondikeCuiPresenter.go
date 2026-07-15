//go:build !js || !wasm || classic

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// DoubleKlondikeCuiPresenter ダブル・クロンダイクのCUIプレゼンタークラス。
type DoubleKlondikeCuiPresenter struct{}

// dkTableauStr renders a tableau column (face-down cards shown as "##").
func dkTableauStr(col []*domain.DoubleKlondikeTableauCard) string {
	if len(col) == 0 {
		return i18n.T("cuiEmptyCol")
	}
	parts := make([]string, len(col))
	for i, tc := range col {
		if tc.FaceUp {
			parts[i] = cuiCardStr(tc.Card)
		} else {
			parts[i] = "##"
		}
	}
	return strings.Join(parts, " ")
}

// dkPileStr renders a foundation/waste pile.
func dkPileStr(pile []*domain.Card) string {
	if len(pile) == 0 {
		return i18n.T("cuiEmptyCol")
	}
	parts := make([]string, len(pile))
	for i, c := range pile {
		parts[i] = cuiCardStr(c)
	}
	return strings.Join(parts, " ")
}

// Output renders the current game state.
func (p *DoubleKlondikeCuiPresenter) Output(g interfaces.DoubleKlondikeGame, lastErr error) string {
	return buildCuiOutput(i18n.T("doubleklondike.helpTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("doubleklondike.stockLine",
			"stock", strconv.Itoa(g.GetStockCount()),
			"waste", dkPileStr(g.GetWaste())) + "\n")
		foundation := g.GetFoundation()
		for i := 0; i < domain.DoubleKlondikeFoundationCnt; i++ {
			sb.WriteString(i18n.Tf("doubleklondike.foundationLabel", "idx", strconv.Itoa(i)))
			sb.WriteString(" " + dkPileStr(foundation[i]) + "\n")
		}
		tableau := g.GetTableau()
		for i := 0; i < domain.DoubleKlondikeTableauCnt; i++ {
			sb.WriteString(i18n.Tf("doubleklondike.tableauLabel", "idx", strconv.Itoa(i)))
			sb.WriteString(" " + dkTableauStr(tableau[i]) + "\n")
		}

		cuiErrorBlock(sb, lastErr)

		switch g.GetPhase() {
		case domain.DoubleKlondikePhasePlaying:
			sb.WriteString(i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(g.GetMoveCount())) + "\n")
		case domain.DoubleKlondikePhaseGameClear:
			sb.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(g.GetMoveCount())) + "\n")
		case domain.DoubleKlondikePhaseGameOver:
			sb.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// doubleKlondikeZoneName localises a hint zone identifier (waste / tableau /
// foundation). Unknown identifiers fall back to the raw string so the hint never
// panics or renders empty.
func doubleKlondikeZoneName(zone string) string {
	switch zone {
	case "waste":
		return i18n.T("doubleklondike.zoneWaste")
	case "tableau":
		return i18n.T("doubleklondike.zoneTableau")
	case "foundation":
		return i18n.T("doubleklondike.zoneFoundation")
	default:
		return zone
	}
}

// HintOutput emits the current hint.
func (p *DoubleKlondikeCuiPresenter) HintOutput(g interfaces.DoubleKlondikeGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	return color.Yellow(i18n.Tf("doubleklondike.hintLine",
		"from", doubleKlondikeZoneName(hint.FromZone),
		"fromCol", strconv.Itoa(hint.FromCol),
		"to", doubleKlondikeZoneName(hint.ToZone),
		"toCol", strconv.Itoa(hint.ToCol))) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *DoubleKlondikeCuiPresenter) ActionLogOutput(g interfaces.DoubleKlondikeGame) string {
	return actionLogOutputText(g)
}
