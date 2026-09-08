//go:build !js || !wasm || extra5

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// omiPlayerStr returns the display string for a single Omi player.
// playable が非 nil のとき、そのインデックスの札に "*" を付ける。
func omiPlayerStr(player *domain.OmiPlayer, i int, playable []int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("omi.playerLine",
		"name", cuiPlayerName(player, i),
		"team", strconv.Itoa(player.GetTeam()),
		"tricks", strconv.Itoa(player.GetTrickCount()),
		"cards", strconv.Itoa(player.GetCardsSize()),
	) + "\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(omiHandStr(player, playable) + "\n")
	}
	return b.String()
}

// omiHandStr renders the hand as an indexed list, starring the cards that
// may legally be played right now.
func omiHandStr(player *domain.OmiPlayer, playable []int) string {
	if len(playable) == 0 {
		return cuiIndexedCardListStr(player)
	}
	mark := make(map[int]bool, len(playable))
	for _, idx := range playable {
		mark[idx] = true
	}
	parts := make([]string, player.GetCardsSize())
	for i := 0; i < player.GetCardsSize(); i++ {
		parts[i] = "[" + strconv.Itoa(i) + "]" + cuiCardStr(player.GetCard(i))
		if mark[i] {
			parts[i] += "*"
		}
	}
	return strings.Join(parts, "  ")
}

// omiPlayableIndices returns the human's legal plays, or nil when it is not
// the human's turn to play a card.
func omiPlayableIndices(e interfaces.OmiGame) []int {
	if e.GetPhase() != domain.OmiPhasePlay {
		return nil
	}
	idx := e.GetCurrentPlayerIdx()
	p := e.GetPlayer(idx)
	if p == nil || !p.GetIsHuman() {
		return nil
	}
	return e.GetValidPlayIndices(idx)
}

// OmiCuiPresenter renders the Omi CUI view.
type OmiCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *OmiCuiPresenter) Output(e interfaces.OmiGame, lastErr error) string {
	return buildCuiOutput(i18n.T("omi.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("omi.header",
			"round", strconv.Itoa(e.GetRoundNumber()),
			"trick", strconv.Itoa(e.GetTrickNumber())) + "\n")
		dealerIdx := e.GetDealerIdx()
		b.WriteString(i18n.Tf("omi.dealer",
			"name", cuiPlayerName(e.GetPlayer(dealerIdx), dealerIdx)) + "\n")

		callerIdx := e.GetTrumpCallerIdx()
		b.WriteString(i18n.Tf("omi.trumpCaller",
			"name", cuiPlayerName(e.GetPlayer(callerIdx), callerIdx)) + "\n")

		if trumpSuit := e.GetTrumpSuit(); trumpSuit > 0 {
			b.WriteString(i18n.Tf("omi.trumpLine",
				"suit", cuiSuitName(trumpSuit),
				"caller", cuiPlayerName(e.GetPlayer(callerIdx), callerIdx),
				"team", strconv.Itoa(e.GetMakerTeam())) + "\n")
		} else {
			b.WriteString(i18n.T("omi.trumpUndecided") + "\n")
		}

		if e.GetPhase() == domain.OmiPhaseCallTrump {
			b.WriteString(i18n.T("omi.dealStageFirst") + "\n")
		} else {
			b.WriteString(i18n.T("omi.dealStageSecond") + "\n")
		}

		b.WriteString(i18n.T("omi.scoringRule") + "\n")

		tricks0 := 0
		tricks1 := 0
		for i := 0; i < e.GetPlayerCnt(); i++ {
			if pl := e.GetPlayer(i); pl != nil {
				if pl.GetTeam() == 0 {
					tricks0 += pl.GetTrickCount()
				} else if pl.GetTeam() == 1 {
					tricks1 += pl.GetTrickCount()
				}
			}
		}
		b.WriteString(i18n.Tf("omi.teamScoreLine",
			"t0", strconv.Itoa(e.GetTeamScore(0)),
			"tricks0", strconv.Itoa(tricks0),
			"t1", strconv.Itoa(e.GetTeamScore(1)),
			"tricks1", strconv.Itoa(tricks1)) + "\n")

		playable := omiPlayableIndices(e)
		for i := 0; i < e.GetPlayerCnt(); i++ {
			b.WriteString(omiPlayerStr(e.GetPlayer(i), i, playable))
		}

		b.WriteString("----------\n")

		// Current trick
		trick := e.GetCurrentTrick()
		cuiTrickBlock(b, trick,
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(e.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		// Game state
		if e.GetGameEndFlag() {
			banner := i18n.Tf("omi.gameEnd", "team", strconv.Itoa(e.GetWinnerTeam()))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch e.GetPhase() {
		case domain.OmiPhaseCallTrump:
			b.WriteString(i18n.Tf("omi.promptCallTrump",
				"name", cuiPlayerName(e.GetPlayer(callerIdx), callerIdx)) + "\n")
			b.WriteString(i18n.T("omi.promptCallTrumpHelp") + "\n")
		case domain.OmiPhasePlay:
			currentIdx := e.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("omi.promptCurrentPlayer",
				"name", cuiPlayerName(e.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("omi.promptPlayHelp") + "\n")
		case domain.OmiPhaseTrickEnd:
			b.WriteString(i18n.T("omi.promptTrickEnd") + "\n")
			b.WriteString(i18n.T("omi.promptTrickEndHelp") + "\n")
		case domain.OmiPhaseRoundEnd:
			b.WriteString(i18n.T("omi.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("omi.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current Omi hint.
func (p *OmiCuiPresenter) HintOutput(e interfaces.OmiGame) string {
	hint := e.GetHint()
	if hint == nil {
		return i18n.T("omi.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, omiHintReasonKeys)
	if hint.Suit != nil {
		return color.Yellow(i18n.Tf("omi.hintCallSuit",
			"suit", cuiSuitName(*hint.Suit),
			"reason", reason)) + "\n"
	}
	if hint.CardIndex == nil {
		return i18n.T("omi.hintNone") + "\n"
	}
	player := e.GetPlayer(0)
	card := player.GetCard(*hint.CardIndex)
	return color.Yellow(i18n.Tf("omi.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", reason)) + "\n"
}

// omiHintReasonKeys maps Omi-specific hint-reason identifiers to their
// i18n keys. Reasons not listed here fall through to cui_common via
// hintReasonStr.
var omiHintReasonKeys = map[string]string{
	"strategic_call": "omi.hintReasonStrategicCall",
	"normal_play":    "omi.hintReasonNormalPlay",
	"lead_trump":     "omi.hintReasonLeadTrump",
	"lead_strong":    "omi.hintReasonLeadStrong",
	"follow_suit":    "omi.hintReasonFollowSuit",
	"trump_cut":      "omi.hintReasonTrumpCut",
	"discard_weak":   "omi.hintReasonDiscardWeak",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *OmiCuiPresenter) ActionLogOutput(e interfaces.OmiGame) string {
	return actionLogOutputTextForSeats[*domain.OmiPlayer](e)
}
