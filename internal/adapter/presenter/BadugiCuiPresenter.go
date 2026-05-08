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

// BadugiCuiPresenter renders Badugi state for the CLI.
type BadugiCuiPresenter struct{}

// Output produces the CUI-rendered state for the active locale (#1699).
func (bcp *BadugiCuiPresenter) Output(g interfaces.BadugiGame, lastErr error) string {
	var b strings.Builder
	players := g.GetPlayers()

	b.WriteString("==========\n")
	b.WriteString(i18n.T("badugi.outputTitle") + "\n")
	b.WriteString("==========\n")

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
					"name", pl.GetHandName()) + "\n")
			} else {
				b.WriteString(i18n.Tf("badugi.humanHand", "cards", handStr) + "\n")
			}
		}
		if !pl.GetIsHuman() && isEnd && !pl.GetFolded() {
			b.WriteString(i18n.Tf("badugi.humanHandWithName",
				"cards", cuiCardListStrEmoji(pl),
				"name", pl.GetHandName()) + "\n")
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

	if lastErr != nil {
		fmt.Fprintf(&b, "%s\n", color.Red(lastErr.Error()))
	}
	if g.GetGameEndFlag() {
		b.WriteString(i18n.T("badugi.gameEnd") + "\n")
	}
	return b.String()
}

// ActionLogOutput renders the action log as plain text.
func (bcp *BadugiCuiPresenter) ActionLogOutput(g interfaces.BadugiGame) string {
	return actionLogOutputText(g)
}
