package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// trucoLevelLabel ベッティングレベルの i18n 表示ラベルを返す。
func trucoLevelLabel(level int) string {
	switch level {
	case domain.TrucoLevelTruco:
		return i18n.T("truco.levelTruco")
	case domain.TrucoLevelRetruco:
		return i18n.T("truco.levelRetruco")
	case domain.TrucoLevelValeCuatro:
		return i18n.T("truco.levelValeCuatro")
	default:
		return i18n.T("truco.levelNone")
	}
}

// trucoPlayerStr returns the display string for a single Truco player.
func trucoPlayerStr(player *domain.TrucoPlayer, idx int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("truco.playerLine",
		"name", cuiPlayerName(player, idx),
		"tricks", strconv.Itoa(player.GetTrickCount()),
		"cards", strconv.Itoa(player.GetCardsSize()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// TrucoCuiPresenter renders the Truco CUI view.
type TrucoCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *TrucoCuiPresenter) Output(g interfaces.TrucoGame, lastErr error) string {
	return buildCuiOutput(i18n.T("truco.helpTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("truco.header",
			"hand", strconv.Itoa(g.GetHandNumber()),
			"baza", strconv.Itoa(g.GetTrickNumber()),
			"target", strconv.Itoa(g.GetMatchTarget())) + "\n")
		sb.WriteString(i18n.Tf("truco.matchLine",
			"p0", strconv.Itoa(g.GetPlayerMatchPoints(0)),
			"p1", strconv.Itoa(g.GetPlayerMatchPoints(1))) + "\n")
		sb.WriteString(i18n.Tf("truco.stakeLine",
			"stake", strconv.Itoa(g.GetHandStake()),
			"level", trucoLevelLabel(g.GetAcceptedLevel())) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			sb.WriteString(trucoPlayerStr(g.GetPlayer(i), i))
		}

		sb.WriteString("----------\n")

		cuiTrickBlock(sb, g.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(sb, lastErr)

		if g.GetGameEndFlag() {
			p0 := g.GetPlayerMatchPoints(0)
			p1 := g.GetPlayerMatchPoints(1)
			key := "truco.gameEndP1"
			if g.GetWinnerIdx() == 0 {
				key = "truco.gameEndP0"
			}
			sb.WriteString(color.Green(i18n.Tf(key, "p0", strconv.Itoa(p0), "p1", strconv.Itoa(p1))) + "\n")
			return
		}

		switch g.GetPhase() {
		case domain.TrucoPhasePlay:
			idx := g.GetCurrentPlayerIdx()
			sb.WriteString(i18n.Tf("truco.promptCurrentPlayer",
				"name", cuiPlayerName(g.GetPlayer(idx), idx)) + "\n")
			sb.WriteString(i18n.T("truco.promptPlay") + "\n")
			if g.CanDeclareTruco() {
				sb.WriteString(i18n.T("truco.promptCanTruco") + "\n")
			}
		case domain.TrucoPhaseRespond:
			caller := g.GetTrucoCallerIdx()
			sb.WriteString(i18n.Tf("truco.promptRespond",
				"name", cuiPlayerName(g.GetPlayer(caller), caller),
				"level", trucoLevelLabel(g.GetPendingLevel())) + "\n")
		case domain.TrucoPhaseTrickEnd:
			sb.WriteString(i18n.T("truco.promptTrickEnd") + "\n")
			sb.WriteString(i18n.T("truco.promptNextHelp") + "\n")
		case domain.TrucoPhaseHandEnd:
			w := g.GetHandWinnerIdx()
			sb.WriteString(i18n.Tf("truco.promptHandEnd",
				"name", cuiPlayerName(g.GetPlayer(w), w),
				"stake", strconv.Itoa(g.GetHandStake())) + "\n")
			sb.WriteString(i18n.T("truco.promptNextHelp") + "\n")
		}
	})
}

// HintOutput emits the current Truco hint.
func (p *TrucoCuiPresenter) HintOutput(g interfaces.TrucoGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("truco.hintNone") + "\n"
	}
	switch hint.Action {
	case "accept":
		return color.Yellow(i18n.T("truco.hintAccept")) + "\n"
	case "decline":
		return color.Yellow(i18n.T("truco.hintDecline")) + "\n"
	case "call":
		return color.Yellow(i18n.T("truco.hintCall")) + "\n"
	default:
		if hint.CardIndex == nil {
			return i18n.T("truco.hintNone") + "\n"
		}
		card := g.GetPlayer(0).GetCard(*hint.CardIndex)
		return color.Yellow(i18n.Tf("truco.hintCard",
			"idx", strconv.Itoa(*hint.CardIndex),
			"card", cuiCardStr(card),
			"reason", hintReasonStr(hint.Reason, trucoHintReasonKeys))) + "\n"
	}
}

// trucoHintReasonKeys maps Truco-specific hint-reason identifiers to i18n keys.
var trucoHintReasonKeys = map[string]string{
	"leadStrong": "truco.hintReasonLeadStrong",
	"leadLow":    "truco.hintReasonLeadLow",
	"followWin":  "truco.hintReasonFollowWin",
	"followDump": "truco.hintReasonFollowDump",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *TrucoCuiPresenter) ActionLogOutput(g interfaces.TrucoGame) string {
	return actionLogOutputTextForSeats[*domain.TrucoPlayer](g)
}
