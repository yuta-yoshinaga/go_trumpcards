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

// sutdaCuiCardStr は札の CUI 表示を返す。
//
// **1・3・8 月だけは 2 枚が別物。** 片方が「光」で、どちらを持っているかで役が
// 変わるので、月だけでは盤面から役が読めない。
func sutdaCuiCardStr(c *domain.Card) string {
	if c == nil {
		return "??"
	}
	s := strconv.Itoa(c.GetDesign()) + i18n.T("sutda.monthSuffix")
	if domain.SutdaIsGwang(c) {
		return color.Yellow(s + i18n.T("sutda.gwangMark"))
	}
	return s
}

// sutdaHandLabel は役の訳名を返す。
func sutdaHandLabel(name string) string {
	if name == "" || name == "none" {
		return i18n.T("sutda.handName.none")
	}
	return i18n.T("sutda.handName." + name)
}

// SutdaCuiPresenter renders the Sutda CUI view.
type SutdaCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *SutdaCuiPresenter) Output(g interfaces.SutdaGame, lastErr error) string {
	return buildCuiOutput(i18n.T("sutda.helpTitle"), func(b *strings.Builder) {
		p.writeHeader(b, g)
		p.writeSeats(b, g)
		b.WriteString("----------\n")
		cuiErrorBlock(b, lastErr)
		if g.GetGameEndFlag() {
			p.writeGameEnd(b, g)
			return
		}
		p.writePrompt(b, g)
	})
}

// writeHeader はハンド番号とポットを書く。
//
// **配り終えたポットは 0 になる。** ショーダウンでそのまま 0 と出すと、直後の
// 「誰が何を取った」と噛み合わずに読めないので、その局面では取られた額を出す。
func (p *SutdaCuiPresenter) writeHeader(b *strings.Builder, g interfaces.SutdaGame) {
	b.WriteString(i18n.Tf("sutda.hand", "n", strconv.Itoa(g.GetHandNumber())) + "\n")
	pot := g.GetPot()
	if g.GetPhase() != domain.SutdaPhaseBet {
		if res := g.GetLastResult(); res != nil {
			pot = res.Pot
		}
	}
	b.WriteString(i18n.Tf("sutda.potLine",
		"pot", strconv.Itoa(pot),
		"bet", strconv.Itoa(g.GetCurrentBet())) + "\n")
}

// writeSeats は席ごとの行を書く。
func (p *SutdaCuiPresenter) writeSeats(b *strings.Builder, g interfaces.SutdaGame) {
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		role := i18n.T("sutda.rolePlayer")
		if i == g.GetDealerIdx() {
			role = i18n.T("sutda.roleDealer")
		}
		state := ""
		if player.IsFolded() {
			state = " " + i18n.T("sutda.folded")
		}
		b.WriteString(i18n.Tf("sutda.playerLine",
			"name", cuiPlayerName(player, i),
			"role", role,
			"chips", strconv.Itoa(player.GetChips()),
			"bet", strconv.Itoa(player.GetBet())) + state + "\n")
		// **開くのはショーダウンだけ。** ベッティング中に相手の札が見えると
		// 賭ける意味が無くなる。
		if player.GetIsHuman() || player.IsRevealed() {
			p.writeHandLine(b, g, i, player)
		}
	}
}

// writeHandLine は席の手札と役を書く。
func (p *SutdaCuiPresenter) writeHandLine(b *strings.Builder, g interfaces.SutdaGame, idx int, player *domain.SutdaPlayer) {
	if player.GetCardsSize() == 0 {
		return
	}
	parts := make([]string, player.GetCardsSize())
	for i := 0; i < player.GetCardsSize(); i++ {
		parts[i] = sutdaCuiCardStr(player.GetCard(i))
	}
	b.WriteString(i18n.Tf("sutda.handLine",
		"cards", strings.Join(parts, " "),
		"hand", sutdaHandLabel(g.GetHandOf(idx).Name)) + "\n")
}

// writeGameEnd は終局の行を書く。
func (p *SutdaCuiPresenter) writeGameEnd(b *strings.Builder, g interfaces.SutdaGame) {
	idx := g.GetWinnerIdx()
	b.WriteString(color.Green(i18n.Tf("sutda.gameEnd",
		"name", cuiPlayerName(g.GetPlayer(idx), idx))) + "\n")
}

// writePrompt はフェーズに応じた案内を書く。
func (p *SutdaCuiPresenter) writePrompt(b *strings.Builder, g interfaces.SutdaGame) {
	switch g.GetPhase() {
	case domain.SutdaPhaseBet:
		idx := g.GetCurrentPlayerIdx()
		b.WriteString(i18n.Tf("sutda.promptBet",
			"name", cuiPlayerName(g.GetPlayer(idx), idx)) + "\n")
		player := g.GetPlayer(idx)
		if player == nil || !player.GetIsHuman() {
			return
		}
		// **コールに要る額を出す。** チェックなのか払うのかが分からないと
		// 押す前に決められない。
		need := g.GetCallAmount(idx)
		if need > 0 {
			b.WriteString(i18n.Tf("sutda.toCall", "amount", strconv.Itoa(need)) + "\n")
		} else {
			b.WriteString(i18n.T("sutda.canCheck") + "\n")
		}
		if g.CanRaise(idx) {
			b.WriteString(i18n.Tf("sutda.canRaise",
				"amount", strconv.Itoa(domain.SutdaBetUnit),
				"left", strconv.Itoa(domain.SutdaMaxRaises-g.GetRaiseCount())) + "\n")
		}
		b.WriteString(i18n.T("sutda.promptBetHelp") + "\n")
	case domain.SutdaPhaseShowdown:
		p.writeShowdown(b, g)
	}
}

// writeShowdown はハンドの結果を書く。
func (p *SutdaCuiPresenter) writeShowdown(b *strings.Builder, g interfaces.SutdaGame) {
	if res := g.GetLastResult(); res != nil {
		names := make([]string, 0, len(res.Winners))
		for _, w := range res.Winners {
			names = append(names, cuiPlayerName(g.GetPlayer(w), w))
		}
		b.WriteString(i18n.Tf("sutda.showdown",
			"names", strings.Join(names, ", "),
			"pot", strconv.Itoa(res.Pot)) + "\n")
	}
	b.WriteString(i18n.T("sutda.promptShowdownHelp") + "\n")
}

// HintOutput emits the current hint.
func (p *SutdaCuiPresenter) HintOutput(g interfaces.SutdaGame) string {
	hint := g.GetHint()
	if hint == nil || hint.Action == "" {
		return i18n.T("sutda.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, sutdaHintReasonKeys)
	return color.Yellow(i18n.Tf("sutda.hintAction",
		"action", i18n.T("sutda.action."+hint.Action), "reason", reason)) + "\n"
}

// sutdaHintReasonKeys はヒント理由と i18n キーの対応。
var sutdaHintReasonKeys = map[string]string{
	"strong_hand": "sutda.hintReasonStrong",
	"stay_in":     "sutda.hintReasonStay",
	"weak_hand":   "sutda.hintReasonWeak",
	"next_hand":   "sutda.hintReasonNextHand",
	"none":        "sutda.hintNone",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *SutdaCuiPresenter) ActionLogOutput(g interfaces.SutdaGame) string {
	return actionLogOutputTextForSeats[*domain.SutdaPlayer](g)
}
