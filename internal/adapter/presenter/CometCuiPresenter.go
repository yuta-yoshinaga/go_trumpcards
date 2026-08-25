//go:build !js || !wasm || solo

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

// CometCuiPresenter renders the Comet CUI view.
type CometCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *CometCuiPresenter) Output(g interfaces.CometGame, lastErr error) string {
	return buildCuiOutput(i18n.T("comet.helpTitle"), func(b *strings.Builder) {
		p.writeHeader(b, g)
		p.writeSeats(b, g)
		b.WriteString("----------\n")
		p.writePile(b, g)
		cuiErrorBlock(b, lastErr)
		if g.GetGameEndFlag() {
			p.writeGameEnd(b, g)
			return
		}
		if g.GetPhase() == domain.CometPhaseRoundEnd {
			p.writeRoundResult(b, g)
			return
		}
		p.writeTurn(b, g)
	})
}

// writeHeader は局と死に手の枚数を書く。
func (p *CometCuiPresenter) writeHeader(b *strings.Builder, g interfaces.CometGame) {
	b.WriteString(i18n.Tf("comet.round",
		"n", strconv.Itoa(g.GetRoundNumber()),
		"target", strconv.Itoa(g.GetConfig().TargetScore)) + "\n")
	// **死に手の枚数は見せる。** ここに眠った札で連なりが止まるので、何枚
	// 伏せてあるかは読みの材料になる ── 中身は見せない。
	b.WriteString(i18n.Tf("comet.dead", "n", strconv.Itoa(g.GetDeadCount())) + "\n")
}

// writeSeats は席ごとの行と人間の手札を書く。
func (p *CometCuiPresenter) writeSeats(b *strings.Builder, g interfaces.CometGame) {
	dealer := g.GetDealerIdx()
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		role := i18n.T("comet.roleNonDealer")
		if i == dealer {
			role = i18n.T("comet.roleDealer")
		}
		b.WriteString(i18n.Tf("comet.seat",
			"name", cometSeatName(player.GetIsHuman(), i),
			"role", role,
			"cards", strconv.Itoa(player.GetCardsSize()),
			"score", strconv.Itoa(player.GetScore())) + "\n")
		if player.GetIsHuman() && player.GetCardsSize() > 0 {
			b.WriteString(cometIndexedHand(player) + "\n")
		}
	}
}

// cometSeatName は席の呼び名を返す。
func cometSeatName(isHuman bool, idx int) string {
	if isHuman {
		return i18n.T("comet.you")
	}
	return i18n.Tf("comet.cpu", "n", strconv.Itoa(idx))
}

// cometIndexedHand は番号付きの手札を返す。コメットには印を付ける。
func cometIndexedHand(player *domain.CometPlayer) string {
	parts := make([]string, 0, player.GetCardsSize())
	for i := 0; i < player.GetCardsSize(); i++ {
		card := player.GetCard(i)
		s := fmt.Sprintf("[%d]%s", i, cuiCardStrEmojiRank(card))
		// **コメットは目立たせる。** どのランクの代わりにもなる特別な 1 枚が
		// ただの ♦9 に見えていると、出せる手を見落とす。
		if domain.IsCometWild(card) {
			s = color.Yellow(s + i18n.T("comet.wildMark"))
		}
		parts = append(parts, s)
	}
	return strings.Join(parts, "  ")
}

// writePile は今の連なりと次に要るランクを書く。
func (p *CometCuiPresenter) writePile(b *strings.Builder, g interfaces.CometGame) {
	pile := g.GetPile()
	// **見せるのは今の連なりだけ。** 局のはじめから全部並べると読めない。
	shown := pile
	if len(shown) > cometPileTail {
		shown = shown[len(shown)-cometPileTail:]
	}
	parts := make([]string, 0, len(shown))
	for _, c := range shown {
		parts = append(parts, cuiCardStrEmojiRank(c))
	}
	if len(parts) == 0 {
		parts = append(parts, i18n.T("comet.emptyPile"))
	}
	b.WriteString(i18n.Tf("comet.pile", "cards", strings.Join(parts, " → ")) + "\n")

	if need := g.GetNeed(); need > 0 {
		b.WriteString(i18n.Tf("comet.need", "rank", cuiRankLabel(need)) + "\n")
	} else {
		b.WriteString(i18n.T("comet.needAny") + "\n")
	}
}

// cometPileTail は連なりの表示に残す枚数。
const cometPileTail = 8

// writeTurn は手番と出せる札を書く。
func (p *CometCuiPresenter) writeTurn(b *strings.Builder, g interfaces.CometGame) {
	cur := g.GetCurrentPlayerIdx()
	player := g.GetPlayer(cur)
	if player == nil {
		return
	}
	b.WriteString(i18n.Tf("comet.turn", "name", cometSeatName(player.GetIsHuman(), cur)) + "\n")
	if !g.IsHumanTurn() {
		return
	}
	idxs := g.PlayableIdxs(cur)
	if len(idxs) == 0 {
		// **出せる札が無いならパスしかない。** 探させない。
		b.WriteString(color.Yellow(i18n.T("comet.mustPass")) + "\n")
		return
	}
	parts := make([]string, 0, len(idxs))
	for _, i := range idxs {
		parts = append(parts, fmt.Sprintf("[%d]%s", i, cuiCardStrEmojiRank(player.GetCard(i))))
	}
	b.WriteString("  " + i18n.Tf("comet.playable", "cards", strings.Join(parts, "  ")) + "\n")
	b.WriteString(i18n.T("comet.commandHint") + "\n")
}

// writeRoundResult は局の集計を書く。
func (p *CometCuiPresenter) writeRoundResult(b *strings.Builder, g interfaces.CometGame) {
	res := g.GetLastResult()
	if res == nil {
		return
	}
	winner := g.GetPlayer(res.WinnerIdx)
	b.WriteString(i18n.Tf("comet.goOut",
		"name", cometSeatName(winner != nil && winner.GetIsHuman(), res.WinnerIdx),
		"points", strconv.Itoa(res.Gained[res.WinnerIdx])) + "\n")
	b.WriteString("  " + i18n.Tf("comet.unplayedKings",
		"n", strconv.Itoa(res.UnplayedKings),
		"points", strconv.Itoa(res.UnplayedKings*domain.CometUnplayedKingPoints)) + "\n")
	if res.HeldWildIdx >= 0 {
		held := g.GetPlayer(res.HeldWildIdx)
		b.WriteString("  " + color.Red(i18n.Tf("comet.heldWild",
			"name", cometSeatName(held != nil && held.GetIsHuman(), res.HeldWildIdx),
			"points", strconv.Itoa(domain.CometHoldingWildPenalty))) + "\n")
	}
	b.WriteString(i18n.T("comet.nextRoundHint") + "\n")
}

// writeGameEnd は終局の行を書く。
func (p *CometCuiPresenter) writeGameEnd(b *strings.Builder, g interfaces.CometGame) {
	winner := g.GetWinnerIdx()
	if winner < 0 {
		return
	}
	player := g.GetPlayer(winner)
	b.WriteString(color.Green(i18n.Tf("comet.winner",
		"name", cometSeatName(player != nil && player.GetIsHuman(), winner))) + "\n")
}

// HintOutput renders the recommended move.
func (p *CometCuiPresenter) HintOutput(g interfaces.CometGame) string {
	return buildCuiOutput(i18n.T("comet.helpTitle"), func(b *strings.Builder) {
		hint := g.GetHint()
		if hint == nil || hint.HandIdx < 0 {
			b.WriteString(i18n.T("comet."+cometNoHintKey(hint)) + "\n")
			return
		}
		player := g.GetPlayer(g.GetCurrentPlayerIdx())
		card := "??"
		if player != nil && hint.HandIdx < player.GetCardsSize() {
			card = cuiCardStrEmojiRank(player.GetCard(hint.HandIdx))
		}
		b.WriteString(i18n.Tf("comet.hintCard",
			"idx", strconv.Itoa(hint.HandIdx), "card", card) + "\n")
		b.WriteString(i18n.Tf("comet.hintReason",
			"reason", i18n.T("comet.reason."+hint.Reason)) + "\n")
	})
}

// cometNoHintKey は勧める手が無いときの文言キーを選ぶ。
func cometNoHintKey(hint *domain.CometHint) string {
	if hint != nil && hint.Reason == "pass" {
		return "mustPass"
	}
	return "noHint"
}

// ActionLogOutput renders the action log.
func (p *CometCuiPresenter) ActionLogOutput(g interfaces.CometGame) string {
	return actionLogOutputTextForSeats[*domain.CometPlayer](g)
}
