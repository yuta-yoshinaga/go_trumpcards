//go:build !js || !wasm || casino

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
	titleKey := "poker.outputTitle"
	if p.GetConfig().IsLowball {
		titleKey = "poker.outputTitleLowball"
	}
	return buildCuiOutput(i18n.T(titleKey), func(b *strings.Builder) {
		// Lowball inverts what a poker player already knows -- the ace is the
		// highest card, and a straight or flush still counts, which makes both
		// bad. The web puts this reference on screen; without it the CUI player
		// is choosing discards against the wrong ranking.
		if p.GetConfig().IsLowball {
			b.WriteString(color.Bold(i18n.T("poker.lowballRankTitle")) + "\n")
			for _, k := range []string{"lowballRankBest", "lowballRankAceHigh", "lowballRankStraightFlush", "lowballRankGoal"} {
				b.WriteString("  " + i18n.T("poker."+k) + "\n")
			}
		}
		players := p.GetPlayers()

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
						"name", cuiPokerHandName(player.GetHandRank())) + "\n")
				} else {
					b.WriteString(i18n.Tf("poker.humanHand", "cards", handStr) + "\n")
				}
			}

			// CPU hands are revealed only at end-of-hand.
			if !player.GetIsHuman() && isEnd && !player.GetFolded() {
				b.WriteString(i18n.Tf("poker.humanHandWithName",
					"cards", cuiCardListStrEmoji(player),
					"name", cuiPokerHandName(player.GetHandRank())) + "\n")
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
						"hand", cuiPokerHandName(r.HandRank),
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

		b.WriteString(i18n.Tf("poker.phaseLine", "phase", pokerPhaseName(p.GetPhase())) + "\n")

		cuiErrorBlock(b, lastErr)

		if p.GetGameEndFlag() {
			b.WriteString(i18n.T("poker.gameEnd") + "\n")
		} else {
			// Tell the human what to do this phase (draw poker: bet, then exchange,
			// then bet again).
			switch p.GetPhase() {
			case domain.PokerPhaseDeal, domain.PokerPhaseSecondBet:
				b.WriteString(i18n.T("poker.promptBet") + "\n")
				// **交換枚数は読まれている。** 閾値未満だと全 CPU のフォールド
				// 閾値が1ランク上がる (calcExchangeWarning)。実在の戦略要素なのに
				// Web にも CUI にも説明が無く、プレイヤーは知りようがなかった (#5475)。
				if p.IsExchangeRead(0) {
					b.WriteString(color.BoldYellow(i18n.T("poker.exchangeRead")) + "\n")
				}
			case domain.PokerPhaseExchange:
				b.WriteString(i18n.T("poker.promptExchange") + "\n")
			}
		}
	})
}

// pokerPhaseName returns the localized name for a Poker phase constant.
func pokerPhaseName(phase int) string {
	switch phase {
	case domain.PokerPhaseDeal:
		return i18n.T("poker.phaseDeal")
	case domain.PokerPhaseExchange:
		return i18n.T("poker.phaseExchange")
	case domain.PokerPhaseSecondBet:
		return i18n.T("poker.phaseSecondBet")
	case domain.PokerPhaseEnd:
		return i18n.T("poker.phaseEnd")
	default:
		return i18n.T("poker.phaseInit")
	}
}

// ActionLogOutput emits the action-log transcript as plain text.
func (pcp *PokerCuiPresenter) ActionLogOutput(p interfaces.PokerGame) string {
	return actionLogOutputTextForSeatList(p, p.GetPlayers())
}

// OutputWithOdds appends the draw-odds table to the standard Output.
// The base Output already ends with the buildCuiOutput "==========\n" footer,
// so this appends only the odds-section body without a redundant divider.
func (pcp *PokerCuiPresenter) OutputWithOdds(p interfaces.PokerGame, lastErr error, odds []domain.PokerDrawOdds) string {
	base := pcp.Output(p, lastErr)
	if len(odds) == 0 {
		return base
	}
	var oddsBuilder strings.Builder
	oddsBuilder.WriteString(color.Bold(i18n.T("poker.drawOddsHeader")) + "\n")
	for _, o := range odds {
		oddsBuilder.WriteString(i18n.Tf("poker.drawOddsLine",
			"name", cuiPokerHandName(o.HandRank),
			"prob", fmt.Sprintf("%.2f", o.Probability*100),
			"count", strconv.Itoa(o.Count),
			"total", strconv.Itoa(o.Total)) + "\n")
	}
	return base + oddsBuilder.String()
}
