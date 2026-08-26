//go:build !js || !wasm || classic

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

// unsunKarutaCuiCardStr は札の CUI 表示を返す。
//
// **数札の強さはスートで逆になる。** 表記だけでは強弱が読めないので、
// スート名も一緒に出す (「こつ 1」は最強の数札、「ぱお 1」は最弱)。
func unsunKarutaCuiCardStr(c *domain.Card) string {
	if c == nil {
		return "??"
	}
	suit := i18n.T("unsunkaruta.suit." + domain.UnsunKarutaSuitName(c.GetDesign()))
	rank := domain.UnsunKarutaRankName(c.GetValue())
	if c.GetValue() > 9 {
		rank = i18n.T("unsunkaruta.rank." + rank)
	}
	s := suit + rank
	if domain.UnsunKarutaIsRoundSuit(c.GetDesign()) {
		return color.Red(s)
	}
	return s
}

// unsunKarutaIndexedHand は人間の手札を番号付きで返す。
func unsunKarutaIndexedHand(p *domain.UnsunKarutaPlayer) string {
	parts := make([]string, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		parts[i] = fmt.Sprintf("[%d]%s", i, unsunKarutaCuiCardStr(p.GetCard(i)))
	}
	return strings.Join(parts, "  ")
}

// UnsunKarutaCuiPresenter renders the Unsun Karuta CUI view.
type UnsunKarutaCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *UnsunKarutaCuiPresenter) Output(g interfaces.UnsunKarutaGame, lastErr error) string {
	return buildCuiOutput(i18n.T("unsunkaruta.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("unsunkaruta.round",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber()),
			"tricks", strconv.Itoa(domain.UnsunKarutaTrickCount)) + "\n")
		// **切り札は返した 1 枚で決まる。** 出さないと、どのスートが強いのかが
		// 画面のどこにも無い。
		b.WriteString(i18n.Tf("unsunkaruta.trump",
			"suit", i18n.T("unsunkaruta.suit."+domain.UnsunKarutaSuitName(g.GetTrumpSuit())),
			"card", unsunKarutaCuiCardStr(g.TrumpCard())) + "\n")
		tricks := g.GetTeamTricks()
		scores := g.GetTeamScores()
		if len(tricks) == domain.UnsunKarutaTeamCnt && len(scores) == domain.UnsunKarutaTeamCnt {
			b.WriteString(i18n.Tf("unsunkaruta.teamLine",
				"a", strconv.Itoa(tricks[0]), "b", strconv.Itoa(tricks[1]),
				"sa", strconv.Itoa(scores[0]), "sb", strconv.Itoa(scores[1])) + "\n")
		}

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(unsunKarutaPlayerStr(g, i))
		}
		b.WriteString("----------\n")

		cuiTrickBlock(b, g.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return unsunKarutaCuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			p.writeGameEnd(b, g)
			return
		}
		p.writePrompt(b, g)
	})
}

// unsunKarutaPlayerStr は 1 席ぶんの行を返す。
func unsunKarutaPlayerStr(g interfaces.UnsunKarutaGame, idx int) string {
	player := g.GetPlayer(idx)
	if player == nil {
		return ""
	}
	role := i18n.T("unsunkaruta.rolePlayer")
	if idx == g.GetDealerIdx() {
		role = i18n.T("unsunkaruta.roleDealer")
	}
	var b strings.Builder
	b.WriteString(i18n.Tf("unsunkaruta.playerLine",
		"name", cuiPlayerName(player, idx),
		"team", strconv.Itoa(domain.UnsunKarutaTeamOf(idx)),
		"role", role,
		"cards", strconv.Itoa(player.GetCardsSize()),
		"tricks", strconv.Itoa(player.GetTrickCount())))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(unsunKarutaIndexedHand(player) + "\n")
	}
	return b.String()
}

// writeGameEnd は終局の行を書く。
func (p *UnsunKarutaCuiPresenter) writeGameEnd(b *strings.Builder, g interfaces.UnsunKarutaGame) {
	if g.GetWinnerTeam() < 0 {
		b.WriteString(color.Green(i18n.T("unsunkaruta.gameEndDraw")) + "\n")
		return
	}
	b.WriteString(color.Green(i18n.Tf("unsunkaruta.gameEnd",
		"team", strconv.Itoa(g.GetWinnerTeam()))) + "\n")
}

// writePrompt はフェーズに応じた案内を書く。
func (p *UnsunKarutaCuiPresenter) writePrompt(b *strings.Builder, g interfaces.UnsunKarutaGame) {
	switch g.GetPhase() {
	case domain.UnsunKarutaPhasePlay:
		idx := g.GetCurrentPlayerIdx()
		b.WriteString(i18n.Tf("unsunkaruta.promptPlay",
			"name", cuiPlayerName(g.GetPlayer(idx), idx)) + "\n")
		// **フォロー義務があるかを出す。** 宣言で生まれる義務なので、
		// 出さないと「なぜこの札しか出せないのか」が読めない。
		if g.IsMustFollow() {
			b.WriteString(i18n.T("unsunkaruta.mustFollow") + "\n")
		}
		if player := g.GetPlayer(idx); player != nil && player.GetIsHuman() {
			var parts []string
			for _, i := range g.GetPlayableIndices(idx) {
				if i >= 0 && i < player.GetCardsSize() {
					parts = append(parts, "["+strconv.Itoa(i)+"]"+unsunKarutaCuiCardStr(player.GetCard(i)))
				}
			}
			if len(parts) > 0 {
				b.WriteString(i18n.Tf("unsunkaruta.playableList", "cards", strings.Join(parts, "  ")) + "\n")
			}
		}
		if g.CanDeclare() {
			b.WriteString(i18n.T("unsunkaruta.promptDeclare") + "\n")
		}
		b.WriteString(i18n.T("unsunkaruta.promptPlayHelp") + "\n")
	case domain.UnsunKarutaPhaseTrickEnd:
		// **誰が取ったのかを名指しする。** 8 枚並んだ盤面から勝者を読むには
		// 「切り札か、無ければ台札の最強」を毎回自分で解くことになる。
		if w := g.GetLastTrickWinner(); w >= 0 {
			b.WriteString(i18n.Tf("unsunkaruta.promptTrickEndWinner",
				"name", cuiPlayerName(g.GetPlayer(w), w)) + "\n")
		} else {
			b.WriteString(i18n.T("unsunkaruta.promptTrickEnd") + "\n")
		}
		b.WriteString(i18n.T("unsunkaruta.promptTrickEndHelp") + "\n")
	case domain.UnsunKarutaPhaseRoundEnd:
		tricks := g.GetTeamTricks()
		if len(tricks) == domain.UnsunKarutaTeamCnt {
			b.WriteString(i18n.Tf("unsunkaruta.promptRoundEnd",
				"a", strconv.Itoa(tricks[0]), "b", strconv.Itoa(tricks[1])) + "\n")
		}
		b.WriteString(i18n.T("unsunkaruta.promptRoundEndHelp") + "\n")
	}
}

// HintOutput emits the current hint.
func (p *UnsunKarutaCuiPresenter) HintOutput(g interfaces.UnsunKarutaGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("unsunkaruta.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, unsunKarutaHintReasonKeys)
	if len(hint.CardIndices) == 0 {
		return color.Yellow(i18n.Tf("unsunkaruta.hintCard", "cards", "-", "reason", reason)) + "\n"
	}
	player := g.GetPlayer(g.GetCurrentPlayerIdx())
	cards := make([]string, len(hint.CardIndices))
	for i, idx := range hint.CardIndices {
		if player != nil && idx >= 0 && idx < player.GetCardsSize() {
			cards[i] = "[" + strconv.Itoa(idx) + "]" + unsunKarutaCuiCardStr(player.GetCard(idx))
			continue
		}
		cards[i] = strconv.Itoa(idx)
	}
	return color.Yellow(i18n.Tf("unsunkaruta.hintCard",
		"cards", strings.Join(cards, ", "), "reason", reason)) + "\n"
}

// unsunKarutaHintReasonKeys はヒント理由と i18n キーの対応。
var unsunKarutaHintReasonKeys = map[string]string{
	"lead_strong": "unsunkaruta.hintReasonLead",
	"follow_play": "unsunkaruta.hintReasonFollow",
	"next_trick":  "unsunkaruta.hintReasonNextTrick",
	"next_round":  "unsunkaruta.hintReasonNextRound",
	"none":        "unsunkaruta.hintNone",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *UnsunKarutaCuiPresenter) ActionLogOutput(g interfaces.UnsunKarutaGame) string {
	return actionLogOutputTextForSeats[*domain.UnsunKarutaPlayer](g)
}
