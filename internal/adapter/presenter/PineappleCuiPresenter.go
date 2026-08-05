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

// pineappleTitleKey selects the CUI title for the shared presenter's variant:
// Irish Poker deals 4 hole cards, Crazy Pineapple discards after flop betting,
// and plain Pineapple discards before the flop.
func pineappleTitleKey(dealCount int, discardAfterFlop bool) string {
	switch {
	case dealCount >= 4:
		return "irishpoker.helpTitle"
	case discardAfterFlop:
		return "crazypineapple.helpTitle"
	default:
		return "pineapple.helpTitle"
	}
}

// PineappleCuiPresenter renders the Pineapple Poker CUI view.
type PineappleCuiPresenter struct{}

// ActionLogOutput emits the action-log transcript as plain text.
func (pp *PineappleCuiPresenter) ActionLogOutput(p interfaces.PineappleGame) string {
	return actionLogOutputText(p)
}

// Output renders the current game state for the active locale (#1699).
func (pp *PineappleCuiPresenter) Output(p interfaces.PineappleGame, lastErr error) string {
	titleKey := pineappleTitleKey(p.GetInitialDealCount(), p.IsDiscardAfterFlopBetting())
	return buildCuiOutput(i18n.T(titleKey), func(b *strings.Builder) {
		cfg := p.GetConfig()
		if cfg.TournamentMode {
			b.WriteString(i18n.Tf("pineapple.tournamentLine",
				"hand", strconv.Itoa(p.GetHandCount()),
				"sb", strconv.Itoa(cfg.SmallBlind),
				"bb", strconv.Itoa(cfg.BigBlind),
				"levelup", strconv.Itoa(cfg.BlindLevelHands)) + "\n")
			if cfg.RebuyEnabled {
				b.WriteString(i18n.Tf("pineapple.rebuyLine",
					"chips", strconv.Itoa(cfg.RebuyChips),
					"max", strconv.Itoa(cfg.RebuyMaxCount),
					"period", strconv.Itoa(cfg.RebuyPeriodHands)) + "\n")
			}
			if cfg.AddonEnabled {
				b.WriteString(i18n.Tf("pineapple.addonLine",
					"chips", strconv.Itoa(cfg.AddonChips),
					"after", strconv.Itoa(cfg.AddonAfterHand)) + "\n")
			}
		}

		b.WriteString(i18n.Tf("pineapple.tableMax", "n", strconv.Itoa(p.GetPlayerCnt())) + "\n")
		b.WriteString(i18n.Tf("pineapple.dealerLine", "idx", strconv.Itoa(p.GetDealerIdx())) + "\n")

		cc := p.GetCommunityCards()
		if len(cc) == 0 {
			b.WriteString(i18n.T("pineapple.communityNone") + "\n")
		} else {
			b.WriteString(i18n.Tf("pineapple.communityCards", "cards", cuiCardSliceStrEmoji(cc)) + "\n")
		}

		b.WriteString(i18n.Tf("pineapple.potLine", "pot", strconv.Itoa(p.GetPot())) + "\n")

		if int(cfg.BettingLimit) < len(domain.BettingLimitNames) {
			b.WriteString(i18n.Tf("pineapple.limitLine", "name", domain.BettingLimitNames[cfg.BettingLimit]) + "\n")
		}

		b.WriteString("----------\n")
		for i := 0; i < p.GetPlayerCnt(); i++ {
			player := p.GetPlayer(i)
			b.WriteString(cuiPlayerNameWithStyle(player, i))
			b.WriteString(i18n.Tf("pineapple.playerChips", "chips", strconv.Itoa(player.GetChips())))

			if player.GetTotalHands() > 0 {
				b.WriteString(i18n.Tf("pineapple.playerStats",
					"vpip", strconv.Itoa(player.GetVPIP()),
					"pfr", strconv.Itoa(player.GetPFR()),
					"tb", strconv.Itoa(player.GetThreeBet()),
					"af", player.GetAFDisplay()))
			}

			if player.GetFolded() {
				b.WriteString(color.BoldYellow(i18n.T("pineapple.playerFolded")))
			} else if player.GetAllIn() {
				b.WriteString(color.BoldYellow(i18n.T("pineapple.playerAllIn")))
			}

			if player.GetCurrentBet() > 0 {
				b.WriteString(i18n.Tf("pineapple.playerBet", "bet", strconv.Itoa(player.GetCurrentBet())))
			}
			b.WriteString("\n")

			if player.GetIsHuman() && !player.GetFolded() {
				cards := cuiCardListStrEmoji(player)
				// The discard prompt asks for a card number, so show an indexed
				// hand until the human has discarded.
				if p.GetPhase() == domain.PineapplePhaseDiscard {
					discardDone := p.GetDiscardDone()
					if i >= len(discardDone) || !discardDone[i] {
						cards = cuiIndexedCardListStrEmoji(player)
					}
				}
				b.WriteString(i18n.Tf("pineapple.humanHand", "cards", cards) + "\n")
			}
		}

		cpuActions := p.GetCpuActions()
		if len(cpuActions) > 0 {
			b.WriteString("----------\n")
			b.WriteString(color.Bold(i18n.T("pineapple.cpuActionsHeader")) + "\n")
			for _, action := range cpuActions {
				b.WriteString(i18n.Tf("pineapple.cpuActionLine",
					"idx", strconv.Itoa(action.PlayerIdx),
					"action", cuiBettingActionName(action.Action)))
				if action.Amount > 0 {
					b.WriteString(i18n.Tf("pineapple.cpuActionAmount", "amount", strconv.Itoa(action.Amount)))
				}
				b.WriteString("\n")
			}
		}

		if p.GetPhase() == domain.PineapplePhaseDiscard {
			b.WriteString("----------\n")
			b.WriteString(i18n.T("pineapple.discardHeader") + "\n")
			// The number of cards to discard is variant-dependent (Pineapple keeps
			// 2 of 3, Irish Poker keeps 2 of 4), so derive it from the deal count.
			discardCount := p.GetInitialDealCount() - 2
			b.WriteString(i18n.Tf("pineapple.discardPrompt", "count", strconv.Itoa(discardCount)) + "\n")
		}

		results := p.GetRoundResults()
		if len(results) > 0 && (p.GetPhase() == domain.PineapplePhaseEnd || p.GetPhase() == domain.PineapplePhaseShowdown) {
			b.WriteString("==========\n")
			b.WriteString(color.Bold(i18n.T("pineapple.resultsHeader")) + "\n")
			for _, r := range results {
				name := cuiPlayerName(p.GetPlayer(r.PlayerIdx), r.PlayerIdx)
				kickers := ""
				if ks := domain.FormatKickers(r.Kickers); ks != "" {
					kickers = i18n.Tf("pineapple.resultKickers", "kickers", ks)
				}
				switch {
				case r.Mucked:
					b.WriteString(i18n.Tf("pineapple.resultMucked", "name", name))
				case r.HandName != "":
					b.WriteString(i18n.Tf("pineapple.resultHand",
						"name", name,
						"hand", cuiPokerHandName(r.HandRank),
						"kickers", kickers))
				default:
					b.WriteString(i18n.Tf("pineapple.resultName", "name", name))
				}
				if r.WonAmount > 0 {
					b.WriteString(i18n.Tf("pineapple.wonAmount", "total", strconv.Itoa(r.WonAmount)))
				}
				b.WriteString("\n")
			}
		}

		if p.IsMuckAvailable() {
			b.WriteString("----------\n")
			b.WriteString(i18n.T("pineapple.muckPrompt") + "\n")
		}

		if p.GetPhase() == domain.PineapplePhaseRebuy {
			b.WriteString("----------\n")
			switch p.GetRebuyPhaseType() {
			case domain.PineappleRebuyPhaseRebuy:
				rebuyCounts := p.GetRebuyCounts()
				humanIdx := -1
				for i := 0; i < p.GetPlayerCnt(); i++ {
					if p.GetPlayer(i).GetIsHuman() {
						humanIdx = i
						break
					}
				}
				if humanIdx >= 0 {
					b.WriteString(i18n.Tf("pineapple.rebuyPrompt",
						"chips", strconv.Itoa(cfg.RebuyChips),
						"used", strconv.Itoa(rebuyCounts[humanIdx]),
						"max", strconv.Itoa(cfg.RebuyMaxCount)) + "\n")
				}
			case domain.PineappleRebuyPhaseAddon:
				b.WriteString(i18n.Tf("pineapple.addonPrompt", "chips", strconv.Itoa(cfg.AddonChips)) + "\n")
			}
		}

		cuiErrorBlock(b, lastErr)

		if p.GetGameEndFlag() {
			b.WriteString(i18n.T("pineapple.gameEnd") + "\n")
		}
	})
}
