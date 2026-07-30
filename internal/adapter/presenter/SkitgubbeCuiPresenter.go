//go:build !js || !wasm || extra3

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func skitgubbeCardListStr(cards []*domain.Card, indexed bool) string {
	if len(cards) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(cards))
	for i, c := range cards {
		if indexed {
			parts = append(parts, "["+strconv.Itoa(i)+"]"+cuiCardStr(c))
			continue
		}
		parts = append(parts, cuiCardStr(c))
	}
	return strings.Join(parts, " ")
}

// skitgubbeTrumpStr は切札を描く。山札から最後に引かれるまでは未確定。
func skitgubbeTrumpStr(suit int) string {
	if suit < 0 {
		return i18n.T("skitgubbe.trumpUndecided")
	}
	return cuiSuitName(suit)
}

// SkitgubbeCuiPresenter renders the Skitgubbe CUI view.
type SkitgubbeCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *SkitgubbeCuiPresenter) Output(c interfaces.SkitgubbeGame, lastErr error) string {
	return buildCuiOutput(i18n.T("skitgubbe.helpTitle"), func(sb *strings.Builder) {
		collect := c.GetPhase() == domain.SkitgubbePhaseCollect
		phaseKey := "skitgubbe.phaseShed"
		if collect {
			phaseKey = "skitgubbe.phaseCollect"
		}
		sb.WriteString(i18n.Tf("skitgubbe.header",
			"phase", i18n.T(phaseKey),
			"stock", strconv.Itoa(c.GetStockCount()),
			"trump", skitgubbeTrumpStr(c.GetTrumpSuit())) + "\n")
		sb.WriteString(i18n.T("skitgubbe.ruleLine") + "\n")

		if collect {
			sb.WriteString(i18n.Tf("skitgubbe.duelLine",
				"cards", skitgubbeCardListStr(c.GetDuel(), false)) + "\n")
		} else {
			sb.WriteString(i18n.Tf("skitgubbe.pileLine",
				"cards", skitgubbeCardListStr(c.GetPile(), false)) + "\n")
		}

		for i, player := range c.GetPlayers() {
			sb.WriteString(i18n.Tf("skitgubbe.playerLine",
				"name", cuiPlayerName(player, i),
				"cards", strconv.Itoa(player.GetCardsSize()),
				"collected", strconv.Itoa(c.GetCollectedCount(i))) + "\n")
			if player.GetIsHuman() && player.GetCardsSize() > 0 {
				hand := make([]*domain.Card, 0, player.GetCardsSize())
				for j := range player.GetCardsSize() {
					hand = append(hand, player.GetCard(j))
				}
				sb.WriteString("  " + skitgubbeCardListStr(hand, true) + "\n")
			}
		}

		sb.WriteString("----------\n")
		cuiErrorBlock(sb, lastErr)

		if c.GetGameEndFlag() {
			banner := i18n.T("skitgubbe.gameEndWin")
			if c.GetLoserIdx() == 0 {
				banner = i18n.T("skitgubbe.gameEndLose")
			}
			sb.WriteString(color.Green(banner) + "\n")
			return
		}

		// 引き取れるときだけ pickup を促す。出せる札があるうちは引き取れない。
		if !collect && len(c.GetPile()) > 0 && len(c.GetValidPlayIndices(0)) == 0 &&
			c.GetCurrentPlayerIdx() == 0 {
			sb.WriteString(i18n.T("skitgubbe.promptPickUp") + "\n")
			return
		}
		sb.WriteString(i18n.T("skitgubbe.promptPlay") + "\n")
	})
}

// HintOutput emits the current Skitgubbe hint.
func (p *SkitgubbeCuiPresenter) HintOutput(c interfaces.SkitgubbeGame) string {
	hint := skitgubbeHint(c)
	key := skitgubbeHintReasonKeys[hint.Reason]
	if key == "" {
		key = "skitgubbe.hintNone"
	}
	switch {
	case hint.PickUp:
		return color.Yellow(i18n.Tf("skitgubbe.hintPickUp", "reason", i18n.T(key))) + "\n"
	case hint.CardIndex != nil:
		return color.Yellow(i18n.Tf("skitgubbe.hintPlay",
			"idx", strconv.Itoa(*hint.CardIndex), "reason", i18n.T(key))) + "\n"
	default:
		return color.Yellow(i18n.T(key)) + "\n"
	}
}

// skitgubbeHintReasonKeys maps the reason identifiers skitgubbeHint returns to
// i18n keys. The Web presenter ships the identifier and the frontend resolves
// it; the CUI must resolve it here or it prints the raw key.
var skitgubbeHintReasonKeys = map[string]string{
	"skitgubbe.hint.game_end":      "skitgubbe.hintReasonGameEnd",
	"skitgubbe.hint.not_your_turn": "skitgubbe.hintReasonNotYourTurn",
	"skitgubbe.hint.duel":          "skitgubbe.hintReasonDuel",
	"skitgubbe.hint.beat":          "skitgubbe.hintReasonBeat",
	"skitgubbe.hint.pickup":        "skitgubbe.hintReasonPickUp",
	"skitgubbe.hint.none":          "skitgubbe.hintNone",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *SkitgubbeCuiPresenter) ActionLogOutput(c interfaces.SkitgubbeGame) string {
	return actionLogOutputText(c)
}
