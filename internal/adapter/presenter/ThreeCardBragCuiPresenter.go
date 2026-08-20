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

// threeCardBragPlayerStr returns the display string for a single player.
func threeCardBragPlayerStr(g interfaces.ThreeCardBragGame, i int) string {
	player := g.GetPlayer(i)
	if player == nil {
		return ""
	}
	reveal := g.IsShowdown()
	showCards := player.GetIsHuman() || (reveal && !player.GetFolded())

	var b strings.Builder
	status := threeCardBragStatusStr(player)
	marker := ""
	if i == g.GetCurrentPlayerIdx() && !g.GetGameEndFlag() {
		marker = i18n.T("threecardbrag.turnMarker")
	}
	b.WriteString(i18n.Tf("threecardbrag.playerLine",
		"name", cuiPlayerName(player, i),
		"chips", strconv.Itoa(player.GetChips()),
		"bet", strconv.Itoa(player.GetRoundBet()),
		"status", status,
		"marker", marker,
	))
	b.WriteString("\n")
	if showCards && player.GetCardsSize() > 0 {
		line := cuiIndexedCardListStr(player)
		if reveal && player.GetCardsSize() == domain.ThreeCardBragHandSize {
			line += "  (" + i18n.T(threeCardBragHandName(player)) + ")"
		}
		b.WriteString(line + "\n")
	}
	return b.String()
}

// threeCardBragStatusStr returns the seen/folded/out status label for a player.
func threeCardBragStatusStr(player *domain.ThreeCardBragPlayer) string {
	switch {
	case player.GetOut():
		return i18n.T("threecardbrag.statusOut")
	case player.GetFolded():
		return i18n.T("threecardbrag.statusFolded")
	case player.GetSeen():
		return i18n.T("threecardbrag.statusSeen")
	default:
		return i18n.T("threecardbrag.statusBlind")
	}
}

// threeCardBragRaiseRangeStr returns the betting-phase raise guidance line for
// the human at turn: the allowed stake range (min = stake+1, max = the largest
// stake the player can afford given the seen/blind call multiplier), or an
// "unavailable" notice when chips are too low. Returns "" when it is not the
// human's turn (a raise prompt would be meaningless for CPU turns).
func threeCardBragRaiseRangeStr(g interfaces.ThreeCardBragGame) string {
	player := g.GetPlayer(g.GetCurrentPlayerIdx())
	if player == nil || !player.GetIsHuman() {
		return ""
	}
	minRaise := g.GetStake() + 1
	maxRaise := player.GetChips()
	if player.GetSeen() {
		// Seen players pay double the stake to call/raise, halving the affordable ceiling.
		maxRaise = player.GetChips() / 2
	}
	if maxRaise < minRaise {
		return i18n.T("threecardbrag.promptRaiseUnavailable") + "\n"
	}
	return i18n.Tf("threecardbrag.promptRaiseRange",
		"min", strconv.Itoa(minRaise),
		"max", strconv.Itoa(maxRaise),
	) + "\n"
}

// ThreeCardBragCuiPresenter renders the Three Card Brag CUI view.
type ThreeCardBragCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *ThreeCardBragCuiPresenter) Output(g interfaces.ThreeCardBragGame, lastErr error) string {
	return buildCuiOutput(i18n.T("threecardbrag.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("threecardbrag.roundLine",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"pot", strconv.Itoa(g.GetPot()),
			"stake", strconv.Itoa(g.GetStake()),
		) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(threeCardBragPlayerStr(g, i))
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			// **勝者だけ席番号のままだった** (#5659)。他のゲームは全部
			// cuiPlayerName で「あなた」「CPU 1」と出しているので、ここだけ
			// 「Player 0 の勝ち」に見えていた。
			winnerIdx := g.GetMatchWinnerIdx()
			banner := i18n.Tf("threecardbrag.gameEnd",
				"player", cuiPlayerName(g.GetPlayer(winnerIdx), winnerIdx))
			b.WriteString(color.Green(banner) + "\n")
			return
		}

		switch g.GetPhase() {
		case domain.ThreeCardBragPhaseBetting:
			b.WriteString(i18n.T("threecardbrag.promptBetting") + "\n")
			b.WriteString(i18n.T("threecardbrag.promptBettingHelp") + "\n")
			b.WriteString(threeCardBragRaiseRangeStr(g))
		case domain.ThreeCardBragPhaseShowdown:
			b.WriteString(i18n.T("threecardbrag.promptShowdown") + "\n")
		case domain.ThreeCardBragPhaseRoundEnd:
			winner := g.GetRoundWinnerIdx()
			b.WriteString(i18n.Tf("threecardbrag.promptRoundEnd",
				"player", cuiPlayerName(g.GetPlayer(winner), winner)) + "\n")
			b.WriteString(i18n.T("threecardbrag.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current hint.
func (p *ThreeCardBragCuiPresenter) HintOutput(g interfaces.ThreeCardBragGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("threecardbrag.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, threeCardBragHintReasonKeys)
	return color.Yellow(i18n.Tf("threecardbrag.hint", "action", hint.Action, "reason", reason)) + "\n"
}

// threeCardBragHintReasonKeys maps hint-reason identifiers to i18n keys.
var threeCardBragHintReasonKeys = map[string]string{
	"see_first":   "threecardbrag.hintReasonSeeFirst",
	"strong_hand": "threecardbrag.hintReasonStrongHand",
	"medium_hand": "threecardbrag.hintReasonMediumHand",
	"weak_hand":   "threecardbrag.hintReasonWeakHand",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *ThreeCardBragCuiPresenter) ActionLogOutput(g interfaces.ThreeCardBragGame) string {
	return actionLogOutputText(g)
}
