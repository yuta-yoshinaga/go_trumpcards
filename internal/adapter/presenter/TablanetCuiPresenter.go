//go:build !js || !wasm || extra3

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// TablanetCuiPresenter renders the Tablanet CUI view.
type TablanetCuiPresenter struct{}

// tablanetTableStr は場札を "[0]♠5 [1]♥J" 形式で返す。
func tablanetTableStr(cards []*domain.Card) string {
	if len(cards) == 0 {
		return "-"
	}
	parts := make([]string, len(cards))
	for i, c := range cards {
		parts[i] = "[" + strconv.Itoa(i) + "]" + cuiCardStr(c)
	}
	return strings.Join(parts, " ")
}

func tablanetPlayerStr(g interfaces.TablanetGame, idx int) string {
	player := g.GetPlayer(idx)
	if player == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString(i18n.Tf("tablanet.playerLine",
		"name", cuiPlayerName(player, idx),
		"hand", strconv.Itoa(player.GetCardsSize()),
		"captured", strconv.Itoa(player.CapturedCount()),
		"tabla", strconv.Itoa(player.GetTablaCount())) + "\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
		// Annotate which table cards each hand card can capture, reusing the
		// domain's capture computation (GetCaptureOptions).
		if line := cuiCaptureHintLine(player, g.GetCaptureOptions(idx), "tablanet.captureHint"); line != "" {
			b.WriteString(line + "\n")
		}
	}
	return b.String()
}

// Output renders the current game state for the active locale.
func (p *TablanetCuiPresenter) Output(g interfaces.TablanetGame, lastErr error) string {
	return buildCuiOutput(i18n.T("tablanet.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("tablanet.deckLine",
			"deal", strconv.Itoa(g.GetRoundNumber()),
			"deck", strconv.Itoa(g.GetRemainingDeck())) + "\n")
		b.WriteString(i18n.Tf("tablanet.tableLine", "table", tablanetTableStr(g.GetTableCards())) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(tablanetPlayerStr(g, i))
		}
		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch g.GetPhase() {
		case domain.TablanetPhasePlay:
			currentIdx := g.GetCurrentTurn()
			b.WriteString(i18n.Tf("tablanet.promptPlay",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
		case domain.TablanetPhaseGameEnd:
			b.WriteString(i18n.T("tablanet.promptGameEnd") + "\n")
		}
		b.WriteString(i18n.T("tablanet.promptHelp") + "\n")
	})
}

// HintOutput emits the current Tablanet hint.
func (p *TablanetCuiPresenter) HintOutput(g interfaces.TablanetGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("tablanet.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, tablanetHintReasonKeys)
	if len(hint.CardIndices) > 0 {
		playerIdx := g.GetCurrentTurn()
		player := g.GetPlayer(playerIdx)
		cards := make([]string, len(hint.CardIndices))
		for i, idx := range hint.CardIndices {
			if player != nil && idx >= 0 && idx < player.GetCardsSize() {
				cards[i] = "[" + strconv.Itoa(idx) + "]" + cuiCardStr(player.GetCard(idx))
			} else {
				cards[i] = strconv.Itoa(idx)
			}
		}
		return color.Yellow(i18n.Tf("tablanet.hintCard",
			"cards", strings.Join(cards, ", "),
			"reason", reason)) + "\n"
	}
	return color.Yellow(i18n.Tf("tablanet.hintCard", "cards", "-", "reason", reason)) + "\n"
}

// tablanetHintReasonKeys maps Tablanet-specific hint-reason identifiers to i18n keys.
var tablanetHintReasonKeys = map[string]string{
	"tabla_sweep": "tablanet.hintReasonTabla",
	"jack_sweep":  "tablanet.hintReasonJack",
	"capture":     "tablanet.hintReasonCapture",
	"trail_low":   "tablanet.hintReasonTrail",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *TablanetCuiPresenter) ActionLogOutput(g interfaces.TablanetGame) string {
	return actionLogOutputText(g)
}
