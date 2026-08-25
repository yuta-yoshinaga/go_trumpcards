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

// ContinentalRummyCuiPresenter はコンチネンタル・ラミーの CUI ビュー。
type ContinentalRummyCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *ContinentalRummyCuiPresenter) Output(g interfaces.ContinentalRummyGame, lastErr error) string {
	return buildCuiOutput(i18n.T("continentalrummy.helpTitle"), func(b *strings.Builder) {
		p.writeHeader(b, g)
		p.writeSeats(b, g)
		cuiErrorBlock(b, lastErr)
		if g.GetGameEndFlag() {
			p.writeGameEnd(b, g)
			return
		}
		switch g.GetPhase() {
		case domain.ContinentalRummyPhaseDraw:
			p.writeDrawPrompt(b, g)
		case domain.ContinentalRummyPhaseDiscard:
			p.writeDiscardPrompt(b, g)
		case domain.ContinentalRummyPhaseRoundEnd:
			p.writeRoundEnd(b, g)
		}
	})
}

// writeHeader はラウンド・山札・捨て札の頭を書く。
func (p *ContinentalRummyCuiPresenter) writeHeader(b *strings.Builder, g interfaces.ContinentalRummyGame) {
	b.WriteString(i18n.Tf("continentalrummy.round",
		"n", strconv.Itoa(g.GetRoundNumber()),
		"total", strconv.Itoa(g.GetConfig().TotalRounds)) + "\n")
	b.WriteString(i18n.Tf("continentalrummy.stock", "n", strconv.Itoa(g.GetStockCount())))
	if top := g.GetDiscardTop(); top != nil {
		b.WriteString(i18n.Tf("continentalrummy.discardTop", "card", cuiCardStrEmojiRank(top)))
	}
	b.WriteString("\n")
	// **認められた形はいつでも見えていること。** 15 枚を 3〜5 枚の並びに
	// どう割るかがこのゲームの全部で、5+5+5 が入っていないのが肝。
	b.WriteString(i18n.Tf("continentalrummy.layouts", "list", continentalLayoutList()) + "\n")
}

// continentalLayoutList は認められた形を "3+3+3+3+3 / 4+4+4+3 / 5+4+3+3" と並べる。
func continentalLayoutList() string {
	rows := make([]string, 0, 3)
	for _, l := range domain.ContinentalRummyLayouts() {
		parts := make([]string, 0, len(l))
		for _, n := range l {
			parts = append(parts, strconv.Itoa(n))
		}
		rows = append(rows, strings.Join(parts, "+"))
	}
	return strings.Join(rows, " / ")
}

// writeSeats は席ごとの手札枚数と得点を書く。
func (p *ContinentalRummyCuiPresenter) writeSeats(b *strings.Builder, g interfaces.ContinentalRummyGame) {
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		name := i18n.Tf("continentalrummy.cpuName", "n", strconv.Itoa(i))
		if player.GetIsHuman() {
			name = i18n.T("continentalrummy.you")
		}
		line := i18n.Tf("continentalrummy.seat",
			"name", name,
			"cards", strconv.Itoa(player.GetCardsSize()),
			"score", strconv.Itoa(player.GetScore()))
		if i == g.GetDealerIdx() {
			line += i18n.T("continentalrummy.dealerMark")
		}
		if i == g.GetCurrentPlayerIdx() && !g.GetGameEndFlag() {
			line = color.Yellow(line + i18n.T("continentalrummy.turnMark"))
		}
		b.WriteString(line + "\n")
		p.writeMelds(b, player)
	}
	p.writeHand(b, g)
}

// writeMelds は上がって並べたシーケンスを書く。
func (p *ContinentalRummyCuiPresenter) writeMelds(b *strings.Builder, player *domain.ContinentalRummyPlayer) {
	for _, run := range player.GetMelds() {
		cards := make([]string, 0, len(run))
		for _, c := range run {
			cards = append(cards, cuiCardStrEmojiRank(c))
		}
		b.WriteString("    " + strings.Join(cards, " ") + "\n")
	}
}

// writeHand は人間の手札を番号付きで書く。
func (p *ContinentalRummyCuiPresenter) writeHand(b *strings.Builder, g interfaces.ContinentalRummyGame) {
	me := g.GetPlayer(domain.ContinentalRummyHumanIdx)
	if me == nil || me.GetCardsSize() == 0 {
		return
	}
	parts := make([]string, 0, me.GetCardsSize())
	for i := 0; i < me.GetCardsSize(); i++ {
		parts = append(parts, "["+strconv.Itoa(i)+"]"+cuiCardStrEmojiRank(me.GetCard(i)))
	}
	b.WriteString(strings.Join(parts, " ") + "\n")
}

// writeDrawPrompt は引く番を促す。
func (p *ContinentalRummyCuiPresenter) writeDrawPrompt(b *strings.Builder, g interfaces.ContinentalRummyGame) {
	if g.GetCurrentPlayerIdx() != domain.ContinentalRummyHumanIdx {
		return
	}
	b.WriteString(i18n.T("continentalrummy.drawPrompt") + "\n")
}

// writeDiscardPrompt は捨てる番を促す。上がれるならそれも言う。
func (p *ContinentalRummyCuiPresenter) writeDiscardPrompt(b *strings.Builder, g interfaces.ContinentalRummyGame) {
	if g.GetCurrentPlayerIdx() != domain.ContinentalRummyHumanIdx {
		return
	}
	// **上がれるときは黙っていない。** 15 枚の分割は目で追いきれないので、
	// 見落としたまま捨ててしまうのが一番つまらない負け方になる。
	if idx, ok := g.CanGoOut(); ok {
		b.WriteString(color.Green(i18n.Tf("continentalrummy.canGoOut", "idx", strconv.Itoa(idx))) + "\n")
	}
	b.WriteString(i18n.T("continentalrummy.discardPrompt") + "\n")
}

// writeRoundEnd はラウンドの決着を書く。
func (p *ContinentalRummyCuiPresenter) writeRoundEnd(b *strings.Builder, g interfaces.ContinentalRummyGame) {
	res := g.GetLastResult()
	if res == nil {
		return
	}
	if res.WinnerIdx < 0 {
		b.WriteString(i18n.T("continentalrummy.washout") + "\n")
		b.WriteString(i18n.T("continentalrummy.nextRoundHint") + "\n")
		return
	}
	name := i18n.Tf("continentalrummy.cpuName", "n", strconv.Itoa(res.WinnerIdx))
	if res.WinnerIdx == domain.ContinentalRummyHumanIdx {
		name = i18n.T("continentalrummy.you")
	}
	b.WriteString(i18n.Tf("continentalrummy.wentOut", "name", name) + "\n")
	// **加点は内訳で見せる。** 合計だけだと、どう上がると得なのかが伝わらない。
	for _, bonus := range res.Bonuses {
		b.WriteString("  " + i18n.Tf("continentalrummy.bonusLine",
			"label", i18n.T("continentalrummy.bonus."+bonus.Key),
			"points", strconv.Itoa(bonus.Points)) + "\n")
	}
	b.WriteString(i18n.Tf("continentalrummy.collected",
		"per", strconv.Itoa(res.PerOpponent),
		"total", strconv.Itoa(res.Total)) + "\n")
	b.WriteString(i18n.T("continentalrummy.nextRoundHint") + "\n")
}

// writeGameEnd は終局を書く。
func (p *ContinentalRummyCuiPresenter) writeGameEnd(b *strings.Builder, g interfaces.ContinentalRummyGame) {
	switch g.GetWinnerIdx() {
	case domain.ContinentalRummyHumanIdx:
		b.WriteString(color.Green(i18n.T("continentalrummy.humanWin")) + "\n")
	case -1:
		b.WriteString(i18n.T("continentalrummy.draw") + "\n")
	default:
		b.WriteString(i18n.Tf("continentalrummy.cpuWin",
			"n", strconv.Itoa(g.GetWinnerIdx())) + "\n")
	}
}

// HintOutput renders the recommended move.
func (p *ContinentalRummyCuiPresenter) HintOutput(g interfaces.ContinentalRummyGame) string {
	return buildCuiOutput(i18n.T("continentalrummy.helpTitle"), func(b *strings.Builder) {
		hint := g.GetHint()
		if hint == nil {
			b.WriteString(i18n.T("continentalrummy.noHint") + "\n")
			return
		}
		b.WriteString(i18n.T("continentalrummy.reason."+hint.Reason) + "\n")
		if hint.DiscardIdx >= 0 {
			key := "continentalrummy.hintDiscard"
			if hint.GoOut {
				key = "continentalrummy.hintGoOut"
			}
			b.WriteString(i18n.Tf(key, "idx", strconv.Itoa(hint.DiscardIdx)) + "\n")
		}
	})
}

// ActionLogOutput renders the action log.
func (p *ContinentalRummyCuiPresenter) ActionLogOutput(g interfaces.ContinentalRummyGame) string {
	return actionLogOutputTextForSeats[*domain.ContinentalRummyPlayer](g)
}
