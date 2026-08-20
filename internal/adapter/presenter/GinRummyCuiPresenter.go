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

// ginRummyBestDeadwood returns the lowest deadwood value the player can reach by
// discarding one card — the value that gates knocking (<= GinRummyKnockThreshold).
func ginRummyBestDeadwood(player *domain.GinRummyPlayer) int {
	n := player.GetCardsSize()
	cards := make([]*domain.Card, n)
	for i := 0; i < n; i++ {
		cards[i] = player.GetCard(i)
	}
	best := -1
	for skip := 0; skip < n; skip++ {
		sub := make([]*domain.Card, 0, n-1)
		for i, c := range cards {
			if i != skip {
				sub = append(sub, c)
			}
		}
		_, deadwood := domain.FindBestMelds(sub)
		if v := domain.CalcDeadwoodValue(deadwood); best < 0 || v < best {
			best = v
		}
	}
	if best < 0 {
		best = 0
	}
	return best
}

// ginRummyPlayerStr returns the display string for a single GinRummy player.
func ginRummyPlayerStr(player *domain.GinRummyPlayer, i int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("ginrummy.playerLine",
		"name", cuiPlayerName(player, i),
		"cum", strconv.Itoa(player.GetCumulativeScore()),
		"round", strconv.Itoa(player.GetRoundScore()),
		"cards", strconv.Itoa(player.GetCardsSize())) + "\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// GinRummyCuiPresenter renders the Gin Rummy CUI view.
type GinRummyCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *GinRummyCuiPresenter) Output(g interfaces.GinRummyGame, lastErr error) string {
	return buildCuiOutput(i18n.T("ginrummy.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("ginrummy.header",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"stock", strconv.Itoa(g.GetDrawPileCount())) + "\n")

		if top := g.GetDiscardTop(); top != nil {
			b.WriteString(i18n.Tf("ginrummy.discardLine", "card", cuiCardStr(top)) + "\n")
		}

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(ginRummyPlayerStr(g.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winnerIdx := g.GetWinnerIdx()
			banner := i18n.Tf("ginrummy.gameEnd",
				"name", cuiPlayerName(g.GetPlayer(winnerIdx), winnerIdx))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.GinRummyPhaseDraw:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("ginrummy.promptDraw",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("ginrummy.promptDrawHelpStock") + "\n")
			b.WriteString(i18n.T("ginrummy.promptDrawHelpDiscard") + "\n")
		case domain.GinRummyPhaseDiscard:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("ginrummy.promptDiscard",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			if cur := g.GetPlayer(currentIdx); cur.GetIsHuman() {
				best := ginRummyBestDeadwood(cur)
				b.WriteString(i18n.Tf("ginrummy.deadwoodLine", "value", strconv.Itoa(best)) + "\n")
				if best <= domain.GinRummyKnockThreshold {
					b.WriteString(color.Yellow(i18n.T("ginrummy.canKnockNow")) + "\n")
				}
			}
			b.WriteString(i18n.T("ginrummy.promptDiscardHelp") + "\n")
			b.WriteString(i18n.T("ginrummy.promptKnockHelp") + "\n")
		case domain.GinRummyPhaseLayoff:
			// The knocker's melds are what the player lays off onto, so they must
			// be visible to make a layoff decision (parity with the web view).
			writeGinRummyKnockerMelds(b, g.GetKnockerMelds())
			b.WriteString(i18n.T("ginrummy.promptLayoff") + "\n")
			b.WriteString(i18n.T("ginrummy.promptLayoffHelp") + "\n")
			b.WriteString(i18n.T("ginrummy.promptLayoffSkip") + "\n")
		case domain.GinRummyPhaseRoundEnd:
			writeGinRummyKnockerMelds(b, g.GetKnockerMelds())
			b.WriteString(i18n.T("ginrummy.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("ginrummy.promptRoundEndHelp") + "\n")
		}
	})
}

// ginRummyMeldIsSet reports whether a meld is a set (same rank) rather than a
// run (a same-suit sequence).
func ginRummyMeldIsSet(meld []*domain.Card) bool {
	if len(meld) == 0 {
		return false
	}
	rank := meld[0].GetValue()
	for _, c := range meld {
		if c.GetValue() != rank {
			return false
		}
	}
	return true
}

// writeGinRummyKnockerMelds lists the knocker's melds with a set/run label so
// the player can see what to lay off onto.
func writeGinRummyKnockerMelds(b *strings.Builder, melds [][]*domain.Card) {
	if len(melds) == 0 {
		return
	}
	b.WriteString(color.Bold(i18n.T("ginrummy.knockerMeldsHeader")) + "\n")
	for i, meld := range melds {
		typeLabel := i18n.T("ginrummy.meldRun")
		if ginRummyMeldIsSet(meld) {
			typeLabel = i18n.T("ginrummy.meldSet")
		}
		b.WriteString(i18n.Tf("ginrummy.knockerMeldLine",
			"idx", strconv.Itoa(i+1),
			"type", typeLabel,
			"cards", formatCardSlice(meld, cuiCardStr, ", ")) + "\n")
	}
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *GinRummyCuiPresenter) ActionLogOutput(g interfaces.GinRummyGame) string {
	return actionLogOutputTextForSeats[*domain.GinRummyPlayer](g)
}
