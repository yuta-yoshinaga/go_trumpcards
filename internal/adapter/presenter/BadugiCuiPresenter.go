//go:build !js || !wasm || casino

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// BadugiCuiPresenter renders Badugi state for the CLI.
type BadugiCuiPresenter struct{}

// Output produces the CUI-rendered state for the active locale (#1699).
func (bcp *BadugiCuiPresenter) Output(g interfaces.BadugiGame, lastErr error) string {
	return buildCuiOutput(i18n.T("badugi.outputTitle"), func(b *strings.Builder) {
		players := g.GetPlayers()

		b.WriteString(i18n.Tf("badugi.dealerLine", "idx", strconv.Itoa(g.GetDealerIdx())) + "\n")
		b.WriteString(i18n.Tf("badugi.potLine", "pot", strconv.Itoa(g.GetPot())) + "\n")

		if drawIdx := g.GetDrawIndex(); drawIdx > 0 {
			b.WriteString(i18n.Tf("badugi.drawLine",
				"idx", strconv.Itoa(drawIdx),
				"max", strconv.Itoa(domain.BadugiMaxDraws)) + "\n")
		} else {
			b.WriteString(i18n.T("badugi.drawPredraw") + "\n")
		}

		cfg := g.GetConfig()
		if int(cfg.BettingLimit) < len(domain.BettingLimitNames) {
			b.WriteString(i18n.Tf("badugi.limitLine", "name", domain.BettingLimitNames[cfg.BettingLimit]) + "\n")
		}

		// During betting, show the minimum raise and (where applicable) the max bet
		// so the human can size a bet without trial-and-error error messages.
		if g.GetPhase() != domain.BadugiPhaseEnd {
			minRaise := strconv.Itoa(g.GetMinRaise())
			switch cfg.BettingLimit {
			case domain.BettingLimitPotLimit:
				_, maxBet := domain.CalculateBettingLimits(cfg.BettingLimit, g.GetPot(), g.GetLastBet())
				b.WriteString(i18n.Tf("badugi.betLimitsMax",
					"min", minRaise, "max", strconv.Itoa(maxBet)) + "\n")
			case domain.BettingLimitNoLimit:
				b.WriteString(i18n.Tf("badugi.betLimitsNoCap", "min", minRaise) + "\n")
			default:
				b.WriteString(i18n.Tf("badugi.betLimitsMin", "min", minRaise) + "\n")
			}
		}

		b.WriteString("----------\n")
		isEnd := g.GetPhase() == domain.BadugiPhaseEnd
		for i, pl := range players {
			b.WriteString(cuiPlayerNameWithStyle(pl, i))
			b.WriteString(i18n.Tf("badugi.playerChips", "chips", strconv.Itoa(pl.GetChips())))
			switch {
			case pl.GetFolded():
				b.WriteString(color.BoldYellow(i18n.T("badugi.playerFolded")))
			case pl.GetAllIn():
				b.WriteString(color.BoldYellow(i18n.T("badugi.playerAllIn")))
			}
			if pl.GetCurrentBet() > 0 {
				b.WriteString(i18n.Tf("badugi.playerBet", "bet", strconv.Itoa(pl.GetCurrentBet())))
			}
			if pl.GetDrawCount() > 0 {
				b.WriteString(i18n.Tf("badugi.playerExchange", "count", strconv.Itoa(pl.GetDrawCount())))
			}
			b.WriteString("\n")

			if pl.GetIsHuman() && !pl.GetFolded() {
				handStr := cuiIndexedCardListStrEmoji(pl)
				if isEnd {
					b.WriteString(i18n.Tf("badugi.humanHandWithName",
						"cards", handStr,
						"name", cuiBadugiHandName(pl.GetBestHand().Size)) + "\n")
				} else {
					b.WriteString(i18n.Tf("badugi.humanHand", "cards", handStr) + "\n")
				}
			}
			if !pl.GetIsHuman() && isEnd && !pl.GetFolded() {
				b.WriteString(i18n.Tf("badugi.humanHandWithName",
					"cards", cuiCardListStrEmoji(pl),
					"name", cuiBadugiHandName(pl.GetBestHand().Size)) + "\n")
			}
		}

		cpuActions := g.GetCpuActions()
		if len(cpuActions) > 0 {
			b.WriteString("----------\n")
			b.WriteString(color.Bold(i18n.T("badugi.cpuActionsHeader")) + "\n")
			for _, a := range cpuActions {
				b.WriteString(i18n.Tf("badugi.cpuActionLine",
					"idx", strconv.Itoa(a.PlayerIdx),
					"round", a.RoundLabel,
					"action", cuiBettingActionName(a.Action)))
				if a.Amount > 0 {
					b.WriteString(i18n.Tf("badugi.cpuActionAmount", "amount", strconv.Itoa(a.Amount)))
				}
				b.WriteString("\n")
			}
		}

		cpuExchanges := g.GetCpuExchanges()
		if len(cpuExchanges) > 0 {
			b.WriteString("----------\n")
			b.WriteString(color.Bold(i18n.T("badugi.cpuExchangesHeader")) + "\n")
			for _, e := range cpuExchanges {
				b.WriteString(i18n.Tf("badugi.cpuExchangeLine",
					"idx", strconv.Itoa(e.PlayerIdx),
					"drawIdx", strconv.Itoa(e.DrawIndex),
					"count", strconv.Itoa(e.ExchangeCount)) + "\n")
			}
		}

		results := g.GetRoundResults()
		if len(results) > 0 && isEnd {
			b.WriteString("==========\n")
			b.WriteString(color.Bold(i18n.T("badugi.resultsHeader")) + "\n")
			for _, r := range results {
				name := cuiPlayerName(players[r.PlayerIdx], r.PlayerIdx)
				if r.HandName != "" {
					b.WriteString(i18n.Tf("badugi.resultHand",
						"name", name,
						"hand", r.HandName))
				} else {
					b.WriteString(i18n.Tf("badugi.resultName", "name", name))
				}
				if r.WonAmount > 0 {
					b.WriteString(i18n.Tf("badugi.wonAmount", "total", strconv.Itoa(r.WonAmount)))
				}
				b.WriteString("\n")
			}
		}

		cuiErrorBlock(b, lastErr)
		if g.GetGameEndFlag() {
			b.WriteString(i18n.T("badugi.gameEnd") + "\n")
		}
	})
}

// ActionLogOutput renders the action log as plain text.
func (bcp *BadugiCuiPresenter) ActionLogOutput(g interfaces.BadugiGame) string {
	return actionLogOutputText(g)
}

// HintOutput emits a draw-phase recommendation: which cards form the current
// best Badugi subset (keep) and which to exchange. Mirrors the CPU's
// best-subset discard logic so the human gets the same guidance.
func (bcp *BadugiCuiPresenter) HintOutput(g interfaces.BadugiGame) string {
	if g.GetPhase() != domain.BadugiPhaseDraw {
		return i18n.T("badugi.hintNone") + "\n"
	}
	turn := g.GetCurrentTurn()
	players := g.GetPlayers()
	if turn < 0 || turn >= len(players) {
		return i18n.T("badugi.hintNone") + "\n"
	}
	pl := players[turn]
	if !pl.GetIsHuman() {
		return i18n.T("badugi.hintNone") + "\n"
	}
	// EvalHand idempotently recomputes the cached best subset from the current
	// cards; the draw phase does not otherwise refresh it for the human.
	pl.EvalHand()
	best := pl.GetBestHand()
	kept := make(map[*domain.Card]bool, len(best.Cards))
	for _, c := range best.Cards {
		kept[c] = true
	}
	keep := make([]string, 0, best.Size)
	discard := make([]string, 0, pl.GetCardsSize())
	for i := 0; i < pl.GetCardsSize(); i++ {
		if kept[pl.GetCard(i)] {
			keep = append(keep, strconv.Itoa(i))
		} else {
			discard = append(discard, strconv.Itoa(i))
		}
	}
	if len(discard) == 0 {
		return color.Yellow(i18n.Tf("badugi.hintStandPat", "size", strconv.Itoa(best.Size))) + "\n"
	}
	return color.Yellow(i18n.Tf("badugi.hintExchange",
		"keep", strings.Join(keep, ","),
		"discard", strings.Join(discard, ","),
		"size", strconv.Itoa(best.Size))) + "\n"
}
