//go:build !js || !wasm || extra3

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// klaberjassReveal は手札を公開する局面かを返す。
func klaberjassReveal(g interfaces.KlaberjassGame) bool {
	phase := g.GetPhase()
	return phase == domain.KlaberjassPhaseHandEnd || phase == domain.KlaberjassPhaseGameEnd
}

// klaberjassSuitName は切札スートの表示名を返す。
func klaberjassSuitName(suit int) string {
	switch suit {
	case domain.CardDesignSpade:
		return "♠"
	case domain.CardDesignClover:
		return "♣"
	case domain.CardDesignHeart:
		return "♥"
	case domain.CardDesignDiamond:
		return "♦"
	}
	return "-"
}

// klaberjassPlayerStr returns the display string for a single Klaberjass player.
func klaberjassPlayerStr(g interfaces.KlaberjassGame, i int) string {
	player := g.GetPlayer(i)
	if player == nil {
		return ""
	}
	hand := i18n.Tf("klaberjass.hiddenHand", "count", strconv.Itoa(player.GetCardsSize()))
	if player.GetIsHuman() || klaberjassReveal(g) {
		var b strings.Builder
		for j := range player.GetCardsSize() {
			b.WriteString("[" + strconv.Itoa(j) + "]" + cuiCardStr(player.GetCard(j)) + " ")
		}
		hand = strings.TrimSpace(b.String())
	}
	role := ""
	if i == g.GetDealerIdx() {
		role += " " + i18n.T("klaberjass.dealerTag")
	}
	if i == g.GetMakerIdx() {
		role += " " + i18n.T("klaberjass.makerTag")
	}
	return i18n.Tf("klaberjass.playerLine",
		"name", cuiPlayerName(player, i),
		"role", role,
		"score", strconv.Itoa(g.GetScore(i)),
		"points", strconv.Itoa(g.GetHandPoints(i)),
		"hand", hand) + "\n"
}

// KlaberjassCuiPresenter renders the Klaberjass CUI view.
type KlaberjassCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *KlaberjassCuiPresenter) Output(g interfaces.KlaberjassGame, lastErr error) string {
	return buildCuiOutput(i18n.T("klaberjass.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("klaberjass.header",
			"deal", strconv.Itoa(g.GetDealNumber()),
			"target", strconv.Itoa(g.GetConfig().TargetScore),
			"trump", klaberjassSuitName(g.GetTrumpSuit())) + "\n")

		if c := g.GetTurnUpCard(); c != nil && g.GetTrumpSuit() == 0 {
			b.WriteString(i18n.Tf("klaberjass.turnUp", "card", cuiCardStr(c)) + "\n")
		}

		for i := range g.GetPlayers() {
			b.WriteString(klaberjassPlayerStr(g, i))
		}

		if trick := g.GetTrick(); len(trick) > 0 {
			var t strings.Builder
			for _, c := range trick {
				t.WriteString(cuiCardStr(c) + " ")
			}
			b.WriteString(i18n.Tf("klaberjass.trick", "cards", strings.TrimSpace(t.String())) + "\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			banner := i18n.Tf("klaberjass.gameEnd",
				"name", cuiPlayerName(g.GetPlayer(g.GetWinnerIdx()), g.GetWinnerIdx()))
			b.WriteString(color.Green(banner) + "\n")
			return
		}

		switch g.GetPhase() {
		case domain.KlaberjassPhaseBidTurnUp:
			idx := g.GetBidPlayerIdx()
			b.WriteString(i18n.Tf("klaberjass.promptBid", "name", cuiPlayerName(g.GetPlayer(idx), idx)) + "\n")
			b.WriteString(i18n.T("klaberjass.promptBidTurnUpHelp") + "\n")
		case domain.KlaberjassPhaseBidFree:
			idx := g.GetBidPlayerIdx()
			b.WriteString(i18n.Tf("klaberjass.promptBid", "name", cuiPlayerName(g.GetPlayer(idx), idx)) + "\n")
			b.WriteString(i18n.T("klaberjass.promptBidFreeHelp") + "\n")
		case domain.KlaberjassPhaseSchmeiss:
			b.WriteString(i18n.T("klaberjass.promptSchmeiss") + "\n")
			b.WriteString(i18n.T("klaberjass.promptSchmeissHelp") + "\n")
		case domain.KlaberjassPhasePlay:
			idx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("klaberjass.promptPlay", "name", cuiPlayerName(g.GetPlayer(idx), idx)) + "\n")
			if g.IsHumanTurn() {
				var v strings.Builder
				for _, i := range g.KlaberjassValidPlays(idx) {
					v.WriteString(strconv.Itoa(i) + " ")
				}
				b.WriteString(i18n.Tf("klaberjass.playable", "indexes", strings.TrimSpace(v.String())) + "\n")
			}
			b.WriteString(i18n.T("klaberjass.promptPlayHelp") + "\n")
		case domain.KlaberjassPhaseHandEnd:
			if g.IsBete() {
				b.WriteString(i18n.T("klaberjass.beteLine") + "\n")
			}
			b.WriteString(i18n.T("klaberjass.promptHandEndHelp") + "\n")
		}
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *KlaberjassCuiPresenter) ActionLogOutput(g interfaces.KlaberjassGame) string {
	return actionLogOutputText(g)
}
