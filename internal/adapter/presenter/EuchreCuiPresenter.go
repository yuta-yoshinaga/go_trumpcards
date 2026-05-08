package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// euchrePlayerStr returns the display string for a single Euchre player.
func euchrePlayerStr(player *domain.EuchrePlayer, i int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("euchre.playerLine",
		"name", cuiPlayerName(player, i),
		"team", strconv.Itoa(player.GetTeam()),
		"tricks", strconv.Itoa(player.GetTrickCount()),
		"cards", strconv.Itoa(player.GetCardsSize()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// EuchreCuiPresenter renders the Euchre CUI view.
type EuchreCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *EuchreCuiPresenter) Output(e interfaces.EuchreGame, lastErr error) string {
	return buildCuiOutput(i18n.T("euchre.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("euchre.header",
			"round", strconv.Itoa(e.GetRoundNumber()),
			"trick", strconv.Itoa(e.GetTrickNumber())) + "\n")
		dealerIdx := e.GetDealerIdx()
		b.WriteString(i18n.Tf("euchre.dealer",
			"name", cuiPlayerName(e.GetPlayer(dealerIdx), dealerIdx)) + "\n")

		if trumpSuit := e.GetTrumpSuit(); trumpSuit > 0 {
			b.WriteString(i18n.Tf("euchre.trumpLine",
				"suit", cuiSuitName(trumpSuit),
				"team", strconv.Itoa(e.GetMakerTeam())) + "\n")
		} else {
			b.WriteString(i18n.T("euchre.trumpUndecided") + "\n")
		}

		if faceUpCard := e.GetFaceUpCard(); faceUpCard != nil {
			b.WriteString(i18n.Tf("euchre.faceUpCard", "card", cuiCardStr(faceUpCard)) + "\n")
		}

		if e.GetGoingAlone() {
			aloneIdx := e.GetGoingAlonePlayerIdx()
			b.WriteString(i18n.Tf("euchre.goingAlone",
				"name", cuiPlayerName(e.GetPlayer(aloneIdx), aloneIdx)) + "\n")
		}

		b.WriteString(i18n.Tf("euchre.teamScoreLine",
			"t0", strconv.Itoa(e.GetTeamScore(0)),
			"t1", strconv.Itoa(e.GetTeamScore(1))) + "\n")

		for i := 0; i < e.GetPlayerCnt(); i++ {
			b.WriteString(euchrePlayerStr(e.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		// Current trick
		trick := e.GetCurrentTrick()
		cuiTrickBlock(b, trick,
			func(tc *domain.EuchreTrickCard) int { return tc.PlayerIdx },
			func(tc *domain.EuchreTrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(e.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		// Game state
		if e.GetGameEndFlag() {
			banner := i18n.Tf("euchre.gameEnd", "team", strconv.Itoa(e.GetWinnerTeam()))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch e.GetPhase() {
		case domain.EuchrePhasePickUp:
			bidIdx := e.GetBidPlayerIdx()
			b.WriteString(i18n.Tf("euchre.promptPickup",
				"name", cuiPlayerName(e.GetPlayer(bidIdx), bidIdx)) + "\n")
			b.WriteString(i18n.T("euchre.promptPickupHelp") + "\n")
		case domain.EuchrePhaseCallTrump:
			bidIdx := e.GetBidPlayerIdx()
			b.WriteString(i18n.Tf("euchre.promptCallTrump",
				"name", cuiPlayerName(e.GetPlayer(bidIdx), bidIdx)) + "\n")
			b.WriteString(i18n.T("euchre.promptCallTrumpHelp") + "\n")
		case domain.EuchrePhaseDiscard:
			b.WriteString(i18n.T("euchre.promptDiscard") + "\n")
			b.WriteString(i18n.T("euchre.promptDiscardHelp") + "\n")
		case domain.EuchrePhasePlay:
			currentIdx := e.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("euchre.promptCurrentPlayer",
				"name", cuiPlayerName(e.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("euchre.promptPlayHelp") + "\n")
		case domain.EuchrePhaseTrickEnd:
			b.WriteString(i18n.T("euchre.promptTrickEnd") + "\n")
			b.WriteString(i18n.T("euchre.promptTrickEndHelp") + "\n")
		case domain.EuchrePhaseRoundEnd:
			b.WriteString(i18n.T("euchre.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("euchre.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current Euchre hint.
func (p *EuchreCuiPresenter) HintOutput(e interfaces.EuchreGame) string {
	hint := e.GetHint()
	if hint == nil {
		return i18n.T("euchre.hintNone") + "\n"
	}
	reason := euchreHintReasonStr(hint.Reason)
	if hint.OrderUp != nil {
		if *hint.OrderUp {
			key := "euchre.hintOrderUp"
			if hint.GoAlone != nil && *hint.GoAlone {
				key = "euchre.hintOrderUpAlone"
			}
			return color.Yellow(i18n.Tf(key, "reason", reason)) + "\n"
		}
		return color.Yellow(i18n.Tf("euchre.hintPass", "reason", reason)) + "\n"
	}
	if hint.Suit != nil {
		key := "euchre.hintCallSuit"
		if hint.GoAlone != nil && *hint.GoAlone {
			key = "euchre.hintCallSuitAlone"
		}
		return color.Yellow(i18n.Tf(key,
			"suit", cuiSuitName(*hint.Suit),
			"reason", reason)) + "\n"
	}
	if hint.CardIndex == nil {
		return i18n.T("euchre.hintNone") + "\n"
	}
	player := e.GetPlayer(0)
	card := player.GetCard(*hint.CardIndex)
	return color.Yellow(i18n.Tf("euchre.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", reason)) + "\n"
}

// euchreHintReasonKeys maps Euchre-specific hint-reason identifiers to their
// i18n keys. Reasons not listed here fall through to cui_common via
// lookupHintReason.
var euchreHintReasonKeys = map[string]string{
	"discard_weakest": "euchre.hintReasonDiscardWeakest",
}

// euchreHintReasonStr resolves a reason via the per-game map first, then the
// shared (cui_common) layer.
func euchreHintReasonStr(reason string) string {
	if key, ok := euchreHintReasonKeys[reason]; ok {
		return i18n.T(key)
	}
	return lookupHintReason(reason, nil)
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *EuchreCuiPresenter) ActionLogOutput(e interfaces.EuchreGame) string {
	return actionLogOutputText(e)
}
