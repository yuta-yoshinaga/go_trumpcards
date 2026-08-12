//go:build !js || !wasm || extra2

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// RikkenCuiPresenter リッケンCUIプレゼンタークラス
type RikkenCuiPresenter struct{}

// Output ゲーム状態を出力
func (rp *RikkenCuiPresenter) Output(r interfaces.RikkenGame, lastErr error) string {
	return buildCuiOutput(i18n.T("rikken.outputTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("rikken.phaseLine", "phase", rp.phaseStr(r.GetPhase())) + "\n")
		sb.WriteString(i18n.Tf("rikken.roundLine",
			"round", strconv.Itoa(r.GetRoundNumber()),
			"total", strconv.Itoa(r.GetConfig().Rounds)) + "\n")
		sb.WriteString(i18n.Tf("rikken.contractLine",
			"contract", rp.contractStr(r.GetContract()),
			"trump", rp.trumpStr(r.GetTrumpSuit())) + "\n")
		if r.GetContract() != domain.RikkenContractNone && r.GetDeclarerIdx() >= 0 {
			sb.WriteString(i18n.Tf("rikken.declarerLine",
				"seat", strconv.Itoa(r.GetDeclarerIdx()),
				"tricks", strconv.Itoa(r.GetDeclarerTricks())) + "\n")
		}
		rp.writeScores(sb, r)
		rp.writeTrick(sb, r)
		rp.writeHand(sb, r)
		cuiErrorBlock(sb, lastErr)

		if r.GetGameEndFlag() {
			msg := i18n.Tf("rikken.winnerLine", "seat", strconv.Itoa(r.GetWinnerIdx()))
			if r.GetWinnerIdx() == 0 {
				sb.WriteString(color.Green(msg) + "\n")
			} else {
				sb.WriteString(color.Red(msg) + "\n")
			}
		}
	})
}

// writeScores は各席の得点を書き出す。**負にもなります。**
func (rp *RikkenCuiPresenter) writeScores(sb *strings.Builder, r interfaces.RikkenGame) {
	parts := make([]string, 0, r.GetPlayerCnt())
	for i := range r.GetPlayerCnt() {
		p := r.GetPlayer(i)
		if p == nil {
			continue
		}
		parts = append(parts, "#"+strconv.Itoa(i)+":"+strconv.Itoa(p.GetScore()))
	}
	sb.WriteString(i18n.Tf("rikken.scoreLine", "scores", strings.Join(parts, " ")) + "\n")
}

// writeTrick は進行中のトリックを書き出す。
func (rp *RikkenCuiPresenter) writeTrick(sb *strings.Builder, r interfaces.RikkenGame) {
	trick := r.GetTrick()
	if len(trick) == 0 {
		return
	}
	sb.WriteString("----------\n")
	for _, tc := range trick {
		sb.WriteString(i18n.Tf("rikken.trickCardLine",
			"seat", strconv.Itoa(tc.PlayerIdx), "card", cuiCardStr(tc.Card)) + "\n")
	}
}

// writeHand は人間の手札を書き出す。出せる札に印を付けます。
//
// **オープンミゼールでは宣言者の手札も見せます。** 手札を公開して 0 トリックを
// 目指す契約なので、隠したままだと名前どおりの仕掛けが働きません。
func (rp *RikkenCuiPresenter) writeHand(sb *strings.Builder, r interfaces.RikkenGame) {
	rp.writeOpenMisereHand(sb, r)
	p := r.GetPlayer(0)
	if p == nil || p.GetCardsSize() == 0 {
		return
	}
	valid := map[int]bool{}
	for _, v := range r.GetValidPlayIndices(0) {
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
	sb.WriteString(i18n.Tf("rikken.handLine", "cards", strings.Join(parts, " ")) + "\n")
}

// ActionLogOutput 棋譜をテキスト出力
func (rp *RikkenCuiPresenter) ActionLogOutput(r interfaces.RikkenGame) string {
	return actionLogOutputText(r)
}

// HintOutput ヒントをテキスト出力
func (rp *RikkenCuiPresenter) HintOutput(r interfaces.RikkenGame) string {
	h := r.GetHint()
	if h == nil {
		return i18n.T("rikken.hintNone")
	}
	if h.Contract != nil {
		return i18n.Tf("rikken.hintBid",
			"contract", rp.contractStr(*h.Contract), "reason", i18n.T("rikken."+h.Reason))
	}
	if h.CardIndex != nil {
		return i18n.Tf("rikken.hintPlay",
			"index", strconv.Itoa(*h.CardIndex), "reason", i18n.T("rikken."+h.Reason))
	}
	return i18n.T("rikken.hintNone")
}

// phaseStr フェーズ文字列
func (rp *RikkenCuiPresenter) phaseStr(phase int) string {
	switch phase {
	case domain.RikkenPhaseBid:
		return i18n.T("rikken.phaseBid")
	case domain.RikkenPhaseCall:
		return i18n.T("rikken.phaseCall")
	case domain.RikkenPhasePlay:
		return i18n.T("rikken.phasePlay")
	case domain.RikkenPhaseRoundEnd:
		return i18n.T("rikken.phaseRoundEnd")
	case domain.RikkenPhaseGameEnd:
		return i18n.T("rikken.phaseGameEnd")
	default:
		return i18n.T("rikken.phaseUnknown")
	}
}

// contractStr 契約の文字列
func (rp *RikkenCuiPresenter) contractStr(contract int) string {
	return i18n.T("rikken.contract" + strings.ToUpper(domain.RikkenContractName(contract)[:1]) +
		domain.RikkenContractName(contract)[1:])
}

// trumpStr 切り札の文字列
func (rp *RikkenCuiPresenter) trumpStr(suit int) string {
	switch suit {
	case domain.CardDesignSpade:
		return i18n.T("rikken.trumpSpade")
	case domain.CardDesignClover:
		return i18n.T("rikken.trumpClover")
	case domain.CardDesignHeart:
		return i18n.T("rikken.trumpHeart")
	case domain.CardDesignDiamond:
		return i18n.T("rikken.trumpDiamond")
	default:
		return i18n.T("rikken.trumpNone")
	}
}

// writeOpenMisereHand はオープンミゼールの宣言者の手札を公開する。
func (rp *RikkenCuiPresenter) writeOpenMisereHand(sb *strings.Builder, r interfaces.RikkenGame) {
	if r.GetContract() != domain.RikkenContractOpenMisere {
		return
	}
	idx := r.GetDeclarerIdx()
	if idx < 0 || idx == 0 {
		return // 人間が宣言者なら下の手札表示で出ます。
	}
	p := r.GetPlayer(idx)
	if p == nil || p.GetCardsSize() == 0 {
		return
	}
	parts := make([]string, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		parts = append(parts, cuiCardStr(p.GetCard(i)))
	}
	sb.WriteString("----------\n")
	sb.WriteString(i18n.Tf("rikken.openHandLine",
		"seat", strconv.Itoa(idx), "cards", strings.Join(parts, " ")) + "\n")
}
