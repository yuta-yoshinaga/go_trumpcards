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

// conquianPlayerStr returns the display string for a single Conquian player.
func conquianPlayerStr(player *domain.ConquianPlayer, i int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("conquian.playerLine",
		"name", cuiPlayerName(player, i),
		"wins", strconv.Itoa(player.GetWins()),
		"cards", strconv.Itoa(player.GetCardsSize())) + "\n")
	melds := player.GetMelds()
	for _, meld := range melds {
		b.WriteString(i18n.Tf("conquian.meldLine", "cards", cuiCardSliceStr(meld)) + "\n")
	}
	// **勝利はメルドを並べきること**なのに、CUI は中身を列挙するだけで、あと何枚
	// なのかは毎回目で数えるしかなかった (#5664)。Web は進捗バーで出しており、
	// 残り2枚以下を警告色にしている。同じ基準で1行にする。
	melded := 0
	for _, meld := range melds {
		melded += len(meld)
	}
	line := i18n.Tf("conquian.meldProgress",
		"count", strconv.Itoa(melded),
		"total", strconv.Itoa(domain.ConquianMeldTarget))
	if remaining := domain.ConquianMeldTarget - melded; remaining > 0 && remaining <= 2 {
		line += "  " + color.BoldYellow(i18n.Tf("conquian.meldRemaining", "count", strconv.Itoa(remaining)))
	}
	b.WriteString(line + "\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// ConquianCuiPresenter renders the Conquian CUI view.
type ConquianCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *ConquianCuiPresenter) Output(g interfaces.ConquianGame, lastErr error) string {
	return buildCuiOutput(i18n.T("conquian.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("conquian.header",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"stock", strconv.Itoa(g.GetDrawPileCount())) + "\n")

		if top := g.GetDiscardTop(); top != nil {
			b.WriteString(i18n.Tf("conquian.discardLine", "card", cuiCardStr(top)) + "\n")
		}

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(conquianPlayerStr(g.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winnerIdx := g.GetWinnerIdx()
			if winnerIdx < 0 {
				b.WriteString(color.Green(i18n.T("conquian.drawBanner")) + "\n")
				return
			}
			banner := i18n.Tf("conquian.gameEnd",
				"name", cuiPlayerName(g.GetPlayer(winnerIdx), winnerIdx))
			b.WriteString(color.Green(banner) + "\n")
			return
		}

		switch g.GetPhase() {
		case domain.ConquianPhaseDraw:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("conquian.promptDraw",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("conquian.promptDrawHelpStock") + "\n")
			b.WriteString(i18n.T("conquian.promptDrawHelpDiscard") + "\n")
		case domain.ConquianPhaseMeld:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("conquian.promptMeld",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			if g.GetTookDiscard() {
				b.WriteString(color.Yellow(i18n.T("conquian.forcedUse")) + "\n")
			}
			b.WriteString(i18n.T("conquian.promptMeldHelp") + "\n")
			b.WriteString(i18n.T("conquian.promptDiscardHelp") + "\n")
		case domain.ConquianPhaseRoundEnd:
			b.WriteString(i18n.T("conquian.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("conquian.promptRoundEndHelp") + "\n")
		}
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *ConquianCuiPresenter) ActionLogOutput(g interfaces.ConquianGame) string {
	return actionLogOutputText(g)
}
