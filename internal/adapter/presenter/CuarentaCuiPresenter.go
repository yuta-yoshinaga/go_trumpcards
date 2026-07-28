//go:build !js || !wasm || extra2

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// CuarentaCuiPresenter renders the Cuarenta CUI view.
type CuarentaCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *CuarentaCuiPresenter) Output(cg interfaces.CuarentaGame, lastErr error) string {
	return buildCuiOutput(i18n.T("cuarenta.helpTitle"), func(b *strings.Builder) {
		// チームスコア表示。
		for t := 0; t < domain.CuarentaTeamCnt; t++ {
			b.WriteString(i18n.Tf("cuarenta.teamScoreLine",
				"team", strconv.Itoa(t),
				"score", strconv.Itoa(cg.GetTeamScore(t))) + "\n")
		}
		b.WriteString("----------\n")

		for i := 0; i < cg.GetPlayerCnt(); i++ {
			b.WriteString(cuarentaPlayerStr(cg.GetPlayer(i), i))
		}
		b.WriteString("----------\n")

		if tableCards := cg.GetTableCards(); len(tableCards) > 0 {
			b.WriteString(i18n.Tf("cuarenta.tableLine", "cards", cuiCardSliceStr(tableCards)) + "\n")
		} else {
			b.WriteString(i18n.T("cuarenta.tableEmpty") + "\n")
		}

		if ha := cg.GetHumanAction(); ha != nil {
			b.WriteString(i18n.Tf("cuarenta.humanActionLine", "text", cuarentaActionStr(ha)) + "\n")
		}
		if cpu := cg.GetCpuActions(); len(cpu) > 0 {
			b.WriteString(color.Bold(i18n.T("cuarenta.cpuActionsHeader")) + "\n")
			for _, a := range cpu {
				b.WriteString(i18n.Tf("cuarenta.cpuActionLine",
					"name", cuiPlayerName(cg.GetPlayer(a.PlayerIdx), a.PlayerIdx),
					"text", cuarentaActionStr(a)) + "\n")
			}
		}

		cuiErrorBlock(b, lastErr)

		if cg.GetGameEndFlag() {
			b.WriteString(i18n.T("cuarenta.gameEnd") + "\n")
			for t := 0; t < domain.CuarentaTeamCnt; t++ {
				b.WriteString(i18n.Tf("cuarenta.scoreEntry",
					"team", strconv.Itoa(t),
					"score", strconv.Itoa(cg.GetTeamScore(t))) + "\n")
			}
			return
		}
		currentTurn := cg.GetCurrentTurn()
		b.WriteString(i18n.Tf("cuarenta.promptCurrentTurn",
			"name", cuiPlayerName(cg.GetPlayer(currentTurn), currentTurn)) + "\n")
		b.WriteString(i18n.T("cuarenta.promptHelp") + "\n")
	})
}

// cuarentaPlayerStr returns the display string for a single Cuarenta player.
func cuarentaPlayerStr(player *domain.CuarentaPlayer, i int) string {
	if player == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString(i18n.Tf("cuarenta.playerLine",
		"name", cuiPlayerName(player, i),
		"team", strconv.Itoa(domain.CuarentaTeamOf(i)),
		"hand", strconv.Itoa(player.GetCardsSize()),
		"captured", strconv.Itoa(player.CapturedCount())) + "\n")
	if player.GetIsHuman() {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// cuarentaActionStr renders an action as a short readable line.
func cuarentaActionStr(a *domain.CuarentaAction) string {
	if a == nil {
		return ""
	}
	if len(a.CapturedCards) > 0 {
		suffix := ""
		if a.IsCaida {
			suffix += i18n.T("cuarenta.actionCaidaSuffix")
		}
		if a.RondaBonus > 0 {
			suffix += i18n.Tf("cuarenta.actionRondaSuffix", "bonus", strconv.Itoa(a.RondaBonus))
		}
		if a.IsLimpia {
			suffix += i18n.T("cuarenta.actionLimpiaSuffix")
		}
		return i18n.Tf("cuarenta.actionCapture",
			"played", cuarentaCardShort(a.PlayedCard),
			"count", strconv.Itoa(len(a.CapturedCards)),
			"suffix", suffix)
	}
	return i18n.Tf("cuarenta.actionLay", "played", cuarentaCardShort(a.PlayedCard))
}

// cuarentaCardShort renders a single card as a short text representation.
func cuarentaCardShort(c *domain.Card) string {
	if c == nil {
		return "-"
	}
	return cuiCardSliceStr([]*domain.Card{c})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *CuarentaCuiPresenter) ActionLogOutput(cg interfaces.CuarentaGame) string {
	return actionLogOutputText(cg)
}
