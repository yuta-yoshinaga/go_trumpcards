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

// teenPattiPlayerStr returns the display string for a single player.
func teenPattiPlayerStr(g interfaces.TeenPattiGame, i int) string {
	player := g.GetPlayer(i)
	if player == nil {
		return ""
	}
	reveal := g.IsShowdown()
	showCards := player.GetIsHuman() || (reveal && !player.GetFolded())

	var b strings.Builder
	status := teenPattiStatusStr(player)
	marker := ""
	if i == g.GetCurrentPlayerIdx() && !g.GetGameEndFlag() {
		marker = i18n.T("teenpatti.turnMarker")
	}
	b.WriteString(i18n.Tf("teenpatti.playerLine",
		"name", cuiPlayerName(player, i),
		"chips", strconv.Itoa(player.GetChips()),
		"bet", strconv.Itoa(player.GetRoundBet()),
		"status", status,
		"marker", marker,
	))
	b.WriteString("\n")
	if showCards && player.GetCardsSize() > 0 {
		line := cuiIndexedCardListStr(player)
		if reveal && player.GetCardsSize() == domain.TeenPattiHandSize {
			line += "  (" + i18n.T(teenPattiHandName(player)) + ")"
		}
		b.WriteString(line + "\n")
	}
	return b.String()
}

// teenPattiStatusStr returns the seen/folded/out status label for a player.
func teenPattiStatusStr(player *domain.TeenPattiPlayer) string {
	switch {
	case player.GetOut():
		return i18n.T("teenpatti.statusOut")
	case player.GetFolded():
		return i18n.T("teenpatti.statusFolded")
	case player.GetSeen():
		return i18n.T("teenpatti.statusSeen")
	default:
		return i18n.T("teenpatti.statusBlind")
	}
}

// teenPattiPlayerNameAt returns the human-readable name for the player at idx,
// falling back to the raw index when the player is missing (e.g. idx == -1).
func teenPattiPlayerNameAt(g interfaces.TeenPattiGame, idx int) string {
	player := g.GetPlayer(idx)
	if player == nil {
		return strconv.Itoa(idx)
	}
	return cuiPlayerName(player, idx)
}

// teenPattiRaiseRangeStr returns the betting-phase raise guidance line for a
// human turn: the affordable min..max, or an "unavailable" notice when the
// player is too short on chips. Empty on a CPU turn (mirrors Three Card Brag).
func teenPattiRaiseRangeStr(g interfaces.TeenPattiGame) string {
	player := g.GetPlayer(g.GetCurrentPlayerIdx())
	if player == nil || !player.GetIsHuman() {
		return ""
	}
	minRaise, maxRaise, ok := g.GetRaiseRange(g.GetCurrentPlayerIdx())
	if !ok {
		return i18n.T("teenpatti.promptRaiseUnavailable") + "\n"
	}
	return i18n.Tf("teenpatti.promptRaiseRange",
		"min", strconv.Itoa(minRaise),
		"max", strconv.Itoa(maxRaise),
	) + "\n"
}

// TeenPattiCuiPresenter renders the Teen Patti CUI view.
type TeenPattiCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *TeenPattiCuiPresenter) Output(g interfaces.TeenPattiGame, lastErr error) string {
	return buildCuiOutput(i18n.T("teenpatti.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("teenpatti.roundLine",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"pot", strconv.Itoa(g.GetPot()),
			"stake", strconv.Itoa(g.GetStake()),
		) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(teenPattiPlayerStr(g, i))
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			banner := i18n.Tf("teenpatti.gameEnd", "player", strconv.Itoa(g.GetMatchWinnerIdx()))
			b.WriteString(color.Green(banner) + "\n")
			return
		}

		switch g.GetPhase() {
		case domain.TeenPattiPhaseBetting:
			b.WriteString(i18n.T("teenpatti.promptBetting") + "\n")
			b.WriteString(i18n.T("teenpatti.promptBettingHelp") + "\n")
			b.WriteString(teenPattiRaiseRangeStr(g))
		case domain.TeenPattiPhaseSideShow:
			targetIdx := g.GetSideShowTarget()
			b.WriteString(i18n.Tf("teenpatti.promptSideShow",
				"requester", teenPattiPlayerNameAt(g, g.GetSideShowRequester()),
				"target", teenPattiPlayerNameAt(g, targetIdx),
			) + "\n")
			if target := g.GetPlayer(targetIdx); target != nil && target.GetIsHuman() {
				b.WriteString(i18n.T("teenpatti.promptSideShowYou") + "\n")
			}
			b.WriteString(i18n.T("teenpatti.promptSideShowHelp") + "\n")
		case domain.TeenPattiPhaseShowdown:
			b.WriteString(i18n.T("teenpatti.promptShowdown") + "\n")
		case domain.TeenPattiPhaseRoundEnd:
			winner := g.GetRoundWinnerIdx()
			b.WriteString(i18n.Tf("teenpatti.promptRoundEnd", "player", strconv.Itoa(winner)) + "\n")
			b.WriteString(i18n.T("teenpatti.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current hint.
func (p *TeenPattiCuiPresenter) HintOutput(g interfaces.TeenPattiGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("teenpatti.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, teenPattiHintReasonKeys)
	return color.Yellow(i18n.Tf("teenpatti.hint", "action", hint.Action, "reason", reason)) + "\n"
}

// teenPattiHintReasonKeys maps hint-reason identifiers to i18n keys.
var teenPattiHintReasonKeys = map[string]string{
	"see_first":   "teenpatti.hintReasonSeeFirst",
	"strong_hand": "teenpatti.hintReasonStrongHand",
	"medium_hand": "teenpatti.hintReasonMediumHand",
	"weak_hand":   "teenpatti.hintReasonWeakHand",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *TeenPattiCuiPresenter) ActionLogOutput(g interfaces.TeenPattiGame) string {
	return actionLogOutputText(g)
}
