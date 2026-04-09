package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// PaiGowCuiPresenter パイガオポーカーCUIプレゼンタークラス
type PaiGowCuiPresenter struct {
}

// Output ゲーム状態を出力
func (pp *PaiGowCuiPresenter) Output(pg interfaces.PaiGowGame, lastErr error) string {
	var sb strings.Builder

	sb.WriteString("----------\n")
	fmt.Fprintf(&sb, "chips: %d\n", pg.GetChips())
	fmt.Fprintf(&sb, "phase: %s\n", pp.phaseStr(pg.GetPhase()))

	// プレイヤーカード
	playerCards := pg.GetPlayerCards()
	if len(playerCards) > 0 {
		sb.WriteString("--- " + color.Bold("PLAYER") + " ---\n")
		sb.WriteString("cards: ")
		parts := make([]string, len(playerCards))
		for i, c := range playerCards {
			parts[i] = fmt.Sprintf("[%d]%s", i, cuiCardStr(c))
		}
		sb.WriteString(strings.Join(parts, " "))
		sb.WriteString("\n")
	}

	// ハイハンド・ローハンド（ENDフェーズ）
	if pg.GetPhase() == domain.PaiGowPhaseEnd {
		highHand := pg.GetPlayerHighHand()
		if len(highHand) > 0 {
			rank := pg.GetPlayerHighRank()
			fmt.Fprintf(&sb, "  high: %s (%s)\n", cuiCardSliceStr(highHand), pp.highHandRankStr(rank))
		}
		lowHand := pg.GetPlayerLowHand()
		if len(lowHand) > 0 {
			rank := pg.GetPlayerLowRank()
			fmt.Fprintf(&sb, "  low:  %s (%s)\n", cuiCardSliceStr(lowHand), pp.lowHandRankStr(rank))
		}

		// ディーラー
		sb.WriteString("--- " + color.Bold("DEALER") + " ---\n")
		dealerHigh := pg.GetDealerHighHand()
		if len(dealerHigh) > 0 {
			rank := pg.GetDealerHighRank()
			fmt.Fprintf(&sb, "  high: %s (%s)\n", cuiCardSliceStr(dealerHigh), pp.highHandRankStr(rank))
		}
		dealerLow := pg.GetDealerLowHand()
		if len(dealerLow) > 0 {
			rank := pg.GetDealerLowRank()
			fmt.Fprintf(&sb, "  low:  %s (%s)\n", cuiCardSliceStr(dealerLow), pp.lowHandRankStr(rank))
		}
	}

	sb.WriteString("----------\n")

	// エラーメッセージ
	if lastErr != nil {
		fmt.Fprintf(&sb, "%s\n", color.Red(lastErr.Error()))
	}

	// ゲーム結果
	if pg.GetGameEndFlag() {
		fmt.Fprintf(&sb, "bet: %d\n", pg.GetBet())
		switch pg.GetResult() {
		case domain.GameResultWin:
			sb.WriteString(color.Green("Player wins!") + "\n")
			fmt.Fprintf(&sb, "payout: %d (commission: %d)\n", pg.GetPayout(), pg.GetCommission())
		case domain.GameResultLose:
			sb.WriteString(color.Red("Dealer wins!") + "\n")
		case domain.GameResultDraw:
			sb.WriteString(color.Yellow("Push!") + "\n")
			fmt.Fprintf(&sb, "payout: %d\n", pg.GetPayout())
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
		return "BET"
	case domain.PaiGowPhaseSetHands:
		return "SET HANDS"
	case domain.PaiGowPhaseEnd:
		return "END"
	default:
		return "UNKNOWN"
	}
}

// highHandRankStr ハイハンドランク文字列
func (pp *PaiGowCuiPresenter) highHandRankStr(rank int) string {
	if rank >= 0 && rank < len(domain.PokerHandNames) {
		return domain.PokerHandNames[rank]
	}
	return "Unknown"
}

// lowHandRankStr ローハンドランク文字列
func (pp *PaiGowCuiPresenter) lowHandRankStr(rank int) string {
	if rank >= 0 && rank < len(domain.PaiGowLowHandNames) {
		return domain.PaiGowLowHandNames[rank]
	}
	return "Unknown"
}
