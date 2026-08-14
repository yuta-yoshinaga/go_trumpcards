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

// ChemindeFerCuiPresenter シュマン・ド・フェールCUIプレゼンタークラス
type ChemindeFerCuiPresenter struct{}

// Output ゲーム状態を出力
func (cp *ChemindeFerCuiPresenter) Output(c interfaces.ChemindeFerGame, lastErr error) string {
	return buildCuiOutput(i18n.T("chemindefer.outputTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("chemindefer.phaseLine", "phase", cp.phaseStr(c.GetPhase())) + "\n")
		sb.WriteString(i18n.Tf("chemindefer.roundLine",
			"round", strconv.Itoa(c.GetRoundNumber()),
			"total", strconv.Itoa(c.GetConfig().Rounds)) + "\n")
		sb.WriteString(i18n.Tf("chemindefer.bankerLine",
			"seat", strconv.Itoa(c.GetBankerIdx()),
			"stake", strconv.Itoa(c.GetStake())) + "\n")

		cp.writeBetting(sb, c)
		cp.writeHands(sb, c)
		cp.writeChips(sb, c)
		cuiErrorBlock(sb, lastErr)
		cp.writeResult(sb, c)
	})
}

// writeBetting は賭けの進み具合を書き出す。
func (cp *ChemindeFerCuiPresenter) writeBetting(sb *strings.Builder, c interfaces.ChemindeFerGame) {
	if c.GetStake() == 0 {
		return
	}
	sb.WriteString(i18n.Tf("chemindefer.betLine",
		"total", strconv.Itoa(c.GetTotalBet()),
		"remaining", strconv.Itoa(c.GetRemainingStake())) + "\n")
	if turn := c.GetBetTurn(); turn >= 0 {
		_, hi := c.BetRangeFor(turn)
		sb.WriteString(i18n.Tf("chemindefer.betTurnLine",
			"seat", strconv.Itoa(turn), "max", strconv.Itoa(hi)) + "\n")
	}
	if idx := c.GetRepresentativeIdx(); idx >= 0 {
		sb.WriteString(i18n.Tf("chemindefer.representativeLine", "seat", strconv.Itoa(idx)) + "\n")
	}
}

// writeHands は親子の手札と合計値を書き出す。
//
// **どちらの手も伏せない。** 1 対 1 で突き合わせるゲームなので、盤面は両者とも公開。
func (cp *ChemindeFerCuiPresenter) writeHands(sb *strings.Builder, c interfaces.ChemindeFerGame) {
	if len(c.GetPunterHand()) == 0 && len(c.GetBankerHand()) == 0 {
		return
	}
	sb.WriteString("----------\n")
	sb.WriteString(i18n.Tf("chemindefer.punterHandLine",
		"cards", chemindeFerCardsStr(c.GetPunterHand()),
		"total", strconv.Itoa(c.GetPunterTotal())) + "\n")
	sb.WriteString(i18n.Tf("chemindefer.bankerHandLine",
		"cards", chemindeFerCardsStr(c.GetBankerHand()),
		"total", strconv.Itoa(c.GetBankerTotal())) + "\n")

	// **選べるのは合計 5 の子だけ。** 選べないことを黙って進めると、
	// 「なぜ引かされたのか」が画面から読み取れない。
	if c.GetPhase() == domain.ChemindeFerPhasePunterDraw && !c.PunterMayChoose() {
		sb.WriteString(i18n.T("chemindefer.punterForcedLine") + "\n")
	}
}

// chemindeFerCardsStr は手札を 1 行の文字列にする。
//
// 共通の cuiCardListStr は GetCard を持つ手札型を取るが、こちらの手は素の
// []*domain.Card なのでここで組み立てる。
func chemindeFerCardsStr(cards []*domain.Card) string {
	parts := make([]string, 0, len(cards))
	for _, c := range cards {
		parts = append(parts, cuiCardStr(c))
	}
	return strings.Join(parts, " ")
}

// writeChips は各席のチップと賭け金を書き出す。
func (cp *ChemindeFerCuiPresenter) writeChips(sb *strings.Builder, c interfaces.ChemindeFerGame) {
	parts := make([]string, 0, domain.ChemindeFerSeatCnt)
	for i := range domain.ChemindeFerSeatCnt {
		p := c.GetPlayer(i)
		if p == nil {
			continue
		}
		mark := ""
		if i == c.GetBankerIdx() {
			mark = "*"
		}
		entry := "#" + strconv.Itoa(i) + mark + ":" + strconv.Itoa(p.GetChips())
		if p.GetBet() > 0 {
			entry += "(" + strconv.Itoa(p.GetBet()) + ")"
		}
		parts = append(parts, entry)
	}
	sb.WriteString("----------\n")
	sb.WriteString(i18n.Tf("chemindefer.chipsLine", "chips", strings.Join(parts, " ")) + "\n")
}

// writeResult は決着と終局を書き出す。
func (cp *ChemindeFerCuiPresenter) writeResult(sb *strings.Builder, c interfaces.ChemindeFerGame) {
	if r := c.GetResult(); r != domain.ChemindeFerResultNone {
		sb.WriteString(i18n.Tf("chemindefer.resultLine", "result", cp.resultStr(r)) + "\n")
	}
	if !c.GetGameEndFlag() {
		return
	}
	me := c.GetPlayer(chemindeFerHumanSeat)
	if me == nil {
		return
	}
	msg := i18n.Tf("chemindefer.gameEndLine", "chips", strconv.Itoa(me.GetChips()))
	if me.GetChips() >= c.GetConfig().InitialChips {
		sb.WriteString(color.Green(msg) + "\n")
	} else {
		sb.WriteString(color.Red(msg) + "\n")
	}
}

// ActionLogOutput 棋譜をテキスト出力
func (cp *ChemindeFerCuiPresenter) ActionLogOutput(c interfaces.ChemindeFerGame) string {
	return actionLogOutputText(c)
}

// HintOutput ヒントをテキスト出力
func (cp *ChemindeFerCuiPresenter) HintOutput(c interfaces.ChemindeFerGame) string {
	h := c.GetHint()
	if h == nil {
		return i18n.T("chemindefer.hintNone")
	}
	action := i18n.T("chemindefer.hintStand")
	if h.Draw {
		action = i18n.T("chemindefer.hintDraw")
	}
	return i18n.Tf("chemindefer.hint",
		"action", action, "reason", i18n.T("chemindefer."+h.Reason))
}

// phaseStr フェーズ文字列
func (cp *ChemindeFerCuiPresenter) phaseStr(phase domain.ChemindeFerPhase) string {
	switch phase {
	case domain.ChemindeFerPhaseStake:
		return i18n.T("chemindefer.phaseStake")
	case domain.ChemindeFerPhaseBet:
		return i18n.T("chemindefer.phaseBet")
	case domain.ChemindeFerPhasePunterDraw:
		return i18n.T("chemindefer.phasePunterDraw")
	case domain.ChemindeFerPhaseBankerDraw:
		return i18n.T("chemindefer.phaseBankerDraw")
	case domain.ChemindeFerPhaseRoundEnd:
		return i18n.T("chemindefer.phaseRoundEnd")
	default:
		return i18n.T("chemindefer.phaseUnknown")
	}
}

// resultStr 決着の文字列
func (cp *ChemindeFerCuiPresenter) resultStr(r domain.ChemindeFerResult) string {
	switch r {
	case domain.ChemindeFerResultBanker:
		return i18n.T("chemindefer.resultBanker")
	case domain.ChemindeFerResultPunter:
		return i18n.T("chemindefer.resultPunter")
	case domain.ChemindeFerResultTie:
		return i18n.T("chemindefer.resultTie")
	default:
		return i18n.T("chemindefer.resultNone")
	}
}
