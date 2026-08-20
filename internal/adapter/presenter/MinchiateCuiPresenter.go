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

// minchiateCardStr renders one card as text.
//
// **cuiCardStr は使えない。**あれは design 5/6 を "UNKNOWN" に落とすので、切り札 21 枚と
// マットが手札一覧で全部同じ文字列になる。切札は番号ではなく呼び名で出す ——
// 40 枚もあると番号だけでは何の札か分からない。
func minchiateCardStr(card *domain.Card) string {
	if card == nil {
		return "??"
	}
	switch card.GetDesign() {
	case domain.MinchiateMattoDesign:
		return i18n.T("minchiate.cardMatto")
	case domain.MinchiateTrumpDesign:
		return i18n.Tf("minchiate.cardTrump",
			"value", strconv.Itoa(card.GetValue()),
			"name", domain.MinchiateTrumpName(card.GetValue()))
	default:
		return cuiSuitName(card.GetDesign()) + " " + strconv.Itoa(card.GetValue())
	}
}

// minchiateIndexedHandStr renders the human hand as "[0]<card>  [1]<card>  ...".
func minchiateIndexedHandStr(player *domain.MinchiatePlayer) string {
	parts := make([]string, 0, player.GetCardsSize())
	for i := 0; i < player.GetCardsSize(); i++ {
		parts = append(parts, "["+strconv.Itoa(i)+"]"+minchiateCardStr(player.GetCard(i)))
	}
	return strings.Join(parts, "  ")
}

// minchiatePlayerStr returns the display string for a single player.
func minchiatePlayerStr(g interfaces.MinchiateGame, idx int) string {
	player := g.GetPlayer(idx)
	if player == nil {
		return ""
	}
	role := ""
	if idx == g.GetDealerIdx() {
		role = " " + i18n.T("minchiate.dealerBadge")
	}
	var b strings.Builder
	b.WriteString(i18n.Tf("minchiate.playerLine",
		"name", cuiPlayerName(player, idx),
		"team", strconv.Itoa(domain.MinchiateTeamOf(idx)),
		"cards", strconv.Itoa(player.GetCardsSize()),
		"tricks", strconv.Itoa(player.GetTrickCount()),
	))
	b.WriteString(role + "\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(minchiateIndexedHandStr(player) + "\n")
	}
	return b.String()
}

// MinchiateCuiPresenter renders the Minchiate CUI view.
type MinchiateCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *MinchiateCuiPresenter) Output(g interfaces.MinchiateGame, lastErr error) string {
	return buildCuiOutput(i18n.T("minchiate.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("minchiate.round",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber())) + "\n")

		scores := g.GetTeamScores()
		b.WriteString(i18n.Tf("minchiate.teamLine",
			"team0", strconv.Itoa(scores[0]),
			"team1", strconv.Itoa(scores[1])) + "\n")

		// **切札が 40 枚あることを毎回出す。**21 枚のタロー系との差が最も効くのは
		// 「まだ上に何枚残っているか」の見積もりで、そこが読めないと切札を切る
		// タイミングを誤る。
		b.WriteString(i18n.T("minchiate.trumpCountNote") + "\n")
		// **マットだけが規則の外にいる。**取らない・フォローしない・リードも定めない
		// という 3 点は、ヒントを切っていると伝わる場所がどこにも無かった (#5715)。
		b.WriteString(i18n.T("minchiate.mattoNote") + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(minchiatePlayerStr(g, i))
		}

		b.WriteString("----------\n")

		cuiTrickBlock(b, g.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return minchiateCardStr(tc.Card) },
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

// writeGameEnd マッチ終了のバナーを書く。
func (p *MinchiateCuiPresenter) writeGameEnd(b *strings.Builder, g interfaces.MinchiateGame) {
	winner := g.GetWinnerTeam()
	if winner < 0 {
		b.WriteString(color.Green(i18n.T("minchiate.gameEndDraw")) + "\n")
		return
	}
	b.WriteString(color.Green(i18n.Tf("minchiate.gameEnd", "team", strconv.Itoa(winner))) + "\n")
}

// writePrompt renders the phase-specific prompt block.
func (p *MinchiateCuiPresenter) writePrompt(b *strings.Builder, g interfaces.MinchiateGame) {
	switch g.GetPhase() {
	case domain.MinchiatePhaseScarto:
		b.WriteString(i18n.Tf("minchiate.promptScarto",
			"count", strconv.Itoa(domain.MinchiateSurplus)) + "\n")
		b.WriteString(i18n.T("minchiate.promptScartoHelp") + "\n")
	case domain.MinchiatePhasePlay:
		currentIdx := g.GetCurrentPlayerIdx()
		b.WriteString(i18n.Tf("minchiate.promptPlay",
			"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
		b.WriteString(i18n.T("minchiate.promptPlayHelp") + "\n")
	case domain.MinchiatePhaseTrickEnd:
		b.WriteString(i18n.T("minchiate.promptTrickEnd") + "\n")
		b.WriteString(i18n.T("minchiate.promptTrickEndHelp") + "\n")
	case domain.MinchiatePhaseRoundEnd:
		b.WriteString(i18n.T("minchiate.promptRoundEnd") + "\n")
		p.writeRoundEndResult(b, g)
		b.WriteString(i18n.T("minchiate.promptRoundEndHelp") + "\n")
	}
}

// writeRoundEndResult appends a per-seat trick tally.
func (p *MinchiateCuiPresenter) writeRoundEndResult(b *strings.Builder, g interfaces.MinchiateGame) {
	entries := make([]string, 0, g.GetPlayerCnt())
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		entries = append(entries, i18n.Tf("minchiate.roundEndTrickEntry",
			"name", cuiPlayerName(player, i),
			"tricks", strconv.Itoa(player.GetTrickCount())))
	}
	b.WriteString(i18n.Tf("minchiate.roundEndTricks", "list", strings.Join(entries, ", ")) + "\n")
}

// HintOutput emits the current Minchiate hint.
func (p *MinchiateCuiPresenter) HintOutput(g interfaces.MinchiateGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("minchiate.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, minchiateHintReasonKeys)
	if len(hint.CardIndices) > 0 {
		player := g.GetPlayer(g.GetCurrentPlayerIdx())
		cards := make([]string, len(hint.CardIndices))
		for i, idx := range hint.CardIndices {
			if player != nil {
				cards[i] = "[" + strconv.Itoa(idx) + "]" + minchiateCardStr(player.GetCard(idx))
			} else {
				cards[i] = strconv.Itoa(idx)
			}
		}
		return color.Yellow(i18n.Tf("minchiate.hintCard",
			"cards", strings.Join(cards, ", "), "reason", reason)) + "\n"
	}
	return color.Yellow(i18n.Tf("minchiate.hintCard", "cards", "-", "reason", reason)) + "\n"
}

// minchiateHintReasonKeys maps Minchiate-specific hint-reason identifiers to i18n keys.
//
// **Minchiate.playHintReason が返す 5 種と 1 対 1 で対応させる。**外れた理由は
// hintReasonStr の既定でキー文字列そのものが画面に出る (#4660 で実際に起きた)。
var minchiateHintReasonKeys = map[string]string{
	"lead_low":     "minchiate.hintReasonLeadLow",
	"lead_trump":   "minchiate.hintReasonLeadTrump",
	"play_matto":   "minchiate.hintReasonPlayMatto",
	"follow_trump": "minchiate.hintReasonFollowTrump",
	"follow_low":   "minchiate.hintReasonFollowLow",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *MinchiateCuiPresenter) ActionLogOutput(g interfaces.MinchiateGame) string {
	return actionLogOutputTextWithNames(g, func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) })
}
