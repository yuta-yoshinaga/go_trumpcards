//go:build !js || !wasm || solo

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// BidWhistCuiPresenter renders the Bid Whist CUI view.
type BidWhistCuiPresenter struct{}

// bidWhistHintReasonKeys maps Bid-Whist-specific hint reasons to i18n keys.
// Reasons not listed fall through to the shared cui_common layer.
var bidWhistHintReasonKeys = map[string]string{
	"pass_recommended": "bidwhist.hintReasonPass",
	"trump_longest":    "bidwhist.hintReasonTrumpLongest",
	"discard_weakest":  "bidwhist.hintReasonDiscardWeakest",
	"lead_trump":       "bidwhist.hintReasonLeadTrump",
	"discard_weak":     "bidwhist.hintReasonDiscardWeak",
}

// bidWhistDirName returns the localized name of a bid direction.
func bidWhistDirName(dir int) string {
	switch dir {
	case domain.BidWhistDirectionUptown:
		return i18n.T("bidwhist.dirUptown")
	case domain.BidWhistDirectionDowntown:
		return i18n.T("bidwhist.dirDowntown")
	case domain.BidWhistDirectionNoTrump:
		return i18n.T("bidwhist.dirNoTrump")
	}
	return "?"
}

// bidWhistBidStr returns a localized label for a bid.
func bidWhistBidStr(b *domain.BidWhistBid) string {
	if b == nil {
		return ""
	}
	return i18n.Tf("bidwhist.bidLabel",
		"tricks", strconv.Itoa(b.Tricks),
		"dir", bidWhistDirName(b.Direction))
}

// bidWhistPlayerStr returns the display string for a single player.
func bidWhistPlayerStr(player *domain.BidWhistPlayer, i int, kitty []int) string {
	var b strings.Builder
	status := ""
	if player.GetIsDeclarer() {
		status = " " + i18n.T("bidwhist.declarerMark")
	} else if player.GetPassed() {
		status = " " + i18n.T("bidwhist.passMark")
	} else if bid := player.GetBid(); bid != nil {
		status = " [" + bidWhistBidStr(bid) + "]"
	}
	b.WriteString(i18n.Tf("bidwhist.playerLine",
		"name", cuiPlayerName(player, i),
		"team", strconv.Itoa(player.GetTeam()),
		"tricks", strconv.Itoa(player.GetTrickCount()),
		"cards", strconv.Itoa(player.GetCardsSize()),
		"status", status,
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		// **どの6枚が交換で入ってきたのかを示す。**Web はバッジで区別しているのに、
		// CUI は覚えておくしかなかった (#5632)。kitty が空なら無印のまま。
		b.WriteString(cuiIndexMarkedCardListStr(player, kitty, CuiKittyMark) + "\n")
	}
	return b.String()
}

// Output renders the current game state for the active locale.
func (p *BidWhistCuiPresenter) Output(g interfaces.BidWhistGame, lastErr error) string {
	return buildCuiOutput(i18n.T("bidwhist.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("bidwhist.header",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber())) + "\n")
		dealerIdx := g.GetDealerIdx()
		b.WriteString(i18n.Tf("bidwhist.dealer",
			"name", cuiPlayerName(g.GetPlayer(dealerIdx), dealerIdx)) + "\n")

		if g.GetDeclarerIdx() >= 0 {
			declIdx := g.GetDeclarerIdx()
			trump := i18n.T("bidwhist.noTrumpLabel")
			if g.GetTrumpSuit() >= domain.CardDesignSpade {
				trump = cuiSuitName(g.GetTrumpSuit())
			}
			b.WriteString(i18n.Tf("bidwhist.contractLine",
				"name", cuiPlayerName(g.GetPlayer(declIdx), declIdx),
				"tricks", strconv.Itoa(g.GetContractTricks()),
				"dir", bidWhistDirName(g.GetContractDirection()),
				"trump", trump) + "\n")
		} else if hb := g.GetHighestBid(); hb != nil {
			b.WriteString(i18n.Tf("bidwhist.highestBid", "bid", bidWhistBidStr(hb)) + "\n")
		} else {
			b.WriteString(i18n.T("bidwhist.contractUndecided") + "\n")
		}

		b.WriteString(i18n.Tf("bidwhist.teamScoreLine",
			"t0", strconv.Itoa(g.GetTeamScore(0)),
			"t1", strconv.Itoa(g.GetTeamScore(1))) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			// 印はキティ交換フェーズの落札者にだけ付く (GetKittyIndices が
			// フェーズと落札者を見て空を返す)。
			var kitty []int
			if g.GetPlayer(i) != nil && g.GetPlayer(i).GetIsHuman() {
				kitty = g.GetKittyIndices()
			}
			b.WriteString(bidWhistPlayerStr(g.GetPlayer(i), i, kitty))
		}

		b.WriteString("----------\n")

		cuiTrickBlock(b, g.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			banner := i18n.Tf("bidwhist.gameEnd", "team", strconv.Itoa(g.GetWinnerTeam()))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.BidWhistPhaseBid:
			bidIdx := g.GetBidPlayerIdx()
			b.WriteString(i18n.Tf("bidwhist.promptBid",
				"name", cuiPlayerName(g.GetPlayer(bidIdx), bidIdx)) + "\n")
			b.WriteString(i18n.T("bidwhist.promptBidHelp") + "\n")
		case domain.BidWhistPhaseTrumpDeclaration:
			b.WriteString(i18n.T("bidwhist.promptTrump") + "\n")
			b.WriteString(i18n.T("bidwhist.promptTrumpHelp") + "\n")
		case domain.BidWhistPhaseKittyExchange:
			b.WriteString(i18n.T("bidwhist.promptKittyExchange") + "\n")
			b.WriteString(i18n.T("bidwhist.promptKittyExchangeHelp") + "\n")
		case domain.BidWhistPhasePlay:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("bidwhist.promptCurrentPlayer",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("bidwhist.promptPlayHelp") + "\n")
		case domain.BidWhistPhaseTrickEnd:
			b.WriteString(i18n.T("bidwhist.promptTrickEnd") + "\n")
			b.WriteString(i18n.T("bidwhist.promptTrickEndHelp") + "\n")
		case domain.BidWhistPhaseRoundEnd:
			b.WriteString(i18n.T("bidwhist.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("bidwhist.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current hint.
func (p *BidWhistCuiPresenter) HintOutput(g interfaces.BidWhistGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("bidwhist.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, bidWhistHintReasonKeys)
	switch {
	case hint.Pass != nil && *hint.Pass:
		return color.Yellow(i18n.Tf("bidwhist.hintPass", "reason", reason)) + "\n"
	case hint.BidTricks != nil && hint.BidDirection != nil:
		bid := &domain.BidWhistBid{Tricks: *hint.BidTricks, Direction: *hint.BidDirection}
		return color.Yellow(i18n.Tf("bidwhist.hintBid",
			"bid", bidWhistBidStr(bid), "reason", reason)) + "\n"
	case hint.TrumpSuit != nil:
		return color.Yellow(i18n.Tf("bidwhist.hintTrump",
			"suit", cuiSuitName(*hint.TrumpSuit), "reason", reason)) + "\n"
	case len(hint.DiscardIndices) > 0:
		return color.Yellow(i18n.Tf("bidwhist.hintDiscard",
			"indices", joinInts(hint.DiscardIndices), "reason", reason)) + "\n"
	case hint.CardIndex != nil:
		card := g.GetPlayer(0).GetCard(*hint.CardIndex)
		return color.Yellow(i18n.Tf("bidwhist.hintCard",
			"idx", strconv.Itoa(*hint.CardIndex),
			"card", cuiCardStr(card),
			"reason", reason)) + "\n"
	}
	return i18n.T("bidwhist.hintNone") + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *BidWhistCuiPresenter) ActionLogOutput(g interfaces.BidWhistGame) string {
	return actionLogOutputText(g)
}
