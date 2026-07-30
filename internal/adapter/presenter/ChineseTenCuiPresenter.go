//go:build !js || !wasm || extra2

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// chineseTenCardStr は 1 枚を描く。得点する赤札には点数を添える -- どの札を
// 取りに行く価値があるかが、このゲームで唯一の判断材料だから。
func chineseTenCardStr(c *domain.Card) string {
	if c == nil {
		return "--"
	}
	s := cuiCardStr(c)
	if pts := domain.ChineseTenCardPoints(c); pts > 0 {
		s += "(" + strconv.Itoa(pts) + ")"
	}
	return s
}

func chineseTenCardListStr(cards []*domain.Card, indexed bool) string {
	if len(cards) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(cards))
	for i, c := range cards {
		if indexed {
			parts = append(parts, "["+strconv.Itoa(i)+"]"+chineseTenCardStr(c))
			continue
		}
		parts = append(parts, chineseTenCardStr(c))
	}
	return strings.Join(parts, " ")
}

// ChineseTenCuiPresenter renders the Chinese Ten CUI view.
type ChineseTenCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *ChineseTenCuiPresenter) Output(c interfaces.ChineseTenGame, lastErr error) string {
	return buildCuiOutput(i18n.T("chineseten.helpTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("chineseten.header",
			"stock", strconv.Itoa(c.GetStockCount()),
			"tie", strconv.Itoa(domain.ChineseTenTieScore)) + "\n")
		sb.WriteString(i18n.T("chineseten.ruleLine") + "\n")
		sb.WriteString(i18n.Tf("chineseten.layoutLine",
			"cards", chineseTenCardListStr(c.GetLayout(), true)) + "\n")

		for i, player := range c.GetPlayers() {
			sb.WriteString(i18n.Tf("chineseten.playerLine",
				"name", cuiPlayerName(player, i),
				"score", strconv.Itoa(c.GetScore(i)),
				"cards", strconv.Itoa(player.GetCardsSize())) + "\n")
			// 取り札は公開情報。隠すと残り札の読みが成立しない。
			sb.WriteString("  " + i18n.Tf("chineseten.capturedLine",
				"cards", chineseTenCardListStr(c.GetCaptured(i), false)) + "\n")
			if player.GetIsHuman() && player.GetCardsSize() > 0 {
				hand := make([]*domain.Card, 0, player.GetCardsSize())
				for j := range player.GetCardsSize() {
					hand = append(hand, player.GetCard(j))
				}
				sb.WriteString("  " + chineseTenCardListStr(hand, true) + "\n")
			}
		}

		sb.WriteString("----------\n")
		cuiErrorBlock(sb, lastErr)

		if c.GetGameEndFlag() {
			var banner string
			switch c.GetWinnerIdx() {
			case 0:
				banner = i18n.T("chineseten.gameEndWin")
			case -1:
				banner = i18n.T("chineseten.gameEndDraw")
			default:
				banner = i18n.T("chineseten.gameEndLose")
			}
			sb.WriteString(color.Green(banner) + "\n")
			return
		}

		if c.GetPhase() == domain.ChineseTenPhaseSelect {
			sb.WriteString(i18n.Tf("chineseten.pendingLine",
				"card", chineseTenCardStr(c.GetPendingCard())) + "\n")
			sb.WriteString(i18n.T("chineseten.promptSelect") + "\n")
			return
		}
		sb.WriteString(i18n.T("chineseten.promptPlay") + "\n")
	})
}

// HintOutput emits the current Chinese Ten hint.
func (p *ChineseTenCuiPresenter) HintOutput(c interfaces.ChineseTenGame) string {
	hint := chineseTenHint(c)
	key := chineseTenHintReasonKeys[hint.Reason]
	if key == "" {
		key = "chineseten.hintNone"
	}
	switch {
	case hint.CardIndex != nil:
		return color.Yellow(i18n.Tf("chineseten.hintPlay",
			"idx", strconv.Itoa(*hint.CardIndex), "reason", i18n.T(key))) + "\n"
	case hint.LayoutIndex != nil:
		return color.Yellow(i18n.Tf("chineseten.hintSelect",
			"idx", strconv.Itoa(*hint.LayoutIndex), "reason", i18n.T(key))) + "\n"
	default:
		return color.Yellow(i18n.T(key)) + "\n"
	}
}

// chineseTenHintReasonKeys maps the reason identifiers chineseTenHint returns
// to i18n keys. The Web presenter ships the identifier and the frontend
// resolves it; the CUI must resolve it here or it prints the raw key.
var chineseTenHintReasonKeys = map[string]string{
	"chineseten.hint.game_end":      "chineseten.hintReasonGameEnd",
	"chineseten.hint.not_your_turn": "chineseten.hintReasonNotYourTurn",
	"chineseten.hint.select":        "chineseten.hintReasonSelect",
	"chineseten.hint.play":          "chineseten.hintReasonPlay",
	"chineseten.hint.none":          "chineseten.hintNone",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *ChineseTenCuiPresenter) ActionLogOutput(c interfaces.ChineseTenGame) string {
	return actionLogOutputText(c)
}
