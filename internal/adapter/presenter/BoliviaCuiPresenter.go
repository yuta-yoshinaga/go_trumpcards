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

// boliviaMeldTypeLabel はメルドの表示ラベルを返す。
//
// **3 種類ある。** エスカレラ (ワイルド無しの同スート連番) と、ワイルドだけの
// メルド ── 7 枚揃うと後者が「ボリビア」。両者を同じ言葉で出すと、上がりに
// 必要なのがどちらなのか読めなくなる。
func boliviaMeldTypeLabel(m *domain.BoliviaMeld) string {
	if m.Kind == domain.BoliviaMeldWild {
		label := i18n.T("bolivia.meldTypeWild")
		if m.IsBoliviaCanasta() {
			label += i18n.T("bolivia.meldTypeBoliviaSuffix")
		}
		return label
	}
	if m.Kind == domain.BoliviaMeldEscalera {
		label := i18n.T("bolivia.meldTypeSequence")
		if m.IsEscalera() {
			label += i18n.T("bolivia.meldTypeEscaleraSuffix")
		}
		return label
	}
	label := i18n.T("bolivia.meldTypeMixed")
	if m.IsNatural {
		label = i18n.T("bolivia.meldTypeNatural")
	}
	if m.IsCanasta() {
		label += i18n.T("bolivia.meldTypeCanastaSuffix")
	}
	return label
}

// boliviaPlayerStr returns the display string for a single Bolivia player.
func boliviaPlayerStr(player *domain.BoliviaPlayer, i int, showCards bool) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("bolivia.playerLine",
		"name", cuiPlayerName(player, i),
		"team", strconv.Itoa(player.GetTeam()),
		"cum", strconv.Itoa(player.GetCumulativeScore()),
		"round", strconv.Itoa(player.GetRoundScore()),
		"cards", strconv.Itoa(player.GetCardsSize())))
	if len(player.GetRed3s()) > 0 {
		b.WriteString(i18n.Tf("bolivia.playerRed3s",
			"count", strconv.Itoa(len(player.GetRed3s()))))
	}
	if player.HasCanasta() {
		b.WriteString(i18n.T("bolivia.playerCanastaTag"))
	}
	if player.HasEscalera() {
		b.WriteString(i18n.T("bolivia.playerEscaleraTag"))
	}
	b.WriteString("\n")

	for _, m := range player.GetMelds() {
		cardStrs := make([]string, len(m.Cards))
		for j, c := range m.Cards {
			cardStrs[j] = cuiCardStr(c)
		}
		b.WriteString(i18n.Tf("bolivia.meldLine",
			"type", boliviaMeldTypeLabel(m),
			"cards", strings.Join(cardStrs, ", ")) + "\n")
	}

	if showCards && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// BoliviaCuiPresenter renders the Bolivia CUI view.
type BoliviaCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *BoliviaCuiPresenter) Output(g interfaces.BoliviaGame, lastErr error) string {
	return buildCuiOutput(i18n.T("bolivia.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("bolivia.header",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"stock", strconv.Itoa(g.GetDrawPileCount()),
			"discard", strconv.Itoa(g.GetDiscardPileCount())))
		if g.GetIsFrozen() {
			b.WriteString(i18n.T("bolivia.frozenTag"))
		}
		b.WriteString("\n")

		b.WriteString(i18n.Tf("bolivia.teamScoreLine",
			"t0", strconv.Itoa(g.GetTeamScore(0)),
			"t1", strconv.Itoa(g.GetTeamScore(1))) + "\n")

		if top := g.GetDiscardTop(); top != nil {
			b.WriteString(i18n.Tf("bolivia.discardLine", "card", cuiCardStr(top)) + "\n")
		}

		phase := g.GetPhase()
		showAllCards := phase == domain.BoliviaPhaseRoundEnd || phase == domain.BoliviaPhaseGameEnd
		for i := 0; i < g.GetPlayerCnt(); i++ {
			player := g.GetPlayer(i)
			showCards := player.GetIsHuman() || showAllCards
			b.WriteString(boliviaPlayerStr(player, i, showCards))
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winnerIdx := g.GetWinnerIdx()
			banner := i18n.Tf("bolivia.gameEnd",
				"name", cuiPlayerName(g.GetPlayer(winnerIdx), winnerIdx))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch phase {
		case domain.BoliviaPhaseDraw:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("bolivia.promptDraw",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("bolivia.promptDrawHelpStock") + "\n")
			b.WriteString(i18n.T("bolivia.promptDrawHelpDiscard") + "\n")
			// While the pile is frozen, taking the discard requires showing two
			// natural matching cards from hand — spell out that extra condition.
			if g.GetIsFrozen() {
				b.WriteString(i18n.T("bolivia.promptDrawHelpFrozen") + "\n")
			}
		case domain.BoliviaPhaseMeld:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("bolivia.promptMeld",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			// 初回メルドの最低点はチーム累積点で 15/50/90/120 と変わる。届いて
			// いるかをメルドしてみるまで分からないのは酷なので、先に出す。
			// 初回メルドを済ませた席には課されないので出さない。
			if player := g.GetPlayer(currentIdx); player != nil && !player.GetHasInitMeld() {
				b.WriteString(i18n.Tf("bolivia.promptMeldMinimum",
					"points", strconv.Itoa(g.GetMinimumMeldValue(currentIdx))) + "\n")
			}
			b.WriteString(i18n.T("bolivia.promptMeldHelp") + "\n")
			b.WriteString(i18n.T("bolivia.promptSkipMeld") + "\n")
		case domain.BoliviaPhaseDiscard:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("bolivia.promptDiscard",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("bolivia.promptDiscardHelp") + "\n")
			b.WriteString(i18n.T("bolivia.promptGoOutHelp") + "\n")
		case domain.BoliviaPhaseRoundEnd:
			b.WriteString(i18n.T("bolivia.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("bolivia.promptRoundEndHelp") + "\n")
		}
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *BoliviaCuiPresenter) ActionLogOutput(g interfaces.BoliviaGame) string {
	return actionLogOutputTextForSeats[*domain.BoliviaPlayer](g)
}
