package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// shitheadPlayerStr returns the display string for a single Shithead player.
func shitheadPlayerStr(player *domain.ShitheadPlayer, idx int, currentTurn int) string {
	var b strings.Builder
	name := cuiPlayerName(player, idx)
	turnSuffix := ""
	if idx == currentTurn {
		turnSuffix = i18n.T("shithead.playerTurnSuffix")
	}
	if player.GetIsFinished() {
		b.WriteString(name + turnSuffix)
		b.WriteString(i18n.Tf("shithead.playerFinished",
			"rank", strconv.Itoa(player.GetRank())) + "\n")
		return b.String()
	}
	b.WriteString(i18n.Tf("shithead.playerLine",
		"name", name+turnSuffix,
		"hand", strconv.Itoa(player.GetCardsSize()),
		"up", strconv.Itoa(player.GetFaceUpSize()),
		"down", strconv.Itoa(player.GetFaceDownSize())) + "\n")
	if player.GetIsHuman() {
		if player.GetCardsSize() > 0 {
			b.WriteString(i18n.T("shithead.handLabel"))
			b.WriteString(cuiIndexedCardListStr(player) + "\n")
		}
		if player.GetFaceUpSize() > 0 {
			b.WriteString(i18n.T("shithead.faceupLabel"))
			b.WriteString(cuiCardSliceStr(player.GetFaceUpCards()) + "\n")
		}
	} else if player.GetFaceUpSize() > 0 {
		b.WriteString(i18n.T("shithead.faceupLabel"))
		b.WriteString(cuiCardSliceStr(player.GetFaceUpCards()) + "\n")
	}
	return b.String()
}

// ShitheadCuiPresenter renders the Shithead CUI view.
type ShitheadCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *ShitheadCuiPresenter) Output(sg interfaces.ShitheadGame, lastErr error) string {
	return buildCuiOutput(i18n.T("shithead.outputTitle"), func(b *strings.Builder) {
		currentTurn := sg.GetCurrentTurn()
		for i := 0; i < sg.GetPlayerCnt(); i++ {
			b.WriteString(shitheadPlayerStr(sg.GetPlayer(i), i, currentTurn))
		}

		b.WriteString("----------\n")

		// Discard pile + stock
		discard := sg.GetDiscardPile()
		if len(discard) > 0 {
			b.WriteString(i18n.Tf("shithead.discardLine",
				"cards", cuiCardSliceStr(discard)) + "\n")
		} else {
			b.WriteString(i18n.T("shithead.discardEmpty") + "\n")
		}
		b.WriteString(i18n.Tf("shithead.stockLine",
			"count", strconv.Itoa(sg.GetStockSize())) + "\n")
		if sg.GetSevenActive() {
			b.WriteString(color.BoldYellow(i18n.T("shithead.noticeSeven")) + "\n")
		}
		if sg.GetSkipNext() {
			b.WriteString(color.BoldYellow(i18n.T("shithead.noticeSkip")) + "\n")
		}

		// Last human action
		if humanAction := sg.GetHumanAction(); humanAction != nil {
			b.WriteString(formatShitheadAction(sg, humanAction))
		}

		// CPU action history
		cpuActions := sg.GetCpuActions()
		if len(cpuActions) > 0 {
			b.WriteString(color.Bold(i18n.T("shithead.cpuActionsHeader")) + "\n")
			for _, action := range cpuActions {
				b.WriteString(formatShitheadAction(sg, action))
			}
		}

		cuiErrorBlock(b, lastErr)

		if sg.GetGameEndFlag() {
			b.WriteString(i18n.T("shithead.gameEnd") + "\n")
			for i := 0; i < sg.GetPlayerCnt(); i++ {
				player := sg.GetPlayer(i)
				rank := player.GetRank()
				suffix := ""
				if rank == sg.GetPlayerCnt() {
					suffix = i18n.T("shithead.rankShithead")
				}
				b.WriteString(i18n.Tf("shithead.rankLine",
					"name", cuiPlayerName(player, i),
					"rank", strconv.Itoa(rank),
					"suffix", suffix) + "\n")
			}
			return
		}
		source := sg.CurrentSource()
		currentName := cuiPlayerName(sg.GetPlayer(currentTurn), currentTurn)
		b.WriteString(i18n.Tf("shithead.promptCurrentTurn",
			"name", currentName,
			"source", source) + "\n")
		// On a blind (face-down) human turn, list the selectable slot indices so
		// the player knows which number to play; the card faces stay hidden.
		if source == domain.ShitheadSourceFaceDown {
			if human := sg.GetPlayer(currentTurn); human != nil && human.GetIsHuman() && human.GetFaceDownSize() > 0 {
				slots := make([]string, human.GetFaceDownSize())
				for i := range slots {
					slots[i] = "[" + strconv.Itoa(i) + "]??"
				}
				b.WriteString(i18n.Tf("shithead.facedownSlots",
					"slots", strings.Join(slots, " ")) + "\n")
			}
		}
		b.WriteString(i18n.T("shithead.promptPlayHelp") + "\n")
	})
}

// formatShitheadAction returns one line describing a player action.
func formatShitheadAction(sg interfaces.ShitheadGame, action *domain.ShitheadCpuAction) string {
	name := cuiPlayerName(sg.GetPlayer(action.PlayerIdx), action.PlayerIdx)
	if action.Pickup {
		return i18n.Tf("shithead.actionPickup", "name", name) + "\n"
	}
	suffix := ""
	if action.Burned {
		suffix += i18n.T("shithead.actionSuffixBurned")
	}
	if action.Skipped {
		suffix += i18n.T("shithead.actionSuffixSkipped")
	}
	return i18n.Tf("shithead.actionPlay",
		"name", name,
		"cards", cuiCardSliceStr(action.PlayedCards),
		"source", action.Source,
		"suffix", suffix) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *ShitheadCuiPresenter) ActionLogOutput(sg interfaces.ShitheadGame) string {
	return actionLogOutputText(sg)
}
