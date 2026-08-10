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

// bridgeBidSuitName localizes a bid-suit value (Club..NT) to a suit name.
// Contract/bid suits are numbered separately from card designs, so they are
// mapped through the card design before reusing cuiSuitName; NT has no suit.
func bridgeBidSuitName(bidSuit int) string {
	switch bidSuit {
	case domain.BridgeBidSuitClub:
		return cuiSuitName(domain.CardDesignClover)
	case domain.BridgeBidSuitDiamond:
		return cuiSuitName(domain.CardDesignDiamond)
	case domain.BridgeBidSuitHeart:
		return cuiSuitName(domain.CardDesignHeart)
	case domain.BridgeBidSuitSpade:
		return cuiSuitName(domain.CardDesignSpade)
	case domain.BridgeBidSuitNT:
		return i18n.T("bridge.bidNoTrump")
	}
	return "UNKNOWN"
}

// bridgePlayerStr returns the display string for a single Bridge player.
func bridgePlayerStr(player *domain.BridgePlayer, i int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("bridge.playerLine",
		"name", cuiPlayerName(player, i),
		"team", strconv.Itoa(player.GetTeam()),
		"tricks", strconv.Itoa(player.GetTrickCount()),
		"cards", strconv.Itoa(player.GetCardsSize()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player))
		b.WriteString("\n")
	}
	return b.String()
}

// BridgeCuiPresenter renders the Bridge CUI view.
type BridgeCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *BridgeCuiPresenter) Output(b interfaces.BridgeGame, lastErr error) string {
	return buildCuiOutput(i18n.T("bridge.helpTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("bridge.round",
			"round", strconv.Itoa(b.GetRoundNumber()),
			"trick", strconv.Itoa(b.GetTrickNumber())) + "\n")
		sb.WriteString(i18n.Tf("bridge.dealer",
			"name", cuiPlayerName(b.GetPlayer(b.GetDealerIdx()), b.GetDealerIdx())) + "\n")

		trumpSuit := b.GetTrumpSuit()
		switch {
		case trumpSuit == -1:
			sb.WriteString(i18n.T("bridge.trumpNoTrump") + "\n")
		case trumpSuit > 0:
			sb.WriteString(i18n.Tf("bridge.trumpSuit", "suit", cuiSuitName(trumpSuit)) + "\n")
		default:
			sb.WriteString(i18n.T("bridge.trumpUndecided") + "\n")
		}

		if contractLevel := b.GetContractLevel(); contractLevel > 0 {
			sb.WriteString(i18n.Tf("bridge.contractLine",
				"level", strconv.Itoa(contractLevel),
				"suit", bridgeBidSuitName(b.GetContractSuit())))
			switch b.GetDoubled() {
			case 1:
				sb.WriteString(i18n.T("bridge.contractDoubled"))
			case 2:
				sb.WriteString(i18n.T("bridge.contractRedoubled"))
			}
			sb.WriteString("\n")

			declarerIdx := b.GetDeclarerIdx()
			if declarerIdx >= 0 {
				sb.WriteString(i18n.Tf("bridge.declarerLine",
					"declarer", cuiPlayerName(b.GetPlayer(declarerIdx), declarerIdx),
					"dummy", cuiPlayerName(b.GetPlayer(b.GetDummyIdx()), b.GetDummyIdx())) + "\n")
			}
		}

		// Vulnerability
		sb.WriteString(i18n.Tf("bridge.vulnerability",
			"a", strconv.FormatBool(b.GetVulnerability(0)),
			"b", strconv.FormatBool(b.GetVulnerability(1))) + "\n")

		// Team scores
		sb.WriteString(i18n.Tf("bridge.teamScores",
			"a", strconv.Itoa(b.GetTeamScore(0)),
			"aGames", strconv.Itoa(b.GetGamesWon(0)),
			"aBelow", strconv.Itoa(b.GetBelowLine(0)),
			"b", strconv.Itoa(b.GetTeamScore(1)),
			"bGames", strconv.Itoa(b.GetGamesWon(1)),
			"bBelow", strconv.Itoa(b.GetBelowLine(1))) + "\n")

		// Player rows
		for i := 0; i < b.GetPlayerCnt(); i++ {
			sb.WriteString(bridgePlayerStr(b.GetPlayer(i), i))
		}

		// Dummy hand (after the opening lead)
		if b.IsOpeningLeadDone() {
			if dummyHand := b.GetDummyHand(); len(dummyHand) > 0 {
				parts := make([]string, len(dummyHand))
				for i, c := range dummyHand {
					parts[i] = cuiCardStr(c)
				}
				sb.WriteString(i18n.Tf("bridge.dummyHand", "cards", strings.Join(parts, ", ")) + "\n")
			}
		}

		sb.WriteString("----------\n")

		// Current trick
		trick := b.GetCurrentTrick()
		cuiTrickBlock(sb, trick,
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(b.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(sb, lastErr)

		// Game state
		if b.GetGameEndFlag() {
			banner := color.Green(i18n.Tf("bridge.winnerBanner",
				"team", strconv.Itoa(b.GetWinnerTeam())))
			sb.WriteString(i18n.Tf("bridge.gameEnd", "banner", banner) + "\n")
			return
		}
		switch b.GetPhase() {
		case domain.BridgePhaseBid:
			// Show the auction so far (partner signals, doubles) so the human need
			// not track it from memory; omitted before the first bid.
			if history := b.GetBidHistory(); len(history) > 0 {
				entries := make([]string, len(history))
				for i, e := range history {
					entries[i] = i18n.Tf("bridge.bidHistoryEntry",
						"name", cuiPlayerName(b.GetPlayer(e.PlayerIdx), e.PlayerIdx),
						"bid", bridgeBidEntryStr(e))
				}
				sb.WriteString(i18n.Tf("bridge.bidHistory",
					"bids", strings.Join(entries, " → ")) + "\n")
			}
			bidIdx := b.GetBidPlayerIdx()
			sb.WriteString(i18n.Tf("bridge.promptBid",
				"name", cuiPlayerName(b.GetPlayer(bidIdx), bidIdx)) + "\n")
			// **どこから上回れるか・ダブルできるかを出す。**Web はボタンを無効化して
			// 理由まで出すのに、CUI は打って拒否されるまで分からなかった (#4903)。
			if lv, st, ok := b.BridgeMinLegalBid(); ok {
				sb.WriteString(i18n.Tf("bridge.minLegalBid",
					"bid", strconv.Itoa(lv)+bridgeBidSuitName(st)) + "\n")
			} else {
				sb.WriteString(i18n.T("bridge.noHigherBid") + "\n")
			}
			switch {
			case b.BridgeCanRedouble(bidIdx):
				sb.WriteString(i18n.T("bridge.canRedouble") + "\n")
			case b.BridgeCanDouble(bidIdx):
				sb.WriteString(i18n.T("bridge.canDouble") + "\n")
			}
			sb.WriteString(i18n.T("bridge.promptBidHelp") + "\n")
		case domain.BridgePhasePlay:
			currentIdx := b.GetCurrentPlayerIdx()
			sb.WriteString(i18n.Tf("bridge.promptPlay",
				"name", cuiPlayerName(b.GetPlayer(currentIdx), currentIdx)) + "\n")
			sb.WriteString(i18n.T("bridge.promptPlayHelp") + "\n")
		case domain.BridgePhaseTrickEnd:
			sb.WriteString(i18n.T("bridge.promptTrickEnd") + "\n")
			sb.WriteString(i18n.T("bridge.promptTrickEndHelp") + "\n")
		case domain.BridgePhaseRoundEnd:
			sb.WriteString(i18n.T("bridge.promptRoundEnd") + "\n")
			sb.WriteString(i18n.T("bridge.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current Bridge hint.
func (p *BridgeCuiPresenter) HintOutput(b interfaces.BridgeGame) string {
	hint := b.GetHint()
	if hint == nil {
		return i18n.T("bridge.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, bridgeHintReasonKeys)
	if hint.BidType != nil {
		bidTypeStr := bridgeBidTypeStr(*hint.BidType)
		if *hint.BidType == int(domain.BridgeBidNormal) && hint.BidLevel != nil && hint.BidSuit != nil {
			return color.Yellow(i18n.Tf("bridge.hintBidWithLevel",
				"type", bidTypeStr,
				"level", strconv.Itoa(*hint.BidLevel),
				"suit", bridgeBidSuitName(*hint.BidSuit),
				"reason", reason)) + "\n"
		}
		return color.Yellow(i18n.Tf("bridge.hintBidNoLevel",
			"type", bidTypeStr,
			"reason", reason)) + "\n"
	}
	if hint.CardIndex == nil {
		return i18n.T("bridge.hintNone") + "\n"
	}
	card := b.GetPlayer(0).GetCard(*hint.CardIndex)
	return color.Yellow(i18n.Tf("bridge.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", reason)) + "\n"
}

// bridgeBidTypeKeys maps a bid type constant to its i18n key.
var bridgeBidTypeKeys = map[domain.BridgeBidType]string{
	domain.BridgeBidPass:     "bridge.bidPass",
	domain.BridgeBidNormal:   "bridge.bidNormal",
	domain.BridgeBidDouble:   "bridge.bidDouble",
	domain.BridgeBidRedouble: "bridge.bidRedouble",
}

// bridgeBidTypeStr returns the localized bid-type label, falling back to a
// debug-friendly numeric form for unknown values.
func bridgeBidTypeStr(bidType int) string {
	if key, ok := bridgeBidTypeKeys[domain.BridgeBidType(bidType)]; ok {
		return i18n.T(key)
	}
	return i18n.Tf("bridge.bidUnknown", "n", strconv.Itoa(bidType))
}

// bridgeBidEntryStr formats one auction entry: "1♣" for a normal bid, or the
// localized pass/double/redouble label.
func bridgeBidEntryStr(e *domain.BridgeBidEntry) string {
	if e.BidType == domain.BridgeBidNormal {
		return strconv.Itoa(e.Level) + bridgeBidSuitName(e.Suit)
	}
	return bridgeBidTypeStr(int(e.BidType))
}

// bridgeHintReasonKeys maps Bridge-specific hint-reason identifiers to
// their i18n keys. Reasons not in this map fall through to
// hintReasonStr → cui_common (e.g. strategic_bid).
var bridgeHintReasonKeys = map[string]string{
	"support_partner": "bridge.hintReasonSupportPartner",
	"competitive_bid": "bridge.hintReasonCompetitiveBid",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *BridgeCuiPresenter) ActionLogOutput(b interfaces.BridgeGame) string {
	return actionLogOutputText(b)
}
