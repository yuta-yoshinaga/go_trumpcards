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

// DeuceToSevenCuiPresenter renders 2-7 Triple Draw state for the CLI.
type DeuceToSevenCuiPresenter struct{}

// Output produces the CUI-rendered state for the active locale.
func (dcp *DeuceToSevenCuiPresenter) Output(g interfaces.DeuceToSevenGame, lastErr error) string {
	return buildCuiOutput(i18n.T("deucetoseven.outputTitle"), func(b *strings.Builder) {
		players := g.GetPlayers()

		b.WriteString(i18n.Tf("deucetoseven.dealerLine", "idx", strconv.Itoa(g.GetDealerIdx())) + "\n")
		b.WriteString(i18n.Tf("deucetoseven.potLine", "pot", strconv.Itoa(g.GetPot())) + "\n")

		if drawIdx := g.GetDrawIndex(); drawIdx > 0 {
			b.WriteString(i18n.Tf("deucetoseven.drawLine",
				"idx", strconv.Itoa(drawIdx),
				"max", strconv.Itoa(domain.DeuceToSevenMaxDraws)) + "\n")
		} else {
			b.WriteString(i18n.T("deucetoseven.drawPredraw") + "\n")
		}

		cfg := g.GetConfig()
		if int(cfg.BettingLimit) < len(domain.BettingLimitNames) {
			b.WriteString(i18n.Tf("deucetoseven.limitLine", "name", domain.BettingLimitNames[cfg.BettingLimit]) + "\n")
		}

		b.WriteString("----------\n")
		isEnd := g.GetPhase() == domain.DeuceToSevenPhaseEnd
		for i, pl := range players {
			b.WriteString(cuiPlayerNameWithStyle(pl, i))
			b.WriteString(i18n.Tf("deucetoseven.playerChips", "chips", strconv.Itoa(pl.GetChips())))
			switch {
			case pl.GetFolded():
				b.WriteString(color.BoldYellow(i18n.T("deucetoseven.playerFolded")))
			case pl.GetAllIn():
				b.WriteString(color.BoldYellow(i18n.T("deucetoseven.playerAllIn")))
			}
			if pl.GetCurrentBet() > 0 {
				b.WriteString(i18n.Tf("deucetoseven.playerBet", "bet", strconv.Itoa(pl.GetCurrentBet())))
			}
			if pl.GetDrawCount() > 0 {
				b.WriteString(i18n.Tf("deucetoseven.playerExchange", "count", strconv.Itoa(pl.GetDrawCount())))
			}
			b.WriteString("\n")

			if pl.GetIsHuman() && !pl.GetFolded() {
				handStr := cuiIndexedCardListStrEmoji(pl)
				if isEnd {
					b.WriteString(i18n.Tf("deucetoseven.humanHandWithName",
						"cards", handStr,
						"name", pl.GetHandName()) + "\n")
				} else {
					b.WriteString(i18n.Tf("deucetoseven.humanHand", "cards", handStr) + "\n")
				}
			}
			if !pl.GetIsHuman() && isEnd && !pl.GetFolded() {
				b.WriteString(i18n.Tf("deucetoseven.humanHandWithName",
					"cards", cuiCardListStrEmoji(pl),
					"name", pl.GetHandName()) + "\n")
			}
		}

		cpuActions := g.GetCpuActions()
		if len(cpuActions) > 0 {
			b.WriteString("----------\n")
			b.WriteString(color.Bold(i18n.T("deucetoseven.cpuActionsHeader")) + "\n")
			for _, a := range cpuActions {
				b.WriteString(i18n.Tf("deucetoseven.cpuActionLine",
					"idx", strconv.Itoa(a.PlayerIdx),
					"round", a.RoundLabel,
					"action", cuiBettingActionName(a.Action)))
				if a.Amount > 0 {
					b.WriteString(i18n.Tf("deucetoseven.cpuActionAmount", "amount", strconv.Itoa(a.Amount)))
				}
				b.WriteString("\n")
			}
		}

		cpuExchanges := g.GetCpuExchanges()
		if len(cpuExchanges) > 0 {
			b.WriteString("----------\n")
			b.WriteString(color.Bold(i18n.T("deucetoseven.cpuExchangesHeader")) + "\n")
			for _, e := range cpuExchanges {
				b.WriteString(i18n.Tf("deucetoseven.cpuExchangeLine",
					"idx", strconv.Itoa(e.PlayerIdx),
					"drawIdx", strconv.Itoa(e.DrawIndex),
					"count", strconv.Itoa(e.ExchangeCount)) + "\n")
			}
		}

		results := g.GetRoundResults()
		if len(results) > 0 && isEnd {
			b.WriteString("==========\n")
			b.WriteString(color.Bold(i18n.T("deucetoseven.resultsHeader")) + "\n")
			for _, r := range results {
				name := cuiPlayerName(players[r.PlayerIdx], r.PlayerIdx)
				if r.HandName != "" {
					b.WriteString(i18n.Tf("deucetoseven.resultHand",
						"name", name,
						"hand", r.HandName))
				} else {
					b.WriteString(i18n.Tf("deucetoseven.resultName", "name", name))
				}
				if r.WonAmount > 0 {
					b.WriteString(i18n.Tf("deucetoseven.wonAmount", "total", strconv.Itoa(r.WonAmount)))
				}
				b.WriteString("\n")
			}
		}

		cuiErrorBlock(b, lastErr)
		if g.GetGameEndFlag() {
			b.WriteString(i18n.T("deucetoseven.gameEnd") + "\n")
		}
	})
}

// HintOutput recommends which cards to exchange (or to stand pat) during the
// human's draw phase, reusing the shared domain suggestion logic.
func (dcp *DeuceToSevenCuiPresenter) HintOutput(g interfaces.DeuceToSevenGame) string {
	if g.GetGameEndFlag() || g.GetPhase() != domain.DeuceToSevenPhaseDraw {
		return i18n.T("deucetoseven.hintNotYourTurn") + "\n"
	}
	players := g.GetPlayers()
	humanSeat := -1
	for i, pl := range players {
		if pl != nil && pl.GetIsHuman() {
			humanSeat = i
			break
		}
	}
	if humanSeat < 0 || g.GetCurrentTurn() != humanSeat {
		return i18n.T("deucetoseven.hintNotYourTurn") + "\n"
	}
	ex := g.SuggestExchange(humanSeat)
	if len(ex) == 0 {
		return i18n.T("deucetoseven.hintStandPat") + "\n"
	}
	human := players[humanSeat]
	idxParts := make([]string, len(ex))
	cardParts := make([]string, len(ex))
	for i, hi := range ex {
		idxParts[i] = strconv.Itoa(hi)
		cardParts[i] = cuiCardStr(human.GetCard(hi))
	}
	return i18n.Tf("deucetoseven.hintExchange",
		"idxs", strings.Join(idxParts, ", "),
		"cards", strings.Join(cardParts, ","),
		"cmd", strings.Join(idxParts, " ")) + "\n"
}

// ActionLogOutput renders the action log as plain text.
func (dcp *DeuceToSevenCuiPresenter) ActionLogOutput(g interfaces.DeuceToSevenGame) string {
	return actionLogOutputText(g)
}
