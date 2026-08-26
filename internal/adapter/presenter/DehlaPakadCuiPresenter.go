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

// DehlaPakadCuiPresenter renders the Dehla Pakad CUI view.
type DehlaPakadCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *DehlaPakadCuiPresenter) Output(g interfaces.DehlaPakadGame, lastErr error) string {
	return buildCuiOutput(i18n.T("dehlapakad.helpTitle"), func(b *strings.Builder) {
		p.writeHeader(b, g)
		p.writeSeats(b, g)
		b.WriteString("----------\n")
		cuiTrickBlock(b, g.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStrEmojiRank(tc.Card) },
			func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) },
		)
		p.writeCentrePile(b, g)
		cuiErrorBlock(b, lastErr)
		if g.GetGameEndFlag() {
			p.writeGameEnd(b, g)
			return
		}
		p.writePrompt(b, g)
	})
}

// writeHeader はハンド番号・切り札・コート数を書く。
func (p *DehlaPakadCuiPresenter) writeHeader(b *strings.Builder, g interfaces.DehlaPakadGame) {
	b.WriteString(i18n.Tf("dehlapakad.hand",
		"n", strconv.Itoa(g.GetHandNumber()),
		"target", strconv.Itoa(g.GetConfig().TargetKots)) + "\n")
	if g.GetTrumpSuit() > 0 {
		b.WriteString(i18n.Tf("dehlapakad.trump",
			"suit", i18n.T("dehlapakad.suit."+domain.DehlaPakadSuitName(g.GetTrumpSuit()))) + "\n")
		b.WriteString(i18n.Tf("dehlapakad.trick",
			"n", strconv.Itoa(g.GetTrickNumber()),
			"total", strconv.Itoa(domain.DehlaPakadTrickCount)) + "\n")
	}
	tens := g.GetTeamTens()
	kots := g.GetTeamKots()
	if len(tens) == domain.DehlaPakadTeamCnt && len(kots) == domain.DehlaPakadTeamCnt {
		human := 0
		for i := 0; i < g.GetPlayerCnt(); i++ {
			if pl := g.GetPlayer(i); pl != nil && pl.GetIsHuman() {
				human = domain.DehlaPakadTeamOf(i)
				break
			}
		}
		b.WriteString(i18n.Tf("dehlapakad.teamLine",
			"a", strconv.Itoa(tens[human]), "b", strconv.Itoa(tens[1-human]),
			"ka", strconv.Itoa(kots[human]), "kb", strconv.Itoa(kots[1-human])) + "\n")
	}
	if g.GetStreakCount() > 1 && g.GetStreakTeam() >= 0 {
		b.WriteString(i18n.Tf("dehlapakad.streak",
			"team", strconv.Itoa(g.GetStreakTeam()),
			"n", strconv.Itoa(g.GetStreakCount()),
			"need", strconv.Itoa(domain.DehlaPakadStreakForKot)) + "\n")
	}
}

// writeSeats は席ごとの行と人間の手札を書く。
func (p *DehlaPakadCuiPresenter) writeSeats(b *strings.Builder, g interfaces.DehlaPakadGame) {
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		role := i18n.T("dehlapakad.rolePlayer")
		if i == g.GetDealerIdx() {
			role = i18n.T("dehlapakad.roleDealer")
		}
		b.WriteString(i18n.Tf("dehlapakad.playerLine",
			"name", cuiPlayerName(player, i),
			"team", strconv.Itoa(domain.DehlaPakadTeamOf(i)),
			"role", role,
			"cards", strconv.Itoa(player.GetCardsSize())) + "\n")
		if player.GetIsHuman() && player.GetCardsSize() > 0 {
			b.WriteString(dehlaPakadIndexedHand(player) + "\n")
		}
	}
}

// dehlaPakadIndexedHand は人間の手札を番号付きで返す。
func dehlaPakadIndexedHand(p *domain.DehlaPakadPlayer) string {
	parts := make([]string, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		parts[i] = fmt.Sprintf("[%d]%s", i, cuiCardStrEmojiRank(p.GetCard(i)))
	}
	return strings.Join(parts, "  ")
}

// writeCentrePile は中央に積み上がった札を書く。
//
// **これがこのゲームの心臓部。** 取っただけでは札は手に入らず、同じ席が
// 2 トリック続けて取ってはじめて山ごと引き取れる ── 山に 10 が何枚乗って
// いるかが見えないと、いま何を賭けているのか読めない。
func (p *DehlaPakadCuiPresenter) writeCentrePile(b *strings.Builder, g interfaces.DehlaPakadGame) {
	pile := g.GetCentrePile()
	if len(pile) == 0 {
		return
	}
	b.WriteString(i18n.Tf("dehlapakad.centrePile",
		"n", strconv.Itoa(len(pile)),
		"tens", strconv.Itoa(g.GetCentrePileTens())) + "\n")
	if prev := g.GetPrevTrickWinner(); prev >= 0 {
		b.WriteString(i18n.Tf("dehlapakad.pileGoesTo",
			"name", cuiPlayerName(g.GetPlayer(prev), prev)) + "\n")
	}
}

// writeGameEnd は終局の行を書く。
func (p *DehlaPakadCuiPresenter) writeGameEnd(b *strings.Builder, g interfaces.DehlaPakadGame) {
	b.WriteString(color.Green(i18n.Tf("dehlapakad.gameEnd",
		"team", strconv.Itoa(g.GetWinnerTeam()))) + "\n")
}

// writePrompt はフェーズに応じた案内を書く。
func (p *DehlaPakadCuiPresenter) writePrompt(b *strings.Builder, g interfaces.DehlaPakadGame) {
	switch g.GetPhase() {
	case domain.DehlaPakadPhaseSelectTrump:
		idx := g.GetTrumpChooserIdx()
		// **決めるのは親の右隣で、見えているのは最初の 5 枚だけ。**
		b.WriteString(i18n.Tf("dehlapakad.promptTrump",
			"name", cuiPlayerName(g.GetPlayer(idx), idx),
			"n", strconv.Itoa(domain.DehlaPakadFirstBatch)) + "\n")
		b.WriteString(i18n.T("dehlapakad.promptTrumpHelp") + "\n")
	case domain.DehlaPakadPhasePlay:
		idx := g.GetCurrentTurn()
		b.WriteString(i18n.Tf("dehlapakad.promptPlay",
			"name", cuiPlayerName(g.GetPlayer(idx), idx)) + "\n")
		player := g.GetPlayer(idx)
		if player == nil || !player.GetIsHuman() {
			return
		}
		parts := make([]string, 0, DehlaPakadMaxHandForList)
		for _, i := range g.GetPlayableIndices(idx) {
			if i >= 0 && i < player.GetCardsSize() {
				parts = append(parts, "["+strconv.Itoa(i)+"]"+cuiCardStrEmojiRank(player.GetCard(i)))
			}
		}
		if len(parts) > 0 {
			b.WriteString(i18n.Tf("dehlapakad.playableList", "cards", strings.Join(parts, "  ")) + "\n")
		}
		b.WriteString(i18n.T("dehlapakad.promptPlayHelp") + "\n")
	case domain.DehlaPakadPhaseHandEnd:
		p.writeHandEnd(b, g)
	}
}

// DehlaPakadMaxHandForList は出せる札の一覧を組むときの初期容量。
const DehlaPakadMaxHandForList = domain.DehlaPakadHandSize

// writeHandEnd はハンドの結果を書く。
func (p *DehlaPakadCuiPresenter) writeHandEnd(b *strings.Builder, g interfaces.DehlaPakadGame) {
	if res := g.GetLastResult(); res != nil {
		b.WriteString(i18n.Tf("dehlapakad.promptHandEnd",
			"team", strconv.Itoa(res.WinnerTeam),
			"a", strconv.Itoa(res.TeamTens[0]),
			"b", strconv.Itoa(res.TeamTens[1])) + "\n")
		if res.Kot {
			b.WriteString(color.Yellow(i18n.T("dehlapakad.kot."+res.KotReason)) + "\n")
		}
	}
	b.WriteString(i18n.T("dehlapakad.promptHandEndHelp") + "\n")
}

// HintOutput emits the current hint.
func (p *DehlaPakadCuiPresenter) HintOutput(g interfaces.DehlaPakadGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("dehlapakad.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, dehlaPakadHintReasonKeys)
	if hint.TrumpSuit > 0 {
		return color.Yellow(i18n.Tf("dehlapakad.hintTrump",
			"suit", i18n.T("dehlapakad.suit."+domain.DehlaPakadSuitName(hint.TrumpSuit)),
			"reason", reason)) + "\n"
	}
	if len(hint.CardIndices) == 0 {
		return color.Yellow(i18n.Tf("dehlapakad.hintCard", "cards", "-", "reason", reason)) + "\n"
	}
	player := g.GetPlayer(g.GetCurrentTurn())
	cards := make([]string, len(hint.CardIndices))
	for i, idx := range hint.CardIndices {
		if player != nil && idx >= 0 && idx < player.GetCardsSize() {
			cards[i] = "[" + strconv.Itoa(idx) + "]" + cuiCardStrEmojiRank(player.GetCard(idx))
			continue
		}
		cards[i] = strconv.Itoa(idx)
	}
	return color.Yellow(i18n.Tf("dehlapakad.hintCard",
		"cards", strings.Join(cards, ", "), "reason", reason)) + "\n"
}

// dehlaPakadHintReasonKeys はヒント理由と i18n キーの対応。
var dehlaPakadHintReasonKeys = map[string]string{
	"call_longest":  "dehlapakad.hintReasonTrump",
	"take_the_ten":  "dehlapakad.hintReasonTakeTen",
	"keep_the_lead": "dehlapakad.hintReasonKeepLead",
	"next_hand":     "dehlapakad.hintReasonNextHand",
	"none":          "dehlapakad.hintNone",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *DehlaPakadCuiPresenter) ActionLogOutput(g interfaces.DehlaPakadGame) string {
	return actionLogOutputTextForSeats[*domain.DehlaPakadPlayer](g)
}
