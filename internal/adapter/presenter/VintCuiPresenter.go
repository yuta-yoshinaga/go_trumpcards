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

// vintReveal は全員の手札を公開する局面かを返す。
func vintReveal(g interfaces.VintGame) bool {
	phase := g.GetPhase()
	return phase == domain.VintPhaseHandEnd || phase == domain.VintPhaseGameEnd
}

// vintDenomName は宣言スートの表示名を返す。
func vintDenomName(denom int) string {
	switch denom {
	case domain.VintDenomSpade:
		return "♠"
	case domain.VintDenomClub:
		return "♣"
	case domain.VintDenomDiamond:
		return "♦"
	case domain.VintDenomHeart:
		return "♥"
	case domain.VintDenomNoTrump:
		return i18n.T("vint.noTrump")
	}
	return "-"
}

// vintPlayerStr returns the display string for a single Vint player.
func vintPlayerStr(g interfaces.VintGame, i int) string {
	player := g.GetPlayer(i)
	if player == nil {
		return ""
	}
	// **ダミーが無いので、プレイ中は誰の手札も見えない。**
	hand := i18n.Tf("vint.hiddenHand", "count", strconv.Itoa(player.GetCardsSize()))
	if player.GetIsHuman() || vintReveal(g) {
		var b strings.Builder
		for j := range player.GetCardsSize() {
			b.WriteString("[" + strconv.Itoa(j) + "]" + cuiCardStr(player.GetCard(j)) + " ")
		}
		hand = strings.TrimSpace(b.String())
	}
	role := ""
	if i == g.GetDealerIdx() {
		role += " " + i18n.T("vint.dealerTag")
	}
	if i == g.GetDeclarerIdx() {
		role += " " + i18n.T("vint.declarerTag")
	}
	return i18n.Tf("vint.playerLine",
		"name", cuiPlayerName(player, i),
		"team", strconv.Itoa(domain.VintTeamOf(i)),
		"role", role,
		"tricks", strconv.Itoa(g.GetTricksWon(i)),
		"hand", hand) + "\n"
}

// VintCuiPresenter renders the Vint CUI view.
type VintCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *VintCuiPresenter) Output(g interfaces.VintGame, lastErr error) string {
	return buildCuiOutput(i18n.T("vint.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("vint.header",
			"hand", strconv.Itoa(g.GetHandNumber()),
			"target", strconv.Itoa(domain.VintGameTarget)) + "\n")
		b.WriteString(i18n.Tf("vint.scoreLine",
			"b0", strconv.Itoa(g.GetBelow(0)), "a0", strconv.Itoa(g.GetAbove(0)), "g0", strconv.Itoa(g.GetGamesWon(0)),
			"b1", strconv.Itoa(g.GetBelow(1)), "a1", strconv.Itoa(g.GetAbove(1)), "g1", strconv.Itoa(g.GetGamesWon(1))) + "\n")

		if hb := g.GetHighBid(); hb != nil {
			b.WriteString(i18n.Tf("vint.contractLine",
				"level", strconv.Itoa(hb.Level),
				"denom", vintDenomName(hb.Denom),
				"value", strconv.Itoa(domain.VintTrickValue(hb.Denom, hb.Level))) + "\n")
		}

		for i := range g.GetPlayers() {
			b.WriteString(vintPlayerStr(g, i))
		}

		if trick := g.GetTrick(); len(trick) > 0 {
			var t strings.Builder
			for _, c := range trick {
				t.WriteString(cuiCardStr(c) + " ")
			}
			b.WriteString(i18n.Tf("vint.trick", "cards", strings.TrimSpace(t.String())) + "\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			banner := i18n.Tf("vint.gameEnd", "team", strconv.Itoa(g.GetWinnerTeam()))
			b.WriteString(color.Green(banner) + "\n")
			return
		}

		switch g.GetPhase() {
		case domain.VintPhaseBid:
			idx := g.GetBidPlayerIdx()
			b.WriteString(i18n.Tf("vint.promptBid", "name", cuiPlayerName(g.GetPlayer(idx), idx)) + "\n")
			b.WriteString(vintLadderLine() + "\n")
			b.WriteString(i18n.T("vint.promptBidHelp") + "\n")
		case domain.VintPhasePlay:
			idx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("vint.promptPlay", "name", cuiPlayerName(g.GetPlayer(idx), idx)) + "\n")
			if g.IsHumanTurn() {
				var v strings.Builder
				for _, i := range g.VintValidPlays(idx) {
					v.WriteString(strconv.Itoa(i) + " ")
				}
				b.WriteString(i18n.Tf("vint.playable", "indexes", strings.TrimSpace(v.String())) + "\n")
			}
			b.WriteString(i18n.T("vint.promptPlayHelp") + "\n")
		case domain.VintPhaseHandEnd:
			if r := g.GetLastResult(); r != nil {
				key := "vint.setLine"
				if r.Made {
					key = "vint.madeLine"
				}
				b.WriteString(i18n.Tf(key, "tricks", strconv.Itoa(r.DeclarerTricks)) + "\n")
				// **両チームが線下に得点する。**守備側の分も出す。
				b.WriteString(i18n.Tf("vint.trickPointsLine",
					"t0", strconv.Itoa(r.TrickPoints[0]),
					"t1", strconv.Itoa(r.TrickPoints[1])) + "\n")
				// **線上の点も局の得点。**オナー・エースは domain が既に持って
				// いるのに CUI は出しておらず、自分の名誉札点が分からなかった。
				b.WriteString(i18n.Tf("vint.honourLine",
					"t0", strconv.Itoa(r.HonourPoints[0]),
					"t1", strconv.Itoa(r.HonourPoints[1])) + "\n")
				b.WriteString(i18n.Tf("vint.aceLine",
					"t0", strconv.Itoa(r.AcePoints[0]),
					"t1", strconv.Itoa(r.AcePoints[1])) + "\n")
				// ペナルティは発生した局だけ (Web と同じ条件)。
				if r.Penalty[0] > 0 || r.Penalty[1] > 0 {
					b.WriteString(i18n.Tf("vint.penaltyLine",
						"t0", strconv.Itoa(r.Penalty[0]),
						"t1", strconv.Itoa(r.Penalty[1])) + "\n")
				}
			}
			b.WriteString(i18n.T("vint.promptHandEndHelp") + "\n")
		}
	})
}

// vintLadderLine は宣言スートの序列を 1 行で示す。
//
// **♠ が最弱で NT が最強。**ブリッジと逆なので必ず見せる。
func vintLadderLine() string {
	var b strings.Builder
	b.WriteString(i18n.T("vint.ladderTitle") + " ")
	for d := range domain.VintDenomCount {
		if d > 0 {
			b.WriteString(" < ")
		}
		b.WriteString(strconv.Itoa(d) + ":" + vintDenomName(d) +
			"(" + strconv.Itoa(domain.VintTrickValue(d, domain.VintMinLevel)) + ")")
	}
	return b.String()
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *VintCuiPresenter) ActionLogOutput(g interfaces.VintGame) string {
	return actionLogOutputText(g)
}
