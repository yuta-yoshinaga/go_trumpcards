//go:build !js || !wasm || extra4

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// mightySuitGlyphs maps the trump-suit constant to a Unicode glyph.
var mightySuitGlyphs = map[int]string{
	domain.CardDesignSpade:   "♠",
	domain.CardDesignClover:  "♣",
	domain.CardDesignHeart:   "♥",
	domain.CardDesignDiamond: "♦",
}

// mightyBidStr renders the bid column for the player line.
func mightyBidStr(bid int, noTrump bool) string {
	switch {
	case bid == 0:
		return i18n.T("mighty.bidPass")
	case bid > 0:
		if noTrump {
			return strconv.Itoa(bid) + " " + i18n.T("mighty.bidNoTrumpTag")
		}
		return strconv.Itoa(bid)
	default:
		return i18n.T("mighty.bidPending")
	}
}

// mightyRoleStr returns the role badge (declarer / partner) prefixed with a
// leading space, or "" when the player has no badge to display.
func mightyRoleStr(player *domain.MightyPlayer, partnerRevealed bool) string {
	if player.GetIsDeclarer() {
		return " " + i18n.T("mighty.roleDeclarer")
	}
	if partnerRevealed && player.GetIsPartner() {
		return " " + i18n.T("mighty.rolePartner")
	}
	return ""
}

// mightyPlayerStr returns the display string for a single Mighty player.
func mightyPlayerStr(player *domain.MightyPlayer, i int, partnerRevealed bool) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("mighty.playerLine",
		"name", cuiPlayerName(player, i),
		"role", mightyRoleStr(player, partnerRevealed),
		"bid", mightyBidStr(player.GetBid(), player.GetBidNoTrump()),
		"tricks", strconv.Itoa(player.GetTrickCount()),
		"points", strconv.Itoa(player.GetPointCards()),
		"cum", strconv.Itoa(player.GetCumulativeScore()),
		"round", strconv.Itoa(player.GetRoundScore()),
		"cards", strconv.Itoa(player.GetCardsSize()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player))
		b.WriteString("\n")
	}
	return b.String()
}

// MightyCuiPresenter renders the Mighty CUI view.
type MightyCuiPresenter struct{}

// mightyKittyLine returns an indexed list of the kitty cards as they now sit in
// the declarer's hand (the kitty is merged in when this phase begins), so the CUI
// player can see exactly which hand indices to discard. Returns "" when there is
// no human-visible kitty to show. The indices shown match those the `e` (exchange)
// command consumes.
func mightyKittyLine(m interfaces.MightyGame) string {
	declarer := m.GetPlayer(m.GetDeclarerIdx())
	if declarer == nil {
		return ""
	}
	kitty := m.GetKitty()
	parts := make([]string, 0, len(kitty))
	for _, kc := range kitty {
		if kc == nil {
			continue
		}
		for i := 0; i < declarer.GetCardsSize(); i++ {
			hc := declarer.GetCard(i)
			if hc != nil && hc.GetDesign() == kc.GetDesign() && hc.GetValue() == kc.GetValue() {
				parts = append(parts, "["+strconv.Itoa(i)+"]"+mightyCuiCardStr(hc))
				break
			}
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return i18n.T("mighty.promptKittyCards") + " " + strings.Join(parts, "  ")
}

// Output renders the current game state for the active locale.
func (p *MightyCuiPresenter) Output(m interfaces.MightyGame, lastErr error) string {
	return buildCuiOutput(i18n.T("mighty.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("mighty.round",
			"round", strconv.Itoa(m.GetRoundNumber()),
			"trick", strconv.Itoa(m.GetTrickNumber())))
		b.WriteString("\n")

		trumpSuit := m.GetTrumpSuit()
		if trumpSuit > 0 {
			b.WriteString(i18n.Tf("mighty.trump", "suit", mightySuitGlyphs[trumpSuit]))
			b.WriteString("\n")
		} else if m.GetWinningBidNoTrump() {
			b.WriteString(i18n.T("mighty.trumpNoTrump"))
			b.WriteString("\n")
		}

		if partnerCard := m.GetPartnerCard(); partnerCard != nil {
			b.WriteString(i18n.Tf("mighty.partnerCard", "card", mightyCuiCardStr(partnerCard)))
			if m.GetPartnerRevealed() {
				b.WriteString(i18n.T("mighty.partnerRevealed"))
			} else {
				b.WriteString(i18n.T("mighty.partnerHidden"))
			}
			b.WriteString("\n")
		}

		if m.GetHighestBid() > 0 {
			b.WriteString(i18n.Tf("mighty.highestBid", "bid", strconv.Itoa(m.GetHighestBid())))
			b.WriteString("\n")
		}

		for i := 0; i < m.GetPlayerCnt(); i++ {
			b.WriteString(mightyPlayerStr(m.GetPlayer(i), i, m.GetPartnerRevealed()))
		}

		b.WriteString("----------\n")

		trick := m.GetCurrentTrick()
		cuiTrickBlock(b, trick,
			func(tc *domain.MightyTrickCard) int { return tc.PlayerIdx },
			func(tc *domain.MightyTrickCard) string { return mightyCuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(m.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		if m.GetGameEndFlag() {
			if m.GetWinnerTeam() == domain.MightyWinnerDeclarer {
				b.WriteString(color.Green(i18n.T("mighty.gameEndDeclarerWin")) + "\n")
			} else {
				b.WriteString(color.Green(i18n.T("mighty.gameEndOppositionWin")) + "\n")
			}
			return
		}
		switch m.GetPhase() {
		case domain.MightyPhaseBid:
			bidIdx := m.GetBidPlayerIdx()
			b.WriteString(i18n.Tf("mighty.promptBid",
				"name", cuiPlayerName(m.GetPlayer(bidIdx), bidIdx)))
			b.WriteString("\n")
			b.WriteString(i18n.T("mighty.promptBidHelp") + "\n")
			// **`nt` を付けると最低ビッドが上がる。**その分だけ数字が要るのに、
			// CUI はコマンド構文しか出しておらず、付けてエラーになって初めて
			// 気づく形だった (#5594)。点数は設定から取る。
			b.WriteString(i18n.Tf("mighty.promptBidNoTrumpExtra",
				"points", strconv.Itoa(m.GetConfig().NoTrumpExtra)) + "\n")
		case domain.MightyPhaseTrumpAndFriend:
			b.WriteString(i18n.T("mighty.promptTrumpHeader") + "\n")
			b.WriteString(i18n.T("mighty.promptTrumpDeclareHelp") + "\n")
			b.WriteString(i18n.T("mighty.promptTrumpSuitLegend") + "\n")
		case domain.MightyPhaseKittyExchange:
			b.WriteString(i18n.T("mighty.promptKittyHeader") + "\n")
			if line := mightyKittyLine(m); line != "" {
				b.WriteString(line + "\n")
			}
			b.WriteString(i18n.T("mighty.promptKittyHelp") + "\n")
		case domain.MightyPhasePlay:
			currentIdx := m.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("mighty.promptPlayTurn",
				"name", cuiPlayerName(m.GetPlayer(currentIdx), currentIdx)))
			b.WriteString("\n")
			b.WriteString(i18n.T("mighty.promptPlayHelp") + "\n")
		case domain.MightyPhaseTrickEnd:
			b.WriteString(i18n.T("mighty.promptTrickEnd") + "\n")
			b.WriteString(i18n.T("mighty.promptTrickEndHelp") + "\n")
		case domain.MightyPhaseRoundEnd:
			b.WriteString(i18n.T("mighty.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("mighty.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current Mighty hint.
func (p *MightyCuiPresenter) HintOutput(m interfaces.MightyGame) string {
	hint := m.GetHint()
	if hint == nil {
		return i18n.T("mighty.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, mightyHintReasonKeys)
	switch {
	case hint.Bid != nil:
		tag := ""
		if hint.BidNoTrump != nil && *hint.BidNoTrump {
			tag = " " + i18n.T("mighty.bidNoTrumpTag")
		}
		return color.Yellow(i18n.Tf("mighty.hintBid",
			"bid", strconv.Itoa(*hint.Bid)+tag,
			"reason", reason)) + "\n"
	case hint.TrumpSuit != nil:
		suit := *hint.TrumpSuit
		suitStr := i18n.T("mighty.trumpNoTrump")
		if g, ok := mightySuitGlyphs[suit]; ok {
			suitStr = g
		}
		return color.Yellow(i18n.Tf("mighty.hintTrump",
			"suit", suitStr,
			"reason", reason)) + "\n"
	case len(hint.DiscardIndices) > 0:
		idxs := make([]string, len(hint.DiscardIndices))
		for i, idx := range hint.DiscardIndices {
			idxs[i] = strconv.Itoa(idx)
		}
		return color.Yellow(i18n.Tf("mighty.hintDiscard",
			"indices", strings.Join(idxs, ","),
			"reason", reason)) + "\n"
	case hint.CardIndex != nil:
		card := m.GetPlayer(0).GetCard(*hint.CardIndex)
		return color.Yellow(i18n.Tf("mighty.hintCard",
			"idx", strconv.Itoa(*hint.CardIndex),
			"card", mightyCuiCardStr(card),
			"reason", reason)) + "\n"
	}
	return i18n.T("mighty.hintNone") + "\n"
}

// mightyCuiCardStr renders a card with Mighty-flavoured rules (Joker → "Joker").
func mightyCuiCardStr(card *domain.Card) string {
	if card.GetDesign() == domain.CardDesignJoker {
		return "Joker"
	}
	return cuiCardStr(card)
}

// mightyHintReasonKeys maps a hint-reason identifier specific to Mighty
// to its i18n key.
var mightyHintReasonKeys = map[string]string{
	"strategic_bid":     "mighty.hintReasonStrategicBid",
	"strategic_declare": "mighty.hintReasonStrategicDeclare",
	"strategic_discard": "mighty.hintReasonStrategicDiscard",
	"joker_lead":        "mighty.hintReasonJokerLead",
	"play_joker":        "mighty.hintReasonPlayJoker",
	"play_mighty":       "mighty.hintReasonPlayMighty",
	"play_low":          "mighty.hintReasonPlayLow",
	"play_trump":        "mighty.hintReasonPlayTrump",
	"play_follow":       "mighty.hintReasonPlayFollow",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *MightyCuiPresenter) ActionLogOutput(m interfaces.MightyGame) string {
	return actionLogOutputTextForSeats[*domain.MightyPlayer](m)
}
