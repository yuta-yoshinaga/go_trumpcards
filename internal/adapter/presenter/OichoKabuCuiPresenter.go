//go:build !js || !wasm || extra

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// OichoKabuCuiPresenter おいちょかぶCUIプレゼンタークラス
type OichoKabuCuiPresenter struct{}

// oichoKabuCardStr はカブ札を数字文字列で表す（フレンチスートは使わない）。
func oichoKabuCardStr(c *domain.Card) string {
	if c == nil {
		return "?"
	}
	return strconv.Itoa(c.GetValue())
}

// oichoKabuHandStr は手札スライスをカンマ区切りの数字列にする。
func oichoKabuHandStr(hand []*domain.Card) string {
	parts := make([]string, len(hand))
	for i, c := range hand {
		parts[i] = oichoKabuCardStr(c)
	}
	return strings.Join(parts, ",")
}

// Output ゲーム状態を出力
func (p *OichoKabuCuiPresenter) Output(o interfaces.OichoKabuGame, lastErr error) string {
	var sb strings.Builder

	sb.WriteString("----------\n")
	sb.WriteString(i18n.Tf("oichokabu.chipsLine", "chips", strconv.Itoa(o.GetChips())) + "\n")
	sb.WriteString(i18n.Tf("oichokabu.phaseLine", "phase", p.phaseStr(o.GetPhase())) + "\n")
	if o.GetBet() > 0 {
		sb.WriteString(i18n.Tf("oichokabu.betLine", "bet", strconv.Itoa(o.GetBet())) + "\n")
	}

	// Give CLI players the same guidance the web UI shows: the bet ceiling
	// (current chips) while betting, and the draw/stand choice while drawing.
	if !o.GetGameEndFlag() {
		switch o.GetPhase() {
		case domain.OichoKabuPhaseBet:
			sb.WriteString(i18n.Tf("oichokabu.maxBetHint", "max", strconv.Itoa(o.GetChips())) + "\n")
		case domain.OichoKabuPhaseDraw:
			sb.WriteString(i18n.T("oichokabu.drawHint") + "\n")
		}
	}

	if hand := o.GetPlayerHand(); len(hand) > 0 {
		sb.WriteString("--- " + color.Bold(i18n.T("oichokabu.playerHeader")) + " ---\n")
		sb.WriteString(i18n.Tf("oichokabu.handLine", "cards", oichoKabuHandStr(hand)) + "\n")
		sb.WriteString(i18n.Tf("oichokabu.rankLine", "rank", strconv.Itoa(o.GetPlayerRank())) + "\n")
	}

	// 親の手は結果まで伏せる。
	sb.WriteString("--- " + color.Bold(i18n.T("oichokabu.bankerHeader")) + " ---\n")
	if o.GetGameEndFlag() {
		sb.WriteString(i18n.Tf("oichokabu.handLine", "cards", oichoKabuHandStr(o.GetBankerHand())) + "\n")
		sb.WriteString(i18n.Tf("oichokabu.rankLine", "rank", strconv.Itoa(o.GetBankerRank())) + "\n")
	} else {
		sb.WriteString(i18n.T("oichokabu.hidden") + "\n")
	}

	sb.WriteString("----------\n")

	if lastErr != nil {
		sb.WriteString(color.Red(lastErr.Error()) + "\n")
	}

	if o.GetGameEndFlag() {
		switch o.GetResult() {
		case domain.OichoKabuResultWin:
			sb.WriteString(color.Green(i18n.T("oichokabu.playerWins")) + "\n")
		case domain.OichoKabuResultLose:
			sb.WriteString(color.Red(i18n.T("oichokabu.bankerWins")) + "\n")
		default:
			sb.WriteString(color.Yellow(i18n.T("oichokabu.push")) + "\n")
		}
		sb.WriteString(i18n.Tf("oichokabu.totalPayoutLine", "payout", strconv.Itoa(o.GetTotalPayout())) + "\n")
		sb.WriteString("----------\n")
	}

	return sb.String()
}

// ActionLogOutput 棋譜をテキスト出力
func (p *OichoKabuCuiPresenter) ActionLogOutput(o interfaces.OichoKabuGame) string {
	return actionLogOutputText(o)
}

// phaseStr フェーズ文字列
func (p *OichoKabuCuiPresenter) phaseStr(phase int) string {
	switch phase {
	case domain.OichoKabuPhaseBet:
		return i18n.T("oichokabu.phaseBet")
	case domain.OichoKabuPhaseDraw:
		return i18n.T("oichokabu.phaseDraw")
	case domain.OichoKabuPhaseEnd:
		return i18n.T("oichokabu.phaseEnd")
	default:
		return i18n.T("oichokabu.phaseUnknown")
	}
}
