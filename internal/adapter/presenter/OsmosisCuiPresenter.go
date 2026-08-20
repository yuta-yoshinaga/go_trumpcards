//go:build !js || !wasm || solo

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

		// Foundation rows (top card of each), each annotated with the ranks that
		// may be placed on it so the player need not run the hint command.
		foundation := o.GetFoundation()
		baseRank := o.GetBaseRank()
		for i := range domain.OsmosisFoundationCnt {
			b.WriteString(i18n.Tf("osmosis.foundationLabel", "idx", strconv.Itoa(i)))
			pile := foundation[i]
			if len(pile) == 0 {
				b.WriteString(" " + i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(" " + cuiCardStr(pile[len(pile)-1]) +
					i18n.Tf("osmosis.foundationCount", "count", strconv.Itoa(len(pile))))
			}
			var allowed string
			if i == 0 && len(pile) > 0 {
				allowed = i18n.T("osmosis.anyRank")
			} else {
				allowed = osmosisRankLabels(osmosisAllowedRanks(foundation, baseRank, i))
			}
			if allowed != "" {
				b.WriteString(" " + i18n.Tf("osmosis.allowedRanks", "ranks", allowed))
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
			// 手詰まりでもフェーズは Playing のままなので、明示しないと
			// プレイヤーは自力で気づくまで無駄にめくり続ける (#4808)。
			if o.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(o.GetMoveCount())) +
				cuiSolitaireUndoHint(o.CanUndo()) + "\n")
		case domain.OsmosisPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(o.GetMoveCount())) + "\n")
		case domain.OsmosisPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// osmosisRanksIn returns the set of card ranks present in a foundation pile.
func osmosisRanksIn(pile []*domain.Card) map[int]bool {
	have := make(map[int]bool, len(pile))
	for _, c := range pile {
		have[c.GetValue()] = true
	}
	return have
}

// osmosisAllowedRanks mirrors the web rule: which ranks may be placed on
// foundation row i (base row 0 accepts any rank once seeded; lower rows accept
// ranks present in the row above that are not yet in the row).
func osmosisAllowedRanks(foundation [domain.OsmosisFoundationCnt][]*domain.Card, baseRank, i int) []int {
	pile := foundation[i]
	if len(pile) == 0 {
		if i >= 1 && len(foundation[i-1]) == 0 {
			return nil
		}
		return []int{baseRank}
	}
	have := osmosisRanksIn(pile)
	if i == 0 {
		var all []int
		for r := 1; r <= 13; r++ {
			if !have[r] {
				all = append(all, r)
			}
		}
		return all
	}
	above := osmosisRanksIn(foundation[i-1])
	var out []int
	for r := 1; r <= 13; r++ {
		if above[r] && !have[r] {
			out = append(out, r)
		}
	}
	return out
}

// osmosisRankLabel renders an Ace-low rank (1=A … 13=K) as a short label.
func osmosisRankLabel(rank int) string {
	switch rank {
	case 1:
		return "A"
	case 11:
		return "J"
	case 12:
		return "Q"
	case 13:
		return "K"
	default:
		return strconv.Itoa(rank)
	}
}

// osmosisRankLabels renders ranks as short labels (A/2-10/J/Q/K).
func osmosisRankLabels(ranks []int) string {
	labels := make([]string, 0, len(ranks))
	for _, r := range ranks {
		labels = append(labels, osmosisRankLabel(r))
	}
	return strings.Join(labels, " ")
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

	foundation := o.GetFoundation()
	var allowed string
	if hint.ToCol == 0 && len(foundation[0]) > 0 {
		allowed = i18n.T("osmosis.anyRank")
	} else {
		allowed = osmosisRankLabels(osmosisAllowedRanks(foundation, o.GetBaseRank(), hint.ToCol))
	}
	return i18n.Tf("osmosis.hintLine", "from", from, "to", to, "allowed", allowed) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *OsmosisCuiPresenter) ActionLogOutput(o interfaces.OsmosisGame) string {
	if o.GetPhase() == domain.OsmosisPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(o.GetActionLog())
}
