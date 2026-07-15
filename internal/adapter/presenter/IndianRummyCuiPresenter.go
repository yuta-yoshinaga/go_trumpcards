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

// indianRummyPlayerStr returns the display string for a single Indian Rummy player.
func indianRummyPlayerStr(g interfaces.IndianRummyGame, i int) string {
	player := g.GetPlayer(i)
	var b strings.Builder
	b.WriteString(i18n.Tf("indianrummy.playerLine",
		"name", cuiPlayerName(player, i),
		"cum", strconv.Itoa(player.GetCumulativeScore()),
		"round", strconv.Itoa(player.GetRoundScore()),
		"cards", strconv.Itoa(player.GetCardsSize())) + "\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// IndianRummyCuiPresenter renders the Indian Rummy CUI view.
type IndianRummyCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *IndianRummyCuiPresenter) Output(g interfaces.IndianRummyGame, lastErr error) string {
	return buildCuiOutput(i18n.T("indianrummy.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("indianrummy.header",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"total", strconv.Itoa(g.GetTargetRounds()),
			"stock", strconv.Itoa(g.GetDrawPileCount())) + "\n")

		if wj := g.GetWildJoker(); wj != nil {
			b.WriteString(i18n.Tf("indianrummy.wildLine", "card", cuiCardStr(wj)) + "\n")
		}

		if top := g.GetDiscardTop(); top != nil {
			b.WriteString(i18n.Tf("indianrummy.discardLine", "card", cuiCardStr(top)) + "\n")
		}

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(indianRummyPlayerStr(g, i))
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winnerIdx := g.GetWinnerIdx()
			banner := i18n.Tf("indianrummy.gameEnd",
				"name", cuiPlayerName(g.GetPlayer(winnerIdx), winnerIdx))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.IndianRummyPhaseDraw:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("indianrummy.promptDraw",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("indianrummy.promptDrawHelpStock") + "\n")
			b.WriteString(i18n.T("indianrummy.promptDrawHelpDiscard") + "\n")
		case domain.IndianRummyPhaseDiscard:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("indianrummy.promptDiscard",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			// A misdeclaration is heavily penalized here, so surface the human's
			// current deadwood and whether the mandatory pure sequence is met
			// (parity with GinRummy's deadwood/knock hints).
			if g.GetPlayer(currentIdx).GetIsHuman() {
				b.WriteString(i18n.Tf("indianrummy.deadwoodLine",
					"value", strconv.Itoa(g.PlayerDeadwoodValue(currentIdx))) + "\n")
				if g.PlayerHasPureSequence(currentIdx) {
					b.WriteString(color.Green(i18n.T("indianrummy.pureSequenceMet")) + "\n")
				} else {
					b.WriteString(color.Yellow(i18n.T("indianrummy.pureSequenceUnmet")) + "\n")
				}
			}
			b.WriteString(i18n.T("indianrummy.promptDiscardHelp") + "\n")
			b.WriteString(i18n.T("indianrummy.promptDeclareHelp") + "\n")
		case domain.IndianRummyPhaseRoundEnd:
			b.WriteString(i18n.T("indianrummy.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("indianrummy.promptRoundEndHelp") + "\n")
		}
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *IndianRummyCuiPresenter) ActionLogOutput(g interfaces.IndianRummyGame) string {
	return actionLogOutputText(g)
}
