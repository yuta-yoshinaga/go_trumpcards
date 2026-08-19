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
			// 選べる値だけをドロップダウンに詰めている (#4899)。判定はドメインの
			// BidEuchreCanBid を通す — 規則を presenter に書き写すと、案内した額が
			// 拒否されることになる。
			if floor, ok := g.BidEuchreMinLegalBid(idx); ok {
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

// bidEuchreHumanIdx は人間の席を返す (居なければ -1)。
func bidEuchreHumanIdx(g interfaces.BidEuchreGame) int {
	for i, p := range g.GetPlayers() {
		if p != nil && p.GetIsHuman() {
			return i
		}
	}
	return -1
}

// bidEuchreLongestSuits は手札で最も長いスートを返す (同数なら複数)。
func bidEuchreLongestSuits(p *domain.BidEuchrePlayer) []int {
	counts := map[int]int{}
	for j := range p.GetCardsSize() {
		if c := p.GetCard(j); c != nil {
			counts[c.GetDesign()]++
		}
	}
	best := 0
	for _, n := range counts {
		if n > best {
			best = n
		}
	}
	var out []int
	for suit := domain.CardDesignSpade; suit <= domain.CardDesignDiamond; suit++ {
		if counts[suit] == best && best > 0 {
			out = append(out, suit)
		}
	}
	return out
}

// bidEuchreHintKey は局面ごとの助言キーと、添える手札インデックス／スートを返す。
//
// **Kaiser (#4938) と同じ構造。**入札 → 切札 → プレイの多段判断はどちらも同じで、
// CUI では出せる札も通せる宣言も自分で数えるしかないため。
func bidEuchreHintKey(g interfaces.BidEuchreGame) (key string, idxs []int, suits []int) {
	if g.GetGameEndFlag() {
		return "bideuchre.hintGameEnd", nil, nil
	}
	human := bidEuchreHumanIdx(g)
	if human < 0 {
		return "bideuchre.hintNone", nil, nil
	}
	p := g.GetPlayer(human)

	switch g.GetPhase() {
	case domain.BidEuchrePhaseBid:
		if g.GetBidPlayerIdx() != human {
			return "bideuchre.hintNotYourTurn", nil, nil
		}
		// 通せる最低額はドメインが知っている。通せないならパスしかない。
		if minBid, ok := g.BidEuchreMinLegalBid(human); ok {
			return i18n.Tf("bideuchre.hintBid", "bid", strconv.Itoa(minBid)), nil, nil
		}
		return "bideuchre.hintPass", nil, nil
	case domain.BidEuchrePhaseChooseTrump:
		if g.GetDeclarerIdx() != human {
			return "bideuchre.hintNotYourTurn", nil, nil
		}
		return "bideuchre.hintTrump", nil, bidEuchreLongestSuits(p)
	case domain.BidEuchrePhasePlay:
		if g.GetCurrentPlayerIdx() != human {
			return "bideuchre.hintNotYourTurn", nil, nil
		}
		plays := g.BidEuchreValidPlays(human)
		switch len(plays) {
		case 0:
			return "bideuchre.hintNone", nil, nil
		case 1:
			return "bideuchre.hintForced", plays, nil
		default:
			return "bideuchre.hintChoose", plays, nil
		}
	case domain.BidEuchrePhaseHandEnd:
		return "bideuchre.hintHandEnd", nil, nil
	}
	return "bideuchre.hintNone", nil, nil
}

// HintOutput emits the current Bid Euchre hint.
func (p *BidEuchreCuiPresenter) HintOutput(g interfaces.BidEuchreGame) string {
	key, idxs, suits := bidEuchreHintKey(g)
	if key == "" {
		key = "bideuchre.hintNone"
	}
	// bidEuchreHintKey は組み立て済みの文を返すことがある (最低宣言額入り)。
	msg := key
	if strings.HasPrefix(key, "bideuchre.") {
		msg = i18n.T(key)
	}
	parts := make([]string, 0, len(idxs)+len(suits))
	for _, v := range idxs {
		parts = append(parts, "["+strconv.Itoa(v)+"]")
	}
	for _, s := range suits {
		parts = append(parts, cuiSuitName(s))
	}
	if len(parts) > 0 {
		msg = i18n.Tf("bideuchre.hintWith", "hint", msg, "list", strings.Join(parts, ", "))
	}
	return color.Yellow(msg) + "\n"
}
