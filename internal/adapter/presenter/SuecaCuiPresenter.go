//go:build !js || !wasm || casino

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// suecaSuitSymbol maps a suit constant (1-4) to its glyph for CUI display.
func suecaSuitSymbol(suit int) string {
	if suit < 1 || suit > 4 {
		return "?"
	}
	return []string{"", "♠", "♣", "♥", "♦"}[suit]
}

func suecaPlayerStr(g interfaces.SuecaGame, idx int) string {
	player := g.GetPlayer(idx)
	if player == nil {
		return ""
	}
	teamPts := g.GetTeamGamePoints()
	team := domain.SuecaTeamOf(idx)
	var b strings.Builder
	b.WriteString(i18n.Tf("sueca.playerLine",
		"name", cuiPlayerName(player, idx),
		"cards", strconv.Itoa(player.GetCardsSize()),
		"teamScore", strconv.Itoa(teamPts[team]),
		"tricks", strconv.Itoa(player.GetTrickCount()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// SuecaCuiPresenter renders the Sueca CUI view.
type SuecaCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *SuecaCuiPresenter) Output(g interfaces.SuecaGame, lastErr error) string {
	return buildCuiOutput(i18n.T("sueca.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("sueca.round",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber()),
			"trump", suecaSuitSymbol(g.GetTrumpSuit())) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(suecaPlayerStr(g, i))
		}

		b.WriteString("----------\n")

		cuiTrickBlock(b, g.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winnerTeam := g.GetWinnerTeam()
			var winnerStr string
			if winnerTeam >= 0 {
				winnerStr = strconv.Itoa(winnerTeam)
			}
			banner := i18n.Tf("sueca.gameEnd", "team", winnerStr)
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.SuecaPhasePlay:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("sueca.promptPlay",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("sueca.promptPlayHelp") + "\n")
		case domain.SuecaPhaseTrickEnd:
			// The trick winner leads the next trick, so name them and their team
			// (parity with the web trick-winner banner).
			winnerIdx := g.GetLeadPlayerIdx()
			teamLabel := "A"
			if domain.SuecaTeamOf(winnerIdx) == 1 {
				teamLabel = "B"
			}
			b.WriteString(i18n.Tf("sueca.trickWinner",
				"name", cuiPlayerName(g.GetPlayer(winnerIdx), winnerIdx),
				"team", teamLabel) + "\n")
			b.WriteString(i18n.T("sueca.promptTrickEnd") + "\n")
			b.WriteString(i18n.T("sueca.promptTrickEndHelp") + "\n")
		case domain.SuecaPhaseRoundEnd:
			pts := g.GetRoundCardPoints()
			b.WriteString(i18n.Tf("sueca.promptRoundEnd",
				"ptsA", strconv.Itoa(pts[0]),
				"ptsB", strconv.Itoa(pts[1])) + "\n")
			b.WriteString(i18n.T("sueca.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current Sueca hint.
func (p *SuecaCuiPresenter) HintOutput(g interfaces.SuecaGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("sueca.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, suecaHintReasonKeys)
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
		return color.Yellow(i18n.Tf("sueca.hintCard",
			"cards", strings.Join(cards, ", "),
			"reason", reason)) + "\n"
	}
	return color.Yellow(i18n.Tf("sueca.hintCard",
		"cards", "-",
		"reason", reason)) + "\n"
}

// suecaHintReasonKeys maps Sueca-specific hint-reason identifiers to i18n keys.
var suecaHintReasonKeys = map[string]string{
	"lead_low":    "sueca.hintReasonLeadLow",
	"follow_win":  "sueca.hintReasonFollowWin",
	"follow_duck": "sueca.hintReasonFollowDuck",
	"discard_low": "sueca.hintReasonDiscardLow",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *SuecaCuiPresenter) ActionLogOutput(g interfaces.SuecaGame) string {
	return actionLogOutputText(g)
}
