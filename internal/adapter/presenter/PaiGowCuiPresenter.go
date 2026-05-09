package presenter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// PaiGowCuiPresenter パイガオポーカーCUIプレゼンタークラス
type PaiGowCuiPresenter struct {
}

// Output ゲーム状態を出力
func (pp *PaiGowCuiPresenter) Output(pg interfaces.PaiGowGame, lastErr error) string {
	var sb strings.Builder

	sb.WriteString("----------\n")
	fmt.Fprintf(&sb, "%s\n", i18n.Tf("paigow.chipsLine", "chips", strconv.Itoa(pg.GetChips())))
	fmt.Fprintf(&sb, "%s\n", i18n.Tf("paigow.phaseLine", "phase", pp.phaseStr(pg.GetPhase())))

	playerCards := pg.GetPlayerCards()
	if len(playerCards) > 0 {
		sb.WriteString("--- " + color.Bold(i18n.T("paigow.playerHeader")) + " ---\n")
		sb.WriteString(i18n.T("paigow.cardsLabel"))
		parts := make([]string, len(playerCards))
		for i, c := range playerCards {
			parts[i] = fmt.Sprintf("[%d]%s", i, cuiCardStr(c))
		}
		sb.WriteString(strings.Join(parts, " "))
		sb.WriteString("\n")
	}

	if pg.GetPhase() == domain.PaiGowPhaseEnd {
		highHand := pg.GetPlayerHighHand()
		if len(highHand) > 0 {
			fmt.Fprintf(&sb, "%s\n", i18n.Tf("paigow.highLine",
				"cards", cuiCardSliceStr(highHand),
				"rank", pp.highHandRankStr(pg.GetPlayerHighRank()),
			))
		}
		lowHand := pg.GetPlayerLowHand()
		if len(lowHand) > 0 {
			fmt.Fprintf(&sb, "%s\n", i18n.Tf("paigow.lowLine",
				"cards", cuiCardSliceStr(lowHand),
				"rank", pp.lowHandRankStr(pg.GetPlayerLowRank()),
			))
		}

		sb.WriteString("--- " + color.Bold(i18n.T("paigow.dealerHeader")) + " ---\n")
		dealerHigh := pg.GetDealerHighHand()
		if len(dealerHigh) > 0 {
			fmt.Fprintf(&sb, "%s\n", i18n.Tf("paigow.highLine",
				"cards", cuiCardSliceStr(dealerHigh),
				"rank", pp.highHandRankStr(pg.GetDealerHighRank()),
			))
		}
		dealerLow := pg.GetDealerLowHand()
		if len(dealerLow) > 0 {
			fmt.Fprintf(&sb, "%s\n", i18n.Tf("paigow.lowLine",
				"cards", cuiCardSliceStr(dealerLow),
				"rank", pp.lowHandRankStr(pg.GetDealerLowRank()),
			))
		}
	}

	sb.WriteString("----------\n")

	if lastErr != nil {
		fmt.Fprintf(&sb, "%s\n", color.Red(lastErr.Error()))
	}

	if pg.GetGameEndFlag() {
		fmt.Fprintf(&sb, "%s\n", i18n.Tf("paigow.betLine", "bet", strconv.Itoa(pg.GetBet())))
		switch pg.GetResult() {
		case domain.GameResultWin:
			sb.WriteString(color.Green(i18n.T("paigow.playerWins")) + "\n")
			fmt.Fprintf(&sb, "%s\n", i18n.Tf("paigow.payoutWithCommissionLine",
				"payout", strconv.Itoa(pg.GetPayout()),
				"commission", strconv.Itoa(pg.GetCommission()),
			))
		case domain.GameResultLose:
			sb.WriteString(color.Red(i18n.T("paigow.dealerWins")) + "\n")
		case domain.GameResultDraw:
			sb.WriteString(color.Yellow(i18n.T("paigow.push")) + "\n")
			fmt.Fprintf(&sb, "%s\n", i18n.Tf("paigow.payoutLine", "payout", strconv.Itoa(pg.GetPayout())))
		default:
		}
		sb.WriteString("----------\n")
	}

	return sb.String()
}

// ActionLogOutput 棋譜をテキスト出力
func (pp *PaiGowCuiPresenter) ActionLogOutput(pg interfaces.PaiGowGame) string {
	return actionLogOutputText(pg)
}

// phaseStr フェーズ文字列
func (pp *PaiGowCuiPresenter) phaseStr(phase int) string {
	switch phase {
	case domain.PaiGowPhaseBet:
		return i18n.T("paigow.phaseBet")
	case domain.PaiGowPhaseSetHands:
		return i18n.T("paigow.phaseSetHands")
	case domain.PaiGowPhaseEnd:
		return i18n.T("paigow.phaseEnd")
	default:
		return i18n.T("paigow.phaseUnknown")
	}
}

// highHandRankStr ハイハンドランク文字列
func (pp *PaiGowCuiPresenter) highHandRankStr(rank int) string {
	if rank >= 0 && rank < len(domain.PokerHandNames) {
		return domain.PokerHandNames[rank]
	}
	return i18n.T("paigow.rankUnknown")
}

// lowHandRankStr ローハンドランク文字列
func (pp *PaiGowCuiPresenter) lowHandRankStr(rank int) string {
	if rank >= 0 && rank < len(domain.PaiGowLowHandNames) {
		return domain.PaiGowLowHandNames[rank]
	}
	return i18n.T("paigow.rankUnknown")
}
