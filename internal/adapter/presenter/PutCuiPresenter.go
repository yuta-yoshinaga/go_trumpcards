//go:build !js || !wasm || extra4

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// putLevelLabel ベッティングレベルの i18n 表示ラベルを返す。
func putLevelLabel(level int) string {
	switch level {
	case domain.PutLevelPut:
		return i18n.T("put.levelPut")
	case domain.PutLevelReput:
		return i18n.T("put.levelReput")
	case domain.PutLevelValeCuatro:
		return i18n.T("put.levelValeCuatro")
	default:
		return i18n.T("put.levelNone")
	}
}

// putPlayerStr returns the display string for a single Put player.
func putPlayerStr(player *domain.PutPlayer, idx int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("put.playerLine",
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

// PutCuiPresenter renders the Put CUI view.
type PutCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *PutCuiPresenter) Output(g interfaces.PutGame, lastErr error) string {
	return buildCuiOutput(i18n.T("put.helpTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("put.header",
			"hand", strconv.Itoa(g.GetHandNumber()),
			"baza", strconv.Itoa(g.GetTrickNumber()),
			"target", strconv.Itoa(g.GetMatchTarget())) + "\n")
		sb.WriteString(i18n.Tf("put.matchLine",
			"p0", strconv.Itoa(g.GetPlayerMatchPoints(0)),
			"p1", strconv.Itoa(g.GetPlayerMatchPoints(1))) + "\n")
		sb.WriteString(i18n.Tf("put.stakeLine",
			"stake", strconv.Itoa(g.GetHandStake()),
			"level", putLevelLabel(g.GetAcceptedLevel())) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			sb.WriteString(putPlayerStr(g.GetPlayer(i), i))
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
			key := "put.gameEndP1"
			if g.GetWinnerIdx() == 0 {
				key = "put.gameEndP0"
			}
			sb.WriteString(color.Green(i18n.Tf(key, "p0", strconv.Itoa(p0), "p1", strconv.Itoa(p1))) + "\n")
			return
		}

		switch g.GetPhase() {
		case domain.PutPhasePlay:
			idx := g.GetCurrentPlayerIdx()
			sb.WriteString(i18n.Tf("put.promptCurrentPlayer",
				"name", cuiPlayerName(g.GetPlayer(idx), idx)) + "\n")
			sb.WriteString(i18n.T("put.promptPlay") + "\n")
			if g.CanDeclarePut() {
				sb.WriteString(i18n.T("put.promptCanPut") + "\n")
			}
		case domain.PutPhaseRespond:
			caller := g.GetPutCallerIdx()
			sb.WriteString(i18n.Tf("put.promptRespond",
				"name", cuiPlayerName(g.GetPlayer(caller), caller),
				"level", putLevelLabel(g.GetPendingLevel())) + "\n")
		case domain.PutPhaseTrickEnd:
			sb.WriteString(i18n.T("put.promptTrickEnd") + "\n")
			sb.WriteString(i18n.T("put.promptNextHelp") + "\n")
		case domain.PutPhaseHandEnd:
			w := g.GetHandWinnerIdx()
			sb.WriteString(i18n.Tf("put.promptHandEnd",
				"name", cuiPlayerName(g.GetPlayer(w), w),
				"stake", strconv.Itoa(g.GetHandStake())) + "\n")
			sb.WriteString(i18n.T("put.promptNextHelp") + "\n")
		}
	})
}

// HintOutput emits the current Put hint.
func (p *PutCuiPresenter) HintOutput(g interfaces.PutGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("put.hintNone") + "\n"
	}
	switch hint.Action {
	case "accept":
		return color.Yellow(i18n.T("put.hintAccept")) + "\n"
	case "decline":
		return color.Yellow(i18n.T("put.hintDecline")) + "\n"
	case "call":
		return color.Yellow(i18n.T("put.hintCall")) + "\n"
	default:
		if hint.CardIndex == nil {
			return i18n.T("put.hintNone") + "\n"
		}
		card := g.GetPlayer(0).GetCard(*hint.CardIndex)
		return color.Yellow(i18n.Tf("put.hintCard",
			"idx", strconv.Itoa(*hint.CardIndex),
			"card", cuiCardStr(card),
			"reason", hintReasonStr(hint.Reason, putHintReasonKeys))) + "\n"
	}
}

// putHintReasonKeys maps Put-specific hint-reason identifiers to i18n keys.
var putHintReasonKeys = map[string]string{
	"leadStrong": "put.hintReasonLeadStrong",
	"leadLow":    "put.hintReasonLeadLow",
	"followWin":  "put.hintReasonFollowWin",
	"followDump": "put.hintReasonFollowDump",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *PutCuiPresenter) ActionLogOutput(g interfaces.PutGame) string {
	return actionLogOutputTextForSeats[*domain.PutPlayer](g)
}
