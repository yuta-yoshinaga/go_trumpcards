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

// CinchCuiPresenter renders the Cinch CUI view.
type CinchCuiPresenter struct{}

// cinchBidLabel はビッド値の表示 (-1=未ビッド, 0=pass, それ以外は数値)。
func cinchBidLabel(bid int) string {
	switch {
	case bid < 0:
		return "-"
	case bid == domain.CinchPassBid:
		return i18n.T("cinch.bidPass")
	default:
		return strconv.Itoa(bid)
	}
}

// cinchTrumpLabel は切り札スートの表示名を返す (0=未確定)。
func cinchTrumpLabel(suit int) string {
	if suit < domain.CardDesignSpade || suit > domain.CardDesignDiamond {
		return "-"
	}
	return suitNames[suit]
}

// cinchBidStrengthLine renders the human hand's held points per candidate trump
// suit. The Web GUI shows this table for the whole bid phase (#4845); the CUI
// printed only the current high bid, so the quantitative half of the bidding
// decision was missing.
func cinchBidStrengthLine(g interfaces.CinchGame) string {
	// 席順は固定でないので GetIsHuman で探す (ドメインの findHumanIdx と同じ)。
	var human *domain.CinchPlayer
	for i := range g.GetPlayerCnt() {
		if p := g.GetPlayer(i); p != nil && p.GetIsHuman() {
			human = p
			break
		}
	}
	if human == nil || human.GetCardsSize() == 0 {
		return ""
	}
	cards := make([]*domain.Card, 0, human.GetCardsSize())
	for i := range human.GetCardsSize() {
		cards = append(cards, human.GetCard(i))
	}
	// 計算はドメインに置いた。ここで書き直すと Left Pedro (同色スートの 5) の
	// 扱いが Web と食い違う。
	points := domain.CinchHandPointsBySuit(cards)
	best := domain.CinchBestTrumpSuit(points)
	parts := make([]string, 0, domain.CardDesignDiamond)
	for suit := domain.CardDesignSpade; suit <= domain.CardDesignDiamond; suit++ {
		parts = append(parts, cinchTrumpLabel(suit)+" "+strconv.Itoa(points[suit]))
	}
	return i18n.Tf("cinch.bidStrength",
		"points", strings.Join(parts, "  "),
		"best", cinchTrumpLabel(best),
		"max", strconv.Itoa(points[best]),
		"total", strconv.Itoa(domain.CinchTotalPoints)) + "\n"
}

func cinchPlayerStr(g interfaces.CinchGame, idx int) string {
	player := g.GetPlayer(idx)
	if player == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString(i18n.Tf("cinch.playerLine",
		"name", cuiPlayerName(player, idx),
		"hand", strconv.Itoa(player.GetCardsSize()),
		"bid", cinchBidLabel(player.GetBid()),
		"tricks", strconv.Itoa(player.GetTrickCount()),
		"total", strconv.Itoa(player.GetTotalScore())) + "\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// Output renders the current game state for the active locale.
func (p *CinchCuiPresenter) Output(g interfaces.CinchGame, lastErr error) string {
	return buildCuiOutput(i18n.T("cinch.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("cinch.dealLine",
			"deal", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber()),
			"bid", strconv.Itoa(g.GetCurrentBid()),
			"trump", cinchTrumpLabel(g.GetTrumpSuit())) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(cinchPlayerStr(g, i))
		}
		b.WriteString("----------\n")

		cuiTrickBlock(b, g.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			b.WriteString(i18n.T("cinch.gameEnd") + "\n")
			for i := 0; i < g.GetPlayerCnt(); i++ {
				player := g.GetPlayer(i)
				if player == nil {
					continue
				}
				b.WriteString(i18n.Tf("cinch.scoreEntry",
					"name", cuiPlayerName(player, i),
					"score", strconv.Itoa(player.GetTotalScore())) + "\n")
			}
			return
		}

		switch g.GetPhase() {
		case domain.CinchPhaseBid:
			bidderIdx := g.GetBidPlayerIdx()
			b.WriteString(i18n.Tf("cinch.promptBid",
				"name", cuiPlayerName(g.GetPlayer(bidderIdx), bidderIdx),
				"bid", strconv.Itoa(g.GetCurrentBid())) + "\n")
			b.WriteString(cinchBidStrengthLine(g))
		case domain.CinchPhaseNameTrump:
			winnerIdx := g.GetBidWinnerIdx()
			b.WriteString(i18n.Tf("cinch.promptTrump",
				"name", cuiPlayerName(g.GetPlayer(winnerIdx), winnerIdx)) + "\n")
		case domain.CinchPhasePlay:
			currentIdx := g.GetCurrentTurn()
			b.WriteString(i18n.Tf("cinch.promptPlay",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
		case domain.CinchPhaseTrickEnd:
			b.WriteString(i18n.T("cinch.promptTrickEnd") + "\n")
		case domain.CinchPhaseRoundEnd:
			b.WriteString(i18n.T("cinch.promptRoundEnd") + "\n")
		}
		b.WriteString(i18n.T("cinch.promptHelp") + "\n")
	})
}

// HintOutput emits the current Cinch hint.
func (p *CinchCuiPresenter) HintOutput(g interfaces.CinchGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("cinch.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, cinchHintReasonKeys)
	if hint.Bid != nil {
		return color.Yellow(i18n.Tf("cinch.hintBid",
			"bid", cinchBidLabel(*hint.Bid),
			"reason", reason)) + "\n"
	}
	if hint.TrumpSuit != nil {
		return color.Yellow(i18n.Tf("cinch.hintTrump",
			"trump", cinchTrumpLabel(*hint.TrumpSuit),
			"reason", reason)) + "\n"
	}
	if len(hint.CardIndices) > 0 {
		playerIdx := g.GetCurrentTurn()
		player := g.GetPlayer(playerIdx)
		cards := make([]string, len(hint.CardIndices))
		for i, idx := range hint.CardIndices {
			if player != nil && idx >= 0 && idx < player.GetCardsSize() {
				cards[i] = "[" + strconv.Itoa(idx) + "]" + cuiCardStr(player.GetCard(idx))
			} else {
				cards[i] = strconv.Itoa(idx)
			}
		}
		return color.Yellow(i18n.Tf("cinch.hintCard",
			"cards", strings.Join(cards, ", "),
			"reason", reason)) + "\n"
	}
	return color.Yellow(i18n.Tf("cinch.hintCard", "cards", "-", "reason", reason)) + "\n"
}

// cinchHintReasonKeys maps Cinch-specific hint-reason identifiers to i18n keys.
var cinchHintReasonKeys = map[string]string{
	"bid_pass":    "cinch.hintReasonBidPass",
	"bid_strong":  "cinch.hintReasonBidStrong",
	"name_trump":  "cinch.hintReasonNameTrump",
	"lead_strong": "cinch.hintReasonLeadStrong",
	"trump_cut":   "cinch.hintReasonTrumpCut",
	"follow_suit": "cinch.hintReasonFollowSuit",
	"discard_low": "cinch.hintReasonDiscardLow",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *CinchCuiPresenter) ActionLogOutput(g interfaces.CinchGame) string {
	return actionLogOutputTextForSeats[*domain.CinchPlayer](g)
}
