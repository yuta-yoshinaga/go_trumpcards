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

// calabresellaBidLabel maps a bid value to its i18n label key.
func calabresellaBidLabel(bid domain.CalabresellaBid) string {
	switch bid {
	case domain.CalabresellaBidChiamo:
		return i18n.T("calabresella.bidChiamo")
	case domain.CalabresellaBidSolo:
		return i18n.T("calabresella.bidSolo")
	default:
		return i18n.T("calabresella.bidNone")
	}
}

func calabresellaPlayerStr(g interfaces.CalabresellaGame, idx int) string {
	player := g.GetPlayer(idx)
	if player == nil {
		return ""
	}
	scores := g.GetPlayerScores()
	role := i18n.T("calabresella.roleCoalition")
	if idx == g.GetSoloistIdx() {
		role = i18n.T("calabresella.roleSoloist")
	}
	var b strings.Builder
	b.WriteString(i18n.Tf("calabresella.playerLine",
		"name", cuiPlayerName(player, idx),
		"role", role,
		"cards", strconv.Itoa(player.GetCardsSize()),
		"score", strconv.Itoa(scores[idx]),
		"tricks", strconv.Itoa(player.GetTrickCount()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// calabresellaMonteCards renders the revealed monte, or "" before it is taken.
//
// **モンテは取得した時点で全員に公開される情報。**ドメインは取得時に g.monte を
// nil にするので、Web の buildMonteOutput と同じ出どころ — 棋譜の最後の
// "monte_take" エントリ — から読む。両者がずれない唯一の読み方 (#4843)。
func calabresellaMonteCards(g interfaces.CalabresellaGame) string {
	if g.GetPhase() == domain.CalabresellaPhaseBid {
		return ""
	}
	log := g.GetActionLog()
	for i := len(log) - 1; i >= 0; i-- {
		entry := log[i]
		if entry == nil || entry.ActionType != "monte_take" {
			continue
		}
		return cuiCardSliceStr(entry.Cards)
	}
	return ""
}

// CalabresellaCuiPresenter renders the Calabresella CUI view.
type CalabresellaCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *CalabresellaCuiPresenter) Output(g interfaces.CalabresellaGame, lastErr error) string {
	return buildCuiOutput(i18n.T("calabresella.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("calabresella.round",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber()),
			"bid", calabresellaBidLabel(g.GetWinningBid())) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(calabresellaPlayerStr(g, i))
		}

		// **モンテは公開情報。**中身の出どころは calabresellaMonteCards を参照。
		if monte := calabresellaMonteCards(g); monte != "" {
			b.WriteString(i18n.Tf("calabresella.monteLine", "cards", monte) + "\n")
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
			banner := i18n.Tf("calabresella.gameEnd", "name", winnerStr)
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.CalabresellaPhaseBid:
			bidderIdx := g.GetCurrentBidderIdx()
			b.WriteString(i18n.Tf("calabresella.promptBid",
				"bid", calabresellaBidLabel(g.GetWinningBid()),
				"name", cuiPlayerName(g.GetPlayer(bidderIdx), bidderIdx)) + "\n")
			b.WriteString(i18n.T("calabresella.promptBidHelp") + "\n")
		case domain.CalabresellaPhaseDiscard:
			soloistIdx := g.GetSoloistIdx()
			soloist := g.GetPlayer(soloistIdx)
			b.WriteString(i18n.Tf("calabresella.promptDiscard",
				"name", cuiPlayerName(soloist, soloistIdx)) + "\n")
			// 固定文言の「4枚」は捨てても減らないので、実際の手札枚数から残りを出す
			// (Web の discardRemaining と同じ式)。12 枚ちょうどになったら出さない。
			if soloist != nil {
				if remaining := soloist.GetCardsSize() - domain.CalabresellaHandSize; remaining > 0 {
					b.WriteString(i18n.Tf("calabresella.promptDiscardRemaining",
						"n", strconv.Itoa(remaining),
						"size", strconv.Itoa(soloist.GetCardsSize()),
						"target", strconv.Itoa(domain.CalabresellaHandSize)) + "\n")
				}
			}
			b.WriteString(i18n.T("calabresella.promptDiscardHelp") + "\n")
		case domain.CalabresellaPhasePlay:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("calabresella.promptPlay",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("calabresella.promptPlayHelp") + "\n")
		case domain.CalabresellaPhaseTrickEnd:
			b.WriteString(i18n.T("calabresella.promptTrickEnd") + "\n")
			b.WriteString(i18n.T("calabresella.promptTrickEndHelp") + "\n")
		case domain.CalabresellaPhaseRoundEnd:
			thirds := g.GetRoundThirds()
			soloist := g.GetSoloistIdx()
			soloistThirds := 0
			if soloist >= 0 && soloist < len(thirds) {
				soloistThirds = thirds[soloist]
			}
			b.WriteString(i18n.Tf("calabresella.promptRoundEnd",
				"soloist", cuiPlayerName(g.GetPlayer(soloist), soloist),
				"thirds", strconv.Itoa(soloistThirds)) + "\n")
			// Break down every player's thirds with their role: the 11 thirds are
			// split between the soloist and the two-player coalition, so the CUI
			// should show each share, not just the soloist's.
			for i := 0; i < g.GetPlayerCnt(); i++ {
				role := i18n.T("calabresella.roleCoalition")
				if i == soloist {
					role = i18n.T("calabresella.roleSoloist")
				}
				b.WriteString(i18n.Tf("calabresella.thirdsLine",
					"name", cuiPlayerName(g.GetPlayer(i), i),
					"role", role,
					"thirds", strconv.Itoa(thirds[i])) + "\n")
			}
			b.WriteString(i18n.T("calabresella.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current Calabresella hint.
func (p *CalabresellaCuiPresenter) HintOutput(g interfaces.CalabresellaGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("calabresella.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, calabresellaHintReasonKeys)
	if len(hint.CardIndices) > 0 {
		playerIdx := g.GetCurrentPlayerIdx()
		player := g.GetPlayer(playerIdx)
		cards := make([]string, len(hint.CardIndices))
		for i, idx := range hint.CardIndices {
			if player != nil && idx >= 0 && idx < player.GetCardsSize() {
				cards[i] = "[" + strconv.Itoa(idx) + "]" + cuiCardStr(player.GetCard(idx))
			} else {
				cards[i] = strconv.Itoa(idx)
			}
		}
		return color.Yellow(i18n.Tf("calabresella.hintCard",
			"cards", strings.Join(cards, ", "),
			"reason", reason)) + "\n"
	}
	return color.Yellow(i18n.Tf("calabresella.hintCard",
		"cards", "-",
		"reason", reason)) + "\n"
}

// calabresellaHintReasonKeys maps Calabresella-specific hint-reason identifiers to i18n keys.
var calabresellaHintReasonKeys = map[string]string{
	"lead_low":     "calabresella.hintReasonLeadLow",
	"follow_win":   "calabresella.hintReasonFollowWin",
	"follow_duck":  "calabresella.hintReasonFollowDuck",
	"give_partner": "calabresella.hintReasonGivePartner",
	"discard_low":  "calabresella.hintReasonDiscardLow",
	"bid_chiamo":   "calabresella.hintReasonBidChiamo",
	"bid_solo":     "calabresella.hintReasonBidSolo",
	"bid_pass":     "calabresella.hintReasonBidPass",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *CalabresellaCuiPresenter) ActionLogOutput(g interfaces.CalabresellaGame) string {
	return actionLogOutputTextWithNames(g, func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) })
}
