//go:build !js || !wasm || solo

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// canastaPlayerStr returns the display string for a single Canasta player.
func canastaPlayerStr(player *domain.CanastaPlayer, i int, showCards bool) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("canasta.playerLine",
		"name", cuiPlayerName(player, i),
		"cum", strconv.Itoa(player.GetCumulativeScore()),
		"round", strconv.Itoa(player.GetRoundScore()),
		"cards", strconv.Itoa(player.GetCardsSize())))
	if len(player.GetRed3s()) > 0 {
		b.WriteString(i18n.Tf("canasta.playerRed3s",
			"count", strconv.Itoa(len(player.GetRed3s()))))
	}
	if player.HasCanasta() {
		b.WriteString(i18n.T("canasta.playerCanastaTag"))
	}
	b.WriteString("\n")

	// Melds
	for _, m := range player.GetMelds() {
		meldType := i18n.T("canasta.meldTypeMixed")
		if m.IsNatural {
			meldType = i18n.T("canasta.meldTypeNatural")
		}
		if m.IsCanasta() {
			meldType += i18n.T("canasta.meldTypeCanastaSuffix")
		}
		cardStrs := make([]string, len(m.Cards))
		for j, c := range m.Cards {
			cardStrs[j] = cuiCardStr(c)
		}
		b.WriteString(i18n.Tf("canasta.meldLine",
			"type", meldType,
			"cards", strings.Join(cardStrs, ", ")) + "\n")
	}

	if showCards && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// CanastaCuiPresenter renders the Canasta CUI view.
type CanastaCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *CanastaCuiPresenter) Output(g interfaces.CanastaGame, lastErr error) string {
	return buildCuiOutput(i18n.T("canasta.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("canasta.header",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"stock", strconv.Itoa(g.GetDrawPileCount()),
			"discard", strconv.Itoa(g.GetDiscardPileCount())))
		if g.GetIsFrozen() {
			b.WriteString(i18n.T("canasta.frozenTag"))
		}
		b.WriteString("\n")

		// Top of discard
		if top := g.GetDiscardTop(); top != nil {
			b.WriteString(i18n.Tf("canasta.discardLine", "card", cuiCardStr(top)) + "\n")
		}

		// Players
		phase := g.GetPhase()
		showAllCards := phase == domain.CanastaPhaseRoundEnd || phase == domain.CanastaPhaseGameEnd
		for i := 0; i < g.GetPlayerCnt(); i++ {
			player := g.GetPlayer(i)
			showCards := player.GetIsHuman() || showAllCards
			b.WriteString(canastaPlayerStr(player, i, showCards))
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winnerIdx := g.GetWinnerIdx()
			banner := i18n.Tf("canasta.gameEnd",
				"name", cuiPlayerName(g.GetPlayer(winnerIdx), winnerIdx))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch phase {
		case domain.CanastaPhaseDraw:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("canasta.promptDraw",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("canasta.promptDrawHelpStock") + "\n")
			b.WriteString(i18n.T("canasta.promptDrawHelpDiscard") + "\n")
		case domain.CanastaPhaseMeld:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("canasta.promptMeld",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("canasta.promptMeldHelp") + "\n")
			b.WriteString(i18n.T("canasta.promptSkipMeld") + "\n")
		case domain.CanastaPhaseDiscard:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("canasta.promptDiscard",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("canasta.promptDiscardHelp") + "\n")
			b.WriteString(i18n.T("canasta.promptGoOutHelp") + "\n")
		case domain.CanastaPhaseRoundEnd:
			b.WriteString(i18n.T("canasta.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("canasta.promptRoundEndHelp") + "\n")
		}
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *CanastaCuiPresenter) ActionLogOutput(g interfaces.CanastaGame) string {
	return actionLogOutputText(g)
}
