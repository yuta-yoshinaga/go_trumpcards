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

// PontoonCuiPresenter ポンツーン CUI プレゼンタークラス
type PontoonCuiPresenter struct{}

// pontoonRankLabel 手の格の表示名。点数だけでは順位が決まらないので、格を明示する。
func pontoonRankLabel(rank domain.PontoonRank) string {
	switch rank {
	case domain.PontoonRankPontoon:
		return i18n.T("pontoon.rankPontoon")
	case domain.PontoonRankFiveCard:
		return i18n.T("pontoon.rankFiveCard")
	case domain.PontoonRankBust:
		return i18n.T("pontoon.rankBust")
	default:
		return ""
	}
}

// pontoonHandLine 1 つの手を 1 行に描く
func pontoonHandLine(p interfaces.PontoonGame, h *domain.PontoonHand, hide bool) string {
	cards := h.GetCards()
	parts := make([]string, len(cards))
	for i, c := range cards {
		if hide {
			parts[i] = i18n.T("pontoon.faceDown")
			continue
		}
		parts[i] = cuiCardStr(c)
	}
	line := strings.Join(parts, " ")
	if hide {
		return line
	}
	line += " " + i18n.Tf("pontoon.totalInline", "total", strconv.Itoa(p.GetHandTotal(cards)))
	if label := pontoonRankLabel(p.GetHandRank(cards)); label != "" {
		line += " " + color.Bold(label)
	}
	return line
}

// Output ゲーム状態を出力
func (pp *PontoonCuiPresenter) Output(p interfaces.PontoonGame, lastErr error) string {
	var sb strings.Builder

	sb.WriteString("----------\n")
	sb.WriteString(i18n.Tf("pontoon.chipsLine", "chips", strconv.Itoa(p.GetChips())) + "\n")
	bankerName := i18n.T("pontoon.bankerIsYou")
	if !p.IsHumanBanker() {
		bankerName = p.GetSeats()[p.GetBankerIdx()].GetName()
	}
	sb.WriteString(i18n.Tf("pontoon.bankerLine", "name", bankerName) + "\n")

	ended := p.GetGameEndFlag()
	if bh := p.GetBankerHand(); bh != nil {
		// 親の手は 2 枚とも伏せる。ここが読めないのがブラックジャックとの違い。
		hide := !ended && !p.IsHumanBanker()
		sb.WriteString(i18n.T("pontoon.bankerHandHeader") + " " +
			pontoonHandLine(p, bh, hide) + "\n")
	}

	sb.WriteString("----------\n")
	for i, s := range p.GetSeats() {
		if i == p.GetBankerIdx() {
			continue
		}
		for j, h := range s.GetHands() {
			marker := "  "
			if i == p.GetActiveSeat() && j == p.GetActiveHand() &&
				p.GetPhase() == domain.PontoonPhasePlayerTurn {
				marker = "> "
			}
			hide := !ended && s.IsCPU()
			sb.WriteString(marker + s.GetName() + " " +
				i18n.Tf("pontoon.betInline", "bet", strconv.Itoa(h.GetBet())) + " " +
				pontoonHandLine(p, h, hide))
			if ended && h.GetPayout() != 0 {
				sb.WriteString(" " + i18n.Tf("pontoon.payoutInline",
					"payout", strconv.Itoa(h.GetPayout())))
			}
			sb.WriteString("\n")
		}
	}
	sb.WriteString("----------\n")

	cuiErrorBlock(&sb, lastErr)

	switch p.GetPhase() {
	case domain.PontoonPhaseBet:
		if p.IsHumanBanker() {
			sb.WriteString(color.Yellow(i18n.T("pontoon.dealAsBanker")) + "\n")
		} else {
			sb.WriteString(i18n.T("pontoon.placeBet") + "\n")
		}
	case domain.PontoonPhasePlayerTurn:
		sb.WriteString(pp.actionHints(p))
	case domain.PontoonPhaseBankerTurn:
		sb.WriteString(color.Yellow(i18n.T("pontoon.bankerTurn")) + "\n")
	case domain.PontoonPhaseEnd:
		sb.WriteString(color.Green(p.GetLastResult()) + "\n")
		if nb := p.GetNextBanker(); nb >= 0 {
			sb.WriteString(color.Yellow(i18n.Tf("pontoon.bankPasses",
				"name", p.GetSeats()[nb].GetName())) + "\n")
		}
	}

	return sb.String()
}

// actionHints 今打てる手だけを並べる。15 未満で Stick を勧めないなど、
// 規則の制約をそのまま案内に落とす。
func (pp *PontoonCuiPresenter) actionHints(p interfaces.PontoonGame) string {
	var opts []string
	if p.CanStick() {
		opts = append(opts, i18n.T("pontoon.optStick"))
	}
	if p.CanTwist() {
		opts = append(opts, i18n.T("pontoon.optTwist"))
	}
	if p.CanBuy() {
		opts = append(opts, i18n.T("pontoon.optBuy"))
	}
	if p.CanSplit() {
		opts = append(opts, i18n.T("pontoon.optSplit"))
	}
	if len(opts) == 0 {
		return ""
	}
	// **相手がどこで止まるかは判断材料の半分。**15 未満は宣言できないという
	// 自分側の制約は出ていたのに、CPU と親が 17 で止まることはどこにも
	// 書かれていなかった (#5565)。数字はドメインの定数から差し込む。
	return i18n.Tf("pontoon.actionsLine", "options", strings.Join(opts, " / ")) + "\n" +
		color.Yellow(i18n.Tf("pontoon.cpuStickLine",
			"cpuMin", strconv.Itoa(domain.PontoonCpuStickMin),
			"min", strconv.Itoa(domain.PontoonStickMin))) + "\n"
}

// ActionLogOutput 棋譜をテキスト出力
func (pp *PontoonCuiPresenter) ActionLogOutput(p interfaces.PontoonGame) string {
	if !p.GetGameEndFlag() {
		return actionLogToText(nil)
	}
	return actionLogToText(p.GetActionLog())
}
