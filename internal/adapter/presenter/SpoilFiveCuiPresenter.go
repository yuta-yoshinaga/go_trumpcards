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

// spoilFiveSuitSymbols maps a suit constant (1-4) to its glyph for CUI display.
var spoilFiveSuitSymbols = [...]string{"?", "♠", "♣", "♥", "♦"}

// spoilFiveSuitSymbol maps a suit constant (1-4) to its glyph for CUI display.
func spoilFiveSuitSymbol(suit int) string {
	if suit < 1 || suit > 4 {
		return "?"
	}
	return spoilFiveSuitSymbols[suit]
}

func spoilFivePlayerStr(g interfaces.SpoilFiveGame, idx int) string {
	player := g.GetPlayer(idx)
	if player == nil {
		return ""
	}
	var b strings.Builder
	// Mark the current lead and (at round end) the round winner so seat status is
	// visible at a glance in the 5-player game, matching the web player panels.
	if idx == g.GetLeadPlayerIdx() {
		b.WriteString(i18n.T("spoilfive.leaderMark"))
	}
	if idx == g.GetRoundWinnerIdx() {
		b.WriteString(i18n.T("spoilfive.winnerMark"))
	}
	b.WriteString(i18n.Tf("spoilfive.playerLine",
		"name", cuiPlayerName(player, idx),
		"cards", strconv.Itoa(player.GetCardsSize()),
		"roundTricks", strconv.Itoa(player.GetRoundTricks()),
		"score", strconv.Itoa(player.GetScore()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// SpoilFiveCuiPresenter renders the Spoil Five CUI view.
type SpoilFiveCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *SpoilFiveCuiPresenter) Output(g interfaces.SpoilFiveGame, lastErr error) string {
	return buildCuiOutput(i18n.T("spoilfive.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("spoilfive.round",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber()),
			"trump", spoilFiveSuitSymbol(g.GetTrumpSuit()),
			"pot", strconv.Itoa(g.GetPot())) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(spoilFivePlayerStr(g, i))
		}

		b.WriteString("----------\n")

		cuiTrickBlock(b, g.GetCurrentTrick(),
			func(tc *domain.SpoilFiveTrickCard) int { return tc.PlayerIdx },
			func(tc *domain.SpoilFiveTrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winner := g.GetWinnerPlayer()
			var winnerStr string
			if winner >= 0 {
				winnerStr = cuiPlayerName(g.GetPlayer(winner), winner)
			}
			banner := i18n.Tf("spoilfive.gameEnd", "name", winnerStr)
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.SpoilFivePhasePlay:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("spoilfive.promptPlay",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("spoilfive.promptPlayHelp") + "\n")
		case domain.SpoilFivePhaseTrickEnd:
			b.WriteString(i18n.T("spoilfive.promptTrickEnd") + "\n")
			b.WriteString(i18n.T("spoilfive.promptTrickEndHelp") + "\n")
		case domain.SpoilFivePhaseRoundEnd:
			if g.GetRoundWinnerIdx() >= 0 {
				winnerIdx := g.GetRoundWinnerIdx()
				b.WriteString(i18n.Tf("spoilfive.promptRoundEnd",
					"name", cuiPlayerName(g.GetPlayer(winnerIdx), winnerIdx),
					"pot", strconv.Itoa(g.GetPot())) + "\n")
			} else {
				b.WriteString(i18n.Tf("spoilfive.promptSpoil",
					"pot", strconv.Itoa(g.GetPot())) + "\n")
			}
			b.WriteString(i18n.T("spoilfive.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current Spoil Five hint.
func (p *SpoilFiveCuiPresenter) HintOutput(g interfaces.SpoilFiveGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("spoilfive.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, spoilFiveHintReasonKeys)
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
		return color.Yellow(i18n.Tf("spoilfive.hintCard",
			"cards", strings.Join(cards, ", "),
			"reason", reason)) + "\n"
	}
	return color.Yellow(i18n.Tf("spoilfive.hintCard",
		"cards", "-",
		"reason", reason)) + "\n"
}

// spoilFiveHintReasonKeys maps Spoil Five-specific hint-reason identifiers to i18n keys.
var spoilFiveHintReasonKeys = map[string]string{
	"lead_high":   "spoilfive.hintReasonLeadHigh",
	"take_trick":  "spoilfive.hintReasonTakeTrick",
	"discard_low": "spoilfive.hintReasonDiscardLow",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *SpoilFiveCuiPresenter) ActionLogOutput(g interfaces.SpoilFiveGame) string {
	return actionLogOutputText(g)
}
