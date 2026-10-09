//go:build !js || !wasm || extra8

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// VideoPokerCuiPresenter ビデオポーカーCUIプレゼンタークラス
type VideoPokerCuiPresenter struct {
}

// Output ゲーム状態を出力
func (vpp *VideoPokerCuiPresenter) Output(vp interfaces.VideoPokerGame, lastErr error) string {
	var sb strings.Builder

	sb.WriteString("----------\n")
	sb.WriteString(i18n.Tf("videopoker.chipsLine", "chips", strconv.Itoa(vp.GetChips())) + "\n")
	sb.WriteString(i18n.Tf("videopoker.phaseLine", "phase", vpp.phaseStr(vp.GetPhase())) + "\n")
	stats := videoPokerSessionStats(vp)
	sb.WriteString(i18n.Tf("videopoker.statsSummary",
		"hands", strconv.Itoa(stats.hands),
		"winRate", strconv.Itoa(stats.winRate),
		"net", formatSigned(stats.net)) + "\n")

	hand := vp.GetHand()
	if len(hand) > 0 {
		sb.WriteString(i18n.T("videopoker.handHeader") + "\n")
		held := vp.GetHeldIndices()
		holdLabel := i18n.T("videopoker.holdLabel")
		parts := make([]string, len(hand))
		for i, card := range hand {
			s := vpp.cardStr(vp, card)
			if held[i] {
				s += " " + holdLabel
			}
			parts[i] = s
		}
		sb.WriteString(strings.Join(parts, ", "))
		sb.WriteString("\n")
	}

	// ベットフェーズではベット額決定の判断材料として配当表を表示する。
	if vp.GetPhase() == domain.VideoPokerPhaseBet {
		// Reset silently tops the balance back up below the minimum bet, so the
		// chips appear from nowhere unless this says otherwise.
		if vp.GetChipsRefilled() {
			sb.WriteString(color.Yellow(i18n.Tf("videopoker.chipsRefilled",
				"chips", strconv.Itoa(domain.VideoPokerDefaultChips))) + "\n")
		}
		sb.WriteString(vpp.paytableStr(vp.GetVariantName()))
	}

	// **いま何の役になっているかをドロー中に出す。** Web は vp-made-hand で
	// リアルタイムに見せているのに、CUI は手札とホールド推奨しか出しておらず、
	// 配当対象かどうかはプレイヤーが自分で判定するしかなかった (#5508)。
	// 評価はバリアント自身の GetResult を通すので、3変種とも下限がずれない。
	if vp.GetPhase() == domain.VideoPokerPhaseDraw {
		if key := vp.GetCurrentHandKey(); key != "" {
			sb.WriteString(i18n.Tf("videopoker.madeHandLine",
				"handName", vpp.translateHandKey(key)) + "\n")
		} else {
			sb.WriteString(i18n.T("videopoker.madeHandNone") + "\n")
		}
	}

	sb.WriteString("----------\n")

	if lastErr != nil {
		sb.WriteString(i18n.MarkErrorLine(color.Red(lastErr.Error())) + "\n")
	}

	if vp.GetGameEndFlag() {
		sb.WriteString(i18n.Tf("videopoker.betLine", "bet", strconv.Itoa(vp.GetBetAmount())) + "\n")
		if vp.GetResult() == domain.GameResultWin {
			sb.WriteString(color.Green(i18n.Tf("videopoker.winLine", "handName", vpp.handNameForWin(vp))) + "\n")
		} else {
			sb.WriteString(color.Red(i18n.T("videopoker.noWin")) + "\n")
		}
		sb.WriteString(i18n.Tf("videopoker.payoutLine", "payout", strconv.Itoa(vp.GetPayout())) + "\n")
		sb.WriteString(i18n.Tf("videopoker.netChangeLine", "net", formatSigned(vp.GetPayout()-vp.GetBetAmount())) + "\n")
		sb.WriteString("----------\n")
	}

	return sb.String()
}

// ActionLogOutput 棋譜をテキスト出力
func (vpp *VideoPokerCuiPresenter) ActionLogOutput(vp interfaces.VideoPokerGame) string {
	return actionLogOutputText(vp)
}

// cardStr ワイルドカードを強調した手札カード文字列（ジョーカーは太字黄、Deuces Wildの2は黄）
func (vpp *VideoPokerCuiPresenter) cardStr(vp interfaces.VideoPokerGame, card *domain.Card) string {
	if card == nil {
		return cuiCardStr(card)
	}
	if card.GetDesign() == domain.CardDesignJoker {
		return color.BoldYellow("JOKER")
	}
	if card.GetValue() == 2 && vp.GetVariantName() == "deuceswild" {
		// 赤スートの通常色（赤）を上書きしないよう、素のスート名から組み立てる
		return color.Yellow(cuiSuitName(card.GetDesign()) + " 2")
	}
	return cuiCardStr(card)
}

// paytableStr はバリアント固有の配当表（役名と 1 コインあたりの倍率）を組み立てる。
// 配当値は domain.VideoPokerPaytable を単一情報源として参照する。
func (vpp *VideoPokerCuiPresenter) paytableStr(variantName string) string {
	var sb strings.Builder
	sb.WriteString(i18n.T("videopoker.payoutTitle") + "\n")
	for _, row := range domain.VideoPokerPaytable(variantName) {
		line := i18n.T("videopoker."+row.HandKey) + " x" + strconv.Itoa(row.Multiplier)
		if row.RoyalJackpot {
			line += " " + i18n.T("videopoker.payoutMaxBetNote")
		}
		sb.WriteString(line + "\n")
	}
	return sb.String()
}

// handNameForWin は勝利行に表示する役名を返す。
//
// **役名はバリアントに依らず共通。**以前は Deuces Wild だけ安定キー
// (GetHandKey) 経由で翻訳し、Joker Poker と Video Poker は英語の GetHandName に
// フォールバックしていたため、日本語ロケールでも勝敗行だけ英語で出ていた
// (#4693)。翻訳表は 3 バリアント共通の `pokerhand` に置いてある。
//
// キーが無い、または訳が無い場合だけ英語名に落とす。
func (vpp *VideoPokerCuiPresenter) handNameForWin(vp interfaces.VideoPokerGame) string {
	key := vp.GetHandKey()
	if key == "" {
		return vp.GetHandName()
	}
	if name := vpp.translateHandKey(key); name != "" {
		return name
	}
	return vp.GetHandName()
}

// translateHandKey は安定キーを pokerhand.* で訳す。訳が無ければ空文字を返す
// -- 生のキーを画面に出さないため。
func (vpp *VideoPokerCuiPresenter) translateHandKey(key string) string {
	full := "pokerhand." + key
	if name := i18n.T(full); name != full {
		return name
	}
	return ""
}

// phaseStr フェーズ文字列
func (vpp *VideoPokerCuiPresenter) phaseStr(phase int) string {
	switch phase {
	case domain.VideoPokerPhaseBet:
		return i18n.T("videopoker.phaseBet")
	case domain.VideoPokerPhaseDraw:
		return i18n.T("videopoker.phaseDraw")
	case domain.VideoPokerPhaseResult:
		return i18n.T("videopoker.phaseResult")
	default:
		return i18n.T("videopoker.phaseUnknown")
	}
}

// HintOutput recommends the domain strategy table's hold during the draw phase.
func (p *VideoPokerCuiPresenter) HintOutput(g interfaces.VideoPokerGame) string {
	if g.GetPhase() != domain.VideoPokerPhaseDraw {
		return i18n.T("videopoker.hintNone") + "\n"
	}
	advice := g.RecommendedHold()
	hand := g.GetHand()
	var parts []string
	for i, held := range advice.Hold {
		if held {
			parts = append(parts, "["+strconv.Itoa(i)+"]"+cuiCardStr(hand[i]))
		}
	}
	if len(parts) == 0 {
		return color.Yellow(i18n.T("videopoker.hintHoldNone")) + "\n"
	}
	return color.Yellow(i18n.Tf("videopoker.hintHold",
		"cards", strings.Join(parts, " "),
		"reason", i18n.T("videopoker.strategy."+advice.RuleKey),
	)) + "\n"
}
