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

// mushiCardStr は 1 枚を「月-番号(種別)」で描く。花札に専用 PNG は無く、CUI では
// 月と種別が読めれば足りる。
func mushiCardStr(c *domain.Card) string {
	if c == nil {
		return "--"
	}
	label := i18n.T("mushi.category." + strconv.Itoa(int(domain.MushiCardCategory(c))))
	s := i18n.Tf("mushi.cardLabel",
		"month", strconv.Itoa(c.GetDesign()),
		"index", strconv.Itoa(c.GetValue()),
		"category", label)
	if domain.MushiIsWild(c) {
		s += "*"
	}
	return s
}

func mushiCardListStr(cards []*domain.Card, indexed bool) string {
	parts := make([]string, 0, len(cards))
	for i, c := range cards {
		if indexed {
			parts = append(parts, "["+strconv.Itoa(i)+"]"+mushiCardStr(c))
			continue
		}
		parts = append(parts, mushiCardStr(c))
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, " ")
}

// MushiCuiPresenter renders the Mushi CUI view.
type MushiCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *MushiCuiPresenter) Output(m interfaces.MushiGame, lastErr error) string {
	return buildCuiOutput(i18n.T("mushi.helpTitle"), func(sb *strings.Builder) {
		cfg := m.GetConfig()
		sb.WriteString(i18n.Tf("mushi.header",
			"round", strconv.Itoa(m.GetRoundNumber()),
			"total", strconv.Itoa(cfg.TargetRounds),
			"stock", strconv.Itoa(m.GetStockCount())) + "\n")

		fieldCards := m.GetField()
		sb.WriteString(i18n.Tf("mushi.fieldLine", "cards", mushiCardListStr(fieldCards, true)) + "\n")

		for i, player := range m.GetPlayers() {
			captured := m.GetCaptured(i)
			pts := 0
			for _, c := range captured {
				pts += domain.MushiCardPoints(c)
			}
			sb.WriteString(i18n.Tf("mushi.playerLine",
				"name", cuiPlayerName(player, i),
				"score", strconv.Itoa(m.GetScore(i)),
				"points", strconv.Itoa(pts),
				"cards", strconv.Itoa(player.GetCardsSize())) + "\n")
			// 取り札は公開情報。伏せると役の読み合いが成立しない。
			sb.WriteString("  " + i18n.Tf("mushi.capturedLine",
				"cards", mushiCardListStr(captured, false)) + "\n")
			if player.GetIsHuman() && player.GetCardsSize() > 0 {
				hand := make([]*domain.Card, 0, player.GetCardsSize())
				for j := range player.GetCardsSize() {
					hand = append(hand, player.GetCard(j))
				}
				sb.WriteString("  " + mushiCardListStr(hand, true) + "\n")
			}
		}

		sb.WriteString("----------\n")
		cuiErrorBlock(sb, lastErr)

		if m.GetGameEndFlag() {
			var banner string
			switch m.GetWinnerIdx() {
			case 0:
				banner = i18n.T("mushi.gameEndWin")
			case -1:
				banner = i18n.T("mushi.gameEndDraw")
			default:
				banner = i18n.T("mushi.gameEndLose")
			}
			sb.WriteString(color.Green(banner) + "\n")
			return
		}

		switch m.GetPhase() {
		case domain.MushiPhaseRoundEnd:
			sb.WriteString(i18n.Tf("mushi.roundEnd",
				"delta", strconv.Itoa(m.GetRoundResult(0))) + "\n")
			sb.WriteString(i18n.T("mushi.promptNext") + "\n")
		case domain.MushiPhaseSelect, domain.MushiPhaseWildSelect:
			sb.WriteString(i18n.Tf("mushi.pendingLine",
				"card", mushiCardStr(m.GetPendingCard())) + "\n")
			sb.WriteString(i18n.T("mushi.promptSelect") + "\n")
		default:
			sb.WriteString(i18n.T("mushi.promptPlay") + "\n")
		}
	})
}

// HintOutput emits the current Mushi hint.
func (p *MushiCuiPresenter) HintOutput(m interfaces.MushiGame) string {
	hint := mushiHint(m)
	key := mushiHintReasonKeys[hint.Reason]
	if key == "" {
		key = "mushi.hintNone"
	}
	switch {
	case hint.CardIndex != nil:
		return color.Yellow(i18n.Tf("mushi.hintPlay",
			"idx", strconv.Itoa(*hint.CardIndex), "reason", i18n.T(key))) + "\n"
	case hint.FieldIndex != nil:
		return color.Yellow(i18n.Tf("mushi.hintSelect",
			"idx", strconv.Itoa(*hint.FieldIndex), "reason", i18n.T(key))) + "\n"
	default:
		return color.Yellow(i18n.T(key)) + "\n"
	}
}

// mushiHintReasonKeys maps the reason identifiers mushiHint returns to i18n
// keys. The Web presenter ships the identifier and the frontend resolves it;
// the CUI has to resolve it here.
var mushiHintReasonKeys = map[string]string{
	"mushi.hint.game_end":      "mushi.hintReasonGameEnd",
	"mushi.hint.round_end":     "mushi.hintReasonRoundEnd",
	"mushi.hint.not_your_turn": "mushi.hintReasonNotYourTurn",
	"mushi.hint.select":        "mushi.hintReasonSelect",
	"mushi.hint.play":          "mushi.hintReasonPlay",
	"mushi.hint.none":          "mushi.hintNone",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *MushiCuiPresenter) ActionLogOutput(m interfaces.MushiGame) string {
	return actionLogOutputText(m)
}
