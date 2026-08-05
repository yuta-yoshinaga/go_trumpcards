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

// bidEuchreReveal は全員の手札を公開する局面かを返す。
func bidEuchreReveal(g interfaces.BidEuchreGame) bool {
	phase := g.GetPhase()
	return phase == domain.BidEuchrePhaseHandEnd || phase == domain.BidEuchrePhaseGameEnd
}

// bidEuchreTrumpName は宣言の表示名を返す。
//
// **ノートランプが 2 種類ある。**ハイとローで序列が逆になるので必ず区別する。
func bidEuchreTrumpName(t domain.BidEuchreTrump) string {
	switch t {
	case domain.BidEuchreTrumpSpade:
		return "♠"
	case domain.BidEuchreTrumpClub:
		return "♣"
	case domain.BidEuchreTrumpDiamond:
		return "♦"
	case domain.BidEuchreTrumpHeart:
		return "♥"
	case domain.BidEuchreTrumpNoHigh:
		return i18n.T("bideuchre.noTrumpHigh")
	case domain.BidEuchreTrumpNoLow:
		return i18n.T("bideuchre.noTrumpLow")
	}
	return "-"
}

// bidEuchrePlayerStr returns the display string for a single Bid Euchre player.
func bidEuchrePlayerStr(g interfaces.BidEuchreGame, i int) string {
	player := g.GetPlayer(i)
	if player == nil {
		return ""
	}
	// **キティが無く、誰の手札も伏せたまま。**
	hand := i18n.Tf("bideuchre.hiddenHand", "count", strconv.Itoa(player.GetCardsSize()))
	if player.GetIsHuman() || bidEuchreReveal(g) {
		var b strings.Builder
		for j := range player.GetCardsSize() {
			b.WriteString("[" + strconv.Itoa(j) + "]" + cuiCardStr(player.GetCard(j)) + " ")
		}
		hand = strings.TrimSpace(b.String())
	}
	role := ""
	if i == g.GetDealerIdx() {
		role += " " + i18n.T("bideuchre.dealerTag")
	}
	if i == g.GetDeclarerIdx() {
		role += " " + i18n.T("bideuchre.declarerTag")
	}
	return i18n.Tf("bideuchre.playerLine",
		"name", cuiPlayerName(player, i),
		"team", strconv.Itoa(domain.BidEuchreTeamOf(i)),
		"role", role,
		"tricks", strconv.Itoa(g.GetTricksWon(i)),
		"hand", hand) + "\n"
}

// BidEuchreCuiPresenter renders the Bid Euchre CUI view.
type BidEuchreCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *BidEuchreCuiPresenter) Output(g interfaces.BidEuchreGame, lastErr error) string {
	return buildCuiOutput(i18n.T("bideuchre.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("bideuchre.header",
			"hand", strconv.Itoa(g.GetHandNumber()),
			"target", strconv.Itoa(domain.BidEuchreGameTarget)) + "\n")
		b.WriteString(i18n.Tf("bideuchre.scoreLine",
			"s0", strconv.Itoa(g.GetScore(0)),
			"s1", strconv.Itoa(g.GetScore(1))) + "\n")

		if hb := g.GetHighBid(); hb != nil {
			trump := i18n.T("bideuchre.trumpUndecided")
			if g.IsTrumpChosen() {
				trump = bidEuchreTrumpName(g.GetTrump())
			}
			b.WriteString(i18n.Tf("bideuchre.contractLine",
				"value", strconv.Itoa(hb.Value), "trump", trump) + "\n")
		}

		for i := range g.GetPlayers() {
			b.WriteString(bidEuchrePlayerStr(g, i))
		}

		if trick := g.GetTrick(); len(trick) > 0 {
			var t strings.Builder
			for _, c := range trick {
				t.WriteString(cuiCardStr(c) + " ")
			}
			b.WriteString(i18n.Tf("bideuchre.trick", "cards", strings.TrimSpace(t.String())) + "\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			banner := i18n.Tf("bideuchre.gameEnd", "team", strconv.Itoa(g.GetWinnerTeam()))
			b.WriteString(color.Green(banner) + "\n")
			return
		}

		switch g.GetPhase() {
		case domain.BidEuchrePhaseBid:
			idx := g.GetBidPlayerIdx()
			b.WriteString(i18n.Tf("bideuchre.promptBid", "name", cuiPlayerName(g.GetPlayer(idx), idx)) + "\n")
			// **「立っている宣言＋1、親なら同額」を毎回暗算させない。**Web は
			// 選べる値だけをドロップダウンに詰めている (#4899)。
			floor := domain.BidEuchreMinBid
			if high := g.GetHighBid(); high != nil && high.Value > 0 {
				floor = high.Value + 1
				// **親だけは同額で奪える。**
				if idx == g.GetDealerIdx() {
					floor = high.Value
				}
			}
			if floor <= domain.BidEuchreMaxBid {
				b.WriteString(i18n.Tf("bideuchre.bidRange",
					"min", strconv.Itoa(floor),
					"max", strconv.Itoa(domain.BidEuchreMaxBid)) + "\n")
			} else {
				b.WriteString(i18n.T("bideuchre.bidRangeNone") + "\n")
			}
			b.WriteString(i18n.T("bideuchre.promptBidHelp") + "\n")
		case domain.BidEuchrePhaseChooseTrump:
			idx := g.GetDeclarerIdx()
			b.WriteString(i18n.Tf("bideuchre.promptTrump", "name", cuiPlayerName(g.GetPlayer(idx), idx)) + "\n")
			b.WriteString(bidEuchreTrumpMenuLine() + "\n")
			b.WriteString(i18n.T("bideuchre.promptTrumpHelp") + "\n")
		case domain.BidEuchrePhasePlay:
			idx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("bideuchre.promptPlay", "name", cuiPlayerName(g.GetPlayer(idx), idx)) + "\n")
			if g.IsHumanTurn() {
				var v strings.Builder
				for _, i := range g.BidEuchreValidPlays(idx) {
					v.WriteString(strconv.Itoa(i) + " ")
				}
				b.WriteString(i18n.Tf("bideuchre.playable", "indexes", strings.TrimSpace(v.String())) + "\n")
			}
			b.WriteString(i18n.T("bideuchre.promptPlayHelp") + "\n")
		case domain.BidEuchrePhaseHandEnd:
			if r := g.GetLastResult(); r != nil {
				key := "bideuchre.setLine"
				if r.Made {
					key = "bideuchre.madeLine"
				}
				b.WriteString(i18n.Tf(key,
					"bid", strconv.Itoa(r.Bid),
					"tricks", strconv.Itoa(r.Tricks[bidEuchreDeclaringTeam(g)])) + "\n")
				// **未達でも守備側は取ったトリックを得点する。**両チーム分を出す。
				b.WriteString(i18n.Tf("bideuchre.pointsLine",
					"p0", strconv.Itoa(r.Points[0]),
					"p1", strconv.Itoa(r.Points[1])) + "\n")
			}
			b.WriteString(i18n.T("bideuchre.promptHandEndHelp") + "\n")
		}
	})
}

// bidEuchreDeclaringTeam は落札側のチームを返す (落札前は 0)。
func bidEuchreDeclaringTeam(g interfaces.BidEuchreGame) int {
	team := domain.BidEuchreTeamOf(g.GetDeclarerIdx())
	if team < 0 {
		return 0
	}
	return team
}

// bidEuchreTrumpMenuLine は宣言できる切札の一覧を 1 行で示す。
func bidEuchreTrumpMenuLine() string {
	var b strings.Builder
	b.WriteString(i18n.T("bideuchre.trumpMenuTitle") + " ")
	for t := range int(domain.BidEuchreTrumpCount) {
		if t > 0 {
			b.WriteString(" / ")
		}
		b.WriteString(strconv.Itoa(t) + ":" + bidEuchreTrumpName(domain.BidEuchreTrump(t)))
	}
	return b.String()
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *BidEuchreCuiPresenter) ActionLogOutput(g interfaces.BidEuchreGame) string {
	return actionLogOutputText(g)
}
