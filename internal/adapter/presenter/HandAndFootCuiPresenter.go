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

// handAndFootPlayerStr returns the display string for a single player.
func handAndFootPlayerStr(player *domain.HandAndFootPlayer, i int, showCards bool) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("handandfoot.playerLine",
		"name", cuiPlayerName(player, i),
		"team", strconv.Itoa(domain.HandAndFootTeamOf(i)),
		"cum", strconv.Itoa(player.GetCumulativeScore()),
		"round", strconv.Itoa(player.GetRoundScore()),
		"cards", strconv.Itoa(player.GetCardsSize()),
		"foot", strconv.Itoa(player.GetFootSize())))
	if player.GetInFoot() {
		b.WriteString(i18n.T("handandfoot.inFootTag"))
	}
	b.WriteString("\n")

	if showCards && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// handAndFootTeamStr returns the meld display string for one team.
func handAndFootTeamStr(g interfaces.HandAndFootGame, team int) string {
	var b strings.Builder
	melds := g.GetTeamMelds(team)
	red3s := g.GetTeamRed3s(team)
	if len(melds) == 0 && len(red3s) == 0 {
		return ""
	}
	b.WriteString(i18n.Tf("handandfoot.teamHeader", "team", strconv.Itoa(team)))
	if len(red3s) > 0 {
		b.WriteString(i18n.Tf("handandfoot.teamRed3s", "count", strconv.Itoa(len(red3s))))
	}
	b.WriteString("\n")
	for _, m := range melds {
		meldType := i18n.T("handandfoot.meldTypeBlack")
		if m.IsNatural {
			meldType = i18n.T("handandfoot.meldTypeRed")
		}
		if m.IsCanasta() {
			meldType += i18n.T("handandfoot.meldTypeCanastaSuffix")
		}
		cardStrs := make([]string, len(m.Cards))
		for j, c := range m.Cards {
			cardStrs[j] = cuiCardStr(c)
		}
		b.WriteString(i18n.Tf("handandfoot.meldLine",
			"type", meldType,
			"cards", strings.Join(cardStrs, ", ")) + "\n")
	}
	return b.String()
}

// HandAndFootCuiPresenter renders the Hand and Foot CUI view.
type HandAndFootCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *HandAndFootCuiPresenter) Output(g interfaces.HandAndFootGame, lastErr error) string {
	return buildCuiOutput(i18n.T("handandfoot.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("handandfoot.header",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"stock", strconv.Itoa(g.GetDrawPileCount()),
			"discard", strconv.Itoa(g.GetDiscardPileCount())))
		if g.GetIsFrozen() {
			b.WriteString(i18n.T("handandfoot.frozenTag"))
		}
		b.WriteString("\n")

		if top := g.GetDiscardTop(); top != nil {
			b.WriteString(i18n.Tf("handandfoot.discardLine", "card", cuiCardStr(top)) + "\n")
		}

		// Team melds
		for t := 0; t < domain.HandAndFootTeamCnt; t++ {
			b.WriteString(handAndFootTeamStr(g, t))
		}

		// Players
		phase := g.GetPhase()
		showAllCards := phase == domain.HandAndFootPhaseRoundEnd || phase == domain.HandAndFootPhaseGameEnd
		for i := 0; i < g.GetPlayerCnt(); i++ {
			player := g.GetPlayer(i)
			showCards := player.GetIsHuman() || showAllCards
			b.WriteString(handAndFootPlayerStr(player, i, showCards))
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			banner := i18n.Tf("handandfoot.gameEnd", "team", strconv.Itoa(g.GetWinnerTeam()))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch phase {
		case domain.HandAndFootPhaseDraw:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("handandfoot.promptDraw",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("handandfoot.promptDrawHelpStock") + "\n")
			b.WriteString(i18n.T("handandfoot.promptDrawHelpDiscard") + "\n")
		case domain.HandAndFootPhaseMeld:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("handandfoot.promptMeld",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("handandfoot.promptMeldHelp") + "\n")
			b.WriteString(i18n.T("handandfoot.promptSkipMeld") + "\n")
		case domain.HandAndFootPhaseDiscard:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("handandfoot.promptDiscard",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("handandfoot.promptDiscardHelp") + "\n")
			b.WriteString(i18n.T("handandfoot.promptGoOutHelp") + "\n")
		case domain.HandAndFootPhaseRoundEnd:
			b.WriteString(i18n.T("handandfoot.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("handandfoot.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput lists the meld groups the human can currently form, reusing the
// shared domain suggestion logic. Returns a "no meld" message when none exist.
func (p *HandAndFootCuiPresenter) HintOutput(g interfaces.HandAndFootGame) string {
	if !g.IsHumanTurn() {
		return i18n.T("handandfoot.hintNotYourTurn") + "\n"
	}
	melds := g.SuggestMelds(g.GetCurrentPlayerIdx())
	if len(melds) == 0 {
		return i18n.T("handandfoot.hintNone") + "\n"
	}
	parts := make([]string, len(melds))
	for i, group := range melds {
		cards := make([]string, len(group))
		for j, c := range group {
			cards[j] = cuiCardStr(c)
		}
		parts[i] = "[" + strings.Join(cards, " ") + "]"
	}
	return color.Yellow(i18n.Tf("handandfoot.hintMeld", "melds", strings.Join(parts, ", "))) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *HandAndFootCuiPresenter) ActionLogOutput(g interfaces.HandAndFootGame) string {
	return actionLogOutputText(g)
}
