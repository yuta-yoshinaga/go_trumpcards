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

// ShortDeckCuiPresenter renders the Short Deck Hold'em CUI view.
type ShortDeckCuiPresenter struct{}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *ShortDeckCuiPresenter) ActionLogOutput(o interfaces.ShortDeckGame) string {
	return actionLogOutputText(o)
}

// Output renders the current game state for the active locale (#1699).
func (p *ShortDeckCuiPresenter) Output(o interfaces.ShortDeckGame, lastErr error) string {
	return buildCuiOutput(i18n.T("shortdeck.outputTitle"), func(b *strings.Builder) {
		// Short Deck reshuffles the hand ranking (flush beats full house) and
		// counts A-6-7-8-9 as a straight; surface it every render so the CUI
		// player is not caught out by the 36-card deck's rules.
		b.WriteString(i18n.T("shortdeck.ruleReminderLine") + "\n")

		cfg := o.GetConfig()
		if cfg.TournamentMode {
			b.WriteString(i18n.Tf("shortdeck.tournamentLine",
				"hand", strconv.Itoa(o.GetHandCount()),
				"sb", strconv.Itoa(cfg.SmallBlind),
				"bb", strconv.Itoa(cfg.BigBlind),
				"levelup", strconv.Itoa(cfg.BlindLevelHands)) + "\n")
			if cfg.RebuyEnabled {
				b.WriteString(i18n.Tf("shortdeck.rebuyLine",
					"chips", strconv.Itoa(cfg.RebuyChips),
					"max", strconv.Itoa(cfg.RebuyMaxCount),
					"period", strconv.Itoa(cfg.RebuyPeriodHands)) + "\n")
			}
			if cfg.AddonEnabled {
				b.WriteString(i18n.Tf("shortdeck.addonLine",
					"chips", strconv.Itoa(cfg.AddonChips),
					"after", strconv.Itoa(cfg.AddonAfterHand)) + "\n")
			}
		}

		b.WriteString(i18n.Tf("shortdeck.tableMax", "n", strconv.Itoa(o.GetPlayerCnt())) + "\n")
		b.WriteString(i18n.Tf("shortdeck.dealerLine", "idx", strconv.Itoa(o.GetDealerIdx())) + "\n")

		cc := o.GetCommunityCards()
		if len(cc) == 0 {
			b.WriteString(i18n.T("shortdeck.communityNone") + "\n")
		} else {
			b.WriteString(i18n.Tf("shortdeck.communityCards", "cards", cuiCardSliceStrEmoji(cc)) + "\n")
		}

		b.WriteString(i18n.Tf("shortdeck.potLine", "pot", strconv.Itoa(o.GetPot())) + "\n")

		if int(cfg.BettingLimit) < len(domain.BettingLimitNames) {
			b.WriteString(i18n.Tf("shortdeck.limitLine", "name", domain.BettingLimitNames[cfg.BettingLimit]) + "\n")
		}

		b.WriteString("----------\n")
		for i := 0; i < o.GetPlayerCnt(); i++ {
			player := o.GetPlayer(i)
			b.WriteString(cuiPlayerNameWithStyle(player, i))
			b.WriteString(i18n.Tf("shortdeck.playerChips", "chips", strconv.Itoa(player.GetChips())))

			if player.GetTotalHands() > 0 {
				b.WriteString(i18n.Tf("shortdeck.playerStats",
					"vpip", strconv.Itoa(player.GetVPIP()),
					"pfr", strconv.Itoa(player.GetPFR()),
					"tb", strconv.Itoa(player.GetThreeBet()),
					"af", player.GetAFDisplay()))
			}

			if player.GetFolded() {
				b.WriteString(color.BoldYellow(i18n.T("shortdeck.playerFolded")))
			} else if player.GetAllIn() {
				b.WriteString(color.BoldYellow(i18n.T("shortdeck.playerAllIn")))
			}

			if player.GetCurrentBet() > 0 {
				b.WriteString(i18n.Tf("shortdeck.playerBet", "bet", strconv.Itoa(player.GetCurrentBet())))
			}
			b.WriteString("\n")

			if player.GetIsHuman() && !player.GetFolded() {
				b.WriteString(i18n.Tf("shortdeck.humanHand", "cards", cuiCardListStrEmoji(player)) + "\n")
			}
		}

		cpuActions := o.GetCpuActions()
		if len(cpuActions) > 0 {
			b.WriteString("----------\n")
			b.WriteString(color.Bold(i18n.T("shortdeck.cpuActionsHeader")) + "\n")
			for _, action := range cpuActions {
				b.WriteString(i18n.Tf("shortdeck.cpuActionLine",
					"idx", strconv.Itoa(action.PlayerIdx),
					"action", cuiBettingActionName(action.Action)))
				if action.Amount > 0 {
					b.WriteString(i18n.Tf("shortdeck.cpuActionAmount", "amount", strconv.Itoa(action.Amount)))
				}
				b.WriteString("\n")
			}
		}

		results := o.GetRoundResults()
		if len(results) > 0 && (o.GetPhase() == domain.ShortDeckPhaseEnd || o.GetPhase() == domain.ShortDeckPhaseShowdown) {
			b.WriteString("==========\n")
			b.WriteString(color.Bold(i18n.T("shortdeck.resultsHeader")) + "\n")
			for _, r := range results {
				name := cuiPlayerName(o.GetPlayer(r.PlayerIdx), r.PlayerIdx)
				kickers := ""
				if ks := domain.FormatKickers(r.Kickers); ks != "" {
					kickers = i18n.Tf("shortdeck.resultKickers", "kickers", ks)
				}
				switch {
				case r.Mucked:
					b.WriteString(i18n.Tf("shortdeck.resultMucked", "name", name))
				case r.HandName != "":
					b.WriteString(i18n.Tf("shortdeck.resultHand",
						"name", name,
						"hand", r.HandName,
						"kickers", kickers))
				default:
					b.WriteString(i18n.Tf("shortdeck.resultName", "name", name))
				}
				if r.WonAmount > 0 {
					b.WriteString(i18n.Tf("shortdeck.wonAmount", "total", strconv.Itoa(r.WonAmount)))
				}
				b.WriteString("\n")
			}
		}

		if o.IsMuckAvailable() {
			b.WriteString("----------\n")
			b.WriteString(i18n.T("shortdeck.muckPrompt") + "\n")
		}

		if o.GetPhase() == domain.ShortDeckPhaseRebuy {
			b.WriteString("----------\n")
			switch o.GetRebuyPhaseType() {
			case domain.ShortDeckRebuyPhaseRebuy:
				rebuyCounts := o.GetRebuyCounts()
				humanIdx := -1
				for i := 0; i < o.GetPlayerCnt(); i++ {
					if o.GetPlayer(i).GetIsHuman() {
						humanIdx = i
						break
					}
				}
				if humanIdx >= 0 {
					b.WriteString(i18n.Tf("shortdeck.rebuyPrompt",
						"chips", strconv.Itoa(cfg.RebuyChips),
						"used", strconv.Itoa(rebuyCounts[humanIdx]),
						"max", strconv.Itoa(cfg.RebuyMaxCount)) + "\n")
				}
			case domain.ShortDeckRebuyPhaseAddon:
				b.WriteString(i18n.Tf("shortdeck.addonPrompt", "chips", strconv.Itoa(cfg.AddonChips)) + "\n")
			}
		}

		cuiErrorBlock(b, lastErr)

		if o.GetGameEndFlag() {
			b.WriteString(i18n.T("shortdeck.gameEnd") + "\n")
		}
	})
}
