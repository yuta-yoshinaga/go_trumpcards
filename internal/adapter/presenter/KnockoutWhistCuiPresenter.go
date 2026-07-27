//go:build !js || !wasm || classic

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// knockoutWhistSuitSymbols maps a suit constant (1-4) to its glyph for CUI display.
var knockoutWhistSuitSymbols = [...]string{"?", "♠", "♣", "♥", "♦"}

// knockoutWhistSuitSymbol maps a suit constant (1-4) to its glyph for CUI display.
func knockoutWhistSuitSymbol(suit int) string {
	if suit < 1 || suit > 4 {
		return "?"
	}
	return knockoutWhistSuitSymbols[suit]
}

func knockoutWhistPlayerStr(g interfaces.KnockoutWhistGame, idx int) string {
	player := g.GetPlayer(idx)
	if player == nil {
		return ""
	}
	var b strings.Builder
	if player.GetEliminated() {
		b.WriteString(i18n.Tf("knockoutwhist.playerLineEliminated",
			"name", cuiPlayerName(player, idx)))
	} else {
		b.WriteString(i18n.Tf("knockoutwhist.playerLine",
			"name", cuiPlayerName(player, idx),
			"cards", strconv.Itoa(player.GetCardsSize()),
			"roundTricks", strconv.Itoa(player.GetRoundTricks()),
			"dogbones", strconv.Itoa(player.GetDogbones()),
		))
	}
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// KnockoutWhistCuiPresenter renders the Knockout Whist CUI view.
type KnockoutWhistCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *KnockoutWhistCuiPresenter) Output(g interfaces.KnockoutWhistGame, lastErr error) string {
	return buildCuiOutput(i18n.T("knockoutwhist.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("knockoutwhist.round",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"handSize", strconv.Itoa(g.GetHandSize()),
			"trick", strconv.Itoa(g.GetTrickNumber()),
			"trump", knockoutWhistSuitSymbol(g.GetTrumpSuit())) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(knockoutWhistPlayerStr(g, i))
		}

		b.WriteString("----------\n")

		cuiTrickBlock(b, g.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winner := g.GetWinnerPlayer()
			var winnerStr string
			if winner >= 0 {
				winnerStr = cuiPlayerName(g.GetPlayer(winner), winner)
			}
			banner := i18n.Tf("knockoutwhist.gameEnd", "name", winnerStr)
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.KnockoutWhistPhaseTrumpSelect:
			leadIdx := g.GetLeadPlayerIdx()
			b.WriteString(i18n.Tf("knockoutwhist.promptTrumpSelect",
				"name", cuiPlayerName(g.GetPlayer(leadIdx), leadIdx)) + "\n")
			b.WriteString(i18n.T("knockoutwhist.promptTrumpSelectHelp") + "\n")
		case domain.KnockoutWhistPhasePlay:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("knockoutwhist.promptPlay",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("knockoutwhist.promptPlayHelp") + "\n")
		case domain.KnockoutWhistPhaseTrickEnd:
			b.WriteString(i18n.T("knockoutwhist.promptTrickEnd") + "\n")
			b.WriteString(i18n.T("knockoutwhist.promptTrickEndHelp") + "\n")
		case domain.KnockoutWhistPhaseRoundEnd:
			b.WriteString(i18n.Tf("knockoutwhist.promptRoundEnd",
				"active", strconv.Itoa(g.GetActiveCount())) + "\n")
			b.WriteString(i18n.T("knockoutwhist.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current Knockout Whist hint.
func (p *KnockoutWhistCuiPresenter) HintOutput(g interfaces.KnockoutWhistGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("knockoutwhist.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, knockoutWhistHintReasonKeys)
	if len(hint.CardIndices) > 0 {
		playerIdx := g.GetCurrentPlayerIdx()
		player := g.GetPlayer(playerIdx)
		cards := make([]string, len(hint.CardIndices))
		for i, idx := range hint.CardIndices {
			if player != nil {
				cards[i] = "[" + strconv.Itoa(idx) + "]" + cuiCardStr(player.GetCard(idx))
			} else {
				cards[i] = strconv.Itoa(idx)
			}
		}
		return color.Yellow(i18n.Tf("knockoutwhist.hintCard",
			"cards", strings.Join(cards, ", "),
			"reason", reason)) + "\n"
	}
	return color.Yellow(i18n.Tf("knockoutwhist.hintCard",
		"cards", "-",
		"reason", reason)) + "\n"
}

// knockoutWhistHintReasonKeys maps Knockout Whist-specific hint-reason identifiers to i18n keys.
var knockoutWhistHintReasonKeys = map[string]string{
	"lead_high":   "knockoutwhist.hintReasonLeadHigh",
	"follow_win":  "knockoutwhist.hintReasonFollowWin",
	"follow_duck": "knockoutwhist.hintReasonFollowDuck",
	"discard_low": "knockoutwhist.hintReasonDiscardLow",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *KnockoutWhistCuiPresenter) ActionLogOutput(g interfaces.KnockoutWhistGame) string {
	return actionLogOutputText(g)
}
