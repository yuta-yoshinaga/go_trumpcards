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

// NiuNiuCuiPresenter 闘牛 CUI プレゼンタークラス
type NiuNiuCuiPresenter struct{}

// niuNiuHandLine 1 つの手を 1 行に描く。
//
// 牛を作った 3 枚に印をつけるのは、5 枚のどれが役になったのかが分からないと
// 結果を読めないため。「なぜこの格なのか」が一目で追える。
func niuNiuHandLine(n interfaces.NiuNiuGame, h *domain.NiuNiuHand, hide bool) string {
	cards := h.GetCards()
	if hide {
		parts := make([]string, len(cards))
		for i := range cards {
			parts[i] = i18n.T("niuniu.faceDown")
		}
		return strings.Join(parts, " ")
	}
	inCombo := make(map[int]bool, domain.NiuNiuComboSize)
	for _, i := range h.GetComboIdx() {
		inCombo[i] = true
	}
	parts := make([]string, len(cards))
	for i, c := range cards {
		s := cuiCardStr(c)
		if inCombo[i] {
			s = "*" + s
		}
		parts[i] = s
	}
	line := strings.Join(parts, " ") + " " + color.Bold(n.GetRankLabel(h.GetRank()))
	if m := n.GetMultiplier(h.GetRank()); m > 1 {
		line += i18n.Tf("niuniu.multiplierInline", "mult", strconv.Itoa(m))
	}
	return line
}

// Output ゲーム状態を出力
func (np *NiuNiuCuiPresenter) Output(n interfaces.NiuNiuGame, lastErr error) string {
	var sb strings.Builder

	sb.WriteString("----------\n")
	sb.WriteString(i18n.Tf("niuniu.chipsLine", "chips", strconv.Itoa(n.GetChips())) + "\n")

	ended := n.GetGameEndFlag()
	if bh := n.GetBankerHand(); bh != nil {
		sb.WriteString(i18n.T("niuniu.bankerHandHeader") + " " +
			niuNiuHandLine(n, bh, !ended) + "\n")
	}

	sb.WriteString("----------\n")
	for i, s := range n.GetSeats() {
		if i == n.GetBankerIdx() || s.GetHand() == nil {
			continue
		}
		h := s.GetHand()
		sb.WriteString("  " + s.GetName() + " " +
			i18n.Tf("niuniu.betInline", "bet", strconv.Itoa(h.GetBet())) + " " +
			niuNiuHandLine(n, h, !ended && s.IsCPU()))
		if ended && h.GetPayout() != 0 {
			sb.WriteString(" " + i18n.Tf("niuniu.payoutInline",
				"payout", strconv.Itoa(h.GetPayout())))
		}
		sb.WriteString("\n")
	}
	sb.WriteString("----------\n")

	cuiErrorBlock(&sb, lastErr)

	if ended {
		sb.WriteString(color.Green(n.GetLastResult()) + "\n")
		sb.WriteString(i18n.T("niuniu.nextRound") + "\n")
	} else {
		sb.WriteString(i18n.T("niuniu.placeBet") + "\n")
	}

	return sb.String()
}

// ActionLogOutput 棋譜をテキスト出力
func (np *NiuNiuCuiPresenter) ActionLogOutput(n interfaces.NiuNiuGame) string {
	if !n.GetGameEndFlag() {
		return actionLogToText(nil)
	}
	return actionLogToText(n.GetActionLog())
}
