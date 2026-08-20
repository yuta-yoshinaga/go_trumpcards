//go:build !js || !wasm || casino

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// tarneebPlayerStr returns the display string for a single Tarneeb player.
func tarneebPlayerStr(player *domain.TarneebPlayer, i int, playable []int) string {
	var b strings.Builder
	bidStr := i18n.T("tarneeb.bidPending")
	switch {
	case player.GetBid() == domain.TarneebPassBid:
		bidStr = i18n.T("tarneeb.bidPass")
	case player.GetBid() > 0:
		bidStr = strconv.Itoa(player.GetBid())
	}
	b.WriteString(i18n.Tf("tarneeb.playerLine",
		"name", cuiPlayerName(player, i),
		"team", strconv.Itoa(player.GetTeam()),
		"bid", bidStr,
		"tricks", strconv.Itoa(player.GetTrickCount()),
		"cards", strconv.Itoa(player.GetCardsSize()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		// **Web は validPlayIndices で出せない札を無効化しているのに、CUI は素の
		// 一覧だけだった** (#5606, CallBreak #5605 と同型)。playable が空のときは
		// 目印を付けない (制限が決まっていない状態と区別する)。
		b.WriteString(cuiPlayableMarkedCardListStr(player, playable) + "\n")
	}
	return b.String()
}

// tarneebTrumpSuitStr returns the localised trump-suit label, or a "not declared" placeholder.
func tarneebTrumpSuitStr(suit int) string {
	switch suit {
	case domain.CardDesignSpade:
		return "♠"
	case domain.CardDesignClover:
		return "♣"
	case domain.CardDesignHeart:
		return "♥"
	case domain.CardDesignDiamond:
		return "♦"
	default:
		return i18n.T("tarneeb.trumpUndeclared")
	}
}

// TarneebCuiPresenter renders the Tarneeb CUI view.
type TarneebCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *TarneebCuiPresenter) Output(t interfaces.TarneebGame, lastErr error) string {
	return buildCuiOutput(i18n.T("tarneeb.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("tarneeb.round",
			"round", strconv.Itoa(t.GetRoundNumber()),
			"trick", strconv.Itoa(t.GetTrickNumber())) + "\n")

		b.WriteString(i18n.Tf("tarneeb.trumpLine", "suit", tarneebTrumpSuitStr(t.GetTrumpSuit())) + "\n")
		b.WriteString(i18n.Tf("tarneeb.scoreLine",
			"team0", strconv.Itoa(t.GetTeamScore(0)),
			"team1", strconv.Itoa(t.GetTeamScore(1))) + "\n")
		if t.GetBidWinnerIdx() >= 0 {
			bw := t.GetPlayer(t.GetBidWinnerIdx())
			b.WriteString(i18n.Tf("tarneeb.bidWinnerLine",
				"name", cuiPlayerName(bw, t.GetBidWinnerIdx()),
				"bid", strconv.Itoa(t.GetHighestBid())) + "\n")
		}
		if t.GetRedealCount() > 0 {
			b.WriteString(i18n.Tf("tarneeb.redealLine", "count", strconv.Itoa(t.GetRedealCount())) + "\n")
		}

		for i := 0; i < t.GetPlayerCnt(); i++ {
			// 目印はプレイフェーズで本人の手番のときだけ。
			var playable []int
			if t.GetPhase() == domain.TarneebPhasePlay && t.GetCurrentPlayerIdx() == i {
				playable = t.GetValidPlayIndices(i)
			}
			b.WriteString(tarneebPlayerStr(t.GetPlayer(i), i, playable))
		}

		b.WriteString("----------\n")

		trick := t.GetCurrentTrick()
		cuiTrickBlock(b, trick,
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(t.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		if t.GetGameEndFlag() {
			banner := i18n.Tf("tarneeb.gameEnd", "team", strconv.Itoa(t.GetWinnerTeam()))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch t.GetPhase() {
		case domain.TarneebPhaseBid:
			bidIdx := t.GetBidPlayerIdx()
			b.WriteString(i18n.Tf("tarneeb.promptBid",
				"name", cuiPlayerName(t.GetPlayer(bidIdx), bidIdx)) + "\n")
			b.WriteString(i18n.T("tarneeb.promptBidHelp") + "\n")
		case domain.TarneebPhaseTrumpDeclaration:
			bw := t.GetBidWinnerIdx()
			b.WriteString(i18n.Tf("tarneeb.promptTrump",
				"name", cuiPlayerName(t.GetPlayer(bw), bw)) + "\n")
			b.WriteString(i18n.T("tarneeb.promptTrumpHelp") + "\n")
		case domain.TarneebPhasePlay:
			currentIdx := t.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("tarneeb.promptPlay",
				"name", cuiPlayerName(t.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("tarneeb.promptPlayHelp") + "\n")
		case domain.TarneebPhaseTrickEnd:
			b.WriteString(i18n.T("tarneeb.promptTrickEnd") + "\n")
			b.WriteString(i18n.T("tarneeb.promptTrickEndHelp") + "\n")
		case domain.TarneebPhaseRoundEnd:
			b.WriteString(i18n.T("tarneeb.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("tarneeb.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current Tarneeb hint.
func (p *TarneebCuiPresenter) HintOutput(t interfaces.TarneebGame) string {
	hint := t.GetHint()
	if hint == nil {
		return i18n.T("tarneeb.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, tarneebHintReasonKeys)
	if hint.Bid != nil {
		return color.Yellow(i18n.Tf("tarneeb.hintBid",
			"bid", strconv.Itoa(*hint.Bid),
			"reason", reason)) + "\n"
	}
	if hint.TrumpSuit != nil {
		return color.Yellow(i18n.Tf("tarneeb.hintTrump",
			"suit", tarneebTrumpSuitStr(*hint.TrumpSuit),
			"reason", reason)) + "\n"
	}
	if hint.CardIndex == nil {
		return i18n.T("tarneeb.hintNone") + "\n"
	}
	card := t.GetPlayer(0).GetCard(*hint.CardIndex)
	return color.Yellow(i18n.Tf("tarneeb.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", reason)) + "\n"
}

// tarneebHintReasonKeys maps Tarneeb-specific hint-reason identifiers to their i18n keys.
var tarneebHintReasonKeys = map[string]string{
	"bid_pass":      "tarneeb.hintReasonBidPass",
	"bid_estimate":  "tarneeb.hintReasonBidEstimate",
	"trump_longest": "tarneeb.hintReasonTrumpLongest",
	"trump_cut":     "tarneeb.hintReasonTrumpCut",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *TarneebCuiPresenter) ActionLogOutput(t interfaces.TarneebGame) string {
	return actionLogOutputTextForSeats[*domain.TarneebPlayer](t)
}
