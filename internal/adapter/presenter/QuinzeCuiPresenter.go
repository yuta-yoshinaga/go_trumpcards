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

// QuinzeCuiPresenter カーンズ CUI プレゼンタークラス
type QuinzeCuiPresenter struct{}

// quinzeHandLine 1 つの手を 1 行に描く
func quinzeHandLine(s interfaces.QuinzeGame, h *domain.QuinzeHand, hide bool) string {
	cards := h.GetCards()
	parts := make([]string, len(cards))
	for i, c := range cards {
		if hide {
			parts[i] = i18n.T("quinze.faceDown")
			continue
		}
		parts[i] = cuiCardStr(c)
	}
	line := strings.Join(parts, " ")
	if hide {
		return line
	}
	line += " " + i18n.Tf("quinze.totalInline", "total", s.FormatPoints(s.GetHandPoints(h)))
	return line
}

// Output ゲーム状態を出力
func (sp *QuinzeCuiPresenter) Output(s interfaces.QuinzeGame, lastErr error) string {
	var sb strings.Builder

	sb.WriteString("----------\n")
	sb.WriteString(i18n.Tf("quinze.chipsLine", "chips", strconv.Itoa(s.GetChips())) + "\n")
	bankerName := i18n.T("quinze.bankerIsYou")
	if !s.IsHumanBanker() {
		bankerName = s.GetSeats()[s.GetBankerIdx()].GetName()
	}
	sb.WriteString(i18n.Tf("quinze.bankerLine", "name", bankerName) + "\n")

	ended := s.GetGameEndFlag()
	if bh := s.GetBankerHand(); bh != nil {
		hide := !ended && !s.IsHumanBanker()
		sb.WriteString(i18n.T("quinze.bankerHandHeader") + " " +
			quinzeHandLine(s, bh, hide) + "\n")
	}

	sb.WriteString("----------\n")
	for i, seat := range s.GetSeats() {
		if i == s.GetBankerIdx() || seat.GetHand() == nil {
			continue
		}
		h := seat.GetHand()
		marker := "  "
		if i == s.GetActiveSeat() && s.GetPhase() == domain.QuinzePhasePlayerTurn {
			marker = "> "
		}
		hide := !ended && seat.IsCPU()
		sb.WriteString(marker + seat.GetName() + " " +
			i18n.Tf("quinze.betInline", "bet", strconv.Itoa(h.GetBet())) + " " +
			quinzeHandLine(s, h, hide))
		if ended && h.GetPayout() != 0 {
			sb.WriteString(" " + i18n.Tf("quinze.payoutInline",
				"payout", strconv.Itoa(h.GetPayout())))
		}
		sb.WriteString("\n")
	}
	sb.WriteString("----------\n")

	cuiErrorBlock(&sb, lastErr)

	switch s.GetPhase() {
	case domain.QuinzePhaseBet:
		if s.IsHumanBanker() {
			sb.WriteString(color.Yellow(i18n.T("quinze.dealAsBanker")) + "\n")
		} else {
			sb.WriteString(i18n.T("quinze.placeBet") + "\n")
		}
	case domain.QuinzePhasePlayerTurn:
		sb.WriteString(sp.actionHints(s))
	case domain.QuinzePhaseBankerTurn:
		sb.WriteString(color.Yellow(i18n.T("quinze.bankerTurn")) + "\n")
	case domain.QuinzePhaseEnd:
		sb.WriteString(color.Green(s.GetLastResult()) + "\n")
		if nb := s.GetNextBanker(); nb >= 0 {
			sb.WriteString(color.Yellow(i18n.Tf("quinze.bankPasses",
				"name", s.GetSeats()[nb].GetName())) + "\n")
		}
	}

	return sb.String()
}

// actionHints 今打てる手だけを並べる
func (sp *QuinzeCuiPresenter) actionHints(s interfaces.QuinzeGame) string {
	var opts []string
	if s.CanHit() {
		opts = append(opts, i18n.T("quinze.optHit"))
	}
	if s.CanStand() {
		opts = append(opts, i18n.T("quinze.optStand"))
	}
	if len(opts) == 0 {
		return ""
	}
	// **相手がいつ引くのをやめるかは、賭け続けるかの判断材料。**その数字は
	// どの画面にも出ていなかった (#5566)。点は FormatPoints に通す ── 半点単位の
	// 内部表現をそのまま出すと 11 と読めてしまう。
	return i18n.Tf("quinze.actionsLine", "options", strings.Join(opts, " / ")) + "\n" +
		color.Yellow(i18n.Tf("quinze.cpuStandLine",
			"total", s.FormatPoints(domain.QuinzeCpuStandPoints),
			"target", s.FormatPoints(domain.QuinzeTarget))) + "\n"
}

// ActionLogOutput 棋譜をテキスト出力
func (sp *QuinzeCuiPresenter) ActionLogOutput(s interfaces.QuinzeGame) string {
	if !s.GetGameEndFlag() {
		return actionLogToText(nil)
	}
	return actionLogToText(s.GetActionLog())
}
