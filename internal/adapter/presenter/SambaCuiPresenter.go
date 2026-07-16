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

// sambaMeldTypeLabel はメルドの表示ラベルを返す。
func sambaMeldTypeLabel(m *domain.SambaMeld) string {
	if m.Kind == domain.SambaMeldSequence {
		label := i18n.T("samba.meldTypeSequence")
		if m.IsSamba() {
			label += i18n.T("samba.meldTypeSambaSuffix")
		}
		return label
	}
	label := i18n.T("samba.meldTypeMixed")
	if m.IsNatural {
		label = i18n.T("samba.meldTypeNatural")
	}
	if m.IsCanasta() {
		label += i18n.T("samba.meldTypeCanastaSuffix")
	}
	return label
}

// sambaPlayerStr returns the display string for a single Samba player.
func sambaPlayerStr(player *domain.SambaPlayer, i int, showCards bool) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("samba.playerLine",
		"name", cuiPlayerName(player, i),
		"team", strconv.Itoa(player.GetTeam()),
		"cum", strconv.Itoa(player.GetCumulativeScore()),
		"round", strconv.Itoa(player.GetRoundScore()),
		"cards", strconv.Itoa(player.GetCardsSize())))
	if len(player.GetRed3s()) > 0 {
		b.WriteString(i18n.Tf("samba.playerRed3s",
			"count", strconv.Itoa(len(player.GetRed3s()))))
	}
	if player.HasCanasta() {
		b.WriteString(i18n.T("samba.playerCanastaTag"))
	}
	if player.HasSamba() {
		b.WriteString(i18n.T("samba.playerSambaTag"))
	}
	b.WriteString("\n")

	for _, m := range player.GetMelds() {
		cardStrs := make([]string, len(m.Cards))
		for j, c := range m.Cards {
			cardStrs[j] = cuiCardStr(c)
		}
		b.WriteString(i18n.Tf("samba.meldLine",
			"type", sambaMeldTypeLabel(m),
			"cards", strings.Join(cardStrs, ", ")) + "\n")
	}

	if showCards && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// SambaCuiPresenter renders the Samba CUI view.
type SambaCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *SambaCuiPresenter) Output(g interfaces.SambaGame, lastErr error) string {
	return buildCuiOutput(i18n.T("samba.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("samba.header",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"stock", strconv.Itoa(g.GetDrawPileCount()),
			"discard", strconv.Itoa(g.GetDiscardPileCount())))
		if g.GetIsFrozen() {
			b.WriteString(i18n.T("samba.frozenTag"))
		}
		b.WriteString("\n")

		b.WriteString(i18n.Tf("samba.teamScoreLine",
			"t0", strconv.Itoa(g.GetTeamScore(0)),
			"t1", strconv.Itoa(g.GetTeamScore(1))) + "\n")

		if top := g.GetDiscardTop(); top != nil {
			b.WriteString(i18n.Tf("samba.discardLine", "card", cuiCardStr(top)) + "\n")
		}

		phase := g.GetPhase()
		showAllCards := phase == domain.SambaPhaseRoundEnd || phase == domain.SambaPhaseGameEnd
		for i := 0; i < g.GetPlayerCnt(); i++ {
			player := g.GetPlayer(i)
			showCards := player.GetIsHuman() || showAllCards
			b.WriteString(sambaPlayerStr(player, i, showCards))
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winnerIdx := g.GetWinnerIdx()
			banner := i18n.Tf("samba.gameEnd",
				"name", cuiPlayerName(g.GetPlayer(winnerIdx), winnerIdx))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch phase {
		case domain.SambaPhaseDraw:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("samba.promptDraw",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("samba.promptDrawHelpStock") + "\n")
			b.WriteString(i18n.T("samba.promptDrawHelpDiscard") + "\n")
			// While the pile is frozen, taking the discard requires showing two
			// natural matching cards from hand — spell out that extra condition.
			if g.GetIsFrozen() {
				b.WriteString(i18n.T("samba.promptDrawHelpFrozen") + "\n")
			}
		case domain.SambaPhaseMeld:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("samba.promptMeld",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("samba.promptMeldHelp") + "\n")
			b.WriteString(i18n.T("samba.promptSkipMeld") + "\n")
		case domain.SambaPhaseDiscard:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("samba.promptDiscard",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("samba.promptDiscardHelp") + "\n")
			b.WriteString(i18n.T("samba.promptGoOutHelp") + "\n")
		case domain.SambaPhaseRoundEnd:
			b.WriteString(i18n.T("samba.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("samba.promptRoundEndHelp") + "\n")
		}
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *SambaCuiPresenter) ActionLogOutput(g interfaces.SambaGame) string {
	return actionLogOutputText(g)
}
