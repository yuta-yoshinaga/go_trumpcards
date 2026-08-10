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

// bostonReveal は全員の手札を公開する局面かを返す。
func bostonReveal(g interfaces.BostonGame) bool {
	phase := g.GetPhase()
	return phase == domain.BostonPhaseHandEnd || phase == domain.BostonPhaseGameEnd
}

// bostonSuitName は切札スートの表示名を返す。
func bostonSuitName(suit int) string {
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
	return i18n.T("boston.noTrump")
}

// bostonBidLabel は宣言の表示名を返す。
func bostonBidLabel(level domain.BostonBidLevel) string {
	return i18n.T("boston.bid." + domain.BostonBidName(level))
}

// bostonPlayerStr returns the display string for a single Boston player.
func bostonPlayerStr(g interfaces.BostonGame, i int) string {
	player := g.GetPlayer(i)
	if player == nil {
		return ""
	}
	// **on the Table の宣言では落札者の手札が第1トリック後に見える。**
	exposedNow := g.IsExposed() && i == g.GetDeclarerIdx() && g.GetTrickNumber() >= 1
	hand := i18n.Tf("boston.hiddenHand", "count", strconv.Itoa(player.GetCardsSize()))
	if player.GetIsHuman() || bostonReveal(g) || exposedNow {
		var b strings.Builder
		for j := range player.GetCardsSize() {
			b.WriteString("[" + strconv.Itoa(j) + "]" + cuiCardStr(player.GetCard(j)) + " ")
		}
		hand = strings.TrimSpace(b.String())
	}
	role := ""
	if i == g.GetDealerIdx() {
		role += " " + i18n.T("boston.dealerTag")
	}
	if i == g.GetDeclarerIdx() {
		role += " " + i18n.T("boston.declarerTag")
	}
	if g.GetPartnerIdx() >= 0 && i == g.GetPartnerIdx() {
		role += " " + i18n.T("boston.partnerTag")
	}
	return i18n.Tf("boston.playerLine",
		"name", cuiPlayerName(player, i),
		"role", role,
		"tricks", strconv.Itoa(g.GetTricksWon(i)),
		"chips", strconv.Itoa(g.GetChips(i)),
		"hand", hand) + "\n"
}

// BostonCuiPresenter renders the Boston CUI view.
type BostonCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *BostonCuiPresenter) Output(g interfaces.BostonGame, lastErr error) string {
	return buildCuiOutput(i18n.T("boston.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("boston.header",
			"hand", strconv.Itoa(g.GetHandNumber()),
			"target", strconv.Itoa(g.GetTargetHands())) + "\n")

		if hb := g.GetHighBid(); hb != nil {
			b.WriteString(i18n.Tf("boston.contractLine",
				"bid", bostonBidLabel(hb.Level),
				"trump", bostonSuitName(g.GetTrumpSuit())) + "\n")
		}

		for i := range g.GetPlayers() {
			b.WriteString(bostonPlayerStr(g, i))
		}

		if trick := g.GetTrick(); len(trick) > 0 {
			var t strings.Builder
			for _, c := range trick {
				t.WriteString(cuiCardStr(c) + " ")
			}
			b.WriteString(i18n.Tf("boston.trick", "cards", strings.TrimSpace(t.String())) + "\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			banner := i18n.Tf("boston.gameEnd",
				"name", cuiPlayerName(g.GetPlayer(g.GetWinnerIdx()), g.GetWinnerIdx()))
			b.WriteString(color.Green(banner) + "\n")
			return
		}

		switch g.GetPhase() {
		case domain.BostonPhaseBid:
			idx := g.GetBidPlayerIdx()
			b.WriteString(i18n.Tf("boston.promptBid", "name", cuiPlayerName(g.GetPlayer(idx), idx)) + "\n")
			b.WriteString(bostonLadderLine() + "\n")
			b.WriteString(i18n.T("boston.promptBidHelp") + "\n")
		case domain.BostonPhaseCallPartner:
			b.WriteString(i18n.T("boston.promptPartner") + "\n")
			b.WriteString(i18n.T("boston.promptPartnerHelp") + "\n")
		case domain.BostonPhasePlay:
			idx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("boston.promptPlay", "name", cuiPlayerName(g.GetPlayer(idx), idx)) + "\n")
			if g.IsHumanTurn() {
				var v strings.Builder
				for _, i := range g.BostonValidPlays(idx) {
					v.WriteString(strconv.Itoa(i) + " ")
				}
				b.WriteString(i18n.Tf("boston.playable", "indexes", strings.TrimSpace(v.String())) + "\n")
			}
			b.WriteString(i18n.T("boston.promptPlayHelp") + "\n")
		case domain.BostonPhaseHandEnd:
			key := "boston.failedLine"
			if g.IsBidMade() {
				key = "boston.madeLine"
			}
			b.WriteString(i18n.Tf(key, "tricks", strconv.Itoa(g.BostonDeclarerTricks())) + "\n")
			b.WriteString(i18n.T("boston.promptHandEndHelp") + "\n")
		}
	})
}

// bostonLadderLine は序列を 1 行で示す。
//
// **ミゼールがトリック宣言の間に挟まる。**並びを見せないと競りの判断ができない。
func bostonLadderLine() string {
	var b strings.Builder
	b.WriteString(i18n.T("boston.ladderTitle") + " ")
	for l := domain.BostonBidFive; l < domain.BostonBidLevelCount; l++ {
		if l > domain.BostonBidFive {
			b.WriteString(" < ")
		}
		b.WriteString(strconv.Itoa(int(l)) + ":" + bostonBidLabel(l))
		// **どの段が手札を晒し、どの段で味方を呼べるかまで見せる。**段の名前
		// だけでは、自分の宣言が第 1 トリックのあとに手札を公開する羽目になるか
		// も、単独で戦うことになるかも分からないまま競らせることになる (#4939)。
		if domain.BostonBidIsExposed(l) {
			b.WriteString(i18n.T("boston.ladderExposedTag"))
		}
		if domain.BostonBidCanCallPartner(l) {
			b.WriteString(i18n.T("boston.ladderPartnerTag"))
		}
	}
	return b.String()
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *BostonCuiPresenter) ActionLogOutput(g interfaces.BostonGame) string {
	return actionLogOutputText(g)
}
