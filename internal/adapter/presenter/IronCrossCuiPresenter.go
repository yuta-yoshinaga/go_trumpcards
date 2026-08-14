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

// IronCrossCuiPresenter アイアンクロスCUIプレゼンタークラス
type IronCrossCuiPresenter struct{}

// Output ゲーム状態を出力
func (cp *IronCrossCuiPresenter) Output(c interfaces.IronCrossGame, lastErr error) string {
	return buildCuiOutput(i18n.T("ironcross.outputTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("ironcross.phaseLine", "phase", cp.phaseStr(c.GetPhase())) + "\n")
		sb.WriteString(i18n.Tf("ironcross.handLine",
			"hand", strconv.Itoa(c.GetHandNumber()),
			"pot", strconv.Itoa(c.GetPot())) + "\n")
		cp.writeCross(sb, c)
		cp.writeSeats(sb, c)
		cp.writeBetLine(sb, c)
		cuiErrorBlock(sb, lastErr)
		cp.writeResult(sb, c)
	})
}

// writeCross は十字を十字の形のまま書き出す。
//
// **1 行に並べない。** 縦と横のどちらを選ぶかがこのゲームの唯一の判断なので、
// どの札が縦でどの札が横なのかが一目で分からないと選びようがない。
func (cp *IronCrossCuiPresenter) writeCross(sb *strings.Builder, c interfaces.IronCrossGame) {
	sb.WriteString("----------\n")
	cross := c.GetCross()
	at := func(i int) string {
		if i < len(cross) && cross[i] != nil {
			return cuiCardStr(cross[i])
		}
		return i18n.T("ironcross.faceDown")
	}
	pad := strings.Repeat(" ", 4)
	sb.WriteString(pad + at(domain.IronCrossTop) + "\n")
	sb.WriteString(at(domain.IronCrossLeft) + " " + at(domain.IronCrossCenter) +
		" " + at(domain.IronCrossRight) + "\n")
	sb.WriteString(pad + at(domain.IronCrossBottom) + "\n")
	sb.WriteString(i18n.Tf("ironcross.revealedLine",
		"revealed", strconv.Itoa(c.GetRevealedCount()),
		"total", strconv.Itoa(domain.IronCrossCommunityCards)) + "\n")
}

// writeSeats は席ごとの状態を書き出す。
//
// **CPU が選んだ列もショーダウンまで伏せる。** 先に見えると降りどころが割れる。
func (cp *IronCrossCuiPresenter) writeSeats(sb *strings.Builder, c interfaces.IronCrossGame) {
	showdown := c.GetPhase() == domain.IronCrossPhaseShowdown || c.GetGameEndFlag()
	for i, p := range c.GetPlayers() {
		mark := " "
		if i == c.GetTurnSeat() && c.GetPhase() == domain.IronCrossPhaseBetting {
			mark = "*"
		}
		cards := i18n.T("ironcross.hidden")
		if p.GetIsHuman() || showdown {
			cards = ironCrossCardsStr(p.GetCards())
		}
		line := ""
		if (p.GetIsHuman() || showdown) && p.GetLine() != domain.IronCrossLineNone {
			line = i18n.T("ironcross.line." + domain.IronCrossLineName(p.GetLine()))
		}
		state := ""
		if p.GetFolded() {
			state = i18n.T("ironcross.foldedMark")
		} else if p.GetAllIn() {
			state = i18n.T("ironcross.allInMark")
		}
		sb.WriteString(i18n.Tf("ironcross.seatLine",
			"mark", mark,
			"name", p.GetName(),
			"state", state,
			"chips", strconv.Itoa(p.GetChips()),
			"bet", strconv.Itoa(p.GetCurrentBet()),
			"line", line,
			"cards", cards) + "\n")
	}
}

// writeBetLine はいま人間が払うべき額、または選ぶ場面であることを書き出す。
func (cp *IronCrossCuiPresenter) writeBetLine(sb *strings.Builder, c interfaces.IronCrossGame) {
	if c.IsChoosing() {
		sb.WriteString(i18n.T("ironcross.chooseLine") + "\n")
		return
	}
	if c.GetPhase() != domain.IronCrossPhaseBetting {
		return
	}
	if toCall := c.GetToCall(); toCall > 0 {
		sb.WriteString(i18n.Tf("ironcross.toCallLine", "amount", strconv.Itoa(toCall)) + "\n")
		return
	}
	sb.WriteString(i18n.T("ironcross.canCheckLine") + "\n")
}

// writeResult は決着を書き出す。
func (cp *IronCrossCuiPresenter) writeResult(sb *strings.Builder, c interfaces.IronCrossGame) {
	if c.GetPhase() != domain.IronCrossPhaseShowdown && !c.GetGameEndFlag() {
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
		sb.WriteString(color.Green(i18n.Tf("ironcross.wonLine",
			"name", players[i].GetName(),
			"line", i18n.T("ironcross.line."+domain.IronCrossLineName(r.Line)),
			"amount", strconv.Itoa(r.WonAmount))) + "\n")
	}
	if c.GetGameEndFlag() {
		sb.WriteString(i18n.Tf("ironcross.gameEndLine",
			"name", players[c.WinnerSeat()].GetName()) + "\n")
	}
}

// ActionLogOutput 棋譜をテキスト出力
func (cp *IronCrossCuiPresenter) ActionLogOutput(c interfaces.IronCrossGame) string {
	return actionLogOutputText(c)
}

// HintOutput ヒントをテキスト出力
func (cp *IronCrossCuiPresenter) HintOutput(c interfaces.IronCrossGame) string {
	h := c.GetHint()
	if h == nil {
		return i18n.T("ironcross.hintNone")
	}
	action := i18n.T("ironcross.action." + h.Action)
	if h.Action == "line" {
		action = i18n.Tf("ironcross.action.lineNamed",
			"line", i18n.T("ironcross.line."+domain.IronCrossLineName(h.Line)))
	}
	return i18n.Tf("ironcross.hint",
		"action", action,
		"reason", i18n.T("ironcross.reason."+h.Reason))
}

// phaseStr フェーズ文字列
func (cp *IronCrossCuiPresenter) phaseStr(phase domain.IronCrossPhase) string {
	switch phase {
	case domain.IronCrossPhaseBetting:
		return i18n.T("ironcross.phaseBetting")
	case domain.IronCrossPhaseChoose:
		return i18n.T("ironcross.phaseChoose")
	case domain.IronCrossPhaseShowdown:
		return i18n.T("ironcross.phaseShowdown")
	case domain.IronCrossPhaseGameEnd:
		return i18n.T("ironcross.phaseGameEnd")
	default:
		return i18n.T("ironcross.phaseUnknown")
	}
}

// ironCrossCardsStr は札の並びを 1 行の文字列にする。
func ironCrossCardsStr(cards []*domain.Card) string {
	parts := make([]string, 0, len(cards))
	for _, c := range cards {
		parts = append(parts, cuiCardStr(c))
	}
	return strings.Join(parts, " ")
}
