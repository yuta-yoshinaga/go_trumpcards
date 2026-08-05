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

// panMeldLine 場に出ているメルドを 1 行にまとめる
func panMeldLine(meld []*domain.Card) string {
	parts := make([]string, 0, len(meld))
	for _, c := range meld {
		parts = append(parts, cuiCardStr(c))
	}
	line := strings.Join(parts, ",")
	// **チップが動いた理由が分かるようにする。**バジェ (3/5/7 のセット) は
	// 各プレイヤーにチップを配る特別ルールなのに、盤面のどのメルドがそれなのか
	// どこにも出ていなかった (#4853)。
	if domain.PanIsValleMeld(meld) {
		line += " " + i18n.T("pan.valleTag")
	}
	return line
}

// panPlayerStr returns the display string for a single Pan player.
func panPlayerStr(player *domain.PanPlayer, i int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("pan.playerLine",
		"name", cuiPlayerName(player, i),
		"cum", strconv.Itoa(player.GetCumulativeScore()),
		"round", strconv.Itoa(player.GetRoundScore()),
		"chips", strconv.Itoa(player.GetChips()),
		"melded", strconv.Itoa(player.GetMeldedCardCount()),
		"cards", strconv.Itoa(player.GetCardsSize())) + "\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	for mi, meld := range player.GetLaidMelds() {
		b.WriteString(i18n.Tf("pan.meldLine",
			"idx", strconv.Itoa(mi),
			"cards", panMeldLine(meld)) + "\n")
	}
	return b.String()
}

// PanCuiPresenter renders the Pan CUI view.
type PanCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *PanCuiPresenter) Output(g interfaces.PanGame, lastErr error) string {
	return buildCuiOutput(i18n.T("pan.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("pan.header",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"total", strconv.Itoa(g.GetTargetRounds()),
			"stock", strconv.Itoa(g.GetDrawPileCount())) + "\n")

		if top := g.GetDiscardTop(); top != nil {
			b.WriteString(i18n.Tf("pan.discardLine", "card", cuiCardStr(top)) + "\n")
		}

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(panPlayerStr(g.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winnerIdx := g.GetWinnerIdx()
			banner := i18n.Tf("pan.gameEnd",
				"name", cuiPlayerName(g.GetPlayer(winnerIdx), winnerIdx))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.PanPhaseDraw:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("pan.promptDraw",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("pan.promptDrawHelpStock") + "\n")
			b.WriteString(i18n.T("pan.promptDrawHelpDiscard") + "\n")
		case domain.PanPhasePlay:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("pan.promptPlay",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("pan.promptPlayHelpMeld") + "\n")
			b.WriteString(i18n.T("pan.promptPlayHelpLayoff") + "\n")
			b.WriteString(i18n.T("pan.promptPlayHelpDiscard") + "\n")
			b.WriteString(i18n.T("pan.promptPlayHelpNote") + "\n")
		case domain.PanPhaseRoundEnd:
			b.WriteString(i18n.T("pan.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("pan.promptRoundEndHelp") + "\n")
		}
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *PanCuiPresenter) ActionLogOutput(g interfaces.PanGame) string {
	return actionLogOutputText(g)
}
