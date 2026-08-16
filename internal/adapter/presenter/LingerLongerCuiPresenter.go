//go:build !js || !wasm || extra

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// lingerLongerPlayerStr returns the display string for a single player.
func lingerLongerPlayerStr(s interfaces.LingerLongerGame, idx int, current bool) string {
	player := s.GetPlayer(idx)
	var b strings.Builder
	marker := " "
	if current {
		marker = ">"
	}
	role := ""
	switch {
	case player.IsEliminated():
		role = i18n.Tf("lingerlonger.roleOut", "rank", strconv.Itoa(player.GetEliminatedAt()))
	case idx == s.GetLastDrawIdx():
		role = i18n.T("lingerlonger.roleDrew")
	}
	b.WriteString(marker + i18n.Tf("lingerlonger.playerLine",
		"name", cuiPlayerName(player, idx),
		"role", role,
		"cards", strconv.Itoa(player.GetCardsSize()),
		"tricks", strconv.Itoa(player.GetTricksWon()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// LingerLongerCuiPresenter renders the Linger Longer CUI view.
type LingerLongerCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *LingerLongerCuiPresenter) Output(s interfaces.LingerLongerGame, lastErr error) string {
	return buildCuiOutput(i18n.T("lingerlonger.helpTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("lingerlonger.header",
			"trick", strconv.Itoa(s.GetTrickNumber()+1),
			"stock", strconv.Itoa(s.GetStockSize())) + "\n")
		// **補充が勝敗そのもの。** 規則を毎回書く。
		sb.WriteString(i18n.T("lingerlonger.rule") + "\n")

		// **山札が尽きたら誰も補充できない。** そこから一気に脱落が進みます。
		if s.GetStockSize() == 0 && !s.GetGameEndFlag() {
			sb.WriteString(color.Yellow(i18n.T("lingerlonger.noStockLine")) + "\n")
		}

		for i := 0; i < s.GetPlayerCnt(); i++ {
			sb.WriteString(lingerLongerPlayerStr(s, i,
				i == s.GetCurrentPlayerIdx() && !s.GetGameEndFlag()))
		}

		sb.WriteString("----------\n")

		cuiTrickBlock(sb, s.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(s.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(sb, lastErr)

		if s.GetGameEndFlag() {
			winner := s.GetWinnerIdx()
			name := cuiPlayerName(s.GetPlayer(winner), winner)
			sb.WriteString(color.Green(lingerLongerEndBanner(s.GetWinReason(), winner, name)) + "\n")
			return
		}

		// **人間が脱落しても盤面は続く。** そうと言わないと打てない理由が分からない。
		if human := s.GetPlayer(0); human != nil && human.IsEliminated() {
			sb.WriteString(i18n.T("lingerlonger.promptEliminated") + "\n")
			return
		}

		currentIdx := s.GetCurrentPlayerIdx()
		sb.WriteString(i18n.Tf("lingerlonger.promptCurrentPlayer",
			"name", cuiPlayerName(s.GetPlayer(currentIdx), currentIdx)) + "\n")
		sb.WriteString(i18n.T("lingerlonger.promptPlay") + "\n")
	})
}

// HintOutput emits the current hint.
func (p *LingerLongerCuiPresenter) HintOutput(s interfaces.LingerLongerGame) string {
	hint := s.GetHint()
	if hint == nil || hint.CardIndex == nil {
		return i18n.T("lingerlonger.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, lingerLongerHintReasonKeys)
	card := s.GetPlayer(s.GetCurrentPlayerIdx()).GetCard(*hint.CardIndex)
	return color.Yellow(i18n.Tf("lingerlonger.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", reason)) + "\n"
}

// lingerLongerHintReasonKeys maps hint-reason identifiers to their i18n keys.
var lingerLongerHintReasonKeys = map[string]string{
	"lingerlongerWinTrick": "lingerlonger.hintReasonWinTrick",
	"lingerlongerNoStock":  "lingerlonger.hintReasonNoStock",
	"lingerlongerDuck":     "lingerlonger.hintReasonDuck",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *LingerLongerCuiPresenter) ActionLogOutput(s interfaces.LingerLongerGame) string {
	return actionLogOutputText(s)
}

// lingerLongerEndBanner は決着の見出しを勝因に合わせて組み立てる。
//
// **勝因を取り違えると規則そのものを誤って説明することになる。** 山札が尽きて
// 全員が同時に手札 0 枚になった局では「最後まで持ち続けた人」は存在せず、勝ちは
// 最後のトリックで決まる。以前はどちらの勝ちでも同じ文言を出していた (#5765)。
// 未知の勝因は「持ち続けた」に寄せる -- 通常勝ちが圧倒的多数なので、そこへ倒す
// ほうが誤りが小さい。
func lingerLongerEndBanner(reason string, winnerIdx int, name string) string {
	switch reason {
	case domain.LingerLongerWinLastTrick:
		if winnerIdx == 0 {
			return i18n.T("lingerlonger.gameEndLastTrickYou")
		}
		return i18n.Tf("lingerlonger.gameEndLastTrickCpu", "name", name)
	case domain.LingerLongerWinGiveUp:
		return i18n.Tf("lingerlonger.gameEndGiveUp", "name", name)
	default:
		if winnerIdx == 0 {
			return i18n.T("lingerlonger.gameEndYou")
		}
		return i18n.Tf("lingerlonger.gameEndCpu", "name", name)
	}
}
