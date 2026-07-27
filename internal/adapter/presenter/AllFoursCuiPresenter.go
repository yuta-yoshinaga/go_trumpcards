package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// allFoursSuitName renders the trump suit using a Unicode glyph. Returns the
// localized "(undecided)" placeholder when no trump has been chosen yet.
func allFoursSuitName(suit int) string {
	switch suit {
	case domain.CardDesignSpade:
		return "♠"
	case domain.CardDesignClover:
		return "♣"
	case domain.CardDesignHeart:
		return "♥"
	case domain.CardDesignDiamond:
		return "♦"
	}
	return i18n.T("allfours.trumpUndecided")
}

// allFoursPlayerStr returns the display string for a single All Fours player.
func allFoursPlayerStr(player *domain.AllFoursPlayer, i int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("allfours.playerLine",
		"name", cuiPlayerName(player, i),
		"tricks", strconv.Itoa(player.GetTrickCount()),
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

// AllFoursCuiPresenter renders the All Fours CUI view.
type AllFoursCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *AllFoursCuiPresenter) Output(s interfaces.AllFoursGame, lastErr error) string {
	return buildCuiOutput(i18n.T("allfours.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("allfours.header",
			"round", strconv.Itoa(s.GetRoundNumber()),
			"trick", strconv.Itoa(s.GetTrickNumber()),
			"dealer", cuiPlayerName(s.GetPlayer(s.GetDealerIdx()), s.GetDealerIdx())) + "\n")
		turnUp := "-"
		if s.GetTurnUp() != nil {
			turnUp = cuiCardStr(s.GetTurnUp())
		}
		b.WriteString(i18n.Tf("allfours.trumpLine",
			"suit", allFoursSuitName(s.GetTrumpSuit()),
			"turnup", turnUp) + "\n")

		for i := 0; i < s.GetPlayerCnt(); i++ {
			b.WriteString(allFoursPlayerStr(s.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		trick := s.GetCurrentTrick()
		cuiTrickBlock(b, trick,
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(s.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		if s.GetGameEndFlag() {
			winnerIdx := s.GetWinnerIdx()
			banner := i18n.Tf("allfours.gameEnd",
				"name", cuiPlayerName(s.GetPlayer(winnerIdx), winnerIdx))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch s.GetPhase() {
		case domain.AllFoursPhaseBeg:
			b.WriteString(i18n.T("allfours.promptBeg") + "\n")
			b.WriteString(i18n.T("allfours.promptBegHelp") + "\n")
		case domain.AllFoursPhaseGift:
			b.WriteString(i18n.T("allfours.promptGift") + "\n")
			b.WriteString(i18n.T("allfours.promptGiftHelp") + "\n")
		case domain.AllFoursPhasePlay:
			currentIdx := s.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("allfours.promptPlay",
				"name", cuiPlayerName(s.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("allfours.promptPlayHelp") + "\n")
		case domain.AllFoursPhaseTrickEnd:
			b.WriteString(i18n.T("allfours.promptTrickEnd") + "\n")
			b.WriteString(i18n.T("allfours.promptTrickEndHelp") + "\n")
		case domain.AllFoursPhaseRoundEnd:
			b.WriteString(i18n.T("allfours.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("allfours.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current All Fours hint.
func (p *AllFoursCuiPresenter) HintOutput(s interfaces.AllFoursGame) string {
	hint := s.GetHint()
	if hint == nil {
		return i18n.T("allfours.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, allFoursHintReasonKeys)
	if hint.Beg != nil {
		action := i18n.T("allfours.hintStand")
		if *hint.Beg {
			action = i18n.T("allfours.hintBeg")
		}
		return color.Yellow(i18n.Tf("allfours.hintDecision", "action", action, "reason", reason)) + "\n"
	}
	if hint.Run != nil {
		action := i18n.T("allfours.hintGift")
		if *hint.Run {
			action = i18n.T("allfours.hintRun")
		}
		return color.Yellow(i18n.Tf("allfours.hintDecision", "action", action, "reason", reason)) + "\n"
	}
	if hint.CardIndex == nil {
		return i18n.T("allfours.hintNone") + "\n"
	}
	humanIdx := -1
	for i := 0; i < s.GetPlayerCnt(); i++ {
		if pl := s.GetPlayer(i); pl != nil && pl.GetIsHuman() {
			humanIdx = i
			break
		}
	}
	if humanIdx < 0 {
		return i18n.T("allfours.hintNone") + "\n"
	}
	card := s.GetPlayer(humanIdx).GetCard(*hint.CardIndex)
	return color.Yellow(i18n.Tf("allfours.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", reason)) + "\n"
}

// allFoursHintReasonKeys maps All Fours-specific hint-reason identifiers to
// their i18n keys. Reasons not in this map fall through to hintReasonStr →
// cui_common.
var allFoursHintReasonKeys = map[string]string{
	"trump_cut":   "allfours.hintReasonTrumpCut",
	"lead_strong": "allfours.hintReasonLeadStrong",
	"beg_beg":     "allfours.hintReasonBegBeg",
	"beg_stand":   "allfours.hintReasonBegStand",
	"gift_gift":   "allfours.hintReasonGiftGift",
	"gift_run":    "allfours.hintReasonGiftRun",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *AllFoursCuiPresenter) ActionLogOutput(s interfaces.AllFoursGame) string {
	return actionLogOutputText(s)
}
