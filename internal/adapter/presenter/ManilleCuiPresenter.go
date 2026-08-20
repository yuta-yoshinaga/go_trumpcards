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

// manilleSuitSymbols maps a suit constant (1-4) to its glyph for CUI display.
var manilleSuitSymbols = [...]string{"?", "♠", "♣", "♥", "♦"}

// manilleSuitSymbol maps a suit constant (1-4) to its glyph for CUI display.
func manilleSuitSymbol(suit int) string {
	if suit < 1 || suit > 4 {
		return "?"
	}
	return manilleSuitSymbols[suit]
}

func manillePlayerStr(g interfaces.ManilleGame, idx int) string {
	player := g.GetPlayer(idx)
	if player == nil {
		return ""
	}
	teamScores := g.GetTeamScores()
	team := domain.ManilleTeamOf(idx)
	var b strings.Builder
	b.WriteString(i18n.Tf("manille.playerLine",
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

// ManilleCuiPresenter renders the Manille CUI view.
type ManilleCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *ManilleCuiPresenter) Output(g interfaces.ManilleGame, lastErr error) string {
	return buildCuiOutput(i18n.T("manille.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("manille.round",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber()),
			"trump", manilleSuitSymbol(g.GetTrumpSuit())) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(manillePlayerStr(g, i))
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
			banner := i18n.Tf("manille.gameEnd", "team", winnerStr)
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.ManillePhasePlay:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("manille.promptPlay",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("manille.promptPlayHelp") + "\n")
			// Manille's rank order is inverted from a normal deck (10 is highest),
			// which is easy to misread since the hand lists cards by strength, not
			// face value — spell out the order and each card's point worth.
			b.WriteString(i18n.T("manille.rankHelp") + "\n")
		case domain.ManillePhaseTrickEnd:
			// **トリックを取ったのが誰かは、次にリードするのが誰かでもある** (#5646)。
			// Web は manille-trick-winner で名前とチームを出しているのに、CUI は
			// 「次のトリックへ」としか言わず、手番を見失いやすかった。姉妹の Sueca
			// は同じ場面で GetLeadPlayerIdx から組み立てている。
			if winnerIdx := g.GetLeadPlayerIdx(); winnerIdx >= 0 {
				teamKey := "manille.teamA"
				if domain.ManilleTeamOf(winnerIdx) == 1 {
					teamKey = "manille.teamB"
				}
				b.WriteString(i18n.Tf("manille.trickWinner",
					"name", cuiPlayerName(g.GetPlayer(winnerIdx), winnerIdx),
					"team", i18n.T(teamKey)) + "\n")
			}
			b.WriteString(i18n.T("manille.promptTrickEnd") + "\n")
			b.WriteString(i18n.T("manille.promptTrickEndHelp") + "\n")
		case domain.ManillePhaseRoundEnd:
			pts := g.GetRoundCardPoints()
			b.WriteString(i18n.Tf("manille.promptRoundEnd",
				"ptsA", strconv.Itoa(pts[0]),
				"ptsB", strconv.Itoa(pts[1])) + "\n")
			b.WriteString(i18n.T("manille.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current Manille hint.
func (p *ManilleCuiPresenter) HintOutput(g interfaces.ManilleGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("manille.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, manilleHintReasonKeys)
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
		return color.Yellow(i18n.Tf("manille.hintCard",
			"cards", strings.Join(cards, ", "),
			"reason", reason)) + "\n"
	}
	return color.Yellow(i18n.Tf("manille.hintCard",
		"cards", "-",
		"reason", reason)) + "\n"
}

// manilleHintReasonKeys maps Manille-specific hint-reason identifiers to i18n keys.
var manilleHintReasonKeys = map[string]string{
	"lead_low":    "manille.hintReasonLeadLow",
	"follow_win":  "manille.hintReasonFollowWin",
	"follow_duck": "manille.hintReasonFollowDuck",
	"discard_low": "manille.hintReasonDiscardLow",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *ManilleCuiPresenter) ActionLogOutput(g interfaces.ManilleGame) string {
	return actionLogOutputTextWithNames(g, func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) })
}
