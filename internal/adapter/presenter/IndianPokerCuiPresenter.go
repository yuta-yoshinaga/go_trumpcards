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

// IndianPokerCuiPresenter renders the Indian Poker CUI view.
type IndianPokerCuiPresenter struct{}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *IndianPokerCuiPresenter) ActionLogOutput(ip interfaces.IndianPokerGame) string {
	return actionLogOutputText(ip)
}

// Output renders the current game state for the active locale (#1699).
func (p *IndianPokerCuiPresenter) Output(ip interfaces.IndianPokerGame, lastErr error) string {
	var b strings.Builder

	b.WriteString("==========\n")
	b.WriteString(i18n.T("indianpoker.outputTitle") + "\n")
	b.WriteString("==========\n")

	b.WriteString(i18n.Tf("indianpoker.dealerLine", "idx", strconv.Itoa(ip.GetDealerIdx())) + "\n")
	b.WriteString(i18n.Tf("indianpoker.potLine", "pot", strconv.Itoa(ip.GetPot())) + "\n")

	cfg := ip.GetConfig()
	if int(cfg.BettingLimit) < len(domain.BettingLimitNames) {
		b.WriteString(i18n.Tf("indianpoker.limitLine", "name", domain.BettingLimitNames[cfg.BettingLimit]) + "\n")
	}

	b.WriteString(i18n.Tf("indianpoker.anteLine", "ante", strconv.Itoa(cfg.Ante)) + "\n")

	b.WriteString("----------\n")
	isShowdown := ip.GetPhase() == domain.IndianPokerPhaseShowdown || ip.GetPhase() == domain.IndianPokerPhaseEnd
	for i := 0; i < ip.GetPlayerCnt(); i++ {
		player := ip.GetPlayer(i)
		b.WriteString(cuiPlayerNameWithStyle(player, i))
		b.WriteString(i18n.Tf("indianpoker.playerChips", "chips", strconv.Itoa(player.GetChips())))

		if player.GetFolded() {
			b.WriteString(color.BoldYellow(i18n.T("indianpoker.playerFolded")))
		} else if player.GetAllIn() {
			b.WriteString(color.BoldYellow(i18n.T("indianpoker.playerAllIn")))
		}

		if player.GetCurrentBet() > 0 {
			b.WriteString(i18n.Tf("indianpoker.playerBet", "bet", strconv.Itoa(player.GetCurrentBet())))
		}
		b.WriteString("\n")

		// Indian Poker: humans never see their own card until showdown,
		// while CPU cards are always visible (everyone-but-yourself rule).
		if player.GetCardsSize() > 0 {
			if player.GetIsHuman() {
				if isShowdown {
					b.WriteString(i18n.Tf("indianpoker.cardLine", "card", cuiCardStrEmoji(player.GetCard(0))) + "\n")
				} else {
					b.WriteString(i18n.T("indianpoker.cardHidden") + "\n")
				}
			} else {
				b.WriteString(i18n.Tf("indianpoker.cardLine", "card", cuiCardStrEmoji(player.GetCard(0))) + "\n")
			}
		}
	}

	cpuActions := ip.GetCpuActions()
	if len(cpuActions) > 0 {
		b.WriteString("----------\n")
		b.WriteString(color.Bold(i18n.T("indianpoker.cpuActionsHeader")) + "\n")
		for _, action := range cpuActions {
			b.WriteString(i18n.Tf("indianpoker.cpuActionLine",
				"idx", strconv.Itoa(action.PlayerIdx),
				"action", cuiBettingActionName(action.Action)))
			if action.Amount > 0 {
				b.WriteString(i18n.Tf("indianpoker.cpuActionAmount", "amount", strconv.Itoa(action.Amount)))
			}
			b.WriteString("\n")
		}
	}

	results := ip.GetRoundResults()
	if len(results) > 0 && isShowdown {
		b.WriteString("==========\n")
		b.WriteString(color.Bold(i18n.T("indianpoker.resultsHeader")) + "\n")
		for _, r := range results {
			name := cuiPlayerName(ip.GetPlayer(r.PlayerIdx), r.PlayerIdx)
			if r.Card != nil {
				b.WriteString(i18n.Tf("indianpoker.resultCard",
					"name", name,
					"card", cuiCardStrEmoji(r.Card)))
			} else {
				b.WriteString(i18n.Tf("indianpoker.resultName", "name", name))
			}
			if r.WonAmount > 0 {
				b.WriteString(i18n.Tf("indianpoker.wonAmount", "total", strconv.Itoa(r.WonAmount)))
			}
			b.WriteString("\n")
		}
	}

	if lastErr != nil {
		fmt.Fprintf(&b, "%s\n", color.Red(lastErr.Error()))
	}

	if ip.GetGameEndFlag() {
		b.WriteString(i18n.T("indianpoker.gameEnd") + "\n")
	}

	return b.String()
}
