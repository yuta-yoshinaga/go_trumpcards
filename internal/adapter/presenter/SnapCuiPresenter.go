//go:build !js || !wasm || solo

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// SnapCuiPresenter renders the Snap CUI view.
type SnapCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *SnapCuiPresenter) Output(s interfaces.SnapGame, lastErr error) string {
	return buildCuiOutput(i18n.T("snap.helpTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("snap.header",
			"pile", strconv.Itoa(s.GetCenterPileSize())) + "\n")
		// **トリガーが動くことがこのゲームの規則そのもの。** 毎回書く。
		sb.WriteString(i18n.T("snap.rule") + "\n")

		if top := s.GetTopCard(); top != nil {
			sb.WriteString(i18n.Tf("snap.topLine", "card", cuiCardStr(top)) + "\n")
		} else {
			sb.WriteString(i18n.T("snap.pileEmpty") + "\n")
		}

		// **成立しているかどうかは一目で分かる必要がある。** 反射ゲームなので。
		if s.IsSnapAvailable() {
			sb.WriteString(color.Yellow(i18n.T("snap.availableLine")) + "\n")
		}

		for i := 0; i < s.GetPlayerCnt(); i++ {
			marker := " "
			if i == s.GetCurrentTurnIdx() && !s.GetGameEndFlag() {
				marker = ">"
			}
			sb.WriteString(marker + i18n.Tf("snap.playerLine",
				"name", cuiPlayerName(s.GetPlayer(i), i),
				"stock", strconv.Itoa(s.GetPlayer(i).GetStockSize())) + "\n")
		}

		sb.WriteString("----------\n")

		// 直近に何が起きたかを出す（反射ゲームなので、盤面だけでは読めない）。
		if line := snapEventLine(s); line != "" {
			sb.WriteString(line + "\n")
		}

		cuiErrorBlock(sb, lastErr)

		if s.GetGameEndFlag() {
			var banner string
			switch {
			case s.GetWinnerIdx() == 0:
				banner = i18n.T("snap.gameEndYou")
			case s.GetWinnerIdx() > 0:
				banner = i18n.Tf("snap.gameEndCpu",
					"name", cuiPlayerName(s.GetPlayer(s.GetWinnerIdx()), s.GetWinnerIdx()))
			default:
				banner = i18n.T("snap.gameEndNone")
			}
			sb.WriteString(color.Green(banner) + "\n")
			return
		}

		currentIdx := s.GetCurrentTurnIdx()
		sb.WriteString(i18n.Tf("snap.promptCurrentPlayer",
			"name", cuiPlayerName(s.GetPlayer(currentIdx), currentIdx)) + "\n")
		sb.WriteString(i18n.T("snap.promptPlay") + "\n")
	})
}

// snapEventLine は直近イベントの 1 行を返す（無ければ空）。
func snapEventLine(s interfaces.SnapGame) string {
	last := s.GetLastEvent()
	if last.Kind == domain.SnapEventNone {
		return ""
	}
	name := cuiPlayerName(s.GetPlayer(last.PlayerIdx), last.PlayerIdx)
	switch last.Kind {
	case domain.SnapEventSnapCorrect:
		return color.Green(i18n.Tf("snap.eventSnapCorrect", "name", name))
	case domain.SnapEventSnapWrong:
		return color.Red(i18n.Tf("snap.eventSnapWrong", "name", name))
	case domain.SnapEventEliminated:
		return i18n.Tf("snap.eventEliminated", "name", name)
	default:
		return i18n.Tf("snap.eventStep", "name", name)
	}
}

// HintOutput emits the current hint.
func (p *SnapCuiPresenter) HintOutput(s interfaces.SnapGame) string {
	hint := s.GetHint()
	if hint == nil {
		return i18n.T("snap.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, snapHintReasonKeys)
	if hint.Snap {
		return color.Yellow(i18n.Tf("snap.hintSnap", "reason", reason)) + "\n"
	}
	return color.Yellow(i18n.Tf("snap.hintWait", "reason", reason)) + "\n"
}

// snapHintReasonKeys maps hint-reason identifiers to their i18n keys.
var snapHintReasonKeys = map[string]string{
	"snapDeclare": "snap.hintReasonDeclare",
	"snapStep":    "snap.hintReasonStep",
	"snapWait":    "snap.hintReasonWait",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *SnapCuiPresenter) ActionLogOutput(s interfaces.SnapGame) string {
	return actionLogOutputText(s)
}
