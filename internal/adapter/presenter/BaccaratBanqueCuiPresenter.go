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

// BaccaratBanqueCuiPresenter renders the Baccarat Banque CUI view.
type BaccaratBanqueCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *BaccaratBanqueCuiPresenter) Output(g interfaces.BaccaratBanqueGame, lastErr error) string {
	return buildCuiOutput(i18n.T("baccaratbanque.helpTitle"), func(b *strings.Builder) {
		p.writeHeader(b, g)
		p.writeSeats(b, g)
		cuiErrorBlock(b, lastErr)
		if g.GetGameEndFlag() {
			p.writeGameEnd(b, g)
			return
		}
		switch g.GetPhase() {
		case domain.BaccaratBanquePhaseBanker:
			p.writeBankerPrompt(b, g)
		case domain.BaccaratBanquePhaseResult:
			p.writeResult(b, g)
		}
	})
}

// writeHeader はクー数・バンクの継続数・シューの残りを書く。
func (p *BaccaratBanqueCuiPresenter) writeHeader(b *strings.Builder, g interfaces.BaccaratBanqueGame) {
	b.WriteString(i18n.Tf("baccaratbanque.coup", "n", strconv.Itoa(g.GetCoupNumber())) + "\n")
	// **バンクの継続数は見せる。** 1 回負けても途切れないのがこの形式の要。
	// ただし終わったあとに「負けても続きます」と出すと嘘になるので、終局後は書かない。
	if !g.GetGameEndFlag() {
		b.WriteString(i18n.Tf("baccaratbanque.bankHeld", "n", strconv.Itoa(g.GetBankHeld())) + "\n")
	}
	b.WriteString(i18n.Tf("baccaratbanque.shoe", "n", strconv.Itoa(g.GetShoeRemaining())) + "\n")
}

// writeSeats は 3 席の札と合計を書く。**バカラは全部表向き。**
func (p *BaccaratBanqueCuiPresenter) writeSeats(b *strings.Builder, g interfaces.BaccaratBanqueGame) {
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		cards := make([]string, 0, player.GetCardsSize())
		for j := 0; j < player.GetCardsSize(); j++ {
			cards = append(cards, cuiCardStrEmojiRank(player.GetCard(j)))
		}
		line := i18n.Tf("baccaratbanque.seat",
			"name", i18n.T("baccaratbanque.role."+baccaratBanqueRole(i)),
			"cards", strings.Join(cards, " "),
			"total", strconv.Itoa(player.GetTotal()),
			"chips", strconv.Itoa(player.GetChips()))
		if i != domain.BaccaratBanqueBankerIdx {
			line += i18n.Tf("baccaratbanque.betSuffix", "bet", strconv.Itoa(player.GetBet()))
		}
		if domain.BaccaratBanqueIsNatural(player.GetHand()) {
			line = color.Yellow(line + i18n.T("baccaratbanque.naturalMark"))
		}
		b.WriteString(line + "\n")
	}
}

// writeBankerPrompt は親の判断を促す。
func (p *BaccaratBanqueCuiPresenter) writeBankerPrompt(b *strings.Builder, g interfaces.BaccaratBanqueGame) {
	banker := g.GetPlayer(domain.BaccaratBanqueBankerIdx)
	if banker == nil {
		return
	}
	// **親はどの合計でも自由。** 固定表が無いことを毎回書く。
	b.WriteString(i18n.Tf("baccaratbanque.bankerPrompt",
		"total", strconv.Itoa(banker.GetTotal())) + "\n")
	b.WriteString(i18n.T("baccaratbanque.commandHint") + "\n")
}

// writeResult はクーの決着を書く。
func (p *BaccaratBanqueCuiPresenter) writeResult(b *strings.Builder, g interfaces.BaccaratBanqueGame) {
	res := g.GetLastResult()
	if res == nil {
		return
	}
	b.WriteString(i18n.Tf("baccaratbanque.resultTitle",
		"total", strconv.Itoa(res.BankerTotal)) + "\n")
	// **左右は別勘定。** どちらがどうなったかを 1 行ずつ書く。
	for _, s := range res.Sides {
		b.WriteString("  " + i18n.Tf("baccaratbanque.sideLine",
			"name", i18n.T("baccaratbanque.role."+baccaratBanqueRole(s.SeatIdx)),
			"outcome", i18n.T("baccaratbanque.outcome."+s.Outcome),
			"delta", strconv.Itoa(s.Delta)) + "\n")
	}
	b.WriteString("  " + i18n.Tf("baccaratbanque.bankerDelta",
		"delta", strconv.Itoa(res.BankerDelta)) + "\n")
	b.WriteString(i18n.T("baccaratbanque.nextCoupHint") + "\n")
}

// writeGameEnd はバンクが終わった理由を書く。
func (p *BaccaratBanqueCuiPresenter) writeGameEnd(b *strings.Builder, g interfaces.BaccaratBanqueGame) {
	key := "baccaratbanque.endBroke"
	switch {
	case g.IsRetired():
		key = "baccaratbanque.endRetired"
	case g.GetWinnerIdx() == domain.BaccaratBanqueBankerIdx:
		key = "baccaratbanque.endAhead"
	}
	line := i18n.Tf(key, "n", strconv.Itoa(g.GetBankHeld()))
	if g.GetWinnerIdx() == domain.BaccaratBanqueBankerIdx {
		line = color.Green(line)
	}
	b.WriteString(line + "\n")
}

// HintOutput renders the recommended move.
func (p *BaccaratBanqueCuiPresenter) HintOutput(g interfaces.BaccaratBanqueGame) string {
	return buildCuiOutput(i18n.T("baccaratbanque.helpTitle"), func(b *strings.Builder) {
		hint := g.GetHint()
		if hint == nil || hint.Reason == "none" {
			b.WriteString(i18n.T("baccaratbanque.noHint") + "\n")
			return
		}
		action := i18n.T("baccaratbanque.actionStand")
		if hint.Draw {
			action = i18n.T("baccaratbanque.actionDraw")
		}
		b.WriteString(i18n.Tf("baccaratbanque.hintAction", "action", action) + "\n")
		b.WriteString(i18n.Tf("baccaratbanque.hintReason",
			"reason", i18n.T("baccaratbanque.reason."+hint.Reason)) + "\n")
	})
}

// ActionLogOutput renders the action log.
//
// **席は盤面と同じ呼び方にする (#5977)。** 既定の名付けは "CPU 1" を返すが、
// 画面ではその席を「右のタブロー」と呼んでいる。棋譜だけ別名になると、どの
// 席が引いたのか読み手が突き合わせられない。
func (p *BaccaratBanqueCuiPresenter) ActionLogOutput(g interfaces.BaccaratBanqueGame) string {
	return actionLogOutputTextNamedBy(g, func(idx int) string {
		return i18n.T("baccaratbanque.role." + baccaratBanqueRole(idx))
	})
}
