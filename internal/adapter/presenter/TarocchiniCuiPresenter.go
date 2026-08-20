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

// tarocchiniCardStr renders one card as text.
//
// **cuiCardStr は使えない。**あれは design 5/6 を "UNKNOWN" に落とすので、切り札 21 枚と
// マットが手札一覧で全部同じ文字列になる。パパは同格なので番号ではなく "Papa" と出す ——
// 番号で出すと 2 が 3 より弱いと読まれてしまう。
func tarocchiniCardStr(card *domain.Card) string {
	if card == nil {
		return "??"
	}
	switch card.GetDesign() {
	case domain.TarocchiniMattoDesign:
		return i18n.T("tarocchini.cardMatto")
	case domain.TarocchiniTrumpDesign:
		if domain.TarocchiniIsPapa(card) {
			return i18n.T("tarocchini.cardPapa")
		}
		return i18n.Tf("tarocchini.cardTrump", "value", strconv.Itoa(card.GetValue()))
	default:
		return cuiSuitName(card.GetDesign()) + " " + strconv.Itoa(card.GetValue())
	}
}

// tarocchiniIndexedHandStr renders the human hand as "[0]<card>  [1]<card>  ...".
func tarocchiniIndexedHandStr(player *domain.TarocchiniPlayer) string {
	parts := make([]string, 0, player.GetCardsSize())
	for i := 0; i < player.GetCardsSize(); i++ {
		parts = append(parts, "["+strconv.Itoa(i)+"]"+tarocchiniCardStr(player.GetCard(i)))
	}
	return strings.Join(parts, "  ")
}

// tarocchiniPlayerStr returns the display string for a single player.
func tarocchiniPlayerStr(g interfaces.TarocchiniGame, idx int) string {
	player := g.GetPlayer(idx)
	if player == nil {
		return ""
	}
	role := ""
	if idx == g.GetDealerIdx() {
		role = " " + i18n.T("tarocchini.dealerBadge")
	}
	var b strings.Builder
	b.WriteString(i18n.Tf("tarocchini.playerLine",
		"name", cuiPlayerName(player, idx),
		"team", strconv.Itoa(domain.TarocchiniTeamOf(idx)),
		"cards", strconv.Itoa(player.GetCardsSize()),
		"tricks", strconv.Itoa(player.GetTrickCount()),
	))
	b.WriteString(role + "\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(tarocchiniIndexedHandStr(player) + "\n")
	}
	return b.String()
}

// TarocchiniCuiPresenter renders the Tarocchini CUI view.
type TarocchiniCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *TarocchiniCuiPresenter) Output(g interfaces.TarocchiniGame, lastErr error) string {
	return buildCuiOutput(i18n.T("tarocchini.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("tarocchini.round",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber())) + "\n")

		scores := g.GetTeamScores()
		b.WriteString(i18n.Tf("tarocchini.teamLine",
			"team0", strconv.Itoa(scores[0]),
			"team1", strconv.Itoa(scores[1])) + "\n")

		// **同格のパパは毎回説明する。**「後から出した方が勝つ」を知らないと、
		// 手札の Papa 4 枚をどう使うか判断できない。
		b.WriteString(i18n.T("tarocchini.papiNote") + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(tarocchiniPlayerStr(g, i))
		}

		b.WriteString("----------\n")

		cuiTrickBlock(b, g.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return tarocchiniCardStr(tc.Card) },
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
func (p *TarocchiniCuiPresenter) writeGameEnd(b *strings.Builder, g interfaces.TarocchiniGame) {
	winner := g.GetWinnerTeam()
	if winner < 0 {
		b.WriteString(color.Green(i18n.T("tarocchini.gameEndDraw")) + "\n")
		return
	}
	b.WriteString(color.Green(i18n.Tf("tarocchini.gameEnd", "team", strconv.Itoa(winner))) + "\n")
}

// writePrompt renders the phase-specific prompt block.
func (p *TarocchiniCuiPresenter) writePrompt(b *strings.Builder, g interfaces.TarocchiniGame) {
	switch g.GetPhase() {
	case domain.TarocchiniPhaseScarto:
		b.WriteString(i18n.Tf("tarocchini.promptScarto",
			"count", strconv.Itoa(domain.TarocchiniSurplus)) + "\n")
		b.WriteString(i18n.T("tarocchini.promptScartoHelp") + "\n")
	case domain.TarocchiniPhasePlay:
		currentIdx := g.GetCurrentPlayerIdx()
		b.WriteString(i18n.Tf("tarocchini.promptPlay",
			"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
		b.WriteString(i18n.T("tarocchini.promptPlayHelp") + "\n")
	case domain.TarocchiniPhaseTrickEnd:
		b.WriteString(i18n.T("tarocchini.promptTrickEnd") + "\n")
		b.WriteString(i18n.T("tarocchini.promptTrickEndHelp") + "\n")
	case domain.TarocchiniPhaseRoundEnd:
		b.WriteString(i18n.T("tarocchini.promptRoundEnd") + "\n")
		p.writeRoundEndResult(b, g)
		b.WriteString(i18n.T("tarocchini.promptRoundEndHelp") + "\n")
	}
}

// writeRoundEndResult appends a per-seat trick tally.
func (p *TarocchiniCuiPresenter) writeRoundEndResult(b *strings.Builder, g interfaces.TarocchiniGame) {
	entries := make([]string, 0, g.GetPlayerCnt())
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		entries = append(entries, i18n.Tf("tarocchini.roundEndTrickEntry",
			"name", cuiPlayerName(player, i),
			"tricks", strconv.Itoa(player.GetTrickCount())))
	}
	b.WriteString(i18n.Tf("tarocchini.roundEndTricks", "list", strings.Join(entries, ", ")) + "\n")

	// 得点はトリック数だけではない (settleRound は最終トリック +2 とスカルト加点も
	// 足す)。内訳を出さないと teamScores の増分と突き合わせて検算できない。
	if winner := g.GetLastTrickWinner(); winner >= 0 {
		b.WriteString(i18n.Tf("tarocchini.roundEndLastTrick",
			"team", i18n.Tf("tarocchini.teamName", "n", strconv.Itoa(domain.TarocchiniTeamOf(winner))),
			"bonus", strconv.Itoa(domain.TarocchiniLastTrickBonus)) + "\n")
	}
	if scarto := g.GetScartoSize(); scarto > 0 {
		b.WriteString(i18n.Tf("tarocchini.roundEndScarto",
			"team", i18n.Tf("tarocchini.teamName", "n", strconv.Itoa(domain.TarocchiniTeamOf(g.GetDealerIdx()))),
			"count", strconv.Itoa(scarto)) + "\n")
	}
}

// HintOutput emits the current Tarocchini hint.
func (p *TarocchiniCuiPresenter) HintOutput(g interfaces.TarocchiniGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("tarocchini.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, tarocchiniHintReasonKeys)
	if len(hint.CardIndices) > 0 {
		player := g.GetPlayer(g.GetCurrentPlayerIdx())
		cards := make([]string, len(hint.CardIndices))
		for i, idx := range hint.CardIndices {
			if player != nil {
				cards[i] = "[" + strconv.Itoa(idx) + "]" + tarocchiniCardStr(player.GetCard(idx))
			} else {
				cards[i] = strconv.Itoa(idx)
			}
		}
		return color.Yellow(i18n.Tf("tarocchini.hintCard",
			"cards", strings.Join(cards, ", "), "reason", reason)) + "\n"
	}
	return color.Yellow(i18n.Tf("tarocchini.hintCard", "cards", "-", "reason", reason)) + "\n"
}

// tarocchiniHintReasonKeys maps Tarocchini-specific hint-reason identifiers to i18n keys.
//
// **Tarocchini.playHintReason が返す 6 種と 1 対 1 で対応させる。**外れた理由は
// hintReasonStr の既定でキー文字列そのものが画面に出る (#4660 で実際に起きた)。
var tarocchiniHintReasonKeys = map[string]string{
	"lead_low":     "tarocchini.hintReasonLeadLow",
	"lead_trump":   "tarocchini.hintReasonLeadTrump",
	"play_papa":    "tarocchini.hintReasonPlayPapa",
	"play_matto":   "tarocchini.hintReasonPlayMatto",
	"follow_trump": "tarocchini.hintReasonFollowTrump",
	"follow_low":   "tarocchini.hintReasonFollowLow",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *TarocchiniCuiPresenter) ActionLogOutput(g interfaces.TarocchiniGame) string {
	return actionLogOutputTextWithNames(g, func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) })
}
