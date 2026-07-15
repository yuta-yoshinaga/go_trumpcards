package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// spadesPlayerStr returns the display string for a single Spades player.
func spadesPlayerStr(player *domain.SpadesPlayer, i, bagThreshold int) string {
	var b strings.Builder
	bidStr := i18n.T("spades.bidPending")
	if player.GetBid() >= 0 {
		bidStr = strconv.Itoa(player.GetBid())
	}
	// Highlight the bag count as it nears the penalty threshold (each threshold
	// of accumulated bags costs 100 points), red within 1 and yellow within 2.
	bagsStr := strconv.Itoa(player.GetBags())
	if bagThreshold > 0 {
		switch remaining := bagThreshold - player.GetBags(); {
		case remaining <= 1:
			bagsStr = color.Red(bagsStr)
		case remaining <= 2:
			bagsStr = color.Yellow(bagsStr)
		}
	}
	b.WriteString(i18n.Tf("spades.playerLine",
		"name", cuiPlayerName(player, i),
		"bid", bidStr,
		"tricks", strconv.Itoa(player.GetTrickCount()),
		"bags", bagsStr,
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

// SpadesCuiPresenter renders the Spades CUI view.
type SpadesCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *SpadesCuiPresenter) Output(s interfaces.SpadesGame, lastErr error) string {
	return buildCuiOutput(i18n.T("spades.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("spades.round",
			"round", strconv.Itoa(s.GetRoundNumber()),
			"trick", strconv.Itoa(s.GetTrickNumber())) + "\n")

		if s.GetSpadesBroken() {
			b.WriteString(i18n.T("spades.spadesBrokenYes") + "\n")
		} else {
			b.WriteString(i18n.T("spades.spadesBrokenNo") + "\n")
		}

		// Point-limit progress: whoever reaches the limit first wins, so surface
		// the limit and the current leader (highest cumulative score).
		cfg := s.GetConfig()
		leaderIdx, maxScore := 0, s.GetPlayer(0).GetCumulativeScore()
		for i := 1; i < s.GetPlayerCnt(); i++ {
			if score := s.GetPlayer(i).GetCumulativeScore(); score > maxScore {
				maxScore, leaderIdx = score, i
			}
		}
		b.WriteString(i18n.Tf("spades.limitProgress",
			"limit", strconv.Itoa(cfg.PointLimit),
			"name", cuiPlayerName(s.GetPlayer(leaderIdx), leaderIdx),
			"score", strconv.Itoa(maxScore)) + "\n")

		for i := 0; i < s.GetPlayerCnt(); i++ {
			b.WriteString(spadesPlayerStr(s.GetPlayer(i), i, cfg.BagPenaltyThreshold))
		}

		b.WriteString("----------\n")

		// Current trick
		trick := s.GetCurrentTrick()
		cuiTrickBlock(b, trick,
			func(tc *domain.SpadesTrickCard) int { return tc.PlayerIdx },
			func(tc *domain.SpadesTrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(s.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		// Game state
		if s.GetGameEndFlag() {
			winnerIdx := s.GetWinnerIdx()
			player := s.GetPlayer(winnerIdx)
			banner := i18n.Tf("spades.gameEnd", "name", cuiPlayerName(player, winnerIdx))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch s.GetPhase() {
		case domain.SpadesPhaseBid:
			bidIdx := s.GetBidPlayerIdx()
			b.WriteString(i18n.Tf("spades.promptBid",
				"name", cuiPlayerName(s.GetPlayer(bidIdx), bidIdx)) + "\n")
			b.WriteString(i18n.T("spades.promptBidHelp") + "\n")
		case domain.SpadesPhasePlay:
			currentIdx := s.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("spades.promptPlay",
				"name", cuiPlayerName(s.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("spades.promptPlayHelp") + "\n")
		case domain.SpadesPhaseTrickEnd:
			b.WriteString(i18n.T("spades.promptTrickEnd") + "\n")
			b.WriteString(i18n.T("spades.promptTrickEndHelp") + "\n")
		case domain.SpadesPhaseRoundEnd:
			b.WriteString(i18n.T("spades.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("spades.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current Spades hint.
func (p *SpadesCuiPresenter) HintOutput(s interfaces.SpadesGame) string {
	hint := s.GetHint()
	if hint == nil {
		return i18n.T("spades.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, spadesHintReasonKeys)
	if hint.Bid != nil {
		return color.Yellow(i18n.Tf("spades.hintBid",
			"bid", strconv.Itoa(*hint.Bid),
			"reason", reason)) + "\n"
	}
	if hint.CardIndex == nil {
		return i18n.T("spades.hintNone") + "\n"
	}
	card := s.GetPlayer(0).GetCard(*hint.CardIndex)
	return color.Yellow(i18n.Tf("spades.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", reason)) + "\n"
}

// spadesHintReasonKeys maps Spades-specific hint-reason identifiers to
// their i18n keys. Reasons not in this map fall through to
// hintReasonStr → cui_common.
var spadesHintReasonKeys = map[string]string{
	"trump_cut": "spades.hintReasonTrumpCut",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *SpadesCuiPresenter) ActionLogOutput(s interfaces.SpadesGame) string {
	return actionLogOutputText(s)
}
