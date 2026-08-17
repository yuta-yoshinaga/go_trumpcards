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

// UltimateTexasHoldemCuiPresenter アルティメット・テキサスホールデムCUIプレゼンタークラス
type UltimateTexasHoldemCuiPresenter struct{}

// Output ゲーム状態を出力
// writePayoutRef はブラインドとトリップスの配当表を並べる。
//
// 倍率はすべて domain の定数から。フラッシュのブラインドだけ 3:2 なので、
// 分子・分母をそのまま出す。
func (up *UltimateTexasHoldemCuiPresenter) writePayoutRef(sb *strings.Builder) {
	sb.WriteString(color.Bold(i18n.T("ultimatetexasholdem.payoutRefTitle")) + "\n")

	blind := []struct {
		key  string
		rate string
	}{
		{"payoutRefBlindRoyalFlush", strconv.Itoa(domain.UltimateTexasHoldemBlindPayRoyalFlush)},
		{"payoutRefBlindStraightFlush", strconv.Itoa(domain.UltimateTexasHoldemBlindPayStraightFlush)},
		{"payoutRefBlindFourOfAKind", strconv.Itoa(domain.UltimateTexasHoldemBlindPayFourOfAKind)},
		{"payoutRefBlindFullHouse", strconv.Itoa(domain.UltimateTexasHoldemBlindPayFullHouse)},
		{"payoutRefBlindFlush", strconv.Itoa(domain.UltimateTexasHoldemBlindPayFlushNum) + ":" +
			strconv.Itoa(domain.UltimateTexasHoldemBlindPayFlushDen)},
		{"payoutRefBlindStraight", strconv.Itoa(domain.UltimateTexasHoldemBlindPayStraight)},
	}
	sb.WriteString("  " + i18n.T("ultimatetexasholdem.payoutRefBlindHeader") + "\n")
	for _, row := range blind {
		sb.WriteString("    " + i18n.Tf("ultimatetexasholdem."+row.key, "rate", row.rate) + "\n")
	}

	trips := []struct {
		key  string
		rate int
	}{
		{"payoutRefTripsRoyalFlush", domain.UltimateTexasHoldemTripsPayRoyalFlush},
		{"payoutRefTripsStraightFlush", domain.UltimateTexasHoldemTripsPayStraightFlush},
		{"payoutRefTripsFourOfAKind", domain.UltimateTexasHoldemTripsPayFourOfAKind},
		{"payoutRefTripsFullHouse", domain.UltimateTexasHoldemTripsPayFullHouse},
		{"payoutRefTripsFlush", domain.UltimateTexasHoldemTripsPayFlush},
		{"payoutRefTripsStraight", domain.UltimateTexasHoldemTripsPayStraight},
		{"payoutRefTripsThreeOfAKind", domain.UltimateTexasHoldemTripsPayThreeOfAKind},
	}
	sb.WriteString("  " + i18n.T("ultimatetexasholdem.payoutRefTripsHeader") + "\n")
	for _, row := range trips {
		sb.WriteString("    " + i18n.Tf("ultimatetexasholdem."+row.key, "rate", strconv.Itoa(row.rate)) + "\n")
	}
}

func (up *UltimateTexasHoldemCuiPresenter) Output(g interfaces.UltimateTexasHoldemGame, lastErr error) string {
	var sb strings.Builder

	sb.WriteString("----------\n")
	fmt.Fprintf(&sb, "%s\n", i18n.Tf("ultimatetexasholdem.chipsLine", "chips", strconv.Itoa(g.GetChips())))
	fmt.Fprintf(&sb, "%s\n", i18n.Tf("ultimatetexasholdem.phaseLine", "phase", up.phaseStr(g.GetPhase())))

	// During play (not the final result block), surface the current bets so the
	// player can size the play-bet multiple; omitted before the ante is placed.
	if !g.GetGameEndFlag() && g.GetAnteBet() > 0 {
		fmt.Fprintf(&sb, "%s\n", i18n.Tf("ultimatetexasholdem.anteLine", "ante", strconv.Itoa(g.GetAnteBet())))
		fmt.Fprintf(&sb, "%s\n", i18n.Tf("ultimatetexasholdem.blindLine", "blind", strconv.Itoa(g.GetBlindBet())))
		if g.GetTripsBet() > 0 {
			fmt.Fprintf(&sb, "%s\n", i18n.Tf("ultimatetexasholdem.tripsLine", "trips", strconv.Itoa(g.GetTripsBet())))
		}
		if play := g.GetPlayBet(); play > 0 {
			fmt.Fprintf(&sb, "%s\n", i18n.Tf("ultimatetexasholdem.playBetLine", "play", strconv.Itoa(play)))
		}
	}

	// **配当を知らないままトリップスの額を決めさせない。**トリップスは
	// フォールドしても評価される特殊なサイドベットなのに、CUI には倍率を知る
	// 手段が無かった (#5589)。Baccarat が #5497 で入れたのと同じ形で、倍率は
	// ドメインの定数から作る ── 文言に書き写すと、配当を変えたとき嘘の表が残る。
	//
	// ベットフェーズだけに出す。賭けた後の卓に配当表を並べても、いま起きた
	// ことが読み取りにくくなるだけ。
	if g.GetPhase() == domain.UltimateTexasHoldemPhaseBet {
		up.writePayoutRef(&sb)
	}

	if community := g.GetCommunity(); len(community) > 0 {
		sb.WriteString("--- " + color.Bold(i18n.T("ultimatetexasholdem.boardHeader")) + " ---\n")
		parts := make([]string, len(community))
		for i, card := range community {
			parts[i] = cuiCardStr(card)
		}
		sb.WriteString(strings.Join(parts, ","))
		sb.WriteString("\n")
	}

	if playerHand := g.GetPlayerHand(); len(playerHand) > 0 {
		sb.WriteString("--- " + color.Bold(i18n.T("ultimatetexasholdem.playerHeader")) + " ---\n")
		if g.GetPhase() == domain.UltimateTexasHoldemPhaseEnd {
			rank := g.GetPlayerHandRank()
			if rank >= 0 && rank < len(domain.PokerHandNames) {
				fmt.Fprintf(&sb, "%s\n", i18n.Tf("ultimatetexasholdem.handLine", "hand", cuiPokerHandName(rank)))
			}
		}
		parts := make([]string, len(playerHand))
		for i, card := range playerHand {
			parts[i] = cuiCardStr(card)
		}
		sb.WriteString(strings.Join(parts, ","))
		sb.WriteString("\n")
	}

	if dealerHand := g.GetDealerHand(); len(dealerHand) > 0 {
		sb.WriteString("--- " + color.Bold(i18n.T("ultimatetexasholdem.dealerHeader")) + " ---\n")
		if g.GetPhase() == domain.UltimateTexasHoldemPhaseEnd {
			rank := g.GetDealerHandRank()
			if rank >= 0 && rank < len(domain.PokerHandNames) {
				fmt.Fprintf(&sb, "%s\n", i18n.Tf("ultimatetexasholdem.handLine", "hand", cuiPokerHandName(rank)))
			}
			parts := make([]string, len(dealerHand))
			for i, card := range dealerHand {
				parts[i] = cuiCardStr(card)
			}
			sb.WriteString(strings.Join(parts, ","))
			sb.WriteString("\n")
		} else {
			parts := make([]string, len(dealerHand))
			for i := range dealerHand {
				parts[i] = "??"
			}
			sb.WriteString(strings.Join(parts, ","))
			sb.WriteString("\n")
		}
	}

	sb.WriteString("----------\n")

	if lastErr != nil {
		fmt.Fprintf(&sb, "%s\n", i18n.MarkErrorLine(color.Red(lastErr.Error())))
	}

	if g.GetGameEndFlag() {
		fmt.Fprintf(&sb, "%s\n", i18n.Tf("ultimatetexasholdem.anteLine", "ante", strconv.Itoa(g.GetAnteBet())))
		fmt.Fprintf(&sb, "%s\n", i18n.Tf("ultimatetexasholdem.blindLine", "blind", strconv.Itoa(g.GetBlindBet())))
		if g.GetTripsBet() > 0 {
			fmt.Fprintf(&sb, "%s\n", i18n.Tf("ultimatetexasholdem.tripsLine", "trips", strconv.Itoa(g.GetTripsBet())))
		}
		if play := g.GetPlayBet(); play > 0 {
			fmt.Fprintf(&sb, "%s\n", i18n.Tf("ultimatetexasholdem.playBetLine", "play", strconv.Itoa(play)))
		}
		if g.GetDealerQualified() {
			sb.WriteString(i18n.T("ultimatetexasholdem.dealerQualified") + "\n")
		} else {
			sb.WriteString(i18n.T("ultimatetexasholdem.dealerNotQualified") + "\n")
		}
		switch g.GetResult() {
		case domain.GameResultWin:
			sb.WriteString(color.Green(i18n.T("ultimatetexasholdem.playerWins")) + "\n")
		case domain.GameResultLose:
			if g.GetFolded() {
				sb.WriteString(color.Red(i18n.T("ultimatetexasholdem.playerFolded")) + "\n")
			} else {
				sb.WriteString(color.Red(i18n.T("ultimatetexasholdem.dealerWins")) + "\n")
			}
		case domain.GameResultDraw:
			sb.WriteString(color.Yellow(i18n.T("ultimatetexasholdem.push")) + "\n")
		default:
		}
		fmt.Fprintf(&sb, "%s\n", i18n.Tf("ultimatetexasholdem.totalPayoutLine", "payout", strconv.Itoa(g.GetTotalPayout())))
		sb.WriteString("----------\n")
	}

	return sb.String()
}

// uthHintKeys は推奨アクションから表示用の i18n キーへの対応。
var uthHintKeys = map[string]string{
	domain.UTHRecommendPlay4x: "ultimatetexasholdem.hintPlay4x",
	domain.UTHRecommendPlay3x: "ultimatetexasholdem.hintPlay3x",
	domain.UTHRecommendPlay2x: "ultimatetexasholdem.hintPlay2x",
	domain.UTHRecommendPlay1x: "ultimatetexasholdem.hintPlay1x",
	domain.UTHRecommendCheck:  "ultimatetexasholdem.hintCheck",
	domain.UTHRecommendFold:   "ultimatetexasholdem.hintFold",
}

// HintOutput は現在のフェーズでの推奨アクションを出力する。
//
// **CUI には 4x/3x/2x/1x/check/fold を選ぶ材料が何も無かった (#4709)。**
// Web はプリフロップの強さで 4x / 3x ボタンを光らせている。判定はドメインの
// RecommendPlay 1か所に置いたので、CUI と Web が違う倍率を指すことはない。
func (up *UltimateTexasHoldemCuiPresenter) HintOutput(g interfaces.UltimateTexasHoldemGame) string {
	key, ok := uthHintKeys[g.RecommendPlay()]
	if !ok {
		return i18n.T("ultimatetexasholdem.hintNone") + "\n"
	}
	return color.Yellow(i18n.T(key)) + "\n"
}

// ActionLogOutput 棋譜をテキスト出力
func (up *UltimateTexasHoldemCuiPresenter) ActionLogOutput(g interfaces.UltimateTexasHoldemGame) string {
	return actionLogOutputText(g)
}

// phaseStr フェーズ文字列
func (up *UltimateTexasHoldemCuiPresenter) phaseStr(phase int) string {
	switch phase {
	case domain.UltimateTexasHoldemPhaseBet:
		return i18n.T("ultimatetexasholdem.phaseBet")
	case domain.UltimateTexasHoldemPhasePreFlop:
		return i18n.T("ultimatetexasholdem.phasePreFlop")
	case domain.UltimateTexasHoldemPhaseFlop:
		return i18n.T("ultimatetexasholdem.phaseFlop")
	case domain.UltimateTexasHoldemPhaseRiver:
		return i18n.T("ultimatetexasholdem.phaseRiver")
	case domain.UltimateTexasHoldemPhaseEnd:
		return i18n.T("ultimatetexasholdem.phaseEnd")
	default:
		return i18n.T("ultimatetexasholdem.phaseUnknown")
	}
}
