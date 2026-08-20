package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// ninetyNinePlayerStr returns the display string for a single Ninety-Nine player.
func ninetyNinePlayerStr(player *domain.NinetyNinePlayer, i int) string {
	var b strings.Builder
	bidStr := i18n.T("ninetynine.bidPending")
	if player.GetBid() >= 0 {
		bidStr = strconv.Itoa(player.GetBid())
	}
	b.WriteString(i18n.Tf("ninetynine.playerLine",
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

// NinetyNineCuiPresenter renders the Ninety-Nine CUI view.
type NinetyNineCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *NinetyNineCuiPresenter) Output(o interfaces.NinetyNineGame, lastErr error) string {
	return buildCuiOutput(i18n.T("ninetynine.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("ninetynine.header",
			"deal", strconv.Itoa(o.GetDealNumber()),
			"target", strconv.Itoa(o.GetTargetScore()),
			"hand", strconv.Itoa(o.GetHandSize()),
			"trick", strconv.Itoa(o.GetTrickNumber())) + "\n")

		b.WriteString(i18n.Tf("ninetynine.trumpSuit", "suit", cuiSuitName(o.GetTrumpSuit())) + "\n")

		dealerIdx := o.GetDealerIdx()
		b.WriteString(i18n.Tf("ninetynine.dealer",
			"name", cuiPlayerName(o.GetPlayer(dealerIdx), dealerIdx)) + "\n")

		for i := 0; i < o.GetPlayerCnt(); i++ {
			b.WriteString(ninetyNinePlayerStr(o.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		trick := o.GetCurrentTrick()
		cuiTrickBlock(b, trick,
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(o.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		if o.GetGameEndFlag() {
			winnerIdx := o.GetWinnerIdx()
			player := o.GetPlayer(winnerIdx)
			banner := i18n.Tf("ninetynine.gameEnd", "name", cuiPlayerName(player, winnerIdx))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch o.GetPhase() {
		case domain.NinetyNinePhaseBid:
			bidIdx := o.GetBidPlayerIdx()
			name := cuiPlayerName(o.GetPlayer(bidIdx), bidIdx)
			b.WriteString(i18n.Tf("ninetynine.promptBid", "name", name) + "\n")
			b.WriteString(i18n.T("ninetynine.promptBidHelp") + "\n")
		case domain.NinetyNinePhasePlay:
			currentIdx := o.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("ninetynine.promptPlay",
				"name", cuiPlayerName(o.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("ninetynine.promptPlayHelp") + "\n")
		case domain.NinetyNinePhaseTrickEnd:
			b.WriteString(i18n.T("ninetynine.promptTrickEnd") + "\n")
			b.WriteString(i18n.T("ninetynine.promptTrickEndHelp") + "\n")
		case domain.NinetyNinePhaseRoundEnd:
			b.WriteString(i18n.T("ninetynine.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("ninetynine.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current Ninety-Nine hint.
func (p *NinetyNineCuiPresenter) HintOutput(o interfaces.NinetyNineGame) string {
	hint := o.GetHint()
	if hint == nil {
		return i18n.T("ninetynine.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, nil)
	if hint.BuryIndices != nil {
		strs := make([]string, len(hint.BuryIndices))
		for i, idx := range hint.BuryIndices {
			strs[i] = strconv.Itoa(idx)
		}
		return color.Yellow(i18n.Tf("ninetynine.hintBury",
			"indices", strings.Join(strs, " "),
			"reason", reason)) + "\n"
	}
	if hint.CardIndex == nil {
		return i18n.T("ninetynine.hintNone") + "\n"
	}
	humanIdx := -1
	for i := 0; i < o.GetPlayerCnt(); i++ {
		if o.GetPlayer(i).GetIsHuman() {
			humanIdx = i
			break
		}
	}
	if humanIdx < 0 {
		return i18n.T("ninetynine.hintNone") + "\n"
	}
	card := o.GetPlayer(humanIdx).GetCard(*hint.CardIndex)
	return color.Yellow(i18n.Tf("ninetynine.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", reason)) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *NinetyNineCuiPresenter) ActionLogOutput(o interfaces.NinetyNineGame) string {
	return actionLogOutputTextForSeats[*domain.NinetyNinePlayer](o)
}
