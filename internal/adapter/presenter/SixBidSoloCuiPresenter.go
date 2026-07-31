//go:build !js || !wasm || extra2

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// sixBidSoloReveal は伏せ札を公開する局面かを返す。
func sixBidSoloReveal(g interfaces.SixBidSoloGame) bool {
	phase := g.GetPhase()
	return phase == domain.SixBidSoloPhaseHandEnd || phase == domain.SixBidSoloPhaseGameEnd
}

// sixBidSoloBidLabel はビッドの表示名を返す。
func sixBidSoloBidLabel(k domain.SixBidSoloBidKind) string {
	switch k {
	case domain.SixBidSoloBidSolo:
		return i18n.T("sixbidsolo.bid.solo")
	case domain.SixBidSoloBidHeartSolo:
		return i18n.T("sixbidsolo.bid.heartSolo")
	case domain.SixBidSoloBidMisere:
		return i18n.T("sixbidsolo.bid.misere")
	case domain.SixBidSoloBidGuarantee:
		return i18n.T("sixbidsolo.bid.guarantee")
	case domain.SixBidSoloBidSpreadMisere:
		return i18n.T("sixbidsolo.bid.spreadMisere")
	case domain.SixBidSoloBidCall:
		return i18n.T("sixbidsolo.bid.callSolo")
	}
	return i18n.T("sixbidsolo.bid.pass")
}

// sixBidSoloSuitLabel はスートの表示名を返す。
func sixBidSoloSuitLabel(suit int) string {
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
	return i18n.T("sixbidsolo.noTrump")
}

// sixBidSoloPlayerStr returns the display string for a single seat.
func sixBidSoloPlayerStr(g interfaces.SixBidSoloGame, i int) string {
	player := g.GetPlayer(i)
	if player == nil {
		return ""
	}
	hand := i18n.Tf("sixbidsolo.hiddenHand", "count", strconv.Itoa(player.GetCardsSize()))
	// **スプレッド・ミゼールでは宣言者の手札が公開される。**それが賭けの中身。
	spread := g.IsSpreadOpen() && i == g.GetDeclarerIdx()
	if player.GetIsHuman() || sixBidSoloReveal(g) || spread {
		var b strings.Builder
		for j := range player.GetCardsSize() {
			b.WriteString("[" + strconv.Itoa(j) + "]" + cuiCardStr(player.GetCard(j)) + " ")
		}
		hand = strings.TrimSpace(b.String())
	}
	role := ""
	if i == g.GetDealerIdx() {
		role += " " + i18n.T("sixbidsolo.dealerTag")
	}
	if i == g.GetDeclarerIdx() {
		role += " " + i18n.T("sixbidsolo.declarerTag")
	}
	return i18n.Tf("sixbidsolo.playerLine",
		"name", cuiPlayerName(player, i),
		"role", role,
		"points", strconv.Itoa(g.GetPoints(i)),
		"tricks", strconv.Itoa(g.GetTricksWon(i)),
		"score", strconv.Itoa(g.GetScore(i)),
		"hand", hand) + "\n"
}

// SixBidSoloCuiPresenter renders the Six-Bid Solo CUI view.
type SixBidSoloCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *SixBidSoloCuiPresenter) Output(g interfaces.SixBidSoloGame, lastErr error) string {
	return buildCuiOutput(i18n.T("sixbidsolo.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("sixbidsolo.header",
			"hand", strconv.Itoa(g.GetHandNumber()),
			"target", strconv.Itoa(g.GetConfig().TargetHands),
			"total", strconv.Itoa(domain.SixBidSoloTotalPoints)) + "\n")

		if hb := g.GetHighBid(); hb != nil {
			trump := i18n.T("sixbidsolo.trumpUndecided")
			if g.IsDeclared() {
				trump = sixBidSoloSuitLabel(g.GetTrumpSuit())
			}
			b.WriteString(i18n.Tf("sixbidsolo.contractLine",
				"bid", sixBidSoloBidLabel(hb.Kind),
				"trump", trump,
				"need", strconv.Itoa(domain.SixBidSoloTargetPoints(hb.Kind, g.GetTrumpSuit()))) + "\n")
			if c := g.GetCalledCard(); c != nil {
				b.WriteString(i18n.Tf("sixbidsolo.calledLine", "card", cuiCardStr(c)) + "\n")
			}
		}

		// **ウィドウは精算まで伏せたまま。**枚数だけ見せる。
		widow := i18n.Tf("sixbidsolo.widowHidden", "count", strconv.Itoa(len(g.GetWidow())))
		if sixBidSoloReveal(g) {
			var w strings.Builder
			for _, c := range g.GetWidow() {
				w.WriteString(cuiCardStr(c) + " ")
			}
			widow = strings.TrimSpace(w.String()) + " (" +
				strconv.Itoa(g.SixBidSoloWidowPoints()) + "pt)"
		}
		b.WriteString(i18n.Tf("sixbidsolo.widowLine", "widow", widow) + "\n")

		for i := range g.GetPlayers() {
			b.WriteString(sixBidSoloPlayerStr(g, i))
		}

		if trick := g.GetTrick(); len(trick) > 0 {
			var t strings.Builder
			for _, c := range trick {
				t.WriteString(cuiCardStr(c) + " ")
			}
			b.WriteString(i18n.Tf("sixbidsolo.trick", "cards", strings.TrimSpace(t.String())) + "\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			banner := i18n.Tf("sixbidsolo.gameEnd", "seat", strconv.Itoa(g.GetWinnerIdx()))
			b.WriteString(color.Green(banner) + "\n")
			return
		}

		switch g.GetPhase() {
		case domain.SixBidSoloPhaseBid:
			idx := g.GetBidPlayerIdx()
			b.WriteString(i18n.Tf("sixbidsolo.promptBid", "name", cuiPlayerName(g.GetPlayer(idx), idx)) + "\n")
			b.WriteString(sixBidSoloLadderLine(g) + "\n")
			b.WriteString(i18n.T("sixbidsolo.promptBidHelp") + "\n")
		case domain.SixBidSoloPhaseDeclare:
			idx := g.GetDeclarerIdx()
			b.WriteString(i18n.Tf("sixbidsolo.promptDeclare", "name", cuiPlayerName(g.GetPlayer(idx), idx)) + "\n")
			b.WriteString(i18n.T("sixbidsolo.promptDeclareHelp") + "\n")
		case domain.SixBidSoloPhasePlay:
			idx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("sixbidsolo.promptPlay", "name", cuiPlayerName(g.GetPlayer(idx), idx)) + "\n")
			if g.IsHumanTurn() {
				var v strings.Builder
				for _, i := range g.SixBidSoloValidPlays(idx) {
					v.WriteString(strconv.Itoa(i) + " ")
				}
				b.WriteString(i18n.Tf("sixbidsolo.playable", "indexes", strings.TrimSpace(v.String())) + "\n")
			}
			b.WriteString(i18n.T("sixbidsolo.promptPlayHelp") + "\n")
		case domain.SixBidSoloPhaseHandEnd:
			if r := g.GetLastResult(); r != nil {
				key := "sixbidsolo.setLine"
				if r.Made {
					key = "sixbidsolo.madeLine"
				}
				b.WriteString(i18n.Tf(key,
					"bid", sixBidSoloBidLabel(r.Kind),
					"points", strconv.Itoa(r.DeclarerPoints),
					"need", strconv.Itoa(r.Target)) + "\n")
				// **ウィドウは宣言者に入る。ミゼール系だけは 0。**
				b.WriteString(i18n.Tf("sixbidsolo.widowCredit", "n", strconv.Itoa(r.WidowPoints)) + "\n")
				b.WriteString(i18n.Tf("sixbidsolo.valueLine", "n", strconv.Itoa(r.Value)) + "\n")
			}
			b.WriteString(i18n.T("sixbidsolo.promptHandEndHelp") + "\n")
		}
	})
}

// sixBidSoloLadderLine は 6 段階のビッドを 1 行で示す。
//
// **目標点も一緒に出す。**ギャランティーがスートで変わることは表でしか読めない。
func sixBidSoloLadderLine(g interfaces.SixBidSoloGame) string {
	var b strings.Builder
	b.WriteString(i18n.T("sixbidsolo.ladderTitle") + " ")
	for k := domain.SixBidSoloMinBid; k <= domain.SixBidSoloMaxBid; k++ {
		if k > domain.SixBidSoloMinBid {
			b.WriteString(" < ")
		}
		b.WriteString(strconv.Itoa(int(k)) + ":" + sixBidSoloBidLabel(k))
	}
	return b.String()
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *SixBidSoloCuiPresenter) ActionLogOutput(g interfaces.SixBidSoloGame) string {
	return actionLogOutputText(g)
}
