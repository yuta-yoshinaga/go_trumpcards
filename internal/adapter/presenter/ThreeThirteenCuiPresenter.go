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

// threeThirteenPlayerStr returns the display string for a single Three Thirteen player.
func threeThirteenPlayerStr(g interfaces.ThreeThirteenGame, player *domain.ThreeThirteenPlayer, i int) string {
	// A CPU's deadwood is hidden information (its hand isn't public); reveal it
	// only for the human, or once all hands are shown at round/game end.
	revealed := g.GetPhase() == domain.ThreeThirteenPhaseRoundEnd || g.GetGameEndFlag()
	deadwoodStr := "?"
	if player.GetIsHuman() || revealed {
		deadwoodStr = strconv.Itoa(g.GetPlayerDeadwoodValue(i))
	}
	var b strings.Builder
	b.WriteString(i18n.Tf("threethirteen.playerLine",
		"name", cuiPlayerName(player, i),
		"cum", strconv.Itoa(player.GetCumulativeScore()),
		"round", strconv.Itoa(player.GetRoundScore()),
		"cards", strconv.Itoa(player.GetCardsSize()),
		"deadwood", deadwoodStr) + "\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		// **ワイルドランクは毎ラウンド変わる** (#5667)。Web は該当札にバッジと
		// リングを付けているのに、CUI はヘッダーの wild 値と 1 枚ずつ照合させて
		// いた。判定はドメインの WildRank に従う。
		var wild []int
		for idx := 0; idx < player.GetCardsSize(); idx++ {
			if player.GetCard(idx).GetValue() == g.WildRank() {
				wild = append(wild, idx)
			}
		}
		b.WriteString(cuiIndexMarkedCardListStr(player, wild, CuiWildMark) + "\n")
		if len(wild) > 0 {
			b.WriteString(i18n.T("threethirteen.wildLegend") + "\n")
		}
		if line := threeThirteenDiscardPreview(g, player, i); line != "" {
			b.WriteString(line + "\n")
		}
	}
	return b.String()
}

// threeThirteenDiscardPreview lists the deadwood the human would be left with
// for each possible discard. Web は 1 枚選ぶたびに同じ値を出している (#4840)。
// ディスカードフェーズ以外では出さない — 捨てられない場面の予測は意味がない。
func threeThirteenDiscardPreview(g interfaces.ThreeThirteenGame, player *domain.ThreeThirteenPlayer, idx int) string {
	if g.GetPhase() != domain.ThreeThirteenPhaseDiscard || g.GetCurrentPlayerIdx() != idx {
		return ""
	}
	parts := make([]string, 0, player.GetCardsSize())
	for i := 0; i < player.GetCardsSize(); i++ {
		after := g.GetDeadwoodAfterDiscard(idx, i)
		if after < 0 {
			continue
		}
		parts = append(parts, "["+strconv.Itoa(i)+"]"+strconv.Itoa(after))
	}
	if len(parts) == 0 {
		return ""
	}
	return i18n.Tf("threethirteen.discardPreview", "values", strings.Join(parts, "  "))
}

// ThreeThirteenCuiPresenter renders the Three Thirteen CUI view.
type ThreeThirteenCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *ThreeThirteenCuiPresenter) Output(g interfaces.ThreeThirteenGame, lastErr error) string {
	return buildCuiOutput(i18n.T("threethirteen.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("threethirteen.header",
			"round", strconv.Itoa(g.GetRound()),
			"wild", strconv.Itoa(g.WildRank()),
			"stock", strconv.Itoa(g.GetDrawPileCount())) + "\n")

		if top := g.GetDiscardTop(); top != nil {
			b.WriteString(i18n.Tf("threethirteen.discardLine", "card", cuiCardStr(top)) + "\n")
		}

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(threeThirteenPlayerStr(g, g.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winnerIdx := g.GetWinnerIdx()
			banner := i18n.Tf("threethirteen.gameEnd",
				"name", cuiPlayerName(g.GetPlayer(winnerIdx), winnerIdx))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.ThreeThirteenPhaseDraw:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("threethirteen.promptDraw",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("threethirteen.promptDrawHelpStock") + "\n")
			b.WriteString(i18n.T("threethirteen.promptDrawHelpDiscard") + "\n")
		case domain.ThreeThirteenPhaseDiscard:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("threethirteen.promptDiscard",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("threethirteen.promptDiscardHelpDiscard") + "\n")
			b.WriteString(i18n.T("threethirteen.promptDiscardHelpKnock") + "\n")
		case domain.ThreeThirteenPhaseRoundEnd:
			b.WriteString(i18n.T("threethirteen.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("threethirteen.promptRoundEndHelp") + "\n")
		}
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *ThreeThirteenCuiPresenter) ActionLogOutput(g interfaces.ThreeThirteenGame) string {
	return actionLogOutputText(g)
}
