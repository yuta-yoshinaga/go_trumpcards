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

// baseballPokerWildValuesStr はワイルドのランクを "3 / 9" で返す。
//
// **説明文もドメインの値から作る** (#5782)。判定は BaseballIsWild が持って
// いるので、文言だけが古い定数を写していると片方だけ嘘になる。
func baseballPokerWildValuesStr() string {
	parts := make([]string, 0, 2)
	for v := 1; v <= 13; v++ {
		if domain.BaseballIsWild(domain.NewCard(domain.CardDesignSpade, v, false)) {
			parts = append(parts, strconv.Itoa(v))
		}
	}
	return strings.Join(parts, " / ")
}

// BaseballPokerCuiPresenter ベースボールポーカーCUIプレゼンタークラス
type BaseballPokerCuiPresenter struct{}

// Output ゲーム状態を出力
func (cp *BaseballPokerCuiPresenter) Output(c interfaces.BaseballPokerGame, lastErr error) string {
	return buildCuiOutput(i18n.T("baseballpoker.outputTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("baseballpoker.phaseLine", "phase", cp.phaseStr(c.GetPhase())) + "\n")
		sb.WriteString(i18n.Tf("baseballpoker.handLine",
			"hand", strconv.Itoa(c.GetHandNumber()),
			"pot", strconv.Itoa(c.GetPot())) + "\n")
		sb.WriteString(i18n.Tf("baseballpoker.streetLine",
			"street", strconv.Itoa(c.GetStreet()),
			"total", strconv.Itoa(domain.BaseballUpCards)) + "\n")
		// **説明文だけ数値を固定しない** (#5782)。規則が変わったとき、
		// 判定は動くのに文言だけ嘘になる。
		sb.WriteString(i18n.Tf("baseballpoker.wildLine",
			"wilds", baseballPokerWildValuesStr(),
			"bonus", strconv.Itoa(domain.BaseballBonusFour),
			"buyIn", strconv.Itoa(domain.BaseballWildThree)) + "\n")
		cp.writeSeats(sb, c)
		cp.writeBetLine(sb, c)
		cuiErrorBlock(sb, lastErr)
		cp.writeResult(sb, c)
	})
}

// writeSeats は席ごとの状態を書き出す。
//
// **表札は全員に見える。それがスタッドの勝負の材料。** 伏せ札だけを隠す ──
// 「CPU の手札を全部隠す」と読み合いの根拠が消え、ただの運になる。
func (cp *BaseballPokerCuiPresenter) writeSeats(sb *strings.Builder, c interfaces.BaseballPokerGame) {
	sb.WriteString("----------\n")
	showdown := c.GetPhase() == domain.BaseballPhaseShowdown || c.GetGameEndFlag()
	for i, p := range c.GetPlayers() {
		mark := " "
		if i == c.GetTurnSeat() && c.GetPhase() == domain.BaseballPhaseBetting {
			mark = "*"
		}
		if i == c.GetBuyerSeat() {
			mark = "$"
		}
		state := ""
		if p.GetFolded() {
			state = i18n.T("baseballpoker.foldedMark")
		} else if p.GetAllIn() {
			state = i18n.T("baseballpoker.allInMark")
		}
		bonus := ""
		if p.GetBonusCards() > 0 {
			bonus = i18n.Tf("baseballpoker.bonusMark", "count", strconv.Itoa(p.GetBonusCards()))
		}
		sb.WriteString(i18n.Tf("baseballpoker.seatLine",
			"mark", mark,
			"name", p.GetName(),
			"state", state,
			"chips", strconv.Itoa(p.GetChips()),
			"bet", strconv.Itoa(p.GetCurrentBet()),
			"bonus", bonus,
			"cards", cp.handStr(p, p.GetIsHuman() || showdown)) + "\n")
	}
}

// handStr は手札を書き出す。showAll が偽なら伏せ札を伏せたまま出す。
func (cp *BaseballPokerCuiPresenter) handStr(p *domain.BaseballPokerPlayer, showAll bool) string {
	cards, faceUp := p.GetCards(), p.GetFaceUp()
	parts := make([]string, 0, len(cards))
	for i, card := range cards {
		visible := showAll || (i < len(faceUp) && faceUp[i])
		if visible {
			parts = append(parts, cuiCardStr(card))
			continue
		}
		parts = append(parts, i18n.T("baseballpoker.hidden"))
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, " ")
}

// writeBetLine はいま人間が求められている判断を書き出す。
func (cp *BaseballPokerCuiPresenter) writeBetLine(sb *strings.Builder, c interfaces.BaseballPokerGame) {
	if c.IsHumanBuying() {
		sb.WriteString(i18n.Tf("baseballpoker.buyInLine",
			"amount", strconv.Itoa(c.GetBuyCost())) + "\n")
		return
	}
	if c.GetPhase() != domain.BaseballPhaseBetting {
		return
	}
	if toCall := c.GetToCall(); toCall > 0 {
		sb.WriteString(i18n.Tf("baseballpoker.toCallLine", "amount", strconv.Itoa(toCall)) + "\n")
		return
	}
	sb.WriteString(i18n.T("baseballpoker.canCheckLine") + "\n")
}

// writeResult は決着を書き出す。
func (cp *BaseballPokerCuiPresenter) writeResult(sb *strings.Builder, c interfaces.BaseballPokerGame) {
	if c.GetPhase() != domain.BaseballPhaseShowdown && !c.GetGameEndFlag() {
		return
	}
	results := c.GetResults()
	players := c.GetPlayers()
	if len(results) == 0 {
		return
	}
	sb.WriteString("----------\n")
	for i, r := range results {
		if i >= len(players) {
			break
		}
		if r.WonAmount <= 0 {
			continue
		}
		wild := ""
		if r.UsedWild {
			wild = i18n.T("baseballpoker.wildMark")
		}
		sb.WriteString(color.Green(i18n.Tf("baseballpoker.wonLine",
			"name", players[i].GetName(),
			"wild", wild,
			"amount", strconv.Itoa(r.WonAmount))) + "\n")
	}
	if c.GetGameEndFlag() {
		sb.WriteString(i18n.Tf("baseballpoker.gameEndLine",
			"name", players[c.WinnerSeat()].GetName()) + "\n")
	}
}

// ActionLogOutput 棋譜をテキスト出力
func (cp *BaseballPokerCuiPresenter) ActionLogOutput(c interfaces.BaseballPokerGame) string {
	return actionLogOutputText(c)
}

// HintOutput ヒントをテキスト出力
func (cp *BaseballPokerCuiPresenter) HintOutput(c interfaces.BaseballPokerGame) string {
	h := c.GetHint()
	if h == nil {
		return i18n.T("baseballpoker.hintNone")
	}
	return i18n.Tf("baseballpoker.hint",
		"action", i18n.T("baseballpoker.action."+h.Action),
		"reason", i18n.T("baseballpoker.reason."+h.Reason))
}

// phaseStr フェーズ文字列
func (cp *BaseballPokerCuiPresenter) phaseStr(phase domain.BaseballPhase) string {
	switch phase {
	case domain.BaseballPhaseBetting:
		return i18n.T("baseballpoker.phaseBetting")
	case domain.BaseballPhaseBuyIn:
		return i18n.T("baseballpoker.phaseBuyIn")
	case domain.BaseballPhaseShowdown:
		return i18n.T("baseballpoker.phaseShowdown")
	case domain.BaseballPhaseGameEnd:
		return i18n.T("baseballpoker.phaseGameEnd")
	default:
		return i18n.T("baseballpoker.phaseUnknown")
	}
}
