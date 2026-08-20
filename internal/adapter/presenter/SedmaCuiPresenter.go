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

func sedmaPlayerStr(g interfaces.SedmaGame, idx int) string {
	player := g.GetPlayer(idx)
	if player == nil {
		return ""
	}
	teamScores := g.GetTeamScores()
	team := domain.SedmaTeamOf(idx)
	var b strings.Builder
	b.WriteString(i18n.Tf("sedma.playerLine",
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

// SedmaCuiPresenter renders the Sedma CUI view.
type SedmaCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *SedmaCuiPresenter) Output(g interfaces.SedmaGame, lastErr error) string {
	return buildCuiOutput(i18n.T("sedma.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("sedma.round",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber())) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(sedmaPlayerStr(g, i))
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
				winnerStr = domain.SedmaTeamName(winnerTeam)
			}
			banner := i18n.Tf("sedma.gameEnd", "team", winnerStr)
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.SedmaPhasePlay:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("sedma.promptPlay",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("sedma.promptPlayHelp") + "\n")
		case domain.SedmaPhaseTrickEnd:
			// The last effective card wins the whole trick, so name the winner
			// (the next lead) and the card points (each A or 10 is worth 10) they
			// just collected — the key information Sedma's CUI previously omitted.
			winnerIdx := g.GetLeadPlayerIdx()
			trickPts := 0
			for _, tc := range g.GetCurrentTrick() {
				if v := tc.Card.GetValue(); v == 1 || v == 10 {
					trickPts += 10
				}
			}
			b.WriteString(i18n.Tf("sedma.trickWinner",
				"name", cuiPlayerName(g.GetPlayer(winnerIdx), winnerIdx),
				"points", strconv.Itoa(trickPts)) + "\n")
			b.WriteString(i18n.T("sedma.promptTrickEnd") + "\n")
			b.WriteString(i18n.T("sedma.promptTrickEndHelp") + "\n")
		case domain.SedmaPhaseRoundEnd:
			pts := g.GetRoundCardPoints()
			b.WriteString(i18n.Tf("sedma.promptRoundEnd",
				"ptsA", strconv.Itoa(pts[0]),
				"ptsB", strconv.Itoa(pts[1])) + "\n")
			b.WriteString(i18n.T("sedma.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current Sedma hint.
func (p *SedmaCuiPresenter) HintOutput(g interfaces.SedmaGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("sedma.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, sedmaHintReasonKeys)
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
		return color.Yellow(i18n.Tf("sedma.hintCard",
			"cards", strings.Join(cards, ", "),
			"reason", reason)) + "\n"
	}
	return color.Yellow(i18n.Tf("sedma.hintCard",
		"cards", "-",
		"reason", reason)) + "\n"
}

// sedmaHintReasonKeys maps Sedma-specific hint-reason identifiers to i18n keys.
var sedmaHintReasonKeys = map[string]string{
	"lead_low":    "sedma.hintReasonLeadLow",
	"capture":     "sedma.hintReasonCapture",
	"discard_low": "sedma.hintReasonDiscardLow",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *SedmaCuiPresenter) ActionLogOutput(g interfaces.SedmaGame) string {
	return actionLogOutputTextWithNames(g, func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) })
}
