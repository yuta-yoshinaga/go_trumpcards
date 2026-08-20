//go:build !js || !wasm || extra4

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// ColourWhistCuiPresenter カラーホイストCUIプレゼンタークラス
type ColourWhistCuiPresenter struct{}

// Output ゲーム状態を出力
func (cp *ColourWhistCuiPresenter) Output(c interfaces.ColourWhistGame, lastErr error) string {
	return buildCuiOutput(i18n.T("colourwhist.outputTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("colourwhist.phaseLine", "phase", cp.phaseStr(c.GetPhase())) + "\n")
		sb.WriteString(i18n.Tf("colourwhist.roundLine",
			"round", strconv.Itoa(c.GetRoundNumber()),
			"total", strconv.Itoa(c.GetConfig().Rounds)) + "\n")
		sb.WriteString(i18n.Tf("colourwhist.contractLine",
			"contract", cp.contractStr(c.GetContract()),
			"trump", cp.trumpStr(c.GetTrumpSuit())) + "\n")
		// **Troel は配りで成立した契約。** 競りで選んだのではないことを明示します。
		if c.IsTroelForced() {
			sb.WriteString(i18n.Tf("colourwhist.troelLine",
				"seat", strconv.Itoa(c.GetDeclarerIdx())) + "\n")
		}
		if c.GetContract() != domain.ColourWhistContractNone && c.GetDeclarerIdx() >= 0 {
			sb.WriteString(i18n.Tf("colourwhist.declarerLine",
				"seat", strconv.Itoa(c.GetDeclarerIdx()),
				"tricks", strconv.Itoa(c.GetDeclarerTricks())) + "\n")
		}
		cp.writeScores(sb, c)
		cp.writeTrick(sb, c)
		cp.writeHand(sb, c)
		cuiErrorBlock(sb, lastErr)

		if c.GetGameEndFlag() {
			msg := i18n.Tf("colourwhist.winnerLine", "seat", strconv.Itoa(c.GetWinnerIdx()))
			if c.GetWinnerIdx() == 0 {
				sb.WriteString(color.Green(msg) + "\n")
			} else {
				sb.WriteString(color.Red(msg) + "\n")
			}
		}
	})
}

// writeScores は各席の得点を書き出す。**負にもなります。**
func (cp *ColourWhistCuiPresenter) writeScores(sb *strings.Builder, c interfaces.ColourWhistGame) {
	parts := make([]string, 0, c.GetPlayerCnt())
	for i := range c.GetPlayerCnt() {
		p := c.GetPlayer(i)
		if p == nil {
			continue
		}
		parts = append(parts, "#"+strconv.Itoa(i)+":"+strconv.Itoa(p.GetScore()))
	}
	sb.WriteString(i18n.Tf("colourwhist.scoreLine", "scores", strings.Join(parts, " ")) + "\n")
}

// writeTrick は進行中のトリックを書き出す。
func (cp *ColourWhistCuiPresenter) writeTrick(sb *strings.Builder, c interfaces.ColourWhistGame) {
	trick := c.GetTrick()
	if len(trick) == 0 {
		return
	}
	sb.WriteString("----------\n")
	for _, tc := range trick {
		sb.WriteString(i18n.Tf("colourwhist.trickCardLine",
			"seat", strconv.Itoa(tc.PlayerIdx), "card", cuiCardStr(tc.Card)) + "\n")
	}
}

// writeHand は人間の手札を書き出す。出せる札に印を付けます。
func (cp *ColourWhistCuiPresenter) writeHand(sb *strings.Builder, c interfaces.ColourWhistGame) {
	p := c.GetPlayer(0)
	if p == nil || p.GetCardsSize() == 0 {
		return
	}
	valid := map[int]bool{}
	for _, v := range c.GetValidPlayIndices(0) {
		valid[v] = true
	}
	parts := make([]string, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		mark := " "
		if valid[i] {
			mark = "*"
		}
		parts = append(parts, "["+strconv.Itoa(i)+mark+"]"+cuiCardStr(p.GetCard(i)))
	}
	sb.WriteString("----------\n")
	sb.WriteString(i18n.Tf("colourwhist.handLine", "cards", strings.Join(parts, " ")) + "\n")
}

// ActionLogOutput 棋譜をテキスト出力
func (cp *ColourWhistCuiPresenter) ActionLogOutput(c interfaces.ColourWhistGame) string {
	return actionLogOutputText(c)
}

// HintOutput ヒントをテキスト出力
func (cp *ColourWhistCuiPresenter) HintOutput(c interfaces.ColourWhistGame) string {
	h := c.GetHint()
	if h == nil {
		return i18n.T("colourwhist.hintNone")
	}
	if h.Contract != nil {
		return i18n.Tf("colourwhist.hintBid",
			"contract", cp.contractStr(*h.Contract), "reason", i18n.T("colourwhist."+h.Reason))
	}
	if h.CardIndex != nil {
		return i18n.Tf("colourwhist.hintPlay",
			"index", strconv.Itoa(*h.CardIndex), "reason", i18n.T("colourwhist."+h.Reason))
	}
	return i18n.T("colourwhist.hintNone")
}

// phaseStr フェーズ文字列
func (cp *ColourWhistCuiPresenter) phaseStr(phase int) string {
	switch phase {
	case domain.ColourWhistPhaseBid:
		return i18n.T("colourwhist.phaseBid")
	case domain.ColourWhistPhaseCall:
		return i18n.T("colourwhist.phaseCall")
	case domain.ColourWhistPhasePlay:
		return i18n.T("colourwhist.phasePlay")
	case domain.ColourWhistPhaseRoundEnd:
		return i18n.T("colourwhist.phaseRoundEnd")
	case domain.ColourWhistPhaseGameEnd:
		return i18n.T("colourwhist.phaseGameEnd")
	default:
		return i18n.T("colourwhist.phaseUnknown")
	}
}

// contractStr 契約の文字列
func (cp *ColourWhistCuiPresenter) contractStr(contract int) string {
	switch contract {
	case domain.ColourWhistContractSamen:
		return i18n.T("colourwhist.contractSamen")
	case domain.ColourWhistContractAlleen:
		return i18n.T("colourwhist.contractAlleen")
	case domain.ColourWhistContractMiserie:
		return i18n.T("colourwhist.contractMiserie")
	case domain.ColourWhistContractTroel:
		return i18n.T("colourwhist.contractTroel")
	default:
		return i18n.T("colourwhist.contractNone")
	}
}

// trumpStr 切り札の文字列
func (cp *ColourWhistCuiPresenter) trumpStr(suit int) string {
	switch suit {
	case domain.CardDesignSpade:
		return i18n.T("colourwhist.trumpSpade")
	case domain.CardDesignClover:
		return i18n.T("colourwhist.trumpClover")
	case domain.CardDesignHeart:
		return i18n.T("colourwhist.trumpHeart")
	case domain.CardDesignDiamond:
		return i18n.T("colourwhist.trumpDiamond")
	default:
		return i18n.T("colourwhist.trumpNone")
	}
}
