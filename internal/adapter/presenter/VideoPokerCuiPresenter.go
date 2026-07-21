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

// VideoPokerCuiPresenter ビデオポーカーCUIプレゼンタークラス
type VideoPokerCuiPresenter struct {
}

// Output ゲーム状態を出力
func (vpp *VideoPokerCuiPresenter) Output(vp interfaces.VideoPokerGame, lastErr error) string {
	var sb strings.Builder

	sb.WriteString("----------\n")
	sb.WriteString(i18n.Tf("videopoker.chipsLine", "chips", strconv.Itoa(vp.GetChips())) + "\n")
	sb.WriteString(i18n.Tf("videopoker.phaseLine", "phase", vpp.phaseStr(vp.GetPhase())) + "\n")

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
		sb.WriteString(vpp.paytableStr(vp.GetVariantName()))
	}

	sb.WriteString("----------\n")

	if lastErr != nil {
		sb.WriteString(color.Red(lastErr.Error()) + "\n")
	}

	if vp.GetGameEndFlag() {
		sb.WriteString(i18n.Tf("videopoker.betLine", "bet", strconv.Itoa(vp.GetBetAmount())) + "\n")
		if vp.GetResult() == domain.GameResultWin {
			sb.WriteString(color.Green(i18n.Tf("videopoker.winLine", "handName", vp.GetHandName())) + "\n")
		} else {
			sb.WriteString(color.Red(i18n.T("videopoker.noWin")) + "\n")
		}
		sb.WriteString(i18n.Tf("videopoker.payoutLine", "payout", strconv.Itoa(vp.GetPayout())) + "\n")
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

// videoPokerDeucesWildHold returns a hold mask for Deuces Wild: always hold the
// wild deuces, plus any rank group of two or more (a made pair or better).
func videoPokerDeucesWildHold(hand []*domain.Card) []bool {
	hold := make([]bool, len(hand))
	rankIdx := map[int][]int{}
	for i, c := range hand {
		if c.GetValue() == 2 { // deuces are wild
			hold[i] = true
			continue
		}
		rankIdx[c.GetValue()] = append(rankIdx[c.GetValue()], i)
	}
	for _, idxs := range rankIdx {
		if len(idxs) >= 2 {
			for _, i := range idxs {
				hold[i] = true
			}
		}
	}
	return hold
}

// HintOutput recommends which cards to hold during the draw phase of Deuces
// Wild (hold all deuces and any made pair-or-better). Other variants and phases
// get no hint.
func (p *VideoPokerCuiPresenter) HintOutput(g interfaces.VideoPokerGame) string {
	if g.GetPhase() != domain.VideoPokerPhaseDraw || g.GetVariantName() != "deuceswild" {
		return i18n.T("videopoker.hintNone") + "\n"
	}
	hand := g.GetHand()
	hold := videoPokerDeucesWildHold(hand)
	var parts []string
	for i, h := range hold {
		if h {
			parts = append(parts, "["+strconv.Itoa(i)+"]"+cuiCardStr(hand[i]))
		}
	}
	if len(parts) == 0 {
		return color.Yellow(i18n.T("videopoker.hintHoldNone")) + "\n"
	}
	return color.Yellow(i18n.Tf("videopoker.hintHold", "cards", strings.Join(parts, " "))) + "\n"
}
