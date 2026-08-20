package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// ohHellPlayerStr returns the display string for a single Oh Hell player.
func ohHellPlayerStr(player *domain.OhHellPlayer, i int) string {
	var b strings.Builder
	bidStr := i18n.T("ohhell.bidPending")
	if player.GetBid() >= 0 {
		bidStr = strconv.Itoa(player.GetBid())
	}
	b.WriteString(i18n.Tf("ohhell.playerLine",
		"name", cuiPlayerName(player, i),
		"bid", bidStr,
		"tricks", strconv.Itoa(player.GetTrickCount()),
		"cum", strconv.Itoa(player.GetCumulativeScore()),
		"round", strconv.Itoa(player.GetRoundScore()),
		"cards", strconv.Itoa(player.GetCardsSize()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// OhHellCuiPresenter renders the Oh Hell CUI view.
type OhHellCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *OhHellCuiPresenter) Output(o interfaces.OhHellGame, lastErr error) string {
	return buildCuiOutput(i18n.T("ohhell.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("ohhell.header",
			"round", strconv.Itoa(o.GetRoundNumber()),
			"total", strconv.Itoa(o.GetTotalRounds()),
			"hand", strconv.Itoa(o.GetHandSize()),
			"trick", strconv.Itoa(o.GetTrickNumber())) + "\n")

		if trumpCard := o.GetTrumpCard(); trumpCard != nil {
			b.WriteString(i18n.Tf("ohhell.trumpCard", "card", cuiCardStr(trumpCard)) + "\n")
		} else {
			b.WriteString(i18n.T("ohhell.trumpNone") + "\n")
		}

		dealerIdx := o.GetDealerIdx()
		b.WriteString(i18n.Tf("ohhell.dealer",
			"name", cuiPlayerName(o.GetPlayer(dealerIdx), dealerIdx)) + "\n")

		for i := 0; i < o.GetPlayerCnt(); i++ {
			b.WriteString(ohHellPlayerStr(o.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		// Current trick
		trick := o.GetCurrentTrick()
		cuiTrickBlock(b, trick,
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(o.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		// Game state
		if o.GetGameEndFlag() {
			winnerIdx := o.GetWinnerIdx()
			player := o.GetPlayer(winnerIdx)
			banner := i18n.Tf("ohhell.gameEnd", "name", cuiPlayerName(player, winnerIdx))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch o.GetPhase() {
		case domain.OhHellPhaseBid:
			// Table bid total vs hand size: mirrors the web bid-total chip so the
			// player can see the over/under/exact tension without summing by hand.
			total := 0
			for i := 0; i < o.GetPlayerCnt(); i++ {
				if bid := o.GetPlayer(i).GetBid(); bid >= 0 {
					total += bid
				}
			}
			hand := o.GetHandSize()
			var state string
			switch {
			case total > hand:
				state = color.Red(i18n.T("ohhell.bidStateOver"))
			case total < hand:
				state = color.Yellow(i18n.T("ohhell.bidStateUnder"))
			default:
				state = color.Green(i18n.T("ohhell.bidStateExact"))
			}
			b.WriteString(i18n.Tf("ohhell.bidSummary",
				"total", strconv.Itoa(total),
				"hand", strconv.Itoa(hand),
				"state", state) + "\n")

			bidIdx := o.GetBidPlayerIdx()
			name := cuiPlayerName(o.GetPlayer(bidIdx), bidIdx)
			if restricted := o.GetRestrictedBid(); restricted >= 0 {
				b.WriteString(i18n.Tf("ohhell.promptBidRestricted",
					"name", name,
					"restricted", strconv.Itoa(restricted)) + "\n")
			} else {
				b.WriteString(i18n.Tf("ohhell.promptBid", "name", name) + "\n")
			}
			b.WriteString(i18n.T("ohhell.promptBidHelp") + "\n")
		case domain.OhHellPhasePlay:
			currentIdx := o.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("ohhell.promptPlay",
				"name", cuiPlayerName(o.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("ohhell.promptPlayHelp") + "\n")
		case domain.OhHellPhaseTrickEnd:
			b.WriteString(i18n.T("ohhell.promptTrickEnd") + "\n")
			b.WriteString(i18n.T("ohhell.promptTrickEndHelp") + "\n")
		case domain.OhHellPhaseRoundEnd:
			b.WriteString(i18n.T("ohhell.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("ohhell.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current Oh Hell hint.
func (p *OhHellCuiPresenter) HintOutput(o interfaces.OhHellGame) string {
	hint := o.GetHint()
	if hint == nil {
		return i18n.T("ohhell.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, nil)
	if hint.Bid != nil {
		return color.Yellow(i18n.Tf("ohhell.hintBid",
			"bid", strconv.Itoa(*hint.Bid),
			"reason", reason)) + "\n"
	}
	if hint.CardIndex == nil {
		return i18n.T("ohhell.hintNone") + "\n"
	}
	humanIdx := -1
	for i := 0; i < o.GetPlayerCnt(); i++ {
		if o.GetPlayer(i).GetIsHuman() {
			humanIdx = i
			break
		}
	}
	if humanIdx < 0 {
		return i18n.T("ohhell.hintNone") + "\n"
	}
	card := o.GetPlayer(humanIdx).GetCard(*hint.CardIndex)
	return color.Yellow(i18n.Tf("ohhell.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", reason)) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *OhHellCuiPresenter) ActionLogOutput(o interfaces.OhHellGame) string {
	return actionLogOutputTextWithNames(o, func(idx int) string { return cuiPlayerName(o.GetPlayer(idx), idx) })
}
