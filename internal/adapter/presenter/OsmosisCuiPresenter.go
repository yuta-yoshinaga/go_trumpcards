package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// OsmosisCuiPresenter renders the Osmosis Solitaire CUI view.
type OsmosisCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *OsmosisCuiPresenter) Output(o interfaces.OsmosisGame, lastErr error) string {
	return buildCuiOutput(i18n.T("osmosis.helpTitle"), func(b *strings.Builder) {
		// Base rank
		b.WriteString(i18n.Tf("osmosis.baseRank", "rank", strconv.Itoa(o.GetBaseRank())) + "\n")

		// Foundation rows (top card of each)
		foundation := o.GetFoundation()
		for i := range domain.OsmosisFoundationCnt {
			b.WriteString(i18n.Tf("osmosis.foundationLabel", "idx", strconv.Itoa(i)))
			pile := foundation[i]
			if len(pile) == 0 {
				b.WriteString(" " + i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(" " + cuiCardStr(pile[len(pile)-1]) +
					i18n.Tf("osmosis.foundationCount", "count", strconv.Itoa(len(pile))))
			}
			b.WriteString("\n")
		}
		b.WriteString("----------\n")

		// Reserve piles (top card of each)
		reserve := o.GetReserve()
		for i := range domain.OsmosisReserveCnt {
			b.WriteString(i18n.Tf("osmosis.reserveLabel", "idx", strconv.Itoa(i)))
			pile := reserve[i]
			if len(pile) == 0 {
				b.WriteString(" " + i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(" " + cuiCardStr(pile[len(pile)-1]) +
					i18n.Tf("osmosis.reserveCount", "count", strconv.Itoa(len(pile))))
			}
			b.WriteString("\n")
		}
		b.WriteString("----------\n")

		// Stock / waste
		b.WriteString(i18n.Tf("osmosis.stockLine", "count", strconv.Itoa(o.GetStockCount())))
		waste := o.GetWaste()
		if len(waste) > 0 {
			b.WriteString(i18n.Tf("osmosis.wasteCard", "card", cuiCardStr(waste[len(waste)-1])))
		} else {
			b.WriteString(i18n.T("osmosis.wasteEmpty"))
		}
		b.WriteString("\n----------\n")

		cuiErrorBlock(b, lastErr)

		switch o.GetPhase() {
		case domain.OsmosisPhasePlaying:
			b.WriteString(i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(o.GetMoveCount())) + "\n")
		case domain.OsmosisPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(o.GetMoveCount())) + "\n")
		case domain.OsmosisPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current Osmosis hint.
func (p *OsmosisCuiPresenter) HintOutput(o interfaces.OsmosisGame) string {
	hint := o.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	var from string
	if hint.FromZone == "reserve" {
		from = i18n.Tf("osmosis.hintFromReserve", "col", strconv.Itoa(hint.FromCol))
	} else {
		from = i18n.T("osmosis.hintFromWaste")
	}
	to := i18n.Tf("osmosis.hintToFoundation", "idx", strconv.Itoa(hint.ToCol))
	return i18n.Tf("osmosis.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *OsmosisCuiPresenter) ActionLogOutput(o interfaces.OsmosisGame) string {
	if o.GetPhase() == domain.OsmosisPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(o.GetActionLog())
}
