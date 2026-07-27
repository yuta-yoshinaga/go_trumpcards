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

// FiveHundredCuiPresenter renders the 500 (Five Hundred) CUI view.
type FiveHundredCuiPresenter struct{}

// fiveHundredHintReasonKeys maps 500-specific hint reasons to i18n keys.
// Reasons not listed fall through to the shared cui_common layer.
var fiveHundredHintReasonKeys = map[string]string{
	"pass_recommended": "fivehundred.hintReasonPass",
	"discard_weakest":  "fivehundred.hintReasonDiscardWeakest",
	"lead_trump":       "fivehundred.hintReasonLeadTrump",
	"discard_weak":     "fivehundred.hintReasonDiscardWeak",
}

// fiveHundredBidStr returns a localized label for a bid.
func fiveHundredBidStr(b *domain.FiveHundredBid) string {
	if b == nil {
		return ""
	}
	switch b.Kind {
	case domain.FiveHundredContractSuit:
		return i18n.Tf("fivehundred.bidSuit",
			"tricks", strconv.Itoa(b.Tricks),
			"suit", cuiSuitName(b.Suit),
			"value", strconv.Itoa(b.Value()))
	case domain.FiveHundredContractNoTrump:
		return i18n.Tf("fivehundred.bidNoTrump",
			"tricks", strconv.Itoa(b.Tricks),
			"value", strconv.Itoa(b.Value()))
	case domain.FiveHundredContractMisere:
		return i18n.Tf("fivehundred.bidMisere", "value", strconv.Itoa(b.Value()))
	case domain.FiveHundredContractOpenMisere:
		return i18n.Tf("fivehundred.bidOpenMisere", "value", strconv.Itoa(b.Value()))
	}
	return ""
}

// fiveHundredPlayerStr returns the display string for a single player.
func fiveHundredPlayerStr(player *domain.FiveHundredPlayer, i int) string {
	var b strings.Builder
	status := ""
	if player.GetIsDeclarer() {
		status = " " + i18n.T("fivehundred.declarerMark")
	} else if player.GetPassed() {
		status = " " + i18n.T("fivehundred.passMark")
	} else if bid := player.GetBid(); bid != nil {
		status = " [" + fiveHundredBidStr(bid) + "]"
	}
	b.WriteString(i18n.Tf("fivehundred.playerLine",
		"name", cuiPlayerName(player, i),
		"team", strconv.Itoa(player.GetTeam()),
		"tricks", strconv.Itoa(player.GetTrickCount()),
		"cards", strconv.Itoa(player.GetCardsSize()),
		"status", status,
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// Output renders the current game state for the active locale.
func (p *FiveHundredCuiPresenter) Output(g interfaces.FiveHundredGame, lastErr error) string {
	return buildCuiOutput(i18n.T("fivehundred.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("fivehundred.header",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber())) + "\n")
		dealerIdx := g.GetDealerIdx()
		b.WriteString(i18n.Tf("fivehundred.dealer",
			"name", cuiPlayerName(g.GetPlayer(dealerIdx), dealerIdx)) + "\n")

		if g.GetContractKind() != int(domain.FiveHundredContractNone) {
			declIdx := g.GetDeclarerIdx()
			b.WriteString(i18n.Tf("fivehundred.contractLine",
				"name", cuiPlayerName(g.GetPlayer(declIdx), declIdx),
				"value", strconv.Itoa(g.GetContractValue())) + "\n")
		} else if hb := g.GetHighestBid(); hb != nil {
			b.WriteString(i18n.Tf("fivehundred.highestBid", "bid", fiveHundredBidStr(hb)) + "\n")
		} else {
			b.WriteString(i18n.T("fivehundred.contractUndecided") + "\n")
		}

		b.WriteString(i18n.Tf("fivehundred.teamScoreLine",
			"t0", strconv.Itoa(g.GetTeamScore(0)),
			"t1", strconv.Itoa(g.GetTeamScore(1))) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(fiveHundredPlayerStr(g.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		cuiTrickBlock(b, g.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			banner := i18n.Tf("fivehundred.gameEnd", "team", strconv.Itoa(g.GetWinnerTeam()))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.FiveHundredPhaseBid:
			bidIdx := g.GetBidPlayerIdx()
			b.WriteString(i18n.Tf("fivehundred.promptBid",
				"name", cuiPlayerName(g.GetPlayer(bidIdx), bidIdx)) + "\n")
			b.WriteString(i18n.T("fivehundred.promptBidHelp") + "\n")
		case domain.FiveHundredPhaseKittyExchange:
			b.WriteString(i18n.T("fivehundred.promptKittyExchange") + "\n")
			b.WriteString(i18n.T("fivehundred.promptKittyExchangeHelp") + "\n")
		case domain.FiveHundredPhasePlay:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("fivehundred.promptCurrentPlayer",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("fivehundred.promptPlayHelp") + "\n")
		case domain.FiveHundredPhaseTrickEnd:
			b.WriteString(i18n.T("fivehundred.promptTrickEnd") + "\n")
			b.WriteString(i18n.T("fivehundred.promptTrickEndHelp") + "\n")
		case domain.FiveHundredPhaseRoundEnd:
			b.WriteString(i18n.T("fivehundred.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("fivehundred.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current hint.
func (p *FiveHundredCuiPresenter) HintOutput(g interfaces.FiveHundredGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("fivehundred.hintNone") + "\n"
	}
	reason := lookupHintReason(hint.Reason, fiveHundredHintReasonKeys)
	switch {
	case hint.Pass != nil && *hint.Pass:
		return color.Yellow(i18n.Tf("fivehundred.hintPass", "reason", reason)) + "\n"
	case hint.BidKind != nil:
		bid := &domain.FiveHundredBid{
			Kind:   domain.FiveHundredContractKind(*hint.BidKind),
			Tricks: ptrOrZero(hint.BidTricks),
			Suit:   ptrOrNeg(hint.BidSuit),
		}
		return color.Yellow(i18n.Tf("fivehundred.hintBid",
			"bid", fiveHundredBidStr(bid), "reason", reason)) + "\n"
	case len(hint.DiscardIndices) > 0:
		return color.Yellow(i18n.Tf("fivehundred.hintDiscard",
			"indices", joinInts(hint.DiscardIndices), "reason", reason)) + "\n"
	case hint.CardIndex != nil:
		card := g.GetPlayer(0).GetCard(*hint.CardIndex)
		return color.Yellow(i18n.Tf("fivehundred.hintCard",
			"idx", strconv.Itoa(*hint.CardIndex),
			"card", cuiCardStr(card),
			"reason", reason)) + "\n"
	}
	return i18n.T("fivehundred.hintNone") + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *FiveHundredCuiPresenter) ActionLogOutput(g interfaces.FiveHundredGame) string {
	return actionLogOutputText(g)
}

// ptrOrZero returns *p or 0 when p is nil.
func ptrOrZero(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}

// ptrOrNeg returns *p or -1 when p is nil.
func ptrOrNeg(p *int) int {
	if p == nil {
		return -1
	}
	return *p
}

// joinInts formats an int slice as a space-separated string.
func joinInts(xs []int) string {
	parts := make([]string, len(xs))
	for i, x := range xs {
		parts[i] = strconv.Itoa(x)
	}
	return strings.Join(parts, " ")
}
