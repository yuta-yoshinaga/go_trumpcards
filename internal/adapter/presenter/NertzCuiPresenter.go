package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// NertzCuiPresenter renders the Nertz / Pounce CUI view.
type NertzCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *NertzCuiPresenter) Output(g interfaces.NertzGame, lastErr error) string {
	return buildCuiOutput(i18n.T("nertz.helpTitle"), func(b *strings.Builder) {
		// Shared foundations
		b.WriteString(i18n.Tf("nertz.header",
			"round", strconv.Itoa(g.GetRoundNo()),
			"moves", strconv.Itoa(g.GetMoveCount())) + "\n")
		// 生スコアだけ出しても、何点で決着するのかが CUI からは分からない。
		// Web はこの同じ値をスコアバーの aria-valuemax にしている (#6374)。
		target := g.GetConfig().TargetScore
		b.WriteString(i18n.Tf("nertz.targetLine", "target", strconv.Itoa(target)) + "\n")
		b.WriteString(i18n.T("nertz.foundationsHeader") + "\n")
		founds := g.GetFoundations()
		maxStr := strconv.Itoa(domain.NertzFoundationMax)
		for i, f := range founds {
			b.WriteString(i18n.Tf("nertz.foundationLabel", "idx", strconv.Itoa(i)))
			if f == nil || f.IsEmpty() {
				b.WriteString(i18n.T("nertz.foundationEmpty"))
			} else {
				b.WriteString(i18n.Tf("nertz.foundationFilled",
					"card", cuiCardStr(f.Top()),
					"count", strconv.Itoa(f.Size()),
					"max", maxStr))
			}
			b.WriteString("\n")
		}
		b.WriteString("----------\n")

		// Per-player
		for i, pl := range g.GetPlayers() {
			if pl == nil {
				continue
			}
			label := i18n.T("nertz.labelHuman")
			if pl.GetIsCpu() {
				label = i18n.T("nertz.labelCpu")
			}
			// 残りは 0 で止める。到達済みの席に負の数を出しても読めない。
			remaining := max(target-pl.GetScore(), 0)
			b.WriteString(i18n.Tf("nertz.playerLine",
				"idx", strconv.Itoa(i),
				"label", label,
				"name", pl.GetName(),
				"score", strconv.Itoa(pl.GetScore()),
				"remaining", strconv.Itoa(remaining)) + "\n")

			// Nertz pile
			if top := pl.NertzTop(); top != nil {
				b.WriteString(i18n.Tf("nertz.nertzLine",
					"card", cuiCardStr(top),
					"count", strconv.Itoa(pl.NertzSize())) + "\n")
			} else {
				b.WriteString(i18n.T("nertz.nertzEmpty") + "\n")
			}

			// Tableau (4 columns)
			for c := range domain.NertzTableauCnt {
				col := pl.GetTableauColumn(c)
				b.WriteString(i18n.Tf("nertz.tableauLine", "idx", strconv.Itoa(c)))
				if len(col) == 0 {
					b.WriteString(i18n.T("nertz.tableauEmpty"))
				} else {
					parts := make([]string, len(col))
					for k, tc := range col {
						parts[k] = cuiCardStr(tc.Card)
					}
					b.WriteString(strings.Join(parts, " "))
				}
				b.WriteString("\n")
			}

			// Waste / stock
			stockStr := strconv.Itoa(pl.StockSize())
			if w := pl.WasteTop(); w != nil {
				b.WriteString(i18n.Tf("nertz.wasteLine",
					"card", cuiCardStr(w),
					"count", strconv.Itoa(pl.WasteSize()),
					"stock", stockStr) + "\n")
			} else {
				b.WriteString(i18n.Tf("nertz.wasteEmpty", "stock", stockStr) + "\n")
			}
		}
		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch g.GetPhase() {
		case domain.NertzPhasePlaying:
			b.WriteString(i18n.T("nertz.playing") + "\n")
		case domain.NertzPhaseRoundEnd:
			banner := i18n.Tf("nertz.roundEnd", "winner", strconv.Itoa(g.GetWinnerIdx()))
			b.WriteString(color.Yellow(banner) + "\n")
		case domain.NertzPhaseGameEnd:
			if g.GetMatchWinner() == 0 {
				b.WriteString(color.Green(i18n.T("nertz.win")) + "\n")
			} else {
				b.WriteString(color.Red(i18n.Tf("nertz.lose",
					"winner", strconv.Itoa(g.GetMatchWinner()))) + "\n")
			}
		}
	})
}

// HintOutput emits the current Nertz hint.
func (p *NertzCuiPresenter) HintOutput(g interfaces.NertzGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("nertz.noHint") + "\n"
	}
	from := nertzHintZoneLabel(hint.FromZone, hint.FromCol, hint.CardIndex)
	to := nertzHintZoneLabel(hint.ToZone, hint.ToCol, -1)
	return i18n.Tf("nertz.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text. The log is
// suppressed during play so the CPU's strategy isn't leaked in real time
// (matches the Web presenter — see PR #1528 review).
func (p *NertzCuiPresenter) ActionLogOutput(g interfaces.NertzGame) string {
	if g.GetPhase() == domain.NertzPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(g.GetActionLog())
}

// nertzHintZoneLabel converts a hint zone identifier into a human label.
func nertzHintZoneLabel(zone string, col, idx int) string {
	switch zone {
	case "nertz":
		return i18n.T("nertz.zoneNertz")
	case "waste":
		return i18n.T("nertz.zoneWaste")
	case "tableau":
		if idx >= 0 {
			return i18n.Tf("nertz.zoneTableauWithIdx",
				"col", strconv.Itoa(col),
				"idx", strconv.Itoa(idx))
		}
		return i18n.Tf("nertz.zoneTableau", "col", strconv.Itoa(col))
	case "foundation":
		return i18n.Tf("nertz.zoneFoundation", "col", strconv.Itoa(col))
	}
	return zone
}
