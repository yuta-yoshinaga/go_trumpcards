//go:build !js || !wasm || classic

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// karnoffelReveal は全員の手札を公開する局面かを返す。
func karnoffelReveal(g interfaces.KarnoffelGame) bool {
	phase := g.GetPhase()
	return phase == domain.KarnoffelPhaseHandEnd || phase == domain.KarnoffelPhaseGameEnd
}

// karnoffelSuitLabel はスートの表示名を返す。
func karnoffelSuitLabel(suit int) string {
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

// karnoffelPlayerStr returns the display string for a single seat.
func karnoffelPlayerStr(g interfaces.KarnoffelGame, i int) string {
	player := g.GetPlayer(i)
	if player == nil {
		return ""
	}
	hand := i18n.Tf("karnoffel.hiddenHand", "count", strconv.Itoa(player.GetCardsSize()))
	if player.GetIsHuman() || karnoffelReveal(g) {
		var b strings.Builder
		for j := range player.GetCardsSize() {
			// **どの札が法王かは毎局変わる。**選ばれたスートの称号札に名前を
			// 付ける。Web はカード下のバッジで出しているのに、CUI だけ
			// スートと数字から自力で照合させていた (#5732)。
			c := player.GetCard(j)
			title := ""
			if key := domain.KarnoffelTitleKey(c, g.GetChosenSuit()); key != "" {
				title = i18n.Tf("karnoffel.handTitle", "title", i18n.T("karnoffel.title."+key))
			}
			b.WriteString("[" + strconv.Itoa(j) + "]" + cuiCardStr(c) + title + " ")
		}
		hand = strings.TrimSpace(b.String())
	}
	role := ""
	if i == g.GetDealerIdx() {
		role += " " + i18n.T("karnoffel.dealerTag")
	}
	// **表向きの札は全員ぶん見える。**切札の根拠がここにある。
	up := "-"
	if c := g.GetUpCard(i); c != nil {
		up = cuiCardStr(c)
	}
	return i18n.Tf("karnoffel.playerLine",
		"name", cuiPlayerName(player, i),
		"team", strconv.Itoa(domain.KarnoffelTeamOf(i)),
		"role", role,
		"up", up,
		"tricks", strconv.Itoa(g.GetTricksWon(i)),
		"hand", hand) + "\n"
}

// KarnoffelCuiPresenter renders the Karnöffel CUI view.
type KarnoffelCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *KarnoffelCuiPresenter) Output(g interfaces.KarnoffelGame, lastErr error) string {
	return buildCuiOutput(i18n.T("karnoffel.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("karnoffel.header",
			"hand", strconv.Itoa(g.GetHandNumber()),
			"target", strconv.Itoa(g.GetConfig().TargetHands),
			"need", strconv.Itoa(domain.KarnoffelTricksToWin)) + "\n")
		b.WriteString(i18n.Tf("karnoffel.chosenLine",
			"suit", karnoffelSuitLabel(g.GetChosenSuit())) + "\n")
		b.WriteString(i18n.T("karnoffel.chosenNote") + "\n")
		b.WriteString(i18n.Tf("karnoffel.scoreLine",
			"t0", strconv.Itoa(g.GetHandsWon(0)),
			"t1", strconv.Itoa(g.GetHandsWon(1)),
			"k0", strconv.Itoa(g.KarnoffelTeamTricks(0)),
			"k1", strconv.Itoa(g.KarnoffelTeamTricks(1))) + "\n")

		for i := range g.GetPlayers() {
			b.WriteString(karnoffelPlayerStr(g, i))
		}

		if trick := g.GetTrick(); len(trick) > 0 {
			var t strings.Builder
			for _, c := range trick {
				t.WriteString(cuiCardStr(c) + " ")
			}
			b.WriteString(i18n.Tf("karnoffel.trick", "cards", strings.TrimSpace(t.String())) + "\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			banner := i18n.Tf("karnoffel.gameEnd", "team", strconv.Itoa(g.GetWinnerTeam()))
			b.WriteString(color.Green(banner) + "\n")
			return
		}

		switch g.GetPhase() {
		case domain.KarnoffelPhasePlay:
			idx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("karnoffel.promptPlay", "name", cuiPlayerName(g.GetPlayer(idx), idx)) + "\n")
			b.WriteString(karnoffelLadderLine() + "\n")
			if g.IsHumanTurn() {
				var v strings.Builder
				for _, i := range g.KarnoffelValidPlays(idx) {
					v.WriteString(strconv.Itoa(i) + " ")
				}
				b.WriteString(i18n.Tf("karnoffel.playable", "indexes", strings.TrimSpace(v.String())) + "\n")
			}
			b.WriteString(i18n.T("karnoffel.promptPlayHelp") + "\n")
		case domain.KarnoffelPhaseHandEnd:
			if r := g.GetLastResult(); r != nil {
				key := "karnoffel.handWonLine"
				if r.WinnerTeam < 0 {
					key = "karnoffel.handDrawnLine"
				}
				b.WriteString(i18n.Tf(key,
					"team", strconv.Itoa(r.WinnerTeam),
					"t0", strconv.Itoa(r.Tricks[0]),
					"t1", strconv.Itoa(r.Tricks[1])) + "\n")
			}
			b.WriteString(i18n.T("karnoffel.promptHandEndHelp") + "\n")
		}
	})
}

// karnoffelLadderLine は選ばれたスート内の序列を 1 行で示す。
//
// **悪魔だけ位置が特殊。**表で見せないと「なぜ負けたのか」が分からない。
func karnoffelLadderLine() string {
	return i18n.T("karnoffel.ladder")
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *KarnoffelCuiPresenter) ActionLogOutput(g interfaces.KarnoffelGame) string {
	return actionLogOutputText(g)
}
