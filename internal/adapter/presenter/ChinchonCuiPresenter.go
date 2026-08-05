//go:build !js || !wasm || extra

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// chinchonPlayerStr returns the display string for a single Chinchón player.
func chinchonPlayerStr(player *domain.ChinchonPlayer, i int) string {
	var b strings.Builder
	line := i18n.Tf("chinchon.playerLine",
		"name", cuiPlayerName(player, i),
		"cum", strconv.Itoa(player.GetCumulativeScore()),
		"round", strconv.Itoa(player.GetRoundScore()),
		"cards", strconv.Itoa(player.GetCardsSize()))
	if player.GetEliminated() {
		line += " " + i18n.T("chinchon.eliminatedTag")
	}
	b.WriteString(line + "\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// chinchonMeldSplitLines renders which of the human's cards are already melded
// and which are deadwood, with the point breakdown. The Web GUI colours the two
// groups and prints "5 + 3 + 2 = 10"; the CUI showed only the total (#4838).
func chinchonMeldSplitLines(g interfaces.ChinchonGame, idx int) string {
	melds, deadwood := g.GetPlayerMeldSplit(idx)
	var b strings.Builder
	if len(melds) > 0 {
		groups := make([]string, 0, len(melds))
		for _, meld := range melds {
			groups = append(groups, cuiCardSliceStr(meld))
		}
		b.WriteString(i18n.Tf("chinchon.meldedLine", "melds", strings.Join(groups, " / ")) + "\n")
	}
	if len(deadwood) > 0 {
		values := make([]string, 0, len(deadwood))
		for _, c := range deadwood {
			values = append(values, strconv.Itoa(domain.CalcDeadwoodValue([]*domain.Card{c})))
		}
		b.WriteString(i18n.Tf("chinchon.deadwoodBreakdown",
			"cards", cuiCardSliceStr(deadwood),
			"sum", strings.Join(values, " + "),
			"total", strconv.Itoa(domain.CalcDeadwoodValue(deadwood))) + "\n")
	}
	return b.String()
}

// ChinchonCuiPresenter renders the Chinchón CUI view.
type ChinchonCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *ChinchonCuiPresenter) Output(g interfaces.ChinchonGame, lastErr error) string {
	return buildCuiOutput(i18n.T("chinchon.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("chinchon.header",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"stock", strconv.Itoa(g.GetDrawPileCount())) + "\n")

		if top := g.GetDiscardTop(); top != nil {
			b.WriteString(i18n.Tf("chinchon.discardLine", "card", cuiCardStr(top)) + "\n")
		}

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(chinchonPlayerStr(g.GetPlayer(i), i))
			// Show the human's current deadwood (the core knock-timing signal),
			// green with a "knock ready" note once at or below the threshold. CPU
			// deadwood stays hidden so hands are not exposed mid-round.
			if player := g.GetPlayer(i); player != nil && player.GetIsHuman() && player.GetCardsSize() > 0 {
				deadwood := g.GetPlayerDeadwoodValue(i)
				threshold := g.GetKnockThreshold()
				line := i18n.Tf("chinchon.deadwoodLine",
					"value", strconv.Itoa(deadwood),
					"threshold", strconv.Itoa(threshold))
				if deadwood <= threshold {
					line = color.Green(line + " " + i18n.T("chinchon.knockReady"))
				}
				b.WriteString(line + "\n")
				b.WriteString(chinchonMeldSplitLines(g, i))
			}
		}

		// レイオフフェーズではノッカーのメルドを表示する。
		if g.GetPhase() == domain.ChinchonPhaseLayoff {
			for _, meld := range g.GetKnockerMelds() {
				b.WriteString(i18n.Tf("chinchon.knockerMeldLine", "cards", cuiCardSliceStr(meld)) + "\n")
			}
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winnerIdx := g.GetWinnerIdx()
			if winnerIdx < 0 {
				b.WriteString(color.Green(i18n.T("chinchon.drawBanner")) + "\n")
				return
			}
			banner := i18n.Tf("chinchon.gameEnd",
				"name", cuiPlayerName(g.GetPlayer(winnerIdx), winnerIdx))
			b.WriteString(color.Green(banner) + "\n")
			return
		}

		switch g.GetPhase() {
		case domain.ChinchonPhaseDraw:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("chinchon.promptDraw",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("chinchon.promptDrawHelpStock") + "\n")
			b.WriteString(i18n.T("chinchon.promptDrawHelpDiscard") + "\n")
		case domain.ChinchonPhaseDiscard:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("chinchon.promptDiscard",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("chinchon.promptDiscardHelp") + "\n")
			b.WriteString(i18n.T("chinchon.promptKnockHelp") + "\n")
		case domain.ChinchonPhaseLayoff:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("chinchon.promptLayoff",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("chinchon.promptLayoffHelp") + "\n")
			b.WriteString(i18n.T("chinchon.promptLayoffSkip") + "\n")
		case domain.ChinchonPhaseRoundEnd:
			b.WriteString(i18n.T("chinchon.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("chinchon.promptRoundEndHelp") + "\n")
		}
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *ChinchonCuiPresenter) ActionLogOutput(g interfaces.ChinchonGame) string {
	return actionLogOutputText(g)
}
