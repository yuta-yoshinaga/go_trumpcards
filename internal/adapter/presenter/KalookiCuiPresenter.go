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

// kalookiPlayerStr returns the display string for a single Kalooki player.
func kalookiPlayerStr(player *domain.KalookiPlayer, i int) string {
	var b strings.Builder
	openStatus := i18n.T("kalooki.notOpened")
	if player.HasOpened() {
		openStatus = i18n.T("kalooki.opened")
	}
	b.WriteString(i18n.Tf("kalooki.playerLine",
		"name", cuiPlayerName(player, i),
		"cum", strconv.Itoa(player.GetCumulativeScore()),
		"round", strconv.Itoa(player.GetRoundScore()),
		"cards", strconv.Itoa(player.GetCardsSize()),
		"status", openStatus) + "\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	if player.GetMeldCount() > 0 {
		b.WriteString(i18n.T("kalooki.meldsHeader"))
		for mi := 0; mi < player.GetMeldCount(); mi++ {
			meld := player.GetMeld(mi)
			if mi > 0 {
				b.WriteString(" | ")
			}
			cardStrs := make([]string, len(meld))
			for ci, c := range meld {
				cardStrs[ci] = cuiCardStr(c)
			}
			b.WriteString(strings.Join(cardStrs, " "))
		}
		b.WriteString("\n")
	}
	return b.String()
}

// KalookiCuiPresenter renders the Kalooki CUI view.
type KalookiCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *KalookiCuiPresenter) Output(g interfaces.KalookiGame, lastErr error) string {
	return buildCuiOutput(i18n.T("kalooki.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("kalooki.header",
			"threshold", strconv.Itoa(g.GetOpeningThreshold()),
			"stock", strconv.Itoa(g.GetDrawPileCount())) + "\n")

		if top := g.GetDiscardTop(); top != nil {
			b.WriteString(i18n.Tf("kalooki.discardLine", "card", cuiCardStr(top)) + "\n")
		}

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(kalookiPlayerStr(g.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winnerIdx := g.GetWinnerIdx()
			banner := i18n.Tf("kalooki.gameEnd",
				"name", cuiPlayerName(g.GetPlayer(winnerIdx), winnerIdx))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.KalookiPhaseDraw:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("kalooki.promptDraw",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("kalooki.promptDrawHelpStock") + "\n")
			b.WriteString(i18n.T("kalooki.promptDrawHelpDiscard") + "\n")
		case domain.KalookiPhaseMeld:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("kalooki.promptMeld",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			if cur := g.GetPlayer(currentIdx); cur != nil && !cur.HasOpened() {
				b.WriteString(i18n.Tf("kalooki.openingHint", "n", strconv.Itoa(g.GetOpeningThreshold())) + "\n")
			}
			b.WriteString(i18n.T("kalooki.promptMeldHelpMeld") + "\n")
			b.WriteString(i18n.T("kalooki.promptMeldHelpLayoff") + "\n")
			b.WriteString(i18n.T("kalooki.promptMeldHelpDiscard") + "\n")
		case domain.KalookiPhaseRoundEnd:
			b.WriteString(i18n.T("kalooki.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("kalooki.promptRoundEndHelp") + "\n")
		}
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *KalookiCuiPresenter) ActionLogOutput(g interfaces.KalookiGame) string {
	return actionLogOutputTextForSeats[*domain.KalookiPlayer](g)
}
