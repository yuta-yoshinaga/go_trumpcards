package presenter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// bjHandStatusStr returns space-prefixed hand status tags like " [DD] [BUST]".
func bjHandStatusStr(hand *domain.BlackJackHand) string {
	var parts []string
	if hand.IsDoubled() {
		parts = append(parts, "[DD]")
	}
	if hand.IsBusted() {
		parts = append(parts, "[BUST]")
	}
	if hand.IsStood() {
		parts = append(parts, "[STAND]")
	}
	if hand.IsBlackJack() {
		parts = append(parts, "[BJ]")
	}
	if hand.IsSurrendered() {
		parts = append(parts, "[SURRENDER]")
	}
	if len(parts) == 0 {
		return ""
	}
	return " " + strings.Join(parts, " ")
}

// bjMultiHandResultStr returns the result string for multi-hand games.
func bjMultiHandResultStr(bj interfaces.BlackJackGame, handCount int) string {
	var b strings.Builder
	for i := 0; i < handCount; i++ {
		result := bj.GameJudgmentForHand(i)
		fmt.Fprintf(&b, "hand %d: ", i+1)
		switch result {
		case domain.GameResultDraw:
			b.WriteString("It is a draw.")
		case domain.GameResultWin:
			b.WriteString("You are the winner.")
		case domain.GameResultLose:
			b.WriteString("It is your loss.")
		}
		b.WriteString("\n")
	}
	return b.String()
}

// BlackJackCuiPresenter ブラックジャックCUIプレゼンタークラス
type BlackJackCuiPresenter struct {
}

// NewBlackJackCuiPresenter コンストラクタ
func NewBlackJackCuiPresenter() *BlackJackCuiPresenter {
	return &BlackJackCuiPresenter{}
}

// Output ゲーム状態を出力
func (bjp *BlackJackCuiPresenter) Output(bj interfaces.BlackJackGame, lastErr error) string {
	player := bj.GetPlayer()
	dealer := bj.GetDealer()
	var b strings.Builder

	b.WriteString("----------\n")

	// チップ情報
	fmt.Fprintf(&b, "chips: player=%d dealer=%d decks=%d\n", player.GetChips(), dealer.GetChips(), bj.GetDeckCount())

	// 設定情報
	config := bj.GetConfig()
	if config.DealerHitsSoft17 {
		b.WriteString("rule: H17 (Dealer hits soft 17)\n")
	}
	if !config.DoubleAfterSplit {
		b.WriteString("rule: No DAS (No double after split)\n")
	}
	if config.DeckPenetration != 0 && config.DeckPenetration != domain.BJDefaultPenetration {
		fmt.Fprintf(&b, "rule: Penetration %d%%\n", config.DeckPenetration)
	}
	if config.CountingEnabled {
		sysName := countingSystemName(config.CountingSystem)
		if domain.IsBalancedCountingSystem(config.CountingSystem) {
			fmt.Fprintf(&b, "count (%s): RC=%d TC=%.1f\n", sysName, bj.GetRunningCount(), bj.GetTrueCount())
		} else {
			fmt.Fprintf(&b, "count (%s): RC=%d TC=N/A\n", sysName, bj.GetRunningCount())
		}
	}

	// マルチハンド情報
	if bj.GetMultiHandCount() > 1 {
		fmt.Fprintf(&b, "multi-hand: %d hands\n", bj.GetMultiHandCount())
	}

	// フェーズ情報
	fmt.Fprintf(&b, "phase: %s\n", bjp.phaseStr(bj.GetPhase()))

	// dealer
	b.WriteString("dealer score ")
	if bj.GetGameEndFlag() {
		fmt.Fprintf(&b, "%d\n", dealer.GetScore())
		b.WriteString(cuiCardListStr(dealer))
		b.WriteString("\n")
	} else {
		b.WriteString("\n")
		if dealer.GetCardsSize() > 0 {
			fmt.Fprintf(&b, "%s,\n", cuiCardStr(dealer.GetCard(0)))
		}
	}
	b.WriteString("----------\n")

	// player hands
	hands := bj.GetPlayerHands()
	for i, hand := range hands {
		prefix := "player"
		if len(hands) > 1 {
			prefix = "hand " + strconv.Itoa(i+1)
		}
		if i == bj.GetCurrentHandIdx() && !bj.GetGameEndFlag() {
			prefix += " (*)"
		}
		fmt.Fprintf(&b, "%s score %d bet=%d%s\n", prefix, hand.GetScore(), hand.GetBet(), bjHandStatusStr(hand))
		b.WriteString(cuiCardListStr(hand))
		b.WriteString("\n")
	}

	// CPUプレイヤー
	cpuPlayers := bj.GetCpuPlayers()
	for cpuIdx, cpu := range cpuPlayers {
		if cpu.GetInsuranceBet() > 0 {
			fmt.Fprintf(&b, "--- CPU %d (chips: %d, insurance: %d) ---\n", cpuIdx+1, cpu.GetPlayer().GetChips(), cpu.GetInsuranceBet())
		} else {
			fmt.Fprintf(&b, "--- CPU %d (chips: %d) ---\n", cpuIdx+1, cpu.GetPlayer().GetChips())
		}
		bjp.writeCpuHands(&b, bj, cpuIdx, cpu)
	}

	b.WriteString("----------\n")

	// インシュランス情報
	if bj.GetInsuranceBet() > 0 {
		fmt.Fprintf(&b, "insurance bet: %d\n", bj.GetInsuranceBet())
	}
	if bj.IsInsuranceAvailable() && bj.GetPhase() == domain.BJPhaseInsurance {
		b.WriteString("Insurance available!\n")
	}

	// サイドベット情報
	sideBetResults := bj.GetSideBetResults()
	for _, r := range sideBetResults {
		if r.Payout > 0 {
			fmt.Fprintf(&b, "side bet [%s]: %s WIN +%d (bet=%d)\n", r.BetTypeName(), r.ResultName, r.Payout, r.BetAmount)
		} else {
			fmt.Fprintf(&b, "side bet [%s]: LOSE -%d\n", r.BetTypeName(), r.BetAmount)
		}
	}

	// ヒント情報
	if bj.IsHintEnabled() {
		suggestion := bj.GetBasicStrategySuggestion()
		if suggestion != domain.BJSuggestNone {
			fmt.Fprintf(&b, "[HINT: %s]\n", bjp.suggestionStr(suggestion))
		}
	}

	// エラーメッセージ（ベット失敗等）
	if lastErr != nil {
		fmt.Fprintf(&b, "%s\n", lastErr.Error())
	}

	if bj.GetGameEndFlag() {
		if len(hands) > 1 {
			b.WriteString(bjMultiHandResultStr(bj, len(hands)))
		} else {
			switch bj.GameJudgment() {
			case domain.GameResultDraw:
				b.WriteString("It is a draw.\n")
			case domain.GameResultWin:
				b.WriteString("You are the winner.\n")
			case domain.GameResultLose:
				b.WriteString("It is your loss.\n")
			}
		}
		b.WriteString("\n----------\n")
	}
	return b.String()
}

// writeCpuHands CPUハンド一覧を出力
func (bjp *BlackJackCuiPresenter) writeCpuHands(b *strings.Builder, bj interfaces.BlackJackGame, cpuIdx int, cpu *domain.BlackJackCpuSeat) {
	for hi, hand := range cpu.GetHands() {
		prefix := fmt.Sprintf("CPU %d", cpuIdx+1)
		if len(cpu.GetHands()) > 1 {
			prefix = fmt.Sprintf("CPU %d hand %d", cpuIdx+1, hi+1)
		}
		fmt.Fprintf(b, "%s score %d bet=%d%s\n", prefix, hand.GetScore(), hand.GetBet(), bjHandStatusStr(hand))
		if bj.GetGameEndFlag() || bj.GetPhase() != domain.BJPhaseBet {
			b.WriteString(cuiCardListStr(hand))
			b.WriteString("\n")
		}
	}
}

// ActionLogOutput 棋譜をテキスト出力
func (bjp *BlackJackCuiPresenter) ActionLogOutput(bj interfaces.BlackJackGame) string {
	return actionLogOutputText(bj)
}

// phaseStr フェーズ文字列
func (bjp *BlackJackCuiPresenter) phaseStr(phase int) string {
	switch phase {
	case domain.BJPhaseBet:
		return "BET"
	case domain.BJPhaseDeal:
		return "DEAL"
	case domain.BJPhaseInsurance:
		return "INSURANCE"
	case domain.BJPhaseAction:
		return "ACTION"
	case domain.BJPhaseEnd:
		return "END"
	case domain.BJPhaseEarlySurrender:
		return "EARLY SURRENDER"
	default:
		return "UNKNOWN"
	}
}

// countingSystemName カウンティングシステム名
func countingSystemName(system int) string {
	switch system {
	case domain.BJCountingKO:
		return "KO"
	case domain.BJCountingZen:
		return "Zen Count"
	case domain.BJCountingOmegaII:
		return "Omega II"
	default:
		return "Hi-Lo"
	}
}

// suggestionStr 推奨アクション文字列
func (bjp *BlackJackCuiPresenter) suggestionStr(s domain.BJSuggestedAction) string {
	switch s {
	case domain.BJSuggestHit:
		return "HIT"
	case domain.BJSuggestStand:
		return "STAND"
	case domain.BJSuggestDouble:
		return "DOUBLE"
	case domain.BJSuggestDoubleStand:
		return "DOUBLE"
	case domain.BJSuggestSplit:
		return "SPLIT"
	case domain.BJSuggestSurrender:
		return "SURRENDER"
	case domain.BJSuggestDeclineInsurance:
		return "DECLINE INSURANCE"
	default:
		return ""
	}
}
