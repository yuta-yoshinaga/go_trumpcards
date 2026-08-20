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

// oldMaidMaxDrawHistory is the maximum number of most-recent draw-history
// entries shown in the CUI; older entries are summarized as an omitted count.
const oldMaidMaxDrawHistory = 10

// oldMaidPlayerStr returns the display string for a single OldMaid player.
func oldMaidPlayerStr(player *domain.OldMaidPlayer, i int) string {
	var b strings.Builder
	name := cuiPlayerName(player, i)
	if player.GetIsFinished() {
		b.WriteString(i18n.Tf("oldmaid.playerLineFinished", "name", name) + "\n")
		return b.String()
	}
	b.WriteString(i18n.Tf("oldmaid.playerLine",
		"name", name,
		"cards", strconv.Itoa(player.GetCardsSize())) + "\n")
	if player.GetIsHuman() {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// OldMaidCuiPresenter renders the Old Maid CUI view.
type OldMaidCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *OldMaidCuiPresenter) Output(om interfaces.OldMaidGame, lastErr error) string {
	titleKey := "oldmaid.outputTitleNormal"
	if om.GetConfig().Mode == domain.OldMaidModeJijiNuki {
		titleKey = "oldmaid.outputTitleJijiNuki"
	}
	return buildCuiOutput(i18n.T(titleKey), func(b *strings.Builder) {
		for i := 0; i < om.GetPlayerCnt(); i++ {
			b.WriteString(oldMaidPlayerStr(om.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		if om.GetHasDrawn() {
			drawPlayerIdx := om.GetLastDrawPlayerIdx()
			drawFromIdx := om.GetLastDrawFromIdx()
			discarded := om.GetLastDiscardedPairs()
			drawer := cuiPlayerName(om.GetPlayer(drawPlayerIdx), drawPlayerIdx)
			from := cuiPlayerName(om.GetPlayer(drawFromIdx), drawFromIdx)
			drawnCard := om.GetLastDrawCard()
			drawPlayer := om.GetPlayer(drawPlayerIdx)
			b.WriteString(i18n.Tf("oldmaid.drawAction",
				"drawer", drawer, "from", from))
			// Only reveal drawn card for human players to preserve CPU game fairness
			if drawnCard != nil && drawPlayer != nil && drawPlayer.GetIsHuman() {
				b.WriteString(i18n.Tf("oldmaid.drawActionWithCard",
					"card", cuiCardStr(drawnCard)))
			}
			if discarded > 0 {
				b.WriteString(i18n.Tf("oldmaid.drawActionDiscard",
					"count", strconv.Itoa(discarded)))
			}
			b.WriteString("\n")
		}

		// CPU action history
		cpuActions := om.GetCpuActions()
		if len(cpuActions) > 0 {
			b.WriteString(color.Bold(i18n.T("oldmaid.cpuActionsHeader")) + "\n")
			for _, action := range cpuActions {
				drawer := cuiPlayerName(om.GetPlayer(action.DrawPlayerIdx), action.DrawPlayerIdx)
				from := cuiPlayerName(om.GetPlayer(action.DrawFromIdx), action.DrawFromIdx)
				b.WriteString(i18n.Tf("oldmaid.drawAction",
					"drawer", drawer, "from", from))
				// CPU drawn card is intentionally hidden to preserve game fairness
				if action.DiscardedPairs > 0 {
					b.WriteString(i18n.Tf("oldmaid.drawActionDiscard",
						"count", strconv.Itoa(action.DiscardedPairs)))
				}
				b.WriteString("\n")
			}
		}

		// Draw history (cap to the most recent entries so a long game doesn't
		// flood the terminal; the full log stays available via ActionLogOutput).
		drawHistory := om.GetDrawHistory()
		if len(drawHistory) > 0 {
			b.WriteString(color.Bold(i18n.T("oldmaid.drawHistoryHeader")) + "\n")
			start := 0
			if len(drawHistory) > oldMaidMaxDrawHistory {
				start = len(drawHistory) - oldMaidMaxDrawHistory
				b.WriteString(i18n.Tf("oldmaid.drawHistoryOmitted", "count", strconv.Itoa(start)) + "\n")
			}
			for i := start; i < len(drawHistory); i++ {
				entry := drawHistory[i]
				drawer := cuiPlayerName(om.GetPlayer(entry.DrawPlayerIdx), entry.DrawPlayerIdx)
				from := cuiPlayerName(om.GetPlayer(entry.DrawFromIdx), entry.DrawFromIdx)
				b.WriteString(i18n.Tf("oldmaid.drawHistoryEntry",
					"idx", strconv.Itoa(i+1),
					"drawer", drawer,
					"from", from))
				if entry.DiscardedPairs > 0 {
					b.WriteString(i18n.Tf("oldmaid.drawHistoryDiscard",
						"count", strconv.Itoa(entry.DiscardedPairs)))
				}
				if entry.DrawerFinished {
					b.WriteString(i18n.Tf("oldmaid.drawHistoryFinished", "name", drawer))
				}
				if entry.TargetFinished {
					b.WriteString(i18n.Tf("oldmaid.drawHistoryFinished", "name", from))
				}
				b.WriteString("\n")
			}
		}

		// Meta AI status
		if profile := om.GetHumanProfile(); profile != nil {
			rate := (profile.PickRate(0) + profile.PickRate(2)) * 100
			b.WriteString(i18n.Tf("oldmaid.metaAILine",
				"games", strconv.Itoa(profile.GamesPlayed),
				"rate", fmt.Sprintf("%.0f", rate)) + "\n")
		}

		cuiErrorBlock(b, lastErr)

		if om.GetGameEndFlag() {
			loserIdx := om.GetLoserIdx()
			if loserIdx >= 0 {
				loserName := cuiPlayerName(om.GetPlayer(loserIdx), loserIdx)
				gameEndLine := i18n.T("oldmaid.gameEndPrefix") +
					color.Red(i18n.Tf("oldmaid.gameEndLoser", "name", loserName))
				if om.GetConfig().Mode == domain.OldMaidModeJijiNuki && om.GetRemovedCard() != nil {
					gameEndLine += i18n.Tf("oldmaid.gameEndJijiNukiSuffix",
						"card", cuiCardStr(om.GetRemovedCard()))
				}
				b.WriteString(gameEndLine + "\n")
			}
			return
		}
		currentTurn := om.GetCurrentTurn()
		currentName := cuiPlayerName(om.GetPlayer(currentTurn), currentTurn)
		targetIdx := om.GetNextDrawTargetIdx()
		if targetIdx >= 0 {
			targetName := cuiPlayerName(om.GetPlayer(targetIdx), targetIdx)
			b.WriteString(i18n.Tf("oldmaid.promptCurrentTurnWithTarget",
				"name", currentName,
				"target", targetName) + "\n")
		} else {
			b.WriteString(i18n.Tf("oldmaid.promptCurrentTurn", "name", currentName) + "\n")
		}
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *OldMaidCuiPresenter) ActionLogOutput(om interfaces.OldMaidGame) string {
	return actionLogOutputTextForSeats[*domain.OldMaidPlayer](om)
}
