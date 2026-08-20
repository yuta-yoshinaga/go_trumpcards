package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// CassinoCuiPresenter renders the Cassino CUI view.
type CassinoCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *CassinoCuiPresenter) Output(cg interfaces.CassinoGame, lastErr error) string {
	return buildCuiOutput(i18n.T("cassino.helpTitle"), func(b *strings.Builder) {
		for i := 0; i < cg.GetPlayerCnt(); i++ {
			b.WriteString(cassinoPlayerStr(cg.GetPlayer(i), i))
		}
		b.WriteString("----------\n")

		if tableCards := cg.GetTableCards(); len(tableCards) > 0 {
			b.WriteString(i18n.Tf("cassino.tableLine",
				"cards", cuiCardSliceStr(tableCards)) + "\n")
		} else {
			b.WriteString(i18n.T("cassino.tableEmpty") + "\n")
		}

		// Builds
		if builds := cg.GetBuilds(); len(builds) > 0 {
			b.WriteString(color.Bold(i18n.T("cassino.buildsHeader")) + "\n")
			for i, build := range builds {
				groupParts := make([]string, len(build.Groups))
				for gi, g := range build.Groups {
					groupParts[gi] = cuiCardSliceStr(g)
				}
				b.WriteString(i18n.Tf("cassino.buildLine",
					"idx", strconv.Itoa(i),
					"value", strconv.Itoa(build.Value),
					"owner", cuiPlayerName(cg.GetPlayer(build.OwnerIdx), build.OwnerIdx),
					"multi", strconv.FormatBool(build.IsMulti),
					"cards", strings.Join(groupParts, " | ")) + "\n")
			}
		}

		// Human action
		if ha := cg.GetHumanAction(); ha != nil {
			b.WriteString(i18n.Tf("cassino.humanActionLine",
				"text", cassinoActionStr(ha)) + "\n")
		}
		// CPU action history
		if cpu := cg.GetCpuActions(); len(cpu) > 0 {
			b.WriteString(color.Bold(i18n.T("cassino.cpuActionsHeader")) + "\n")
			for _, a := range cpu {
				b.WriteString(i18n.Tf("cassino.cpuActionLine",
					"name", cuiPlayerName(cg.GetPlayer(a.PlayerIdx), a.PlayerIdx),
					"text", cassinoActionStr(a)) + "\n")
			}
		}

		cuiErrorBlock(b, lastErr)

		if cg.GetGameEndFlag() {
			b.WriteString(i18n.T("cassino.gameEnd") + "\n")
			for i := 0; i < cg.GetPlayerCnt(); i++ {
				pl := cg.GetPlayer(i)
				if pl == nil {
					continue
				}
				b.WriteString(i18n.Tf("cassino.scoreEntry",
					"name", cuiPlayerName(pl, i),
					"score", strconv.Itoa(pl.GetTotalScore())) + "\n")
			}
			return
		}
		currentTurn := cg.GetCurrentTurn()
		b.WriteString(i18n.Tf("cassino.promptCurrentTurn",
			"name", cuiPlayerName(cg.GetPlayer(currentTurn), currentTurn)) + "\n")
		b.WriteString(i18n.T("cassino.promptHelp") + "\n")
	})
}

// cassinoPlayerStr returns the display string for a single Cassino player.
func cassinoPlayerStr(player *domain.CassinoPlayer, i int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("cassino.playerLine",
		"name", cuiPlayerName(player, i),
		"hand", strconv.Itoa(player.GetCardsSize()),
		"captured", strconv.Itoa(player.CapturedCount()),
		"sweep", strconv.Itoa(player.GetSweepCount()),
		"total", strconv.Itoa(player.GetTotalScore())) + "\n")
	if player.GetIsHuman() {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// cassinoActionStr renders an action as a short readable line.
func cassinoActionStr(a *domain.CassinoAction) string {
	if a == nil {
		return ""
	}
	switch a.Type {
	case domain.CassinoActionTake:
		suffix := ""
		if a.IsSweep {
			suffix = i18n.T("cassino.actionTakeSweepSuffix")
		}
		return i18n.Tf("cassino.actionTake",
			"played", cassinoCardShort(a.PlayedCard),
			"count", strconv.Itoa(len(a.CapturedCards)),
			"suffix", suffix)
	case domain.CassinoActionBuild:
		return i18n.Tf("cassino.actionBuild",
			"value", strconv.Itoa(a.BuildValue),
			"played", cassinoCardShort(a.PlayedCard))
	case domain.CassinoActionTrail:
		return i18n.Tf("cassino.actionTrail",
			"played", cassinoCardShort(a.PlayedCard))
	default:
		return string(a.Type)
	}
}

// cassinoCardShort renders a single card as a short text representation.
func cassinoCardShort(c *domain.Card) string {
	if c == nil {
		return "-"
	}
	return cuiCardSliceStr([]*domain.Card{c})
}

// HintOutput recommends a take / build / trail for the human's turn, reusing
// the shared domain suggestion logic.
func (p *CassinoCuiPresenter) HintOutput(cg interfaces.CassinoGame) string {
	if cg.GetPhase() != domain.CassinoPhasePlayerTurn || !cg.IsHumanTurn() {
		return i18n.T("cassino.hintNotYourTurn") + "\n"
	}
	human := cg.GetPlayer(cg.GetCurrentTurn())
	if human == nil {
		return i18n.T("cassino.hintNotYourTurn") + "\n"
	}
	hand := make([]*domain.Card, human.GetCardsSize())
	for i := range hand {
		hand[i] = human.GetCard(i)
	}
	hint := domain.SuggestCassinoMove(hand, cg.GetTableCards(), cg.GetBuilds())
	if hint == nil || hint.HandIdx < 0 || hint.HandIdx >= len(hand) {
		return i18n.T("cassino.hintNoMove") + "\n"
	}
	card := cuiCardStr(hand[hint.HandIdx])
	switch hint.Action {
	case domain.CassinoHintTake:
		return i18n.Tf("cassino.hintTake",
			"card", card, "idxs", cassinoJoinIdxs(hint.TableIdxs)) + "\n"
	case domain.CassinoHintBuild:
		return i18n.Tf("cassino.hintBuild",
			"card", card, "value", strconv.Itoa(hint.Value)) + "\n"
	default:
		return i18n.Tf("cassino.hintTrail", "card", card) + "\n"
	}
}

// cassinoJoinIdxs renders 0-based table indices as a comma-separated string,
// or a dash when the capture is builds-only.
func cassinoJoinIdxs(idxs []int) string {
	if len(idxs) == 0 {
		return "-"
	}
	parts := make([]string, len(idxs))
	for i, v := range idxs {
		parts[i] = strconv.Itoa(v)
	}
	return strings.Join(parts, ", ")
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *CassinoCuiPresenter) ActionLogOutput(cg interfaces.CassinoGame) string {
	return actionLogOutputTextForSeats[*domain.CassinoPlayer](cg)
}
