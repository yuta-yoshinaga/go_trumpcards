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

// MonteBankCuiPresenter モンテバンクCUIプレゼンタークラス
type MonteBankCuiPresenter struct{}

// Output ゲーム状態を出力
func (cp *MonteBankCuiPresenter) Output(c interfaces.MonteBankGame, lastErr error) string {
	return buildCuiOutput(i18n.T("montebank.outputTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("montebank.phaseLine", "phase", cp.phaseStr(c.GetPhase())) + "\n")
		sb.WriteString(i18n.Tf("montebank.roundLine",
			"round", strconv.Itoa(c.GetRoundNumber()),
			"chips", strconv.Itoa(c.GetChips()),
			"remaining", strconv.Itoa(c.GetRemainingCards())) + "\n")
		cp.writeLayout(sb, c)
		cp.writeGate(sb, c)
		cuiErrorBlock(sb, lastErr)
		cp.writeResult(sb, c)
	})
}

// writeLayout は場札と、そのスートが何枚出ているかを書き出す。
//
// **同じスートが何枚出ているかを必ず添える。** それが賭けの良し悪しを決める
// 唯一の数字なので、伏せると勝負が運任せになる。
func (cp *MonteBankCuiPresenter) writeLayout(sb *strings.Builder, c interfaces.MonteBankGame) {
	layout := c.GetLayout()
	if len(layout) == 0 {
		return
	}
	sb.WriteString("----------\n")
	for i, card := range layout {
		if card == nil {
			continue
		}
		mark := " "
		if i == c.GetPick() {
			mark = "*"
		}
		count := c.SuitCountInLayout(card.GetDesign())
		note := i18n.Tf("montebank.suitDupNote", "count", strconv.Itoa(count))
		if count == 1 {
			note = i18n.T("montebank.suitEvenNote")
		}
		sb.WriteString(i18n.Tf("montebank.layoutLine",
			"mark", mark,
			"idx", strconv.Itoa(i+1),
			"card", cuiCardStr(card),
			"note", note) + "\n")
	}
}

// writeGate はめくった 1 枚を書き出す。
func (cp *MonteBankCuiPresenter) writeGate(sb *strings.Builder, c interfaces.MonteBankGame) {
	gate := c.GetGate()
	if gate == nil {
		return
	}
	sb.WriteString(i18n.Tf("montebank.gateLine", "card", cuiCardStr(gate)) + "\n")
}

// writeResult は決着と収支を書き出す。
func (cp *MonteBankCuiPresenter) writeResult(sb *strings.Builder, c interfaces.MonteBankGame) {
	if c.GetResult() == domain.MonteBankResultNone {
		return
	}
	net := c.GetPayout() - c.GetBet()
	msg := i18n.Tf("montebank.resultLine",
		"result", i18n.T("montebank.result."+domain.MonteBankResultName(c.GetResult())),
		"net", strconv.Itoa(net))
	if net >= 0 {
		sb.WriteString(color.Green(msg) + "\n")
	} else {
		sb.WriteString(color.Red(msg) + "\n")
	}
	if c.GetGameEndFlag() {
		sb.WriteString(i18n.Tf("montebank.gameEndLine", "chips", strconv.Itoa(c.GetChips())) + "\n")
	}
}

// ActionLogOutput 棋譜をテキスト出力
func (cp *MonteBankCuiPresenter) ActionLogOutput(c interfaces.MonteBankGame) string {
	return actionLogOutputText(c)
}

// HintOutput ヒントをテキスト出力
func (cp *MonteBankCuiPresenter) HintOutput(c interfaces.MonteBankGame) string {
	h := c.GetHint()
	if h == nil {
		return i18n.T("montebank.hintNone")
	}
	return i18n.Tf("montebank.hint",
		"idx", strconv.Itoa(h.PickIdx+1),
		"reason", i18n.T("montebank.reason."+h.Reason))
}

// phaseStr フェーズ文字列
func (cp *MonteBankCuiPresenter) phaseStr(phase domain.MonteBankPhase) string {
	switch phase {
	case domain.MonteBankPhaseBet:
		return i18n.T("montebank.phaseBet")
	case domain.MonteBankPhaseResult:
		return i18n.T("montebank.phaseResult")
	case domain.MonteBankPhaseGameEnd:
		return i18n.T("montebank.phaseGameEnd")
	default:
		return i18n.T("montebank.phaseUnknown")
	}
}
