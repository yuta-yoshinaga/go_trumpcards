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

// sjavsCardListStr は切札に印を付けて並べる。trumpSuit が -1 (未確定) なら
// 何も付けない。
//
// **6 枚の常時切札 (♣Q ♠Q ♣J ♠J ♥J ♦J) はスートを見ても分からない。**規則文は
// 出ているが、手札のどれがそれかは暗記に頼らせていた (#5575)。判定は合法手と
// 強さを決めている domain.SjavsIsTrump をそのまま呼ぶ ── 6 枚の一覧を書き写すと、
// 印と実際の強さがずれても誰も気づけない。
func sjavsCardListStr(cards []*domain.Card, indexed bool, trumpSuit int) string {
	if len(cards) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(cards))
	for i, c := range cards {
		str := cuiCardStr(c)
		if trumpSuit >= 0 && domain.SjavsIsTrump(c, trumpSuit) {
			str += color.Yellow(i18n.T("sjavs.trumpMark"))
		}
		if indexed {
			str = "[" + strconv.Itoa(i) + "]" + str
		}
		parts = append(parts, str)
	}
	return strings.Join(parts, " ")
}

// sjavsTrumpStr は切札を描く。ビッドが終わるまでは未確定。
func sjavsTrumpStr(suit int) string {
	if suit < 0 {
		return i18n.T("sjavs.trumpUndecided")
	}
	return cuiSuitName(suit)
}

// SjavsCuiPresenter renders the Sjavs CUI view.
type SjavsCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *SjavsCuiPresenter) Output(c interfaces.SjavsGame, lastErr error) string {
	return buildCuiOutput(i18n.T("sjavs.helpTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("sjavs.header",
			"trump", sjavsTrumpStr(c.GetTrumpSuit()),
			"a", strconv.Itoa(c.GetRemaining(0)),
			"b", strconv.Itoa(c.GetRemaining(1))) + "\n")
		// 常時切札 6 枚はこのゲームの骨格なので毎回出す。切札スートの札しか
		// 切札でないと思い込むと、♣Q が飛んでくる理由が分からない。
		sb.WriteString(i18n.T("sjavs.ruleLine") + "\n")
		if c.GetTrumpSuit() >= 0 {
			sb.WriteString(i18n.Tf("sjavs.trumpCountLine",
				"n", strconv.Itoa(domain.SjavsTrumpCount(c.GetTrumpSuit()))) + "\n")
		}
		sb.WriteString(i18n.Tf("sjavs.pointsLine",
			"a", strconv.Itoa(c.GetTeamPoints(0)),
			"b", strconv.Itoa(c.GetTeamPoints(1))) + "\n")

		if c.GetPhase() == domain.SjavsPhasePlay {
			trick := make([]*domain.Card, 0, len(c.GetTrick()))
			for _, tc := range c.GetTrick() {
				trick = append(trick, tc.Card)
			}
			sb.WriteString(i18n.Tf("sjavs.trickLine",
				"cards", sjavsCardListStr(trick, false, c.GetTrumpSuit())) + "\n")
		}

		bids := c.GetBids()
		for i, player := range c.GetPlayers() {
			line := i18n.Tf("sjavs.playerLine",
				"name", cuiPlayerName(player, i),
				"team", strconv.Itoa(domain.SjavsTeamOf(i)),
				"cards", strconv.Itoa(player.GetCardsSize()))
			if i < len(bids) && bids[i] > 0 {
				line += " " + i18n.Tf("sjavs.bidMark", "n", strconv.Itoa(bids[i]))
			}
			sb.WriteString(line + "\n")
			if player.GetIsHuman() && player.GetCardsSize() > 0 {
				hand := make([]*domain.Card, 0, player.GetCardsSize())
				for j := range player.GetCardsSize() {
					hand = append(hand, player.GetCard(j))
				}
				sb.WriteString("  " + sjavsCardListStr(hand, true, c.GetTrumpSuit()) + "\n")
			}
		}

		sb.WriteString("----------\n")
		cuiErrorBlock(sb, lastErr)
		sb.WriteString(p.promptBlock(c))
	})
}

// promptBlock はフェーズごとの案内と、終局・ハンド終了の表示を返す。
func (p *SjavsCuiPresenter) promptBlock(c interfaces.SjavsGame) string {
	if c.GetGameEndFlag() {
		key := "sjavs.gameEndLose"
		if c.GetWinnerTeam() == domain.SjavsTeamOf(0) {
			key = "sjavs.gameEndWin"
		}
		banner := i18n.T(key)
		if c.IsDoubleVictory() {
			banner += " " + i18n.T("sjavs.doubleVictory")
		}
		return color.Green(banner) + "\n"
	}

	var sb strings.Builder
	if hr := c.GetHandResult(); hr != nil && c.GetPhase() == domain.SjavsPhaseHandEnd {
		if hr.ScoringTeam < 0 {
			sb.WriteString(i18n.T("sjavs.handTie") + "\n")
		} else {
			sb.WriteString(i18n.Tf("sjavs.handScore",
				"team", strconv.Itoa(hr.ScoringTeam),
				"amount", strconv.Itoa(hr.Amount)) + "\n")
		}
		sb.WriteString(i18n.T("sjavs.promptNext") + "\n")
		return sb.String()
	}

	if c.GetPhase() == domain.SjavsPhaseBid {
		sb.WriteString(i18n.Tf("sjavs.promptBid",
			"min", strconv.Itoa(domain.SjavsMinBid),
			"longest", strconv.Itoa(c.LongestTrumpLength(0))) + "\n")
		return sb.String()
	}
	sb.WriteString(i18n.T("sjavs.promptPlay") + "\n")
	return sb.String()
}

// HintOutput emits the current Sjavs hint.
func (p *SjavsCuiPresenter) HintOutput(c interfaces.SjavsGame) string {
	hint := sjavsHint(c)
	key := sjavsHintReasonKeys[hint.Reason]
	if key == "" {
		key = "sjavs.hintNone"
	}
	switch {
	case hint.BidLength != nil:
		return color.Yellow(i18n.Tf("sjavs.hintBid",
			"n", strconv.Itoa(*hint.BidLength), "reason", i18n.T(key))) + "\n"
	case hint.CardIndex != nil:
		return color.Yellow(i18n.Tf("sjavs.hintPlay",
			"idx", strconv.Itoa(*hint.CardIndex), "reason", i18n.T(key))) + "\n"
	default:
		return color.Yellow(i18n.T(key)) + "\n"
	}
}

// sjavsHintReasonKeys maps the reason identifiers sjavsHint returns to i18n
// keys. The Web presenter ships the identifier and the frontend resolves it;
// the CUI must resolve it here or it prints the raw key.
var sjavsHintReasonKeys = map[string]string{
	"sjavs.hint.game_end":      "sjavs.hintReasonGameEnd",
	"sjavs.hint.not_your_turn": "sjavs.hintReasonNotYourTurn",
	"sjavs.hint.bid":           "sjavs.hintReasonBid",
	"sjavs.hint.pass":          "sjavs.hintReasonPass",
	"sjavs.hint.play":          "sjavs.hintReasonPlay",
	"sjavs.hint.none":          "sjavs.hintNone",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *SjavsCuiPresenter) ActionLogOutput(c interfaces.SjavsGame) string {
	return actionLogOutputTextForSeats[*domain.SjavsPlayer](c)
}
