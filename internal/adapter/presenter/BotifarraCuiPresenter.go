//go:build !js || !wasm || classic

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// BotifarraCuiPresenter ボティファラCUIプレゼンタークラス
type BotifarraCuiPresenter struct{}

// Output ゲーム状態を出力
func (bp *BotifarraCuiPresenter) Output(b interfaces.BotifarraGame, lastErr error) string {
	return buildCuiOutput(i18n.T("botifarra.outputTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("botifarra.phaseLine", "phase", bp.phaseStr(b.GetPhase())) + "\n")
		sb.WriteString(i18n.Tf("botifarra.scoreLine",
			"team0", strconv.Itoa(b.GetScore(0)),
			"team1", strconv.Itoa(b.GetScore(1)),
			"target", strconv.Itoa(b.GetConfig().TargetScore)) + "\n")
		sb.WriteString(i18n.Tf("botifarra.trumpLine",
			"trump", bp.trumpStr(b.GetTrumpSuit()),
			"multiplier", strconv.Itoa(b.GetMultiplier())) + "\n")
		sb.WriteString(i18n.Tf("botifarra.roundPointsLine",
			"team0", strconv.Itoa(b.GetRoundPoints(0)),
			"team1", strconv.Itoa(b.GetRoundPoints(1)),
			"total", strconv.Itoa(domain.BotifarraTotalPoints)) + "\n")

		bp.writeTrick(sb, b)
		bp.writeHand(sb, b)
		cuiErrorBlock(sb, lastErr)

		if b.GetGameEndFlag() {
			msg := i18n.Tf("botifarra.winnerLine", "team", strconv.Itoa(b.GetWinnerTeam()))
			if b.GetWinnerTeam() == domain.BotifarraTeamOf(0) {
				sb.WriteString(color.Green(msg) + "\n")
			} else {
				sb.WriteString(color.Red(msg) + "\n")
			}
		}
	})
}

// writeTrick は進行中のトリックを書き出す。
func (bp *BotifarraCuiPresenter) writeTrick(sb *strings.Builder, b interfaces.BotifarraGame) {
	trick := b.GetTrick()
	if len(trick) == 0 {
		return
	}
	sb.WriteString("----------\n")
	for _, tc := range trick {
		sb.WriteString(i18n.Tf("botifarra.trickCardLine",
			"seat", strconv.Itoa(tc.PlayerIdx),
			"card", cuiCardStr(tc.Card)) + "\n")
	}
}

// writeHand は人間の手札を書き出す。
//
// **出せる札に印を付けます。** 「勝てるなら勝たなければならない」ので、手札の
// 大半が出せない場面がふつうに起きます。印が無いと理由が分かりません。
func (bp *BotifarraCuiPresenter) writeHand(sb *strings.Builder, b interfaces.BotifarraGame) {
	p := b.GetPlayer(0)
	if p == nil || p.GetCardsSize() == 0 {
		return
	}
	valid := map[int]bool{}
	for _, v := range b.GetValidPlayIndices(0) {
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
	sb.WriteString(i18n.Tf("botifarra.handLine", "cards", strings.Join(parts, " ")) + "\n")
	sb.WriteString(i18n.T("botifarra.handLegend") + "\n")
}

// ActionLogOutput 棋譜をテキスト出力
func (bp *BotifarraCuiPresenter) ActionLogOutput(b interfaces.BotifarraGame) string {
	return actionLogOutputText(b)
}

// HintOutput ヒントをテキスト出力
func (bp *BotifarraCuiPresenter) HintOutput(b interfaces.BotifarraGame) string {
	h := b.GetHint()
	if h == nil {
		return i18n.T("botifarra.hintNone")
	}
	if h.Suit != nil {
		return i18n.Tf("botifarra.hintDeclare",
			"trump", bp.trumpStr(*h.Suit), "reason", i18n.T("botifarra."+h.Reason))
	}
	if h.CardIndex != nil {
		return i18n.Tf("botifarra.hintPlay",
			"index", strconv.Itoa(*h.CardIndex), "reason", i18n.T("botifarra."+h.Reason))
	}
	return i18n.T("botifarra.hintNone")
}

// phaseStr フェーズ文字列
func (bp *BotifarraCuiPresenter) phaseStr(phase int) string {
	switch phase {
	case domain.BotifarraPhaseDeclare:
		return i18n.T("botifarra.phaseDeclare")
	case domain.BotifarraPhaseDelegated:
		return i18n.T("botifarra.phaseDelegated")
	case domain.BotifarraPhaseDouble:
		return i18n.T("botifarra.phaseDouble")
	case domain.BotifarraPhasePlay:
		return i18n.T("botifarra.phasePlay")
	case domain.BotifarraPhaseRoundEnd:
		return i18n.T("botifarra.phaseRoundEnd")
	case domain.BotifarraPhaseGameEnd:
		return i18n.T("botifarra.phaseGameEnd")
	default:
		return i18n.T("botifarra.phaseUnknown")
	}
}

// trumpStr 切り札の文字列
func (bp *BotifarraCuiPresenter) trumpStr(suit int) string {
	switch suit {
	case domain.CardDesignSpade:
		return i18n.T("botifarra.trumpSpade")
	case domain.CardDesignClover:
		return i18n.T("botifarra.trumpClover")
	case domain.CardDesignHeart:
		return i18n.T("botifarra.trumpHeart")
	case domain.CardDesignDiamond:
		return i18n.T("botifarra.trumpDiamond")
	default:
		return i18n.T("botifarra.trumpNone")
	}
}
