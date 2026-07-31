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

// kaiserReveal は手札を公開する局面かを返す。
func kaiserReveal(g interfaces.KaiserGame) bool {
	phase := g.GetPhase()
	return phase == domain.KaiserPhaseHandEnd || phase == domain.KaiserPhaseGameEnd
}

// kaiserContractName は契約種別の表示名キーを返す。
func kaiserContractName(c domain.KaiserContract) string {
	switch c {
	case domain.KaiserContractNoTrump:
		return i18n.T("kaiser.noTrump")
	case domain.KaiserContractLowNoTrump:
		return i18n.T("kaiser.lowNoTrump")
	}
	return i18n.T("kaiser.withTrump")
}

// kaiserSuitName は切札スートの表示名を返す。
func kaiserSuitName(suit int) string {
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

// kaiserPlayerStr returns the display string for a single Kaiser player.
func kaiserPlayerStr(g interfaces.KaiserGame, i int) string {
	player := g.GetPlayer(i)
	if player == nil {
		return ""
	}
	hand := i18n.Tf("kaiser.hiddenHand", "count", strconv.Itoa(player.GetCardsSize()))
	if player.GetIsHuman() || kaiserReveal(g) {
		var b strings.Builder
		for j := range player.GetCardsSize() {
			b.WriteString("[" + strconv.Itoa(j) + "]" + cuiCardStr(player.GetCard(j)) + " ")
		}
		hand = strings.TrimSpace(b.String())
	}
	role := ""
	if i == g.GetDealerIdx() {
		role += " " + i18n.T("kaiser.dealerTag")
	}
	if i == g.GetDeclarerIdx() {
		role += " " + i18n.T("kaiser.declarerTag")
	}
	return i18n.Tf("kaiser.playerLine",
		"name", cuiPlayerName(player, i),
		"team", strconv.Itoa(domain.KaiserTeamOf(i)),
		"role", role,
		"hand", hand) + "\n"
}

// KaiserCuiPresenter renders the Kaiser CUI view.
type KaiserCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *KaiserCuiPresenter) Output(g interfaces.KaiserGame, lastErr error) string {
	return buildCuiOutput(i18n.T("kaiser.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("kaiser.header",
			"hand", strconv.Itoa(g.GetHandNumber()),
			"target", strconv.Itoa(g.GetTargetScore()),
			"t0", strconv.Itoa(g.GetScore(0)),
			"t1", strconv.Itoa(g.GetScore(1))) + "\n")

		if hb := g.GetHighBid(); hb != nil {
			b.WriteString(i18n.Tf("kaiser.contractLine",
				"value", strconv.Itoa(hb.Value),
				"kind", kaiserContractName(g.GetContract()),
				"trump", kaiserSuitName(g.GetTrumpSuit())) + "\n")
		}

		for i := range g.GetPlayers() {
			b.WriteString(kaiserPlayerStr(g, i))
		}

		if trick := g.GetTrick(); len(trick) > 0 {
			var t strings.Builder
			for _, c := range trick {
				t.WriteString(cuiCardStr(c) + " ")
			}
			b.WriteString(i18n.Tf("kaiser.trick", "cards", strings.TrimSpace(t.String())) + "\n")
		}

		b.WriteString(i18n.Tf("kaiser.handPoints",
			"t0", strconv.Itoa(g.GetHandPoints(0)),
			"t1", strconv.Itoa(g.GetHandPoints(1))) + "\n")

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			banner := i18n.Tf("kaiser.gameEnd", "team", strconv.Itoa(g.GetWinnerTeam()))
			b.WriteString(color.Green(banner) + "\n")
			return
		}

		switch g.GetPhase() {
		case domain.KaiserPhaseBid:
			idx := g.GetBidPlayerIdx()
			b.WriteString(i18n.Tf("kaiser.promptBid", "name", cuiPlayerName(g.GetPlayer(idx), idx)) + "\n")
			b.WriteString(i18n.T("kaiser.promptBidHelp") + "\n")
		case domain.KaiserPhaseDiscard:
			if g.GetContract() == domain.KaiserContractTrump && g.GetTrumpSuit() == 0 {
				b.WriteString(i18n.T("kaiser.promptTrump") + "\n")
			} else {
				b.WriteString(i18n.T("kaiser.promptDiscard") + "\n")
			}
		case domain.KaiserPhasePlay:
			idx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("kaiser.promptPlay", "name", cuiPlayerName(g.GetPlayer(idx), idx)) + "\n")
			if g.IsHumanTurn() {
				var v strings.Builder
				for _, i := range g.KaiserValidPlays(idx) {
					v.WriteString(strconv.Itoa(i) + " ")
				}
				b.WriteString(i18n.Tf("kaiser.playable", "indexes", strings.TrimSpace(v.String())) + "\n")
			}
			b.WriteString(i18n.T("kaiser.promptPlayHelp") + "\n")
		case domain.KaiserPhaseHandEnd:
			if g.IsBidMade() {
				b.WriteString(i18n.T("kaiser.madeLine") + "\n")
			} else {
				b.WriteString(i18n.T("kaiser.setLine") + "\n")
			}
			b.WriteString(i18n.T("kaiser.promptHandEndHelp") + "\n")
		}
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *KaiserCuiPresenter) ActionLogOutput(g interfaces.KaiserGame) string {
	return actionLogOutputText(g)
}
