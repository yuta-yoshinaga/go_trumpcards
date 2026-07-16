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

// doubtPlayerStr returns the display string for a single Doubt player.
func doubtPlayerStr(player *domain.DoubtPlayer, i int) string {
	var b strings.Builder
	name := cuiPlayerName(player, i)
	if player.GetIsFinished() {
		b.WriteString(i18n.Tf("doubt.playerLineFinished", "name", name) + "\n")
		return b.String()
	}
	b.WriteString(i18n.Tf("doubt.playerLine",
		"name", name,
		"cards", strconv.Itoa(player.GetCardsSize())) + "\n")
	if player.GetIsHuman() {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// DoubtCuiPresenter renders the Doubt CUI view.
type DoubtCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *DoubtCuiPresenter) Output(d interfaces.DoubtGame, lastErr error) string {
	return buildCuiOutput(i18n.T("doubt.helpTitle"), func(b *strings.Builder) {
		// Players
		for i := 0; i < d.GetPlayerCnt(); i++ {
			b.WriteString(doubtPlayerStr(d.GetPlayer(i), i))
		}

		b.WriteString("----------\n")
		b.WriteString(i18n.Tf("doubt.tableLine",
			"count", strconv.Itoa(d.GetTableCardCount())) + "\n")

		// Last play
		if lastAction := d.GetLastAction(); lastAction != nil {
			b.WriteString(color.Bold(i18n.T("doubt.lastActionLabel")) + " ")
			b.WriteString(i18n.Tf("doubt.lastActionLine",
				"name", cuiPlayerName(d.GetPlayer(lastAction.PlayerIdx), lastAction.PlayerIdx),
				"value", strconv.Itoa(lastAction.ClaimedValue),
				"count", strconv.Itoa(lastAction.CardCount)) + "\n")
		}

		// Doubt result
		if lastResult := d.GetLastDoubtResult(); lastResult != nil {
			doubterName := cuiPlayerName(d.GetPlayer(lastResult.DoubterIdx), lastResult.DoubterIdx)
			cardPlayerName := cuiPlayerName(d.GetPlayer(lastResult.CardPlayerIdx), lastResult.CardPlayerIdx)
			loserName := cuiPlayerName(d.GetPlayer(lastResult.LoserIdx), lastResult.LoserIdx)
			b.WriteString(color.Bold(i18n.T("doubt.doubtLabel")) + " ")
			b.WriteString(i18n.Tf("doubt.doubtResultLeft",
				"doubter", doubterName,
				"cardPlayer", cardPlayerName) + " → ")
			if lastResult.WasLying {
				b.WriteString(color.Green(i18n.T("doubt.verdictLying")))
			} else {
				b.WriteString(color.Red(i18n.T("doubt.verdictHonest")))
			}
			b.WriteString(" " + i18n.Tf("doubt.doubtResultRight",
				"loser", loserName,
				"count", strconv.Itoa(lastResult.CardCount)) + "\n")
			if lastResult.DiscardedCount > 0 {
				b.WriteString(i18n.Tf("doubt.doubtDiscard",
					"count", strconv.Itoa(lastResult.DiscardedCount)) + "\n")
			}
			if len(lastResult.RevealedCards) > 0 {
				b.WriteString(i18n.Tf("doubt.doubtRevealed",
					"cards", cuiCardSliceStr(lastResult.RevealedCards)) + "\n")
			}
		}

		// Human action history
		if humanAction := d.GetHumanAction(); humanAction != nil {
			b.WriteString(i18n.Tf("doubt.humanActionLine",
				"value", strconv.Itoa(humanAction.ClaimedValue),
				"count", strconv.Itoa(humanAction.CardCount)) + "\n")
		}

		// CPU action history
		if cpuActions := d.GetCpuActions(); len(cpuActions) > 0 {
			b.WriteString(color.Bold(i18n.T("doubt.cpuActionsHeader")) + "\n")
			for _, action := range cpuActions {
				// Surface the CPU's tell (nervous behaviour) so the CLI player has
				// the same read-the-opponent cue the web UI shows.
				line := i18n.Tf("doubt.cpuActionLine",
					"name", cuiPlayerName(d.GetPlayer(action.PlayerIdx), action.PlayerIdx),
					"value", strconv.Itoa(action.ClaimedValue),
					"count", strconv.Itoa(action.CardCount))
				if action.HasTell {
					line += " " + color.Yellow(i18n.T("doubt.tell"))
				}
				b.WriteString(line + "\n")
			}
		}

		// Meta AI status
		if profile := d.GetHumanProfile(); profile != nil {
			b.WriteString(i18n.Tf("doubt.metaAILine",
				"games", strconv.Itoa(profile.GamesPlayed),
				"bluff", fmt.Sprintf("%.0f", profile.BluffRate(1)*100),
				"accuracy", fmt.Sprintf("%.0f", profile.DoubtAccuracy()*100)) + "\n")
		}

		cuiErrorBlock(b, lastErr)

		// Game state
		if d.GetGameEndFlag() {
			winnerIdx := d.GetWinnerIdx()
			banner := i18n.Tf("doubt.gameEnd",
				"name", cuiPlayerName(d.GetPlayer(winnerIdx), winnerIdx))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		currentTurn := d.GetCurrentTurn()
		if d.GetPhase() == domain.DoubtPhaseDoubt {
			if lastAct := d.GetLastAction(); lastAct != nil {
				b.WriteString(i18n.Tf("doubt.promptDoubtPhase",
					"name", cuiPlayerName(d.GetPlayer(lastAct.PlayerIdx), lastAct.PlayerIdx)) + "\n")
			} else {
				b.WriteString(i18n.T("doubt.promptDoubtPhaseGeneric") + "\n")
			}
			b.WriteString(i18n.T("doubt.promptDoubtHelp") + "\n")
			return
		}
		b.WriteString(i18n.Tf("doubt.promptCurrentPlayer",
			"name", cuiPlayerName(d.GetPlayer(currentTurn), currentTurn)) + "\n")
		b.WriteString(i18n.T("doubt.promptPlayHelp") + "\n")
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *DoubtCuiPresenter) ActionLogOutput(d interfaces.DoubtGame) string {
	return actionLogOutputText(d)
}
