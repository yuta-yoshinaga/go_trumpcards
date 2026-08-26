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

// gleekSuitLabel maps a suit value to its i18n label key.
func gleekSuitLabel(suit int) string {
	switch suit {
	case domain.CardDesignSpade:
		return i18n.T("gleek.suitSpade")
	case domain.CardDesignClover:
		return i18n.T("gleek.suitClub")
	case domain.CardDesignHeart:
		return i18n.T("gleek.suitHeart")
	case domain.CardDesignDiamond:
		return i18n.T("gleek.suitDiamond")
	default:
		return i18n.T("gleek.suitNone")
	}
}

// gleekRankLabel maps a meld rank to its i18n label key.
func gleekRankLabel(rank int) string {
	switch rank {
	case 1:
		return i18n.T("gleek.rankAce")
	case 13:
		return i18n.T("gleek.rankKing")
	case 12:
		return i18n.T("gleek.rankQueen")
	case 11:
		return i18n.T("gleek.rankJack")
	default:
		return "-"
	}
}

// gleekPlayerStr renders one seat's line.
func gleekPlayerStr(g interfaces.GleekGame, idx int) string {
	player := g.GetPlayer(idx)
	if player == nil {
		return ""
	}
	scores := g.GetPlayerScores()
	trickPoints := g.GetTrickPoints()
	role := i18n.T("gleek.roleDefender")
	if idx == g.GetBuyerIdx() {
		role = i18n.T("gleek.roleBuyer")
	}
	var b strings.Builder
	b.WriteString(i18n.Tf("gleek.playerLine",
		"name", cuiPlayerName(player, idx),
		"role", role,
		"cards", strconv.Itoa(player.GetCardsSize()),
		"score", strconv.Itoa(scores[idx]),
		"tricks", strconv.Itoa(player.GetTrickCount()),
		"points", strconv.Itoa(trickPoints[idx])) + "\n")
	if player.GetIsHuman() {
		b.WriteString(gleekIndexedHandStr(player, g.GetTrumpSuit()) + "\n")
	}
	return b.String()
}

// gleekIndexedHandStr は手札をインデックス付きで並べ、切り札の名札に点を注記する。
//
// **名札は切り札スートでしか点にならない。** 一覧に出さないと、15 点の Tib を
// 相手のトリックに放り込んだ理由が最後まで分からない。
func gleekIndexedHandStr(player cuiCardList, trump int) string {
	return formatCardList(player, func(c *domain.Card) string {
		s := cuiCardStr(c)
		if v := domain.GleekHonourValueForTest(c, trump); v > 0 {
			s += "(" + strconv.Itoa(v) + ")"
		}
		return s
	}, "  ", true)
}

// GleekCuiPresenter renders the Gleek CUI view.
type GleekCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *GleekCuiPresenter) Output(g interfaces.GleekGame, lastErr error) string {
	return buildCuiOutput(i18n.T("gleek.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("gleek.round",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber()),
			"trump", gleekSuitLabel(g.GetTrumpSuit())) + "\n")
		b.WriteString(gleekStockLine(g))
		b.WriteString(gleekStageLine(g))

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(gleekPlayerStr(g, i))
		}

		b.WriteString("----------\n")

		cuiTrickBlock(b, g.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winner := g.GetWinnerPlayer()
			var winnerStr string
			if winner >= 0 {
				winnerStr = cuiPlayerName(g.GetPlayer(winner), winner)
			}
			b.WriteString(color.Green(i18n.Tf("gleek.gameEnd", "name", winnerStr)) + "\n")
			return
		}
		p.writePrompt(b, g)
	})
}

// writePrompt フェーズごとの案内を書く。
func (p *GleekCuiPresenter) writePrompt(b *strings.Builder, g interfaces.GleekGame) {
	switch g.GetPhase() {
	case domain.GleekPhaseBid:
		bidderIdx := g.GetCurrentBidderIdx()
		b.WriteString(i18n.Tf("gleek.promptBid",
			"bid", strconv.Itoa(g.HighestBid()),
			"name", cuiPlayerName(g.GetPlayer(bidderIdx), bidderIdx)) + "\n")
		// **置ける額だけを案内する。** 上限に達した卓で「競り上げる」と言うと、
		// その通り打った人間だけが弾かれる。
		if next := g.NextBidAmount(); next > 0 {
			b.WriteString(i18n.Tf("gleek.promptBidNext", "next", strconv.Itoa(next)) + "\n")
		} else {
			b.WriteString(i18n.T("gleek.promptBidCapped") + "\n")
		}
		b.WriteString(i18n.T("gleek.promptBidHelp") + "\n")
	case domain.GleekPhaseDiscard:
		b.WriteString(i18n.Tf("gleek.promptDiscard",
			"name", cuiPlayerName(g.GetPlayer(g.GetBuyerIdx()), g.GetBuyerIdx()),
			"count", strconv.Itoa(domain.GleekSwapSize)) + "\n")
		b.WriteString(i18n.T("gleek.promptDiscardHelp") + "\n")
	case domain.GleekPhasePlay:
		currentIdx := g.GetCurrentPlayerIdx()
		b.WriteString(i18n.Tf("gleek.promptPlay",
			"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx),
			"trump", gleekSuitLabel(g.GetTrumpSuit())) + "\n")
		b.WriteString(i18n.T("gleek.promptPlayHelp") + "\n")
	case domain.GleekPhaseTrickEnd:
		b.WriteString(i18n.T("gleek.promptTrickEnd") + "\n")
		b.WriteString(i18n.T("gleek.promptTrickEndHelp") + "\n")
	case domain.GleekPhaseRoundEnd:
		b.WriteString(i18n.Tf("gleek.promptRoundEnd",
			"total", strconv.Itoa(g.DealPoints()),
			"par", strconv.Itoa(g.Par())) + "\n")
		b.WriteString(i18n.T("gleek.promptRoundEndHelp") + "\n")
	}
}

// gleekStockLine 表向きの札と落札の行を組み立てる。
func gleekStockLine(g interfaces.GleekGame) string {
	turnUp := "-"
	if c := g.GetTurnUp(); c != nil {
		turnUp = cuiCardStr(c)
	}
	if g.GetBuyerIdx() < 0 {
		return i18n.Tf("gleek.stockUnsold", "turnup", turnUp) + "\n"
	}
	return i18n.Tf("gleek.stockSold",
		"turnup", turnUp,
		"name", cuiPlayerName(g.GetPlayer(g.GetBuyerIdx()), g.GetBuyerIdx()),
		"bid", strconv.Itoa(g.GetWinningBid())) + "\n"
}

// gleekStageLine ラフとメルドの結果行を組み立てる。
//
// **段階の点は盤に出さないと見えない。** 競りとトリックの間で動いた点を出さないと、
// 累積点だけが理由なく増減しているように見える。
func gleekStageLine(g interfaces.GleekGame) string {
	if g.GetRuffWinnerIdx() < 0 && len(g.GetMelds()) == 0 {
		return ""
	}
	var b strings.Builder
	if idx := g.GetRuffWinnerIdx(); idx >= 0 {
		total := 0
		suit := -1
		for _, r := range g.GetRuffs() {
			if r != nil && r.PlayerIdx == idx {
				total, suit = r.Total, r.Suit
			}
		}
		b.WriteString(i18n.Tf("gleek.ruffLine",
			"name", cuiPlayerName(g.GetPlayer(idx), idx),
			"total", strconv.Itoa(total),
			"suit", gleekSuitLabel(suit)) + "\n")
	}
	for _, m := range g.GetMelds() {
		if m == nil {
			continue
		}
		key := "gleek.meldGleek"
		if m.Count >= 4 {
			key = "gleek.meldMournival"
		}
		b.WriteString(i18n.Tf(key,
			"name", cuiPlayerName(g.GetPlayer(m.PlayerIdx), m.PlayerIdx),
			"rank", gleekRankLabel(m.Rank),
			"value", strconv.Itoa(m.Value)) + "\n")
	}
	return b.String()
}

// HintOutput emits the current Gleek hint.
func (p *GleekCuiPresenter) HintOutput(g interfaces.GleekGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("gleek.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, gleekHintReasonKeys)
	// **競りのヒントは札でなく額を指す。** 索引を出しても意味が無い。
	if hint.Reason == "bid_raise" || hint.Reason == "bid_pass" {
		action := i18n.T("gleek.hintActionPass")
		if hint.Bid > 0 {
			action = i18n.Tf("gleek.hintActionRaise", "amount", strconv.Itoa(hint.Bid))
		}
		return color.Yellow(i18n.Tf("gleek.hintDecision", "action", action, "reason", reason)) + "\n"
	}
	if len(hint.CardIndices) > 0 {
		playerIdx := g.GetCurrentPlayerIdx()
		if hint.Reason == "discard_stock" {
			playerIdx = g.GetBuyerIdx()
		}
		player := g.GetPlayer(playerIdx)
		cards := make([]string, len(hint.CardIndices))
		for i, idx := range hint.CardIndices {
			if player != nil && idx >= 0 && idx < player.GetCardsSize() {
				cards[i] = "[" + strconv.Itoa(idx) + "]" + cuiCardStr(player.GetCard(idx))
			} else {
				cards[i] = strconv.Itoa(idx)
			}
		}
		return color.Yellow(i18n.Tf("gleek.hintCard",
			"cards", strings.Join(cards, ", "),
			"reason", reason)) + "\n"
	}
	return color.Yellow(i18n.Tf("gleek.hintCard", "cards", "-", "reason", reason)) + "\n"
}

// gleekHintReasonKeys maps Gleek-specific hint-reason identifiers to i18n keys.
var gleekHintReasonKeys = map[string]string{
	"bid_raise":      "gleek.hintReasonBidRaise",
	"bid_pass":       "gleek.hintReasonBidPass",
	"discard_stock":  "gleek.hintReasonDiscardStock",
	"lead_high":      "gleek.hintReasonLeadHigh",
	"follow_win":     "gleek.hintReasonFollowWin",
	"follow_duck":    "gleek.hintReasonFollowDuck",
	"discard_honour": "gleek.hintReasonDiscardHonour",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *GleekCuiPresenter) ActionLogOutput(g interfaces.GleekGame) string {
	return actionLogOutputTextForSeats[*domain.GleekPlayer](g)
}
