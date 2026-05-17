package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// rummy500MeldLine 場に出ているメルドを1行にまとめる
func rummy500MeldLine(meld []*domain.Card) string {
	parts := make([]string, 0, len(meld))
	for _, c := range meld {
		parts = append(parts, cuiCardStr(c))
	}
	return strings.Join(parts, ",")
}

// rummy500PlayerStr returns the display string for a single Rummy500 player.
func rummy500PlayerStr(player *domain.Rummy500Player, i int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("rummy500.playerLine",
		"name", cuiPlayerName(player, i),
		"cum", strconv.Itoa(player.GetCumulativeScore()),
		"round", strconv.Itoa(player.GetRoundScore()),
		"cards", strconv.Itoa(player.GetCardsSize())) + "\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	for mi, meld := range player.GetLaidMelds() {
		b.WriteString(i18n.Tf("rummy500.meldLine",
			"idx", strconv.Itoa(mi),
			"cards", rummy500MeldLine(meld)) + "\n")
	}
	return b.String()
}

// rummy500DiscardLine 捨て札の山を表示する
func rummy500DiscardLine(g interfaces.Rummy500Game) string {
	pile := g.GetDiscardPile()
	if len(pile) == 0 {
		return i18n.T("rummy500.discardEmpty")
	}
	parts := make([]string, 0, len(pile))
	for i, c := range pile {
		parts = append(parts, "["+strconv.Itoa(i)+"]"+cuiCardStr(c))
	}
	return i18n.Tf("rummy500.discardLine", "cards", strings.Join(parts, " "))
}

// Rummy500CuiPresenter renders the Rummy 500 CUI view.
type Rummy500CuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *Rummy500CuiPresenter) Output(g interfaces.Rummy500Game, lastErr error) string {
	return buildCuiOutput(i18n.T("rummy500.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("rummy500.header",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"stock", strconv.Itoa(g.GetDrawPileCount())) + "\n")
		b.WriteString(rummy500DiscardLine(g) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(rummy500PlayerStr(g.GetPlayer(i), i))
		}

		b.WriteString("----------\n")
		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winnerIdx := g.GetWinnerIdx()
			banner := i18n.Tf("rummy500.gameEnd",
				"name", cuiPlayerName(g.GetPlayer(winnerIdx), winnerIdx))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.Rummy500PhaseDraw:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("rummy500.promptDraw",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("rummy500.promptDrawHelpStock") + "\n")
			b.WriteString(i18n.T("rummy500.promptDrawHelpDiscard") + "\n")
		case domain.Rummy500PhasePlay:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("rummy500.promptPlay",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("rummy500.promptPlayHelpMeld") + "\n")
			b.WriteString(i18n.T("rummy500.promptPlayHelpLayoff") + "\n")
			b.WriteString(i18n.T("rummy500.promptPlayHelpDiscard") + "\n")
		case domain.Rummy500PhaseRoundEnd:
			b.WriteString(i18n.T("rummy500.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("rummy500.promptRoundEndHelp") + "\n")
		}
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *Rummy500CuiPresenter) ActionLogOutput(g interfaces.Rummy500Game) string {
	return actionLogOutputText(g)
}
