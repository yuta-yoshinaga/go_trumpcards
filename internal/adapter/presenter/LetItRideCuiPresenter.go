//go:build !js || !wasm || casino

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

// LetItRideCuiPresenter レット・イット・ライドCUIプレゼンタークラス
type LetItRideCuiPresenter struct {
}

// Output ゲーム状態を出力
func (lp *LetItRideCuiPresenter) Output(lir interfaces.LetItRideGame, lastErr error) string {
	var sb strings.Builder

	sb.WriteString("----------\n")
	fmt.Fprintf(&sb, "%s\n", i18n.Tf("letitride.chipsLine", "chips", strconv.Itoa(lir.GetChips())))
	fmt.Fprintf(&sb, "%s\n", i18n.Tf("letitride.phaseLine", "phase", lp.phaseStr(lir.GetPhase())))

	if lir.GetBetAmount() > 0 {
		fmt.Fprintf(&sb, "%s\n", i18n.Tf("letitride.betLine",
			"amount", strconv.Itoa(lir.GetBetAmount()),
			"active", strconv.Itoa(lp.activeBetCount(lir)),
		))
	}

	playerHand := lir.GetPlayerHand()
	if len(playerHand) > 0 {
		sb.WriteString("--- " + color.Bold(i18n.T("letitride.playerHeader")) + " ---\n")
		if lir.GetPhase() == domain.LetItRidePhaseEnd {
			rank := lir.GetHandRank()
			if rank >= 0 && rank < len(domain.PokerHandNames) {
				fmt.Fprintf(&sb, "%s\n", i18n.Tf("letitride.handLine", "hand", domain.PokerHandNames[rank]))
			}
		}
		parts := make([]string, len(playerHand))
		for i, card := range playerHand {
			parts[i] = cuiCardStr(card)
		}
		sb.WriteString(strings.Join(parts, ","))
		sb.WriteString("\n")
	}

	communityCards := lir.GetCommunityCards()
	if len(communityCards) > 0 {
		sb.WriteString("--- " + color.Bold(i18n.T("letitride.communityHeader")) + " ---\n")
		parts := make([]string, len(communityCards))
		switch lir.GetPhase() {
		case domain.LetItRidePhaseBet, domain.LetItRidePhaseFirstDecision:
			for i := range communityCards {
				parts[i] = "??"
			}
		case domain.LetItRidePhaseSecondDecision:
			parts[0] = cuiCardStr(communityCards[0])
			for i := 1; i < len(communityCards); i++ {
				parts[i] = "??"
			}
		default:
			for i, card := range communityCards {
				parts[i] = cuiCardStr(card)
			}
		}
		sb.WriteString(strings.Join(parts, ","))
		sb.WriteString("\n")
	}

	sb.WriteString("----------\n")

	if lastErr != nil {
		fmt.Fprintf(&sb, "%s\n", i18n.MarkErrorLine(color.Red(lastErr.Error())))
	}

	if lir.GetGameEndFlag() {
		fmt.Fprintf(&sb, "%s\n", i18n.Tf("letitride.betsLine",
			"bet1", lp.betStatusStr(lir.GetBet1Active()),
			"bet2", lp.betStatusStr(lir.GetBet2Active()),
			"bet3", lp.betStatusStr(lir.GetBet3Active()),
		))
		switch lir.GetResult() {
		case domain.GameResultWin:
			sb.WriteString(color.Green(i18n.T("letitride.playerWins")) + "\n")
		case domain.GameResultLose:
			sb.WriteString(color.Red(i18n.T("letitride.playerLoses")) + "\n")
		default:
		}
		fmt.Fprintf(&sb, "%s\n", i18n.Tf("letitride.totalPayoutLine", "payout", strconv.Itoa(lir.GetTotalPayout())))
		sb.WriteString("----------\n")
	}

	return sb.String()
}

// PullConfirmOutput は Pull を実行する前の確認内容を出力する。
//
// **Web は専用の確認ダイアログでリスクの前後を見せてから実行する** のに、CUI は
// "p" で即座に取り下げていた (#4699)。戻る額と場に残る額が分からないまま、
// 取り消せない操作を打つことになる。
func (lp *LetItRideCuiPresenter) PullConfirmOutput(lir interfaces.LetItRideGame) string {
	pv := lir.GetPullPreview()
	if pv == nil {
		return i18n.T("letitride.pullUnavailable") + "\n"
	}
	return color.BoldYellow(i18n.Tf("letitride.pullConfirm",
		"amount", strconv.Itoa(pv.Returned),
		"risk", strconv.Itoa(pv.RiskBefore),
		"newRisk", strconv.Itoa(pv.RiskAfter))) + "\n" +
		i18n.T("letitride.pullConfirmPrompt") + "\n"
}

// ActionLogOutput 棋譜をテキスト出力
func (lp *LetItRideCuiPresenter) ActionLogOutput(lir interfaces.LetItRideGame) string {
	return actionLogOutputText(lir)
}

// phaseStr フェーズ文字列
func (lp *LetItRideCuiPresenter) phaseStr(phase int) string {
	switch phase {
	case domain.LetItRidePhaseBet:
		return i18n.T("letitride.phaseBet")
	case domain.LetItRidePhaseFirstDecision:
		return i18n.T("letitride.phaseFirstDecision")
	case domain.LetItRidePhaseSecondDecision:
		return i18n.T("letitride.phaseSecondDecision")
	case domain.LetItRidePhaseEnd:
		return i18n.T("letitride.phaseEnd")
	default:
		return i18n.T("letitride.phaseUnknown")
	}
}

// betStatusStr ベット状態文字列
func (lp *LetItRideCuiPresenter) betStatusStr(active bool) string {
	if active {
		return color.Green(i18n.T("letitride.betStatusRide"))
	}
	return color.Yellow(i18n.T("letitride.betStatusPull"))
}

// activeBetCount アクティブベット数
func (lp *LetItRideCuiPresenter) activeBetCount(lir interfaces.LetItRideGame) int {
	count := 0
	if lir.GetBet1Active() {
		count++
	}
	if lir.GetBet2Active() {
		count++
	}
	if lir.GetBet3Active() {
		count++
	}
	return count
}
