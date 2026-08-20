//go:build !js || !wasm || solo

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// TuSacCuiPresenter 四色牌CUIプレゼンタークラス
type TuSacCuiPresenter struct{}

// tuSacMeldPointsTableStr は組み合わせの点数を "同色同種3枚 2点 / ..." で返す。
func tuSacMeldPointsTableStr() string {
	parts := make([]string, 0, int(domain.TuSacMeldKindMax))
	for k := domain.TuSacMeldNone + 1; k <= domain.TuSacMeldKindMax; k++ {
		parts = append(parts, i18n.T("tusac.meld."+domain.TuSacMeldKindName(k))+" "+
			i18n.Tf("tusac.meld.points", "points", strconv.Itoa(domain.TuSacMeldPoints(k))))
	}
	return strings.Join(parts, " / ")
}

// Output ゲーム状態を出力
func (cp *TuSacCuiPresenter) Output(c interfaces.TuSacGame, lastErr error) string {
	return buildCuiOutput(i18n.T("tusac.outputTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("tusac.roundLine",
			"round", strconv.Itoa(c.GetRoundNumber()),
			"total", strconv.Itoa(c.GetConfig().Rounds)) + "\n")
		sb.WriteString(i18n.Tf("tusac.stockLine",
			"stock", strconv.Itoa(c.GetStockCount()),
			"discard", tuSacCardStr(c.GetDiscardTop())) + "\n")
		// **5 枚の卒を揃える価値は、狙う前に知りたい** (#5784)。点数は
		// ドメインの TuSacMeldPoints から作るので、写した表にならない。
		sb.WriteString(i18n.Tf("tusac.meldPointsLine", "table", tuSacMeldPointsTableStr()) + "\n")
		cp.writeSeats(sb, c)
		cp.writeHand(sb, c)
		cp.writePrompt(sb, c)
		cuiErrorBlock(sb, lastErr)
		cp.writeResult(sb, c)
	})
}

// writeSeats は席ごとの状態を書き出す。
//
// **相手の手札は最後まで見えない。** 枚数だけを出す ── 場に出た組み合わせは
// 全員ぶん見えるので、そこから読むのがこのゲームの筋。
func (cp *TuSacCuiPresenter) writeSeats(sb *strings.Builder, c interfaces.TuSacGame) {
	sb.WriteString("----------\n")
	for i, p := range c.GetPlayers() {
		mark := " "
		if i == c.GetTurnSeat() {
			mark = "*"
		}
		melds := make([]string, 0, len(p.GetMelds()))
		for _, m := range p.GetMelds() {
			melds = append(melds, i18n.T("tusac.meld."+domain.TuSacMeldKindName(m.Kind))+
				"("+tuSacCardsStr(m.Cards)+")")
		}
		meldStr := "-"
		if len(melds) > 0 {
			meldStr = strings.Join(melds, " ")
		}
		sb.WriteString(i18n.Tf("tusac.seatLine",
			"mark", mark,
			"name", p.GetName(),
			"count", strconv.Itoa(len(p.GetCards())),
			"score", strconv.Itoa(p.GetScore()),
			"melds", meldStr) + "\n")
	}
}

// writeHand は人間の手札を番号つきで書き出す。
//
// **番号を振る。** 同じ色・同じ駒が 4 枚あるので、札の名前では指定できない。
func (cp *TuSacCuiPresenter) writeHand(sb *strings.Builder, c interfaces.TuSacGame) {
	p := c.GetPlayers()[c.HumanSeat()]
	if len(p.GetCards()) == 0 {
		return
	}
	sb.WriteString("----------\n")
	parts := make([]string, 0, len(p.GetCards()))
	for i, card := range p.GetCards() {
		parts = append(parts, strconv.Itoa(i+1)+":"+tuSacCardStr(card))
	}
	sb.WriteString(i18n.Tf("tusac.handLine", "cards", strings.Join(parts, " ")) + "\n")
}

// writePrompt はいま人間に求められている操作を書き出す。
func (cp *TuSacCuiPresenter) writePrompt(sb *strings.Builder, c interfaces.TuSacGame) {
	if c.GetGameEndFlag() || c.GetPhase() == domain.TuSacPhaseRoundEnd || !c.IsHumanTurn() {
		return
	}
	if c.GetPhase() == domain.TuSacPhaseDraw {
		sb.WriteString(i18n.T("tusac.drawPrompt") + "\n")
		return
	}
	sb.WriteString(i18n.T("tusac.discardPrompt") + "\n")
}

// writeResult は決着を書き出す。
func (cp *TuSacCuiPresenter) writeResult(sb *strings.Builder, c interfaces.TuSacGame) {
	if c.GetPhase() == domain.TuSacPhaseDraw || c.GetPhase() == domain.TuSacPhaseDiscard {
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
		line := i18n.Tf("tusac.resultLine",
			"name", players[i].GetName(),
			"meld", strconv.Itoa(r.MeldPoints),
			"penalty", strconv.Itoa(r.HandPenalty),
			"score", strconv.Itoa(r.RoundScore))
		if r.WentOut {
			line = color.Green(line + i18n.T("tusac.wentOutMark"))
		}
		sb.WriteString(line + "\n")
	}
	if c.GetGameEndFlag() {
		sb.WriteString(i18n.Tf("tusac.gameEndLine",
			"name", players[c.WinnerSeat()].GetName()) + "\n")
	}
}

// ActionLogOutput 棋譜をテキスト出力
func (cp *TuSacCuiPresenter) ActionLogOutput(c interfaces.TuSacGame) string {
	return actionLogOutputText(c)
}

// HintOutput ヒントをテキスト出力
func (cp *TuSacCuiPresenter) HintOutput(c interfaces.TuSacGame) string {
	h := c.GetHint()
	if h == nil {
		return i18n.T("tusac.hintNone")
	}
	// **薦める札は 1 始まりで見せる。** 画面の番号と揃えないと押し間違える。
	shown := make([]string, 0, len(h.Indexes))
	for _, i := range h.Indexes {
		shown = append(shown, strconv.Itoa(i+1))
	}
	return i18n.Tf("tusac.hint",
		"action", i18n.T("tusac.action."+h.Action),
		"cards", strings.Join(shown, " "),
		"reason", i18n.T("tusac.reason."+h.Reason))
}

// tuSacCardStr は 1 枚を「色+駒」で書き出す。
func tuSacCardStr(c *domain.Card) string {
	if c == nil {
		return "-"
	}
	return i18n.T("tusac.color."+domain.TuSacColorName(c.GetDesign())) +
		i18n.T("tusac.piece."+domain.TuSacPieceName(c.GetValue()))
}

// tuSacCardsStr は札の並びを 1 行にする。
func tuSacCardsStr(cards []*domain.Card) string {
	parts := make([]string, 0, len(cards))
	for _, c := range cards {
		parts = append(parts, tuSacCardStr(c))
	}
	return strings.Join(parts, " ")
}
