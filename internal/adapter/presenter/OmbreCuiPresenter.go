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

// ombreBidLabel maps a bid value to its i18n label key.
func ombreBidLabel(bid domain.OmbreBid) string {
	switch bid {
	case domain.OmbreBidEntrar:
		return i18n.T("ombre.bidEntrar")
	case domain.OmbreBidSolo:
		return i18n.T("ombre.bidSolo")
	default:
		return i18n.T("ombre.bidNone")
	}
}

// ombreTrumpLabel maps a trump suit value to its i18n label key.
func ombreTrumpLabel(suit int) string {
	switch suit {
	case domain.CardDesignSpade:
		return i18n.T("ombre.suitSpade")
	case domain.CardDesignClover:
		return i18n.T("ombre.suitClub")
	case domain.CardDesignHeart:
		return i18n.T("ombre.suitHeart")
	case domain.CardDesignDiamond:
		return i18n.T("ombre.suitDiamond")
	default:
		return i18n.T("ombre.suitNone")
	}
}

func ombrePlayerStr(g interfaces.OmbreGame, idx int) string {
	player := g.GetPlayer(idx)
	if player == nil {
		return ""
	}
	scores := g.GetPlayerScores()
	role := i18n.T("ombre.roleCoalition")
	if idx == g.GetOmbreIdx() {
		role = i18n.T("ombre.roleOmbre")
	}
	var b strings.Builder
	b.WriteString(i18n.Tf("ombre.playerLine",
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

// OmbreCuiPresenter renders the Ombre CUI view.
type OmbreCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *OmbreCuiPresenter) Output(g interfaces.OmbreGame, lastErr error) string {
	return buildCuiOutput(i18n.T("ombre.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("ombre.round",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber()),
			"trump", ombreTrumpLabel(g.GetTrumpSuit())) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(ombrePlayerStr(g, i))
		}

		b.WriteString("----------\n")

		cuiTrickBlock(b, g.GetCurrentTrick(),
			func(tc *domain.OmbreTrickCard) int { return tc.PlayerIdx },
			func(tc *domain.OmbreTrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winner := g.GetWinnerPlayer()
			var winnerStr string
			if winner >= 0 {
				winnerStr = cuiPlayerName(g.GetPlayer(winner), winner)
			}
			banner := i18n.Tf("ombre.gameEnd", "name", winnerStr)
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.OmbrePhaseBid:
			bidderIdx := g.GetCurrentBidderIdx()
			b.WriteString(i18n.Tf("ombre.promptBid",
				"bid", ombreBidLabel(g.GetWinningBid()),
				"name", cuiPlayerName(g.GetPlayer(bidderIdx), bidderIdx)) + "\n")
			b.WriteString(i18n.T("ombre.promptBidHelp") + "\n")
		case domain.OmbrePhasePlay:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("ombre.promptPlay",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx),
				"trump", ombreTrumpLabel(g.GetTrumpSuit())) + "\n")
			b.WriteString(i18n.T("ombre.promptPlayHelp") + "\n")
		case domain.OmbrePhaseTrickEnd:
			b.WriteString(i18n.T("ombre.promptTrickEnd") + "\n")
			b.WriteString(i18n.T("ombre.promptTrickEndHelp") + "\n")
		case domain.OmbrePhaseRoundEnd:
			b.WriteString(i18n.Tf("ombre.promptRoundEnd",
				"ombre", cuiPlayerName(g.GetPlayer(g.GetOmbreIdx()), g.GetOmbreIdx()),
				"outcome", ombreOutcomeLabel(g.GetOutcome())) + "\n")
			b.WriteString(i18n.T("ombre.promptRoundEndHelp") + "\n")
		}
	})
}

// ombreOutcomeLabel maps a deal outcome to its i18n label key.
func ombreOutcomeLabel(o domain.OmbreOutcome) string {
	switch o {
	case domain.OmbreOutcomeSacar:
		return i18n.T("ombre.outcomeSacar")
	case domain.OmbreOutcomePuesta:
		return i18n.T("ombre.outcomePuesta")
	case domain.OmbreOutcomeCodille:
		return i18n.T("ombre.outcomeCodille")
	default:
		return i18n.T("ombre.outcomeNone")
	}
}

// HintOutput emits the current Ombre hint.
func (p *OmbreCuiPresenter) HintOutput(g interfaces.OmbreGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("ombre.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, ombreHintReasonKeys)
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
		return color.Yellow(i18n.Tf("ombre.hintCard",
			"cards", strings.Join(cards, ", "),
			"reason", reason)) + "\n"
	}
	// Bid-phase decisions carry no cards; render them as an action recommendation
	// instead of the meaningless "recommended cards: -" line.
	if actionKey, ok := ombreBidActionKeys[hint.Reason]; ok {
		return color.Yellow(i18n.Tf("ombre.hintDecision",
			"action", i18n.T(actionKey),
			"reason", reason)) + "\n"
	}
	return color.Yellow(i18n.Tf("ombre.hintCard",
		"cards", "-",
		"reason", reason)) + "\n"
}

// ombreBidActionKeys maps bid-phase hint-reason identifiers to the i18n key for
// the recommended action name (Entrar / Solo / Pass).
var ombreBidActionKeys = map[string]string{
	"bid_entrar": "ombre.hintActionEntrar",
	"bid_solo":   "ombre.hintActionSolo",
	"bid_pass":   "ombre.hintActionPass",
}

// ombreHintReasonKeys maps Ombre-specific hint-reason identifiers to i18n keys.
var ombreHintReasonKeys = map[string]string{
	"lead_high":    "ombre.hintReasonLeadHigh",
	"lead_low":     "ombre.hintReasonLeadLow",
	"follow_win":   "ombre.hintReasonFollowWin",
	"follow_duck":  "ombre.hintReasonFollowDuck",
	"give_partner": "ombre.hintReasonGivePartner",
	"discard_low":  "ombre.hintReasonDiscardLow",
	"bid_entrar":   "ombre.hintReasonBidEntrar",
	"bid_solo":     "ombre.hintReasonBidSolo",
	"bid_pass":     "ombre.hintReasonBidPass",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *OmbreCuiPresenter) ActionLogOutput(g interfaces.OmbreGame) string {
	return actionLogOutputText(g)
}
