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

// courtPieceTeamLabels maps a team index (0/1) to its display name.
var courtPieceTeamLabels = [domain.CourtPieceTeamCnt]string{"A", "B"}

// courtPieceTeamLabel returns the team label (A/B) for a team index, or "?" when out of range.
func courtPieceTeamLabel(team int) string {
	if team < 0 || team >= len(courtPieceTeamLabels) {
		return "?"
	}
	return courtPieceTeamLabels[team]
}

// courtPiecePlayerStr returns the display string for a single Court Piece player.
func courtPiecePlayerStr(player *domain.CourtPiecePlayer, i int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("courtpiece.playerLine",
		"name", cuiPlayerName(player, i),
		"team", courtPieceTeamLabel(player.GetTeam()),
		"tricks", strconv.Itoa(player.GetTrickCount()),
		"cards", strconv.Itoa(player.GetCardsSize()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// courtPieceTrumpSuitStr returns the localised trump-suit label, or a "not declared" placeholder.
func courtPieceTrumpSuitStr(suit int) string {
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
		return i18n.T("courtpiece.trumpUndeclared")
	}
}

// CourtPieceCuiPresenter renders the Court Piece CUI view.
type CourtPieceCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *CourtPieceCuiPresenter) Output(t interfaces.CourtPieceGame, lastErr error) string {
	return buildCuiOutput(i18n.T("courtpiece.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("courtpiece.round",
			"round", strconv.Itoa(t.GetRoundNumber()),
			"trick", strconv.Itoa(t.GetTrickNumber())) + "\n")

		b.WriteString(i18n.Tf("courtpiece.trumpLine", "suit", courtPieceTrumpSuitStr(t.GetTrumpSuit())) + "\n")
		b.WriteString(i18n.Tf("courtpiece.scoreLine",
			"team0", strconv.Itoa(t.GetTeamScore(0)),
			"team1", strconv.Itoa(t.GetTeamScore(1))) + "\n")
		callerIdx := t.GetCallerIdx()
		if callerIdx >= 0 {
			caller := t.GetPlayer(callerIdx)
			b.WriteString(i18n.Tf("courtpiece.callerLine",
				"name", cuiPlayerName(caller, callerIdx),
				"team", courtPieceTeamLabel(caller.GetTeam())) + "\n")
		}

		for i := 0; i < t.GetPlayerCnt(); i++ {
			b.WriteString(courtPiecePlayerStr(t.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		trick := t.GetCurrentTrick()
		cuiTrickBlock(b, trick,
			func(tc *domain.CourtPieceTrickCard) int { return tc.PlayerIdx },
			func(tc *domain.CourtPieceTrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(t.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		if t.GetGameEndFlag() {
			banner := i18n.Tf("courtpiece.gameEnd", "team", courtPieceTeamLabel(t.GetWinnerTeam()))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch t.GetPhase() {
		case domain.CourtPiecePhaseTrumpDeclaration:
			b.WriteString(i18n.Tf("courtpiece.promptTrump",
				"name", cuiPlayerName(t.GetPlayer(callerIdx), callerIdx)) + "\n")
			b.WriteString(i18n.T("courtpiece.promptTrumpHelp") + "\n")
		case domain.CourtPiecePhasePlay:
			currentIdx := t.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("courtpiece.promptPlay",
				"name", cuiPlayerName(t.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("courtpiece.promptPlayHelp") + "\n")
		case domain.CourtPiecePhaseTrickEnd:
			b.WriteString(i18n.T("courtpiece.promptTrickEnd") + "\n")
			b.WriteString(i18n.T("courtpiece.promptTrickEndHelp") + "\n")
		case domain.CourtPiecePhaseRoundEnd:
			b.WriteString(i18n.T("courtpiece.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("courtpiece.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current Court Piece hint.
func (p *CourtPieceCuiPresenter) HintOutput(t interfaces.CourtPieceGame) string {
	hint := t.GetHint()
	if hint == nil {
		return i18n.T("courtpiece.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, courtPieceHintReasonKeys)
	if hint.TrumpSuit != nil {
		return color.Yellow(i18n.Tf("courtpiece.hintTrump",
			"suit", courtPieceTrumpSuitStr(*hint.TrumpSuit),
			"reason", reason)) + "\n"
	}
	if hint.CardIndex == nil {
		return i18n.T("courtpiece.hintNone") + "\n"
	}
	card := t.GetPlayer(0).GetCard(*hint.CardIndex)
	return color.Yellow(i18n.Tf("courtpiece.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", reason)) + "\n"
}

// courtPieceHintReasonKeys maps Court Piece-specific hint-reason identifiers to their i18n keys.
var courtPieceHintReasonKeys = map[string]string{
	"trump_longest": "courtpiece.hintReasonTrumpLongest",
	"lead_strong":   "courtpiece.hintReasonLeadStrong",
	"follow_suit":   "courtpiece.hintReasonFollowSuit",
	"trump_cut":     "courtpiece.hintReasonTrumpCut",
	"discard_high":  "courtpiece.hintReasonDiscardHigh",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *CourtPieceCuiPresenter) ActionLogOutput(t interfaces.CourtPieceGame) string {
	return actionLogOutputText(t)
}
