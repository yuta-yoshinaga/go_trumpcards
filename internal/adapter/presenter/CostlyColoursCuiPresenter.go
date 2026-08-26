//go:build !js || !wasm || extra

package presenter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// CostlyColoursCuiPresenter renders the Costly Colours CUI view.
type CostlyColoursCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *CostlyColoursCuiPresenter) Output(g interfaces.CostlyColoursGame, lastErr error) string {
	return buildCuiOutput(i18n.T("costlycolours.helpTitle"), func(b *strings.Builder) {
		p.writeHeader(b, g)
		p.writeSeats(b, g)
		b.WriteString("----------\n")
		p.writeCount(b, g)
		cuiErrorBlock(b, lastErr)
		if g.GetGameEndFlag() {
			p.writeGameEnd(b, g)
			return
		}
		switch g.GetPhase() {
		case domain.CostlyColoursPhaseMog:
			p.writeMogPrompt(b)
		case domain.CostlyColoursPhaseShow:
			p.writeShow(b, g)
		default:
			p.writeTurn(b, g)
		}
	})
}

// writeHeader はディールと表に返した 1 枚を書く。
func (p *CostlyColoursCuiPresenter) writeHeader(b *strings.Builder, g interfaces.CostlyColoursGame) {
	b.WriteString(i18n.Tf("costlycolours.deal",
		"n", strconv.Itoa(g.GetDealNumber()),
		"target", strconv.Itoa(g.GetConfig().TargetScore)) + "\n")
	// **表の 1 枚は常に見せる。** これがトランプで、ショーの色役も
	// J / 2 の 4 点もこれ次第。
	turn := i18n.T("costlycolours.noTurnUp")
	if t := g.GetTurnUp(); t != nil {
		turn = cuiCardStrEmojiRank(t)
	}
	b.WriteString(i18n.Tf("costlycolours.turnUp", "card", turn) + "\n")
}

// writeSeats は席ごとの行と人間の手札を書く。
func (p *CostlyColoursCuiPresenter) writeSeats(b *strings.Builder, g interfaces.CostlyColoursGame) {
	dealer := g.GetDealerIdx()
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		role := i18n.T("costlycolours.roleElder")
		if i == dealer {
			role = i18n.T("costlycolours.roleDealer")
		}
		b.WriteString(i18n.Tf("costlycolours.seat",
			"name", costlySeatName(player.GetIsHuman(), i),
			"role", role,
			"cards", strconv.Itoa(player.GetCardsSize()),
			"score", strconv.Itoa(player.GetScore())) + "\n")
		if player.GetIsHuman() && player.GetCardsSize() > 0 {
			b.WriteString(costlyIndexedHand(player) + "\n")
		}
	}
}

// costlySeatName は席の呼び名を返す。
func costlySeatName(isHuman bool, idx int) string {
	if isHuman {
		return i18n.T("costlycolours.you")
	}
	return i18n.Tf("costlycolours.cpu", "n", strconv.Itoa(idx))
}

// costlyIndexedHand は番号付きの手札を返す。J と 2 には印を付ける。
func costlyIndexedHand(player *domain.CostlyColoursPlayer) string {
	parts := make([]string, 0, player.GetCardsSize())
	for i := 0; i < player.GetCardsSize(); i++ {
		card := player.GetCard(i)
		s := fmt.Sprintf("[%d]%s", i, cuiCardStrEmojiRank(card))
		// **J と 2 は持っているだけで点になる。** 印が無いと、数え上げで
		// 気軽に切ってしまう。
		if domain.CostlyIsJackOrDeuce(card) {
			s = color.Yellow(s + i18n.T("costlycolours.pegMark"))
		}
		parts = append(parts, s)
	}
	return strings.Join(parts, "  ")
}

// writeCount は今の数え上げを書く。
func (p *CostlyColoursCuiPresenter) writeCount(b *strings.Builder, g interfaces.CostlyColoursGame) {
	pile := g.GetPile()
	parts := make([]string, 0, len(pile))
	for _, c := range pile {
		parts = append(parts, cuiCardStrEmojiRank(c))
	}
	if len(parts) == 0 {
		parts = append(parts, i18n.T("costlycolours.emptyPile"))
	}
	b.WriteString(i18n.Tf("costlycolours.count",
		"cards", strings.Join(parts, " "),
		"total", strconv.Itoa(g.GetTotal())) + "\n")
}

// writeMogPrompt は交換の案内を書く。
func (p *CostlyColoursCuiPresenter) writeMogPrompt(b *strings.Builder) {
	b.WriteString(i18n.T("costlycolours.mogPrompt") + "\n")
	b.WriteString(i18n.T("costlycolours.mogHint") + "\n")
}

// writeTurn は手番と出せる札を書く。
func (p *CostlyColoursCuiPresenter) writeTurn(b *strings.Builder, g interfaces.CostlyColoursGame) {
	cur := g.GetCurrentPlayerIdx()
	player := g.GetPlayer(cur)
	if player == nil {
		return
	}
	b.WriteString(i18n.Tf("costlycolours.turn", "name", costlySeatName(player.GetIsHuman(), cur)) + "\n")
	if !g.IsHumanTurn() {
		return
	}
	idxs := g.PlayableIdxs(cur)
	if len(idxs) == 0 {
		// **出せる札が無いなら「ゴー」。** 探させない。
		b.WriteString(color.Yellow(i18n.T("costlycolours.mustGo")) + "\n")
		return
	}
	parts := make([]string, 0, len(idxs))
	for _, i := range idxs {
		parts = append(parts, fmt.Sprintf("[%d]%s", i, cuiCardStrEmojiRank(player.GetCard(i))))
	}
	b.WriteString("  " + i18n.Tf("costlycolours.playable", "cards", strings.Join(parts, "  ")) + "\n")
	b.WriteString(i18n.T("costlycolours.commandHint") + "\n")
}

// writeShow はディールの集計を書く。
func (p *CostlyColoursCuiPresenter) writeShow(b *strings.Builder, g interfaces.CostlyColoursGame) {
	res := g.GetLastResult()
	if res == nil {
		return
	}
	b.WriteString(i18n.T("costlycolours.showTitle") + "\n")
	for _, l := range res.Lines {
		points := make([]string, 0, len(l.Points))
		for _, pt := range l.Points {
			points = append(points, strconv.Itoa(pt))
		}
		b.WriteString("  " + i18n.Tf("costlycolours.showLine",
			"key", i18n.T("costlycolours.score."+l.Key),
			"points", strings.Join(points, " - ")) + "\n")
	}
	// **どの色役が付いたのかを名指す。** 点だけだと梯子のどこに乗ったか
	// 分からない。
	for i, combo := range res.Combos {
		if combo == domain.CostlyComboNone {
			continue
		}
		seat := g.GetPlayer(i)
		b.WriteString("  " + color.Yellow(i18n.Tf("costlycolours.comboLine",
			"name", costlySeatName(seat != nil && seat.GetIsHuman(), i),
			"combo", i18n.T("costlycolours.combo."+combo))) + "\n")
	}
	b.WriteString("  " + i18n.Tf("costlycolours.showTotal",
		"a", strconv.Itoa(res.Totals[0]), "b", strconv.Itoa(res.Totals[1])) + "\n")
	b.WriteString(i18n.T("costlycolours.nextDealHint") + "\n")
}

// writeGameEnd は終局の行を書く。
func (p *CostlyColoursCuiPresenter) writeGameEnd(b *strings.Builder, g interfaces.CostlyColoursGame) {
	winner := g.GetWinnerIdx()
	if winner < 0 {
		return
	}
	player := g.GetPlayer(winner)
	b.WriteString(color.Green(i18n.Tf("costlycolours.winner",
		"name", costlySeatName(player != nil && player.GetIsHuman(), winner))) + "\n")
}

// HintOutput renders the recommended move.
func (p *CostlyColoursCuiPresenter) HintOutput(g interfaces.CostlyColoursGame) string {
	return buildCuiOutput(i18n.T("costlycolours.helpTitle"), func(b *strings.Builder) {
		hint := g.GetHint()
		if hint == nil || hint.Reason == "none" {
			b.WriteString(i18n.T("costlycolours.noHint") + "\n")
			return
		}
		if hint.HandIdx < 0 {
			// 交換の可否、または「ゴー」。
			b.WriteString(i18n.Tf("costlycolours.hintReason",
				"reason", i18n.T("costlycolours.reason."+hint.Reason)) + "\n")
			return
		}
		player := g.GetPlayer(g.GetCurrentPlayerIdx())
		card := "??"
		if player != nil && hint.HandIdx < player.GetCardsSize() {
			card = cuiCardStrEmojiRank(player.GetCard(hint.HandIdx))
		}
		b.WriteString(i18n.Tf("costlycolours.hintCard",
			"idx", strconv.Itoa(hint.HandIdx), "card", card) + "\n")
		b.WriteString(i18n.Tf("costlycolours.hintReason",
			"reason", i18n.T("costlycolours.reason."+hint.Reason)) + "\n")
	})
}

// ActionLogOutput renders the action log.
func (p *CostlyColoursCuiPresenter) ActionLogOutput(g interfaces.CostlyColoursGame) string {
	return actionLogOutputTextForSeats[*domain.CostlyColoursPlayer](g)
}
