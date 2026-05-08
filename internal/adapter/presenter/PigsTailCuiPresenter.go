package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// PigsTailCuiPresenter renders the Pig's Tail CUI view.
type PigsTailCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *PigsTailCuiPresenter) Output(pt interfaces.PigsTailGame, lastErr error) string {
	return buildCuiOutput(i18n.T("pigtail.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("pigtail.header",
			"stock", strconv.Itoa(pt.GetCircleCount()),
			"center", strconv.Itoa(len(pt.GetCenter()))) + "\n")

		if topCard := pt.GetCenterTopCard(); topCard != nil {
			b.WriteString(i18n.Tf("pigtail.topCardLine", "card", cuiCardStr(topCard)) + "\n")
		}

		b.WriteString("----------\n")

		for i := 0; i < pt.GetPlayerCnt(); i++ {
			player := pt.GetPlayer(i)
			b.WriteString(i18n.Tf("pigtail.playerLine",
				"name", cuiPlayerName(player, i),
				"cards", strconv.Itoa(player.GetCardsSize())) + "\n")
		}

		b.WriteString("----------\n")

		// CPU action history
		if cpuActions := pt.GetCpuActions(); len(cpuActions) > 0 {
			b.WriteString(color.Bold(i18n.T("pigtail.cpuActionsHeader")) + "\n")
			for _, action := range cpuActions {
				name := cuiPlayerName(pt.GetPlayer(action.DrawPlayerIdx), action.DrawPlayerIdx)
				if action.PenaltyFlag {
					b.WriteString(i18n.Tf("pigtail.cpuActionPenalty",
						"name", name,
						"count", strconv.Itoa(action.PenaltyCount)) + "\n")
				} else {
					b.WriteString(i18n.Tf("pigtail.cpuActionSafe", "name", name) + "\n")
				}
			}
		}

		cuiErrorBlock(b, lastErr)

		if pt.GetGameEndFlag() {
			loserIdx := pt.GetLoserIdx()
			if loserIdx >= 0 {
				loser := pt.GetPlayer(loserIdx)
				loserName := cuiPlayerName(loser, loserIdx)
				banner := i18n.T("pigtail.gameEndPrefix") +
					color.Red(i18n.Tf("pigtail.gameEndLoser", "name", loserName)) +
					i18n.Tf("pigtail.gameEndCardSuffix",
						"count", strconv.Itoa(loser.GetCardsSize()))
				b.WriteString(banner + "\n")
			}
			return
		}
		currentTurn := pt.GetCurrentTurn()
		b.WriteString(i18n.Tf("pigtail.promptCurrentTurn",
			"name", cuiPlayerName(pt.GetPlayer(currentTurn), currentTurn)) + "\n")
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *PigsTailCuiPresenter) ActionLogOutput(pt interfaces.PigsTailGame) string {
	return actionLogOutputText(pt)
}
