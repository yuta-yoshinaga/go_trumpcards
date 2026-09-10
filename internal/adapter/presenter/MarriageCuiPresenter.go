//go:build !js || !wasm || extra5

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// marriagePlayerStr returns the display string for a single Marriage player.
func marriagePlayerStr(g interfaces.MarriageGame, i int) string {
	player := g.GetPlayer(i)
	phase := g.GetPhase()
	revealAll := phase == domain.MarriagePhaseRoundEnd || phase == domain.MarriagePhaseGameEnd
	playerLineKey := "marriage.playerLine"
	if revealAll || player.GetIsHuman() {
		playerLineKey = "marriage.playerLineWithMaal"
	}
	var b strings.Builder
	params := []string{
		"name", cuiPlayerName(player, i),
		"cum", strconv.Itoa(player.GetCumulativeScore()),
		"round", strconv.Itoa(player.GetRoundScore()),
		"cards", strconv.Itoa(player.GetCardsSize()),
	}
	if playerLineKey == "marriage.playerLineWithMaal" {
		params = append(params, "maal", strconv.Itoa(g.PlayerMaalValue(i)))
	}
	b.WriteString(i18n.Tf(playerLineKey, params...) + "\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// MarriageCuiPresenter renders the Marriage CUI view.
type MarriageCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *MarriageCuiPresenter) Output(g interfaces.MarriageGame, lastErr error) string {
	return buildCuiOutput(i18n.T("marriage.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("marriage.header",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"total", strconv.Itoa(g.GetTargetRounds()),
			"stock", strconv.Itoa(g.GetDrawPileCount())) + "\n")

		if wj := g.GetWildJoker(); wj != nil {
			b.WriteString(i18n.Tf("marriage.wildLine", "card", cuiCardStr(wj)) + "\n")
		}

		if top := g.GetDiscardTop(); top != nil {
			b.WriteString(i18n.Tf("marriage.discardLine", "card", cuiCardStr(top)) + "\n")
		}

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(marriagePlayerStr(g, i))
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winnerIdx := g.GetWinnerIdx()
			banner := i18n.Tf("marriage.gameEnd",
				"name", cuiPlayerName(g.GetPlayer(winnerIdx), winnerIdx))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.MarriagePhaseDraw:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("marriage.promptDraw",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("marriage.promptDrawHelpStock") + "\n")
			b.WriteString(i18n.T("marriage.promptDrawHelpDiscard") + "\n")
		case domain.MarriagePhaseDiscard:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("marriage.promptDiscard",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			// A misdeclaration is heavily penalized here, so surface the human's
			// current deadwood and whether the mandatory pure sequence is met
			// (parity with GinRummy's deadwood/knock hints).
			if g.GetPlayer(currentIdx).GetIsHuman() {
				b.WriteString(i18n.Tf("marriage.deadwoodLine",
					"value", strconv.Itoa(g.PlayerDeadwoodValue(currentIdx))) + "\n")
				// The number above is meaningless without the scale, and this
				// scale is the one a Gin Rummy player gets wrong: the ace is 10
				// here, not 1, and a wild costs nothing. The web prints the same
				// legend beside its deadwood figure (#5501).
				b.WriteString(i18n.T("marriage.pointsLegend") + "\n")
				if g.PlayerHasPureSequence(currentIdx) {
					b.WriteString(color.Green(i18n.T("marriage.pureSequenceMet")) + "\n")
				} else {
					b.WriteString(color.Yellow(i18n.T("marriage.pureSequenceUnmet")) + "\n")
				}
			}
			b.WriteString(i18n.T("marriage.promptDiscardHelp") + "\n")
			b.WriteString(i18n.T("marriage.promptDeclareHelp") + "\n")
		case domain.MarriagePhaseRoundEnd:
			b.WriteString(i18n.T("marriage.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("marriage.promptRoundEndHelp") + "\n")
		}
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *MarriageCuiPresenter) ActionLogOutput(g interfaces.MarriageGame) string {
	return actionLogOutputTextForSeats[*domain.MarriagePlayer](g)
}
