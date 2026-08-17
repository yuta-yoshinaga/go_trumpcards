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

// SetteEMezzoCuiPresenter セッテ・エ・メッツォ CUI プレゼンタークラス
type SetteEMezzoCuiPresenter struct{}

// setteEMezzoHandLine 1 つの手を 1 行に描く
func setteEMezzoHandLine(s interfaces.SetteEMezzoGame, h *domain.SetteEMezzoHand, hide bool) string {
	cards := h.GetCards()
	parts := make([]string, len(cards))
	for i, c := range cards {
		if hide {
			parts[i] = i18n.T("settemezzo.faceDown")
			continue
		}
		parts[i] = cuiCardStr(c)
	}
	line := strings.Join(parts, " ")
	if hide {
		return line
	}
	line += " " + i18n.Tf("settemezzo.totalInline", "total", s.FormatHalves(s.GetHandHalves(h)))
	// マッタを持っているなら、いま何点として数えているかを併記する。
	// 値を付け替えられるのがマッタの要点なので、現在値が見えないと選べない。
	if h.HasMatta() {
		v := h.GetMattaHalves()
		if v == 0 {
			v = 1
		}
		line += " " + i18n.Tf("settemezzo.mattaInline", "value", s.FormatHalves(v))
	}
	return line
}

// Output ゲーム状態を出力
func (sp *SetteEMezzoCuiPresenter) Output(s interfaces.SetteEMezzoGame, lastErr error) string {
	var sb strings.Builder

	sb.WriteString("----------\n")
	sb.WriteString(i18n.Tf("settemezzo.chipsLine", "chips", strconv.Itoa(s.GetChips())) + "\n")
	bankerName := i18n.T("settemezzo.bankerIsYou")
	if !s.IsHumanBanker() {
		bankerName = s.GetSeats()[s.GetBankerIdx()].GetName()
	}
	sb.WriteString(i18n.Tf("settemezzo.bankerLine", "name", bankerName) + "\n")

	ended := s.GetGameEndFlag()
	if bh := s.GetBankerHand(); bh != nil {
		hide := !ended && !s.IsHumanBanker()
		sb.WriteString(i18n.T("settemezzo.bankerHandHeader") + " " +
			setteEMezzoHandLine(s, bh, hide) + "\n")
	}

	sb.WriteString("----------\n")
	for i, seat := range s.GetSeats() {
		if i == s.GetBankerIdx() || seat.GetHand() == nil {
			continue
		}
		h := seat.GetHand()
		marker := "  "
		if i == s.GetActiveSeat() && s.GetPhase() == domain.SetteEMezzoPhasePlayerTurn {
			marker = "> "
		}
		hide := !ended && seat.IsCPU()
		sb.WriteString(marker + seat.GetName() + " " +
			i18n.Tf("settemezzo.betInline", "bet", strconv.Itoa(h.GetBet())) + " " +
			setteEMezzoHandLine(s, h, hide))
		if ended && h.GetPayout() != 0 {
			sb.WriteString(" " + i18n.Tf("settemezzo.payoutInline",
				"payout", strconv.Itoa(h.GetPayout())))
		}
		sb.WriteString("\n")
	}
	sb.WriteString("----------\n")

	cuiErrorBlock(&sb, lastErr)

	switch s.GetPhase() {
	case domain.SetteEMezzoPhaseBet:
		if s.IsHumanBanker() {
			sb.WriteString(color.Yellow(i18n.T("settemezzo.dealAsBanker")) + "\n")
		} else {
			sb.WriteString(i18n.T("settemezzo.placeBet") + "\n")
		}
	case domain.SetteEMezzoPhasePlayerTurn:
		sb.WriteString(sp.actionHints(s))
	case domain.SetteEMezzoPhaseBankerTurn:
		sb.WriteString(color.Yellow(i18n.T("settemezzo.bankerTurn")) + "\n")
	case domain.SetteEMezzoPhaseEnd:
		sb.WriteString(color.Green(s.GetLastResult()) + "\n")
		if nb := s.GetNextBanker(); nb >= 0 {
			sb.WriteString(color.Yellow(i18n.Tf("settemezzo.bankPasses",
				"name", s.GetSeats()[nb].GetName())) + "\n")
		}
	}

	return sb.String()
}

// actionHints 今打てる手だけを並べる
func (sp *SetteEMezzoCuiPresenter) actionHints(s interfaces.SetteEMezzoGame) string {
	var opts []string
	if s.CanHit() {
		opts = append(opts, i18n.T("settemezzo.optHit"))
	}
	if s.CanStand() {
		opts = append(opts, i18n.T("settemezzo.optStand"))
	}
	if s.CanSetMatta() {
		opts = append(opts, i18n.T("settemezzo.optMatta"))
	}
	if len(opts) == 0 {
		return ""
	}
	// **相手がいつ引くのをやめるかは、賭け続けるかの判断材料。**その数字は
	// どの画面にも出ていなかった (#5566)。点は FormatHalves に通す ── 半点単位の
	// 内部表現をそのまま出すと 11 と読めてしまう。
	return i18n.Tf("settemezzo.actionsLine", "options", strings.Join(opts, " / ")) + "\n" +
		color.Yellow(i18n.Tf("settemezzo.cpuStandLine",
			"total", s.FormatHalves(domain.SetteEMezzoCpuStandHalves),
			"target", s.FormatHalves(domain.SetteEMezzoTargetHalves))) + "\n"
}

// ActionLogOutput 棋譜をテキスト出力
func (sp *SetteEMezzoCuiPresenter) ActionLogOutput(s interfaces.SetteEMezzoGame) string {
	if !s.GetGameEndFlag() {
		return actionLogToText(nil)
	}
	return actionLogToText(s.GetActionLog())
}
