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

// klaverjasSuitSymbols maps a suit constant (1-4) to its glyph for CUI display.
var klaverjasSuitSymbols = [...]string{"?", "♠", "♣", "♥", "♦"}

// klaverjasSuitSymbol maps a suit constant (1-4) to its glyph for CUI display.
func klaverjasSuitSymbol(suit int) string {
	if suit < 1 || suit > 4 {
		return "?"
	}
	return klaverjasSuitSymbols[suit]
}

func klaverjasPlayerStr(g interfaces.KlaverjasGame, idx int) string {
	player := g.GetPlayer(idx)
	if player == nil {
		return ""
	}
	teamScores := g.GetTeamScores()
	team := domain.KlaverjasTeamOf(idx)
	var b strings.Builder
	b.WriteString(i18n.Tf("klaverjas.playerLine",
		"name", cuiPlayerName(player, idx),
		"cards", strconv.Itoa(player.GetCardsSize()),
		"teamScore", strconv.Itoa(teamScores[team]),
		"tricks", strconv.Itoa(player.GetTrickCount()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// KlaverjasCuiPresenter renders the Klaverjas CUI view.
type KlaverjasCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *KlaverjasCuiPresenter) Output(g interfaces.KlaverjasGame, lastErr error) string {
	return buildCuiOutput(i18n.T("klaverjas.helpTitle"), func(b *strings.Builder) {
		roem := g.GetRoundRoem()
		b.WriteString(i18n.Tf("klaverjas.round",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber()),
			"trump", klaverjasSuitSymbol(g.GetTrumpSuit()),
			"roemA", strconv.Itoa(roem[0]),
			"roemB", strconv.Itoa(roem[1])) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(klaverjasPlayerStr(g, i))
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
			banner := i18n.Tf("klaverjas.gameEnd", "team", winnerStr)
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.KlaverjasPhasePlay:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("klaverjas.promptPlay",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("klaverjas.promptPlayHelp") + "\n")
			// **切り札と非切り札で順序が丸ごと違う。** Web は
			// klaverjas-strength-legend の 2 表を常時出しているのに、CUI には
			// 説明が一切なかった (#5645)。姉妹の Manille は rankHelp で解決済み。
			b.WriteString(i18n.T("klaverjas.rankHelpTrump") + "\n")
			b.WriteString(i18n.T("klaverjas.rankHelpPlain") + "\n")
		case domain.KlaverjasPhaseTrickEnd:
			b.WriteString(i18n.T("klaverjas.promptTrickEnd") + "\n")
			b.WriteString(i18n.T("klaverjas.promptTrickEndHelp") + "\n")
		case domain.KlaverjasPhaseRoundEnd:
			pts := g.GetRoundCardPoints()
			roem := g.GetRoundRoem()
			b.WriteString(i18n.Tf("klaverjas.promptRoundEnd",
				"ptsA", strconv.Itoa(pts[0]),
				"ptsB", strconv.Itoa(pts[1])) + "\n")
			// Card points alone omit the Roem bonuses; show each team's Roem and the
			// card-points+Roem total so the final round score is complete.
			b.WriteString(i18n.Tf("klaverjas.promptRoundEndRoem",
				"roemA", strconv.Itoa(roem[0]),
				"roemB", strconv.Itoa(roem[1]),
				"totalA", strconv.Itoa(pts[0]+roem[0]),
				"totalB", strconv.Itoa(pts[1]+roem[1])) + "\n")
			b.WriteString(i18n.T("klaverjas.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current Klaverjas hint.
func (p *KlaverjasCuiPresenter) HintOutput(g interfaces.KlaverjasGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("klaverjas.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, klaverjasHintReasonKeys)
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
		return color.Yellow(i18n.Tf("klaverjas.hintCard",
			"cards", strings.Join(cards, ", "),
			"reason", reason)) + "\n"
	}
	return color.Yellow(i18n.Tf("klaverjas.hintCard",
		"cards", "-",
		"reason", reason)) + "\n"
}

// klaverjasHintReasonKeys maps Klaverjas-specific hint-reason identifiers to i18n keys.
var klaverjasHintReasonKeys = map[string]string{
	"lead_low":    "klaverjas.hintReasonLeadLow",
	"follow_win":  "klaverjas.hintReasonFollowWin",
	"follow_duck": "klaverjas.hintReasonFollowDuck",
	"discard_low": "klaverjas.hintReasonDiscardLow",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *KlaverjasCuiPresenter) ActionLogOutput(g interfaces.KlaverjasGame) string {
	return actionLogOutputText(g)
}
