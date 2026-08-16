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

// HoldemCuiPresenter renders the Texas Hold'em CUI view.
type HoldemCuiPresenter struct{}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *HoldemCuiPresenter) ActionLogOutput(h interfaces.HoldemGame) string {
	return actionLogOutputText(h)
}

// Output renders the current game state for the active locale (#1699).
func (p *HoldemCuiPresenter) Output(h interfaces.HoldemGame, lastErr error) string {
	return buildCuiOutput(i18n.T("holdem.outputTitle"), func(b *strings.Builder) {
		cfg := h.GetConfig()
		if cfg.TournamentMode {
			b.WriteString(i18n.Tf("holdem.tournamentLine",
				"hand", strconv.Itoa(h.GetHandCount()),
				"sb", strconv.Itoa(cfg.SmallBlind),
				"bb", strconv.Itoa(cfg.BigBlind),
				"levelup", strconv.Itoa(cfg.BlindLevelHands)) + "\n")
			if cfg.RebuyEnabled {
				b.WriteString(i18n.Tf("holdem.rebuyLine",
					"chips", strconv.Itoa(cfg.RebuyChips),
					"max", strconv.Itoa(cfg.RebuyMaxCount),
					"period", strconv.Itoa(cfg.RebuyPeriodHands)) + "\n")
			}
			if cfg.AddonEnabled {
				b.WriteString(i18n.Tf("holdem.addonLine",
					"chips", strconv.Itoa(cfg.AddonChips),
					"after", strconv.Itoa(cfg.AddonAfterHand)) + "\n")
			}
		} else {
			// **ブラインド額はモードによらず効いている** (minRaise = BigBlind)。
			// tournamentLine の中にしか無かったので、トーナメントでない大半の
			// プレイでは額が一切出ていなかった。Web のヘッダーは常に出している。
			// トーナメント側では tournamentLine が同じ SB/BB を含むので重ねない。
			b.WriteString(i18n.Tf("holdem.blindsLine",
				"sb", strconv.Itoa(cfg.SmallBlind),
				"bb", strconv.Itoa(cfg.BigBlind)) + "\n")
		}

		b.WriteString(i18n.Tf("holdem.tableMax", "n", strconv.Itoa(h.GetPlayerCnt())) + "\n")
		b.WriteString(i18n.Tf("holdem.dealerLine", "idx", strconv.Itoa(h.GetDealerIdx())) + "\n")

		cc := h.GetCommunityCards()
		if len(cc) == 0 {
			b.WriteString(i18n.T("holdem.communityNone") + "\n")
		} else {
			b.WriteString(i18n.Tf("holdem.communityCards", "cards", cuiCardSliceStrEmoji(cc)) + "\n")
		}

		b.WriteString(i18n.Tf("holdem.potLine", "pot", strconv.Itoa(h.GetPot())) + "\n")

		if int(cfg.BettingLimit) < len(domain.BettingLimitNames) {
			b.WriteString(i18n.Tf("holdem.limitLine", "name", domain.BettingLimitNames[cfg.BettingLimit]) + "\n")
		}

		b.WriteString("----------\n")
		for i := 0; i < h.GetPlayerCnt(); i++ {
			player := h.GetPlayer(i)
			b.WriteString(cuiPlayerNameWithStyle(player, i))
			b.WriteString(i18n.Tf("holdem.playerChips", "chips", strconv.Itoa(player.GetChips())))

			if player.GetTotalHands() > 0 {
				b.WriteString(i18n.Tf("holdem.playerStats",
					"vpip", strconv.Itoa(player.GetVPIP()),
					"pfr", strconv.Itoa(player.GetPFR()),
					"tb", strconv.Itoa(player.GetThreeBet()),
					"af", player.GetAFDisplay()))
			}

			if player.GetFolded() {
				b.WriteString(color.BoldYellow(i18n.T("holdem.playerFolded")))
			} else if player.GetAllIn() {
				b.WriteString(color.BoldYellow(i18n.T("holdem.playerAllIn")))
			}

			if player.GetCurrentBet() > 0 {
				b.WriteString(i18n.Tf("holdem.playerBet", "bet", strconv.Itoa(player.GetCurrentBet())))
			}
			b.WriteString("\n")

			if player.GetIsHuman() && !player.GetFolded() {
				b.WriteString(i18n.Tf("holdem.humanHand", "cards", cuiCardListStrEmoji(player)) + "\n")
			}
		}

		// Learning mode: show equity (win probability) and pot odds on the
		// human's turn while equity is available (pre-flop through river, human
		// not folded). GetEquity returns nil otherwise, hiding the display.
		if h.IsHumanTurn() {
			if eq := h.GetEquity(); eq != nil {
				potOdds := h.GetPotOdds()
				b.WriteString("----------\n")
				b.WriteString(color.Bold(i18n.T("holdem.learningHeader")) + "\n")
				b.WriteString(i18n.Tf("holdem.learningLine",
					"equity", fmt.Sprintf("%.1f", eq.Equity*100),
					"potodds", fmt.Sprintf("%.1f", potOdds)) + "\n")
				if potOdds > 0 {
					if eq.Equity*100 > potOdds {
						b.WriteString(i18n.T("holdem.learningEvPlus") + "\n")
					} else {
						b.WriteString(i18n.T("holdem.learningEvMinus") + "\n")
					}
				}
			}
		}

		cpuActions := h.GetCpuActions()
		if len(cpuActions) > 0 {
			b.WriteString("----------\n")
			b.WriteString(color.Bold(i18n.T("holdem.cpuActionsHeader")) + "\n")
			for _, action := range cpuActions {
				b.WriteString(i18n.Tf("holdem.cpuActionLine",
					"idx", strconv.Itoa(action.PlayerIdx),
					"action", cuiBettingActionName(action.Action)))
				if action.Amount > 0 {
					b.WriteString(i18n.Tf("holdem.cpuActionAmount", "amount", strconv.Itoa(action.Amount)))
				}
				b.WriteString("\n")
			}
		}

		results := h.GetRoundResults()
		if len(results) > 0 && (h.GetPhase() == domain.HoldemPhaseEnd || h.GetPhase() == domain.HoldemPhaseShowdown) {
			b.WriteString("==========\n")
			b.WriteString(color.Bold(i18n.T("holdem.resultsHeader")) + "\n")
			for _, r := range results {
				name := cuiPlayerName(h.GetPlayer(r.PlayerIdx), r.PlayerIdx)
				kickers := ""
				if ks := domain.FormatKickers(r.Kickers); ks != "" {
					kickers = i18n.Tf("holdem.resultKickers", "kickers", ks)
				}
				switch {
				case r.Mucked:
					b.WriteString(i18n.Tf("holdem.resultMucked", "name", name))
				case r.HandName != "":
					b.WriteString(i18n.Tf("holdem.resultHand",
						"name", name,
						"hand", cuiPokerHandName(r.HandRank),
						"kickers", kickers))
					// **Web は勝利役を構成する5枚をハイライトしているのに、CUI は
					// 役名とキッカーだけだった (#4679)。**僅差の役 (ツーペアの
					// キッカー勝負など) でどのカードが決め手か分からない。
					// ショーダウン時点で bestHand は確定済みなので、そのまま出す。
					if best := h.GetPlayer(r.PlayerIdx).GetBestHand(); len(best) > 0 {
						b.WriteString(i18n.Tf("holdem.resultBestFive",
							"cards", cuiCardSliceStrEmoji(best)))
					}
				default:
					b.WriteString(i18n.Tf("holdem.resultName", "name", name))
				}
				if r.WonAmount > 0 {
					b.WriteString(i18n.Tf("holdem.wonAmount", "total", strconv.Itoa(r.WonAmount)))
				}
				b.WriteString("\n")
			}
		}

		if h.IsMuckAvailable() {
			b.WriteString("----------\n")
			b.WriteString(i18n.T("holdem.muckPrompt") + "\n")
		}

		if h.GetPhase() == domain.HoldemPhaseRebuy {
			b.WriteString("----------\n")
			switch h.GetRebuyPhaseType() {
			case domain.HoldemRebuyPhaseRebuy:
				rebuyCounts := h.GetRebuyCounts()
				humanIdx := -1
				for i := 0; i < h.GetPlayerCnt(); i++ {
					if h.GetPlayer(i).GetIsHuman() {
						humanIdx = i
						break
					}
				}
				if humanIdx >= 0 {
					b.WriteString(i18n.Tf("holdem.rebuyPrompt",
						"chips", strconv.Itoa(cfg.RebuyChips),
						"used", strconv.Itoa(rebuyCounts[humanIdx]),
						"max", strconv.Itoa(cfg.RebuyMaxCount)) + "\n")
				}
			case domain.HoldemRebuyPhaseAddon:
				b.WriteString(i18n.Tf("holdem.addonPrompt", "chips", strconv.Itoa(cfg.AddonChips)) + "\n")
			}
		}

		cuiErrorBlock(b, lastErr)

		if h.GetGameEndFlag() {
			b.WriteString(i18n.T("holdem.gameEnd") + "\n")
		}
	})
}
