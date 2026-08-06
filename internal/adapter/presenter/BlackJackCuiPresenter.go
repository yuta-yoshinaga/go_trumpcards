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

// bjHandStatusStr returns space-prefixed hand status tags like " [DD] [BUST]".
func bjHandStatusStr(hand *domain.BlackJackHand) string {
	var parts []string
	if hand.IsDoubled() {
		parts = append(parts, color.BoldYellow(i18n.T("blackjack.handStatusDD")))
	}
	if hand.IsBusted() {
		parts = append(parts, color.BoldYellow(i18n.T("blackjack.handStatusBust")))
	}
	if hand.IsStood() {
		parts = append(parts, color.BoldYellow(i18n.T("blackjack.handStatusStand")))
	}
	if hand.IsBlackJack() {
		parts = append(parts, color.BoldYellow(i18n.T("blackjack.handStatusBJ")))
	}
	if hand.IsSurrendered() {
		parts = append(parts, color.BoldYellow(i18n.T("blackjack.handStatusSurrender")))
	}
	if len(parts) == 0 {
		return ""
	}
	return " " + strings.Join(parts, " ")
}

// bjMultiHandResultStr returns the result string for multi-hand games.
func bjMultiHandResultStr(bj interfaces.BlackJackGame, handCount int) string {
	var b strings.Builder
	for i := range handCount {
		result := bj.GameJudgmentForHand(i)
		b.WriteString(i18n.Tf("blackjack.multiHandResultPrefix", "idx", strconv.Itoa(i+1)))
		switch result {
		case domain.GameResultDraw:
			b.WriteString(color.Yellow(i18n.T("blackjack.resultDraw")))
		case domain.GameResultWin:
			b.WriteString(color.Green(i18n.T("blackjack.resultWin")))
		case domain.GameResultLose:
			b.WriteString(color.Red(i18n.T("blackjack.resultLose")))
		}
		b.WriteString("\n")
	}
	return b.String()
}

// BlackJackCuiPresenter ブラックジャックCUIプレゼンタークラス
type BlackJackCuiPresenter struct {
}

// Output ゲーム状態を出力
func (bjp *BlackJackCuiPresenter) Output(bj interfaces.BlackJackGame, lastErr error) string {
	player := bj.GetPlayer()
	dealer := bj.GetDealer()
	var b strings.Builder

	b.WriteString("----------\n")

	fmt.Fprintf(&b, "%s\n", i18n.Tf("blackjack.chipsLine",
		"player", strconv.Itoa(player.GetChips()),
		"dealer", strconv.Itoa(dealer.GetChips()),
		"decks", strconv.Itoa(bj.GetDeckCount()),
	))

	config := bj.GetConfig()
	if config.DealerHitsSoft17 {
		fmt.Fprintf(&b, "%s\n", i18n.T("blackjack.ruleH17"))
	}
	if !config.DoubleAfterSplit {
		fmt.Fprintf(&b, "%s\n", i18n.T("blackjack.ruleNoDAS"))
	}
	if config.DeckPenetration != 0 && config.DeckPenetration != domain.BJDefaultPenetration {
		fmt.Fprintf(&b, "%s\n", i18n.Tf("blackjack.rulePenetration", "percent", strconv.Itoa(config.DeckPenetration)))
	}
	if config.CountingEnabled {
		sysName := countingSystemName(config.CountingSystem)
		if domain.IsBalancedCountingSystem(config.CountingSystem) {
			fmt.Fprintf(&b, "%s\n", i18n.Tf("blackjack.countWithTrueCount",
				"system", sysName,
				"rc", strconv.Itoa(bj.GetRunningCount()),
				"tc", fmt.Sprintf("%.1f", bj.GetTrueCount()),
			))
		} else {
			fmt.Fprintf(&b, "%s\n", i18n.Tf("blackjack.countNoTrueCount",
				"system", sysName,
				"rc", strconv.Itoa(bj.GetRunningCount()),
			))
		}
	}

	if bj.GetMultiHandCount() > 1 {
		fmt.Fprintf(&b, "%s\n", i18n.Tf("blackjack.multiHandLine", "count", strconv.Itoa(bj.GetMultiHandCount())))
	}

	fmt.Fprintf(&b, "%s\n", i18n.Tf("blackjack.phaseLine", "phase", bjp.phaseStr(bj.GetPhase())))

	// **ベット前に配当を見せる。**Web はベットフェーズに配当表の details を出して
	// いるのに、CUI はチップ・デッキ・ルールしか出していなかった (#4677)。
	if bj.GetPhase() == domain.BJPhaseBet && !bj.GetGameEndFlag() {
		b.WriteString(bjp.payoutTable(bj))
	}

	b.WriteString(color.Bold(i18n.T("blackjack.dealerLabel")) + i18n.T("blackjack.scoreSuffix"))
	if bj.GetGameEndFlag() {
		fmt.Fprintf(&b, "%d\n", dealer.GetScore())
		b.WriteString(cuiCardListStr(dealer))
		b.WriteString("\n")
	} else {
		b.WriteString("\n")
		if dealer.GetCardsSize() > 0 {
			fmt.Fprintf(&b, "%s, %s\n", cuiCardStr(dealer.GetCard(0)), i18n.T("blackjack.hiddenCard"))
		}
	}
	b.WriteString("----------\n")

	hands := bj.GetPlayerHands()
	for i, hand := range hands {
		prefix := color.Bold(i18n.T("blackjack.playerLabel"))
		if len(hands) > 1 {
			prefix = color.Bold(i18n.Tf("blackjack.handLabel", "idx", strconv.Itoa(i+1)))
		}
		if i == bj.GetCurrentHandIdx() && !bj.GetGameEndFlag() {
			prefix += i18n.T("blackjack.currentHandMarker")
		}
		fmt.Fprintf(&b, "%s\n", i18n.Tf("blackjack.playerHandLine",
			"prefix", prefix,
			"score", strconv.Itoa(hand.GetScore()),
			"bet", strconv.Itoa(hand.GetBet()),
			"status", bjHandStatusStr(hand),
		))
		b.WriteString(cuiCardListStr(hand))
		b.WriteString("\n")
	}

	cpuPlayers := bj.GetCpuPlayers()
	for cpuIdx, cpu := range cpuPlayers {
		cpuName := color.Bold(i18n.Tf("blackjack.cpuLabel", "idx", strconv.Itoa(cpuIdx+1)))
		if cpu.GetInsuranceBet() > 0 {
			fmt.Fprintf(&b, "%s\n", i18n.Tf("blackjack.cpuHeaderWithInsurance",
				"name", cpuName,
				"chips", strconv.Itoa(cpu.GetPlayer().GetChips()),
				"insurance", strconv.Itoa(cpu.GetInsuranceBet()),
			))
		} else {
			fmt.Fprintf(&b, "%s\n", i18n.Tf("blackjack.cpuHeader",
				"name", cpuName,
				"chips", strconv.Itoa(cpu.GetPlayer().GetChips()),
			))
		}
		writeBJCpuHands(&b, bj, cpuIdx, cpu)
	}

	b.WriteString("----------\n")

	if bj.GetInsuranceBet() > 0 {
		fmt.Fprintf(&b, "%s\n", i18n.Tf("blackjack.insuranceBetLine", "bet", strconv.Itoa(bj.GetInsuranceBet())))
	}
	if bj.IsInsuranceAvailable() && bj.GetPhase() == domain.BJPhaseInsurance {
		fmt.Fprintf(&b, "%s\n", i18n.T("blackjack.insuranceAvailable"))
	}

	sideBetResults := bj.GetSideBetResults()
	for _, r := range sideBetResults {
		if r.Payout > 0 {
			fmt.Fprintf(&b, "%s\n", i18n.Tf("blackjack.sideBetWin",
				"type", r.BetTypeName(),
				"result", r.ResultName,
				"payout", strconv.Itoa(r.Payout),
				"bet", strconv.Itoa(r.BetAmount),
			))
		} else {
			fmt.Fprintf(&b, "%s\n", i18n.Tf("blackjack.sideBetLose",
				"type", r.BetTypeName(),
				"bet", strconv.Itoa(r.BetAmount),
			))
		}
	}

	if bj.IsHintEnabled() {
		suggestion := bj.GetBasicStrategySuggestion()
		if suggestion != domain.BJSuggestNone {
			fmt.Fprintf(&b, "%s\n", color.Yellow(i18n.Tf("blackjack.hintLine", "action", bjp.suggestionStr(suggestion))))
		}
	}

	if lastErr != nil {
		fmt.Fprintf(&b, "%s\n", color.Red(lastErr.Error()))
	}

	if bj.GetGameEndFlag() {
		if len(hands) > 1 {
			b.WriteString(bjMultiHandResultStr(bj, len(hands)))
		} else {
			switch bj.GameJudgment() {
			case domain.GameResultDraw:
				b.WriteString(color.Yellow(i18n.T("blackjack.resultDraw")) + "\n")
			case domain.GameResultWin:
				b.WriteString(color.Green(i18n.T("blackjack.resultWin")) + "\n")
			case domain.GameResultLose:
				b.WriteString(color.Red(i18n.T("blackjack.resultLose")) + "\n")
			}
		}
		// Variant bonuses (e.g. Spanish 21's 5/6/7-card 21, 6-7-8, 7-7-7). Empty
		// for standard Blackjack, so this leaves the normal output unchanged.
		if bonusKeys := bj.GetBonusKeys(); len(bonusKeys) > 0 {
			b.WriteString(color.BoldYellow(i18n.T("blackjack.bonusHeader")) + "\n")
			for _, key := range bonusKeys {
				b.WriteString(color.Yellow(i18n.Tf("blackjack.bonusLine", "name", i18n.T(key))) + "\n")
			}
		}
		b.WriteString("\n----------\n")
	}
	return b.String()
}

// writeBJCpuHands CPUハンド一覧を出力
func writeBJCpuHands(b *strings.Builder, bj interfaces.BlackJackGame, cpuIdx int, cpu *domain.BlackJackCpuSeat) {
	for hi, hand := range cpu.GetHands() {
		var prefix string
		if len(cpu.GetHands()) > 1 {
			prefix = i18n.Tf("blackjack.cpuHandLabel", "cpu", strconv.Itoa(cpuIdx+1), "idx", strconv.Itoa(hi+1))
		} else {
			prefix = i18n.Tf("blackjack.cpuLabel", "idx", strconv.Itoa(cpuIdx+1))
		}
		fmt.Fprintf(b, "%s\n", i18n.Tf("blackjack.cpuLine",
			"prefix", prefix,
			"score", strconv.Itoa(hand.GetScore()),
			"bet", strconv.Itoa(hand.GetBet()),
			"status", bjHandStatusStr(hand),
		))
		if bj.GetGameEndFlag() || bj.GetPhase() != domain.BJPhaseBet {
			b.WriteString(cuiCardListStr(hand))
			b.WriteString("\n")
		}
	}
}

// payoutTable はベットフェーズの配当一覧を返す。Spanish 21 ではボーナス配当も
// 並べる。文言は Web の payoutRef.* と同じ内容 (#4677)。
func (bjp *BlackJackCuiPresenter) payoutTable(bj interfaces.BlackJackGame) string {
	keys := []string{"blackjack", "win", "insurance", "push", "surrender", "bust"}
	ns := "blackjack.payoutRef"
	if v := bj.GetVariant(); v != nil && v.Name == domain.BJVariantSpanish21 {
		ns = "spanish21.payoutRef"
		keys = append(keys,
			"bonusFiveCard21", "bonusSixCard21", "bonusSevenCard21", "bonus678", "bonus777")
	}
	var b strings.Builder
	b.WriteString(color.Bold(i18n.T(ns+".title")) + "\n")
	for _, k := range keys {
		b.WriteString("  " + i18n.T(ns+"."+k) + "\n")
	}
	return b.String()
}

// ActionLogOutput 棋譜をテキスト出力
func (bjp *BlackJackCuiPresenter) ActionLogOutput(bj interfaces.BlackJackGame) string {
	return actionLogOutputText(bj)
}

// phaseStr フェーズ文字列
func (bjp *BlackJackCuiPresenter) phaseStr(phase int) string {
	switch phase {
	case domain.BJPhaseBet:
		return i18n.T("blackjack.phaseBet")
	case domain.BJPhaseDeal:
		return i18n.T("blackjack.phaseDeal")
	case domain.BJPhaseInsurance:
		return i18n.T("blackjack.phaseInsurance")
	case domain.BJPhaseAction:
		return i18n.T("blackjack.phaseAction")
	case domain.BJPhaseEnd:
		return i18n.T("blackjack.phaseEnd")
	case domain.BJPhaseEarlySurrender:
		return i18n.T("blackjack.phaseEarlySurrender")
	default:
		return i18n.T("blackjack.phaseUnknown")
	}
}

// countingSystemName カウンティングシステム名
func countingSystemName(system int) string {
	switch system {
	case domain.BJCountingKO:
		return i18n.T("blackjack.countingKO")
	case domain.BJCountingZen:
		return i18n.T("blackjack.countingZen")
	case domain.BJCountingOmegaII:
		return i18n.T("blackjack.countingOmegaII")
	default:
		return i18n.T("blackjack.countingHiLo")
	}
}

// suggestionStr 推奨アクション文字列
func (bjp *BlackJackCuiPresenter) suggestionStr(s domain.BJSuggestedAction) string {
	switch s {
	case domain.BJSuggestHit:
		return i18n.T("blackjack.suggestHit")
	case domain.BJSuggestStand:
		return i18n.T("blackjack.suggestStand")
	case domain.BJSuggestDouble:
		return i18n.T("blackjack.suggestDouble")
	case domain.BJSuggestDoubleStand:
		return i18n.T("blackjack.suggestDouble")
	case domain.BJSuggestSplit:
		return i18n.T("blackjack.suggestSplit")
	case domain.BJSuggestSurrender:
		return i18n.T("blackjack.suggestSurrender")
	case domain.BJSuggestDeclineInsurance:
		return i18n.T("blackjack.suggestDeclineInsurance")
	default:
		return ""
	}
}
