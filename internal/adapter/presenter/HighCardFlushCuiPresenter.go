//go:build !js || !wasm || casino

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// HighCardFlushCuiPresenter ハイカードフラッシュCUIプレゼンタークラス
type HighCardFlushCuiPresenter struct {
}

// Output ゲーム状態を出力
func (hp *HighCardFlushCuiPresenter) Output(hcf interfaces.HighCardFlushGame, lastErr error) string {
	var sb strings.Builder

	sb.WriteString("----------\n")
	sb.WriteString(i18n.Tf("highcardflush.chipsLine", "chips", strconv.Itoa(hcf.GetChips())) + "\n")
	sb.WriteString(i18n.Tf("highcardflush.phaseLine", "phase", hp.phaseStr(hcf.GetPhase())) + "\n")

	playerHand := hcf.GetPlayerHand()
	if len(playerHand) > 0 {
		sb.WriteString("--- " + color.Bold(i18n.T("highcardflush.playerHeader")) + " ---\n")
		sb.WriteString(i18n.Tf("highcardflush.flushLine", "len", strconv.Itoa(hcf.GetPlayerFlushLen())) + "\n")
		parts := make([]string, len(playerHand))
		for i, card := range playerHand {
			parts[i] = cuiCardStr(card)
		}
		sb.WriteString(strings.Join(parts, ","))
		sb.WriteString("\n")
	}

	dealerHand := hcf.GetDealerHand()
	if len(dealerHand) > 0 && hcf.GetPhase() == domain.HighCardFlushPhaseEnd {
		sb.WriteString("--- " + color.Bold(i18n.T("highcardflush.dealerHeader")) + " ---\n")
		sb.WriteString(i18n.Tf("highcardflush.flushLine", "len", strconv.Itoa(hcf.GetDealerFlushLen())) + "\n")
		if hcf.GetDealerQualified() {
			sb.WriteString(i18n.T("highcardflush.qualified") + "\n")
		} else {
			sb.WriteString(i18n.T("highcardflush.notQualified") + "\n")
		}
		parts := make([]string, len(dealerHand))
		for i, card := range dealerHand {
			parts[i] = cuiCardStr(card)
		}
		sb.WriteString(strings.Join(parts, ","))
		sb.WriteString("\n")
	}

	sb.WriteString("----------\n")

	if lastErr != nil {
		sb.WriteString(color.Red(lastErr.Error()) + "\n")
	}

	if hcf.GetGameEndFlag() {
		sb.WriteString(i18n.Tf("highcardflush.anteLine", "ante", strconv.Itoa(hcf.GetAnteBet())) + "\n")
		if hcf.GetRaiseBet() > 0 {
			sb.WriteString(i18n.Tf("highcardflush.raiseLine", "raise", strconv.Itoa(hcf.GetRaiseBet())) + "\n")
		}
		switch hcf.GetResult() {
		case domain.GameResultWin:
			sb.WriteString(color.Green(i18n.T("highcardflush.playerWins")) + "\n")
		case domain.GameResultLose:
			if hcf.GetRaiseBet() == 0 {
				sb.WriteString(color.Red(i18n.T("highcardflush.playerFolded")) + "\n")
			} else {
				sb.WriteString(color.Red(i18n.T("highcardflush.dealerWins")) + "\n")
			}
		case domain.GameResultDraw:
			sb.WriteString(color.Yellow(i18n.T("highcardflush.push")) + "\n")
		default:
		}
		sb.WriteString(i18n.Tf("highcardflush.totalPayoutLine", "payout", strconv.Itoa(hcf.GetTotalPayout())) + "\n")
		sb.WriteString("----------\n")
	}

	return sb.String()
}

// ActionLogOutput 棋譜をテキスト出力
func (hp *HighCardFlushCuiPresenter) ActionLogOutput(hcf interfaces.HighCardFlushGame) string {
	return actionLogOutputText(hcf)
}

// phaseStr フェーズ文字列
func (hp *HighCardFlushCuiPresenter) phaseStr(phase int) string {
	switch phase {
	case domain.HighCardFlushPhaseBet:
		return i18n.T("highcardflush.phaseBet")
	case domain.HighCardFlushPhaseAction:
		return i18n.T("highcardflush.phaseAction")
	case domain.HighCardFlushPhaseEnd:
		return i18n.T("highcardflush.phaseEnd")
	default:
		return i18n.T("highcardflush.phaseUnknown")
	}
}

// highCardFlushMarginalHigh は3枚フラッシュで賭けに回れる最低の最高札 (Q)。
const highCardFlushMarginalHigh = 12

// HintOutput emits a raise/fold recommendation during the action phase.
//
// **倍率まで言う (#4714)。**Web には 1x / 2x / 3x の3つのボタンが並ぶのに、
// CUI は「レイズ」か「フォールド」の二択しか返しておらず、いくら賭けるべきかは
// 分からなかった。段階はフロントの getHighCardFlushHint と同じ:
// 6枚以上=3x、5枚=2x、4枚=1x、3枚は最高札 Q 以上なら 1x、それ以外はフォールド。
//
// **3枚フラッシュを一律レイズにしない。**以前はディーラーの成立条件
// (3枚) に届いた時点でレイズと言っていたが、それは弱い3枚でも押す助言だった。
func (hp *HighCardFlushCuiPresenter) HintOutput(hcf interfaces.HighCardFlushGame) string {
	if hcf.GetPhase() != domain.HighCardFlushPhaseAction {
		return i18n.T("highcardflush.hintNone") + "\n"
	}
	switch n := hcf.GetPlayerFlushLen(); {
	case n >= 6:
		return color.Yellow(i18n.T("highcardflush.hintRaise3x")) + "\n"
	case n == 5:
		return color.Yellow(i18n.T("highcardflush.hintRaise2x")) + "\n"
	case n == 4:
		return color.Yellow(i18n.T("highcardflush.hintRaise1x")) + "\n"
	case n == 3 && highCardFlushBestFlushHigh(hcf.GetPlayerHand()) >= highCardFlushMarginalHigh:
		return color.Yellow(i18n.T("highcardflush.hintRaise1xMarginal")) + "\n"
	}
	return color.Yellow(i18n.T("highcardflush.hintFold")) + "\n"
}

// highCardFlushBestFlushHigh は3枚以上そろっているスートの中で最も高い札の値を
// 返す。**エースは 14 として数える。**フロントの getHighCardFlushHint と同じ。
func highCardFlushBestFlushHigh(hand []*domain.Card) int {
	bySuit := map[int][]int{}
	for _, c := range hand {
		v := c.GetValue()
		if v == 1 {
			v = 14
		}
		bySuit[c.GetDesign()] = append(bySuit[c.GetDesign()], v)
	}
	best := 0
	for _, vals := range bySuit {
		if len(vals) < domain.HighCardFlushDealerMinFlushLen {
			continue
		}
		for _, v := range vals {
			if v > best {
				best = v
			}
		}
	}
	return best
}
