package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// BaccaratCuiPresenter バカラCUIプレゼンタークラス
type BaccaratCuiPresenter struct {
}

// Output ゲーム状態を出力
func (bp *BaccaratCuiPresenter) Output(b interfaces.BaccaratGame, lastErr error) string {
	var sb strings.Builder

	sb.WriteString("----------\n")
	sb.WriteString(i18n.Tf("baccarat.chipsLine", "chips", strconv.Itoa(b.GetChips())) + "\n")
	sb.WriteString(i18n.Tf("baccarat.phaseLine", "phase", bp.phaseStr(b.GetPhase())) + "\n")

	playerHand := b.GetPlayerHand()
	if len(playerHand) > 0 {
		sb.WriteString("--- " + color.Bold(i18n.T("baccarat.playerHeader")) + " ---\n")
		sb.WriteString(i18n.Tf("baccarat.valueLine", "value", strconv.Itoa(b.GetPlayerHandValue())) + "\n")
		parts := make([]string, len(playerHand))
		for i, card := range playerHand {
			parts[i] = cuiCardStr(card)
		}
		sb.WriteString(strings.Join(parts, ","))
		sb.WriteString("\n")
	}

	bankerHand := b.GetBankerHand()
	if len(bankerHand) > 0 {
		sb.WriteString("--- " + color.Bold(i18n.T("baccarat.bankerHeader")) + " ---\n")
		sb.WriteString(i18n.Tf("baccarat.valueLine", "value", strconv.Itoa(b.GetBankerHandValue())) + "\n")
		parts := make([]string, len(bankerHand))
		for i, card := range bankerHand {
			parts[i] = cuiCardStr(card)
		}
		sb.WriteString(strings.Join(parts, ","))
		sb.WriteString("\n")
	}

	sb.WriteString("----------\n")

	if lastErr != nil {
		sb.WriteString(color.Red(lastErr.Error()) + "\n")
	}

	if b.GetGameEndFlag() {
		sb.WriteString(i18n.Tf("baccarat.betLine",
			"amount", strconv.Itoa(b.GetBetAmount()),
			"type", bp.betTypeStr(b.GetBetType()),
		) + "\n")
		switch b.GetResult() {
		case domain.GameResultWin:
			sb.WriteString(color.Green(i18n.T("baccarat.playerWins")) + "\n")
		case domain.GameResultLose:
			sb.WriteString(color.Red(i18n.T("baccarat.bankerWins")) + "\n")
		case domain.GameResultDraw:
			sb.WriteString(color.Yellow(i18n.T("baccarat.tie")) + "\n")
		default:
		}
		sb.WriteString(i18n.Tf("baccarat.payoutLine", "payout", strconv.Itoa(b.GetPayout())) + "\n")
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
		return i18n.T("baccarat.phaseBet")
	case domain.BaccaratPhaseEnd:
		return i18n.T("baccarat.phaseEnd")
	default:
		return i18n.T("baccarat.phaseUnknown")
	}
}

// betTypeStr ベットタイプ文字列
func (bp *BaccaratCuiPresenter) betTypeStr(betType int) string {
	switch betType {
	case domain.BaccaratBetPlayer:
		return i18n.T("baccarat.betTypePlayer")
	case domain.BaccaratBetBanker:
		return i18n.T("baccarat.betTypeBanker")
	case domain.BaccaratBetTie:
		return i18n.T("baccarat.betTypeTie")
	default:
		return i18n.T("baccarat.betTypeUnknown")
	}
}
