package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// BaccaratCuiPresenter バカラCUIプレゼンタークラス
type BaccaratCuiPresenter struct {
}

// Output ゲーム状態を出力
func (bp *BaccaratCuiPresenter) Output(b interfaces.BaccaratGame, lastErr error) string {
	var sb strings.Builder

	sb.WriteString("----------\n")
	fmt.Fprintf(&sb, "chips: %d\n", b.GetChips())
	fmt.Fprintf(&sb, "phase: %s\n", bp.phaseStr(b.GetPhase()))

	// プレイヤーハンド
	playerHand := b.GetPlayerHand()
	if len(playerHand) > 0 {
		sb.WriteString("--- " + color.Bold("PLAYER") + " ---\n")
		fmt.Fprintf(&sb, "value: %d\n", b.GetPlayerHandValue())
		parts := make([]string, len(playerHand))
		for i, card := range playerHand {
			parts[i] = cuiCardStr(card)
		}
		sb.WriteString(strings.Join(parts, ","))
		sb.WriteString("\n")
	}

	// バンカーハンド
	bankerHand := b.GetBankerHand()
	if len(bankerHand) > 0 {
		sb.WriteString("--- " + color.Bold("BANKER") + " ---\n")
		fmt.Fprintf(&sb, "value: %d\n", b.GetBankerHandValue())
		parts := make([]string, len(bankerHand))
		for i, card := range bankerHand {
			parts[i] = cuiCardStr(card)
		}
		sb.WriteString(strings.Join(parts, ","))
		sb.WriteString("\n")
	}

	sb.WriteString("----------\n")

	// エラーメッセージ
	if lastErr != nil {
		fmt.Fprintf(&sb, "%s\n", color.Red(lastErr.Error()))
	}

	// ゲーム結果
	if b.GetGameEndFlag() {
		fmt.Fprintf(&sb, "bet: %d on %s\n", b.GetBetAmount(), bp.betTypeStr(b.GetBetType()))
		switch b.GetResult() {
		case domain.GameResultWin:
			sb.WriteString(color.Green("Player wins!") + "\n")
		case domain.GameResultLose:
			sb.WriteString(color.Red("Banker wins!") + "\n")
		case domain.GameResultDraw:
			sb.WriteString(color.Yellow("Tie!") + "\n")
		default:
		}
		fmt.Fprintf(&sb, "payout: %d\n", b.GetPayout())
		sb.WriteString("----------\n")
	}

	return sb.String()
}

// ActionLogOutput 棋譜をテキスト出力
func (bp *BaccaratCuiPresenter) ActionLogOutput(b interfaces.BaccaratGame) string {
	return actionLogOutputText(b)
}

// phaseStr フェーズ文字列
func (bp *BaccaratCuiPresenter) phaseStr(phase int) string {
	switch phase {
	case domain.BaccaratPhaseBet:
		return "BET"
	case domain.BaccaratPhaseEnd:
		return "END"
	default:
		return "UNKNOWN"
	}
}

// betTypeStr ベットタイプ文字列
func (bp *BaccaratCuiPresenter) betTypeStr(betType int) string {
	switch betType {
	case domain.BaccaratBetPlayer:
		return "PLAYER"
	case domain.BaccaratBetBanker:
		return "BANKER"
	case domain.BaccaratBetTie:
		return "TIE"
	default:
		return "UNKNOWN"
	}
}
