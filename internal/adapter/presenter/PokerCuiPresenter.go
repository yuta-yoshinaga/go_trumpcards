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

// PokerCuiPresenter renders the 5-Card Draw Poker CUI view.
type PokerCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (pcp *PokerCuiPresenter) Output(p interfaces.PokerGame, lastErr error) string {
	var b strings.Builder
	players := p.GetPlayers()

	b.WriteString("==========\n")
	if p.GetConfig().IsLowball {
		b.WriteString(i18n.T("poker.outputTitleLowball") + "\n")
	} else {
		b.WriteString(i18n.T("poker.outputTitle") + "\n")
	}
	b.WriteString("==========\n")

	b.WriteString(i18n.Tf("poker.dealerLine", "idx", strconv.Itoa(p.GetDealerIdx())) + "\n")
	b.WriteString(i18n.Tf("poker.potLine", "pot", strconv.Itoa(p.GetPot())) + "\n")

	if p.GetConfig().JokerCount > 0 {
		b.WriteString(i18n.Tf("poker.jokerLine", "count", strconv.Itoa(p.GetConfig().JokerCount)) + "\n")
	}

	cfg := p.GetConfig()
	if int(cfg.BettingLimit) < len(domain.BettingLimitNames) {
		b.WriteString(i18n.Tf("poker.limitLine", "name", domain.BettingLimitNames[cfg.BettingLimit]) + "\n")
	}

	b.WriteString("----------\n")
	isEnd := p.GetPhase() == domain.PokerPhaseEnd
	for i, player := range players {
		b.WriteString(cuiPlayerNameWithStyle(player, i))
		b.WriteString(i18n.Tf("poker.playerChips", "chips", strconv.Itoa(player.GetChips())))

		if player.GetFolded() {
			b.WriteString(color.BoldYellow(i18n.T("poker.playerFolded")))
		} else if player.GetAllIn() {
			b.WriteString(color.BoldYellow(i18n.T("poker.playerAllIn")))
		}

		if player.GetCurrentBet() > 0 {
			b.WriteString(i18n.Tf("poker.playerBet", "bet", strconv.Itoa(player.GetCurrentBet())))
		}

		if player.GetExchangeCount() > 0 && (p.GetPhase() == domain.PokerPhaseSecondBet || isEnd) {
			b.WriteString(i18n.Tf("poker.playerExchange", "count", strconv.Itoa(player.GetExchangeCount())))
		}
		b.WriteString("\n")

		// Human's hand is always shown.
		if player.GetIsHuman() && !player.GetFolded() {
			handStr := cuiIndexedCardListStrEmoji(player)
			if isEnd {
				b.WriteString(i18n.Tf("poker.humanHandWithName",
					"cards", handStr,
					"name", player.GetHandName()) + "\n")
			} else {
				b.WriteString(i18n.Tf("poker.humanHand", "cards", handStr) + "\n")
			}
		}

		// CPU hands are revealed only at end-of-hand.
		if !player.GetIsHuman() && isEnd && !player.GetFolded() {
			b.WriteString(i18n.Tf("poker.humanHandWithName",
				"cards", cuiCardListStrEmoji(player),
				"name", player.GetHandName()) + "\n")
		}
	}

	cpuActions := p.GetCpuActions()
	if len(cpuActions) > 0 {
		b.WriteString("----------\n")
		b.WriteString(color.Bold(i18n.T("poker.cpuActionsHeader")) + "\n")
		for _, action := range cpuActions {
			b.WriteString(i18n.Tf("poker.cpuActionLine",
				"idx", strconv.Itoa(action.PlayerIdx),
				"action", cuiBettingActionName(action.Action)))
			if action.Amount > 0 {
				b.WriteString(i18n.Tf("poker.cpuActionAmount", "amount", strconv.Itoa(action.Amount)))
			}
			b.WriteString("\n")
		}
	}

	cpuExchanges := p.GetCpuExchanges()
	if len(cpuExchanges) > 0 {
		b.WriteString("----------\n")
		b.WriteString(color.Bold(i18n.T("poker.cpuExchangesHeader")) + "\n")
		for _, ex := range cpuExchanges {
			b.WriteString(i18n.Tf("poker.cpuExchangeLine",
				"idx", strconv.Itoa(ex.PlayerIdx),
				"count", strconv.Itoa(ex.ExchangeCount)) + "\n")
		}
	}

	results := p.GetRoundResults()
	if len(results) > 0 && isEnd {
		b.WriteString("==========\n")
		b.WriteString(color.Bold(i18n.T("poker.resultsHeader")) + "\n")
		for _, r := range results {
			name := cuiPlayerName(players[r.PlayerIdx], r.PlayerIdx)
			kickers := ""
			if ks := domain.FormatKickers(r.Kickers); ks != "" {
				kickers = i18n.Tf("poker.resultKickers", "kickers", ks)
			}
			if r.HandName != "" {
				b.WriteString(i18n.Tf("poker.resultHand",
					"name", name,
					"hand", r.HandName,
					"kickers", kickers))
			} else {
				b.WriteString(i18n.Tf("poker.resultName", "name", name))
			}
			if r.WonAmount > 0 {
				b.WriteString(i18n.Tf("poker.wonAmount", "total", strconv.Itoa(r.WonAmount)))
			}
			b.WriteString("\n")
		}
	}

	if lastErr != nil {
		fmt.Fprintf(&b, "%s\n", color.Red(lastErr.Error()))
	}

	if p.GetGameEndFlag() {
		b.WriteString(i18n.T("poker.gameEnd") + "\n")
	}

	return b.String()
}

// ActionLogOutput emits the action-log transcript as plain text.
func (pcp *PokerCuiPresenter) ActionLogOutput(p interfaces.PokerGame) string {
	return actionLogOutputText(p)
}

// OutputWithOdds appends the draw-odds table to the standard Output.
func (pcp *PokerCuiPresenter) OutputWithOdds(p interfaces.PokerGame, lastErr error, odds []domain.PokerDrawOdds) string {
	base := pcp.Output(p, lastErr)
	if len(odds) == 0 {
		return base
	}
	var oddsBuilder strings.Builder
	oddsBuilder.WriteString("==========\n")
	oddsBuilder.WriteString(color.Bold(i18n.T("poker.drawOddsHeader")) + "\n")
	for _, o := range odds {
		oddsBuilder.WriteString(i18n.Tf("poker.drawOddsLine",
			"name", o.HandName,
			"prob", fmt.Sprintf("%.2f", o.Probability*100),
			"count", strconv.Itoa(o.Count),
			"total", strconv.Itoa(o.Total)) + "\n")
	}
	return base + oddsBuilder.String()
}
