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

// toepenPlayerStr は 1 人ぶんの表示行を返す。CPU の手札は枚数のみ。
func toepenPlayerStr(t interfaces.ToepenGame, idx int) string {
	player := t.GetPlayer(idx)
	var b strings.Builder
	status := ""
	switch {
	case t.IsEliminated(idx):
		status = " " + i18n.T("toepen.statusOut")
	case t.IsFolded(idx):
		status = " " + i18n.T("toepen.statusFolded")
	}
	b.WriteString(i18n.Tf("toepen.playerLine",
		"name", cuiPlayerName(player, idx),
		"lives", strconv.Itoa(t.GetLives(idx)),
		"max", strconv.Itoa(domain.ToepenMaxLives),
		"cards", strconv.Itoa(player.GetCardsSize())) + status + "\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// ToepenCuiPresenter renders the Toepen CUI view.
type ToepenCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *ToepenCuiPresenter) Output(t interfaces.ToepenGame, lastErr error) string {
	return buildCuiOutput(i18n.T("toepen.helpTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("toepen.header",
			"hand", strconv.Itoa(t.GetHandNumber()),
			"trick", strconv.Itoa(t.GetTrickNumber()),
			"stake", strconv.Itoa(t.GetStake())) + "\n")
		sb.WriteString(i18n.T("toepen.rankLine") + "\n")

		for i := range t.GetPlayers() {
			sb.WriteString(toepenPlayerStr(t, i))
		}

		sb.WriteString("----------\n")
		cuiTrickBlock(sb, t.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(t.GetPlayer(idx), idx) },
		)
		cuiErrorBlock(sb, lastErr)

		if t.GetGameEndFlag() {
			banner := i18n.T("toepen.gameEndLose")
			if t.GetWinnerIdx() == 0 {
				banner = i18n.T("toepen.gameEndWin")
			}
			sb.WriteString(color.Green(banner) + "\n")
			return
		}

		switch t.GetPhase() {
		case domain.ToepenPhaseHandEnd:
			sb.WriteString(i18n.T("toepen.promptNext") + "\n")
		case domain.ToepenPhaseRespond:
			sb.WriteString(i18n.Tf("toepen.respondLine",
				"name", cuiPlayerName(t.GetPlayer(t.GetKnockerIdx()), t.GetKnockerIdx()),
				"stake", strconv.Itoa(t.GetStake())) + "\n")
			sb.WriteString(i18n.T("toepen.promptRespond") + "\n")
		default:
			sb.WriteString(i18n.Tf("toepen.promptCurrentPlayer",
				"name", cuiPlayerName(t.GetPlayer(t.GetCurrentPlayerIdx()), t.GetCurrentPlayerIdx())) + "\n")
			sb.WriteString(i18n.T("toepen.promptPlay") + "\n")
			if t.CanRedeal(0) {
				sb.WriteString(i18n.T("toepen.promptRedeal") + "\n")
			}
		}
	})
}

// HintOutput emits the current Toepen hint.
func (p *ToepenCuiPresenter) HintOutput(t interfaces.ToepenGame) string {
	hint := toepenHint(t)
	key := toepenHintReasonKeys[hint.Reason]
	if key == "" {
		key = "toepen.hintNone"
	}
	if hint.CardIndex != nil {
		return color.Yellow(i18n.Tf("toepen.hintPlay",
			"idx", strconv.Itoa(*hint.CardIndex), "reason", i18n.T(key))) + "\n"
	}
	return color.Yellow(i18n.T(key)) + "\n"
}

// toepenHintReasonKeys maps the reason identifiers toepenHint returns to i18n
// keys. The Web presenter ships the identifier and the frontend resolves it;
// the CUI has to resolve it here or it prints the raw key at the player.
var toepenHintReasonKeys = map[string]string{
	"toepen.hint.game_end":      "toepen.hintReasonGameEnd",
	"toepen.hint.hand_end":      "toepen.hintReasonHandEnd",
	"toepen.hint.not_your_turn": "toepen.hintReasonNotYourTurn",
	"toepen.hint.stay":          "toepen.hintReasonStay",
	"toepen.hint.fold":          "toepen.hintReasonFold",
	"toepen.hint.play":          "toepen.hintReasonPlay",
	"toepen.hint.none":          "toepen.hintNone",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *ToepenCuiPresenter) ActionLogOutput(t interfaces.ToepenGame) string {
	return actionLogOutputTextForSeats[*domain.ToepenPlayer](t)
}
