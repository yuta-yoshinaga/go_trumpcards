package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// twoTenJackTeamCnt is the number of partnerships: seats alternate, so a seat's
// team is its index parity. twoTenJackPlayerStr labels each row with the same
// i%2, and the totals below fold on it, so the rows and the totals cannot
// disagree about who is on which side.
const twoTenJackTeamCnt = 2

// twoTenJackPlayerStr returns the display string for a single TwoTenJack player.
func twoTenJackPlayerStr(player *domain.TwoTenJackPlayer, i int) string {
	var b strings.Builder
	team := i % twoTenJackTeamCnt
	b.WriteString(i18n.Tf("twotenjack.playerLine",
		"name", cuiPlayerName(player, i),
		"team", strconv.Itoa(team),
		"tricks", strconv.Itoa(player.GetTrickCount()),
		"points", strconv.Itoa(player.GetCapturedPointCards()),
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

// twoTenJackSuitLabel returns a human-readable suit label.
func twoTenJackSuitLabel(suit int) string {
	switch suit {
	case domain.CardDesignSpade:
		return "SPADE"
	case domain.CardDesignClover:
		return "CLUB"
	case domain.CardDesignHeart:
		return "HEART"
	case domain.CardDesignDiamond:
		return "DIAMOND"
	}
	return i18n.T("twotenjack.trumpUnset")
}

// TwoTenJackCuiPresenter renders the Two Ten Jack CUI view.
type TwoTenJackCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *TwoTenJackCuiPresenter) Output(s interfaces.TwoTenJackGame, lastErr error) string {
	return buildCuiOutput(i18n.T("twotenjack.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("twotenjack.header",
			"round", strconv.Itoa(s.GetRoundNumber()),
			"trick", strconv.Itoa(s.GetTrickNumber())) + "\n")
		b.WriteString(i18n.Tf("twotenjack.trumpLine",
			"suit", twoTenJackSuitLabel(s.GetTrumpSuit()),
			"declarer", strconv.Itoa(s.GetDeclarerIdx())) + "\n")

		var teamCum, teamPoints [twoTenJackTeamCnt]int
		for i := 0; i < s.GetPlayerCnt(); i++ {
			pl := s.GetPlayer(i)
			b.WriteString(twoTenJackPlayerStr(pl, i))
			team := i % twoTenJackTeamCnt
			teamCum[team] += pl.GetCumulativeScore()
			teamPoints[team] += pl.GetCapturedPointCards()
		}

		for team := range twoTenJackTeamCnt {
			b.WriteString(i18n.Tf("twotenjack.teamLine",
				"team", strconv.Itoa(team),
				"points", strconv.Itoa(teamPoints[team]),
				"cum", strconv.Itoa(teamCum[team])) + "\n")
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
			banner := i18n.Tf("twotenjack.gameEnd", "team", strconv.Itoa(s.GetWinnerTeam()))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch s.GetPhase() {
		case domain.TwoTenJackPhaseDeclare:
			declIdx := s.GetDeclarerIdx()
			b.WriteString(i18n.Tf("twotenjack.promptDeclare",
				"name", cuiPlayerName(s.GetPlayer(declIdx), declIdx)) + "\n")
			b.WriteString(i18n.T("twotenjack.promptDeclareHelp") + "\n")
		case domain.TwoTenJackPhasePlay:
			currentIdx := s.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("twotenjack.promptCurrentPlayer",
				"name", cuiPlayerName(s.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("twotenjack.promptPlay") + "\n")
		case domain.TwoTenJackPhaseTrickEnd:
			b.WriteString(i18n.T("twotenjack.promptTrickEnd") + "\n")
			b.WriteString(i18n.T("twotenjack.promptTrickEndHelp") + "\n")
		case domain.TwoTenJackPhaseRoundEnd:
			b.WriteString(i18n.T("twotenjack.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("twotenjack.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current Two Ten Jack hint.
func (p *TwoTenJackCuiPresenter) HintOutput(s interfaces.TwoTenJackGame) string {
	hint := s.GetHint()
	if hint == nil {
		return i18n.T("twotenjack.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, twoTenJackHintReasonKeys)
	if hint.TrumpSuit != nil {
		return color.Yellow(i18n.Tf("twotenjack.hintTrump",
			"suit", twoTenJackSuitLabel(*hint.TrumpSuit),
			"reason", reason)) + "\n"
	}
	if hint.CardIndex == nil {
		return i18n.T("twotenjack.hintNone") + "\n"
	}
	player := s.GetPlayer(0)
	card := player.GetCard(*hint.CardIndex)
	return color.Yellow(i18n.Tf("twotenjack.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", reason)) + "\n"
}

// twoTenJackHintReasonKeys maps Two Ten Jack-specific hint-reason identifiers
// to their i18n keys. Reasons not listed here fall through to cui_common via
// hintReasonStr.
var twoTenJackHintReasonKeys = map[string]string{
	"strategic_trump": "twotenjack.hintReasonStrategicTrump",
	"lead":            "twotenjack.hintReasonLead",
	"follow_suit":     "twotenjack.hintReasonFollowSuit",
	"trump_cut":       "twotenjack.hintReasonTrumpCut",
	"discard":         "twotenjack.hintReasonDiscard",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *TwoTenJackCuiPresenter) ActionLogOutput(s interfaces.TwoTenJackGame) string {
	return actionLogOutputTextForSeats[*domain.TwoTenJackPlayer](s)
}
