//go:build !js || !wasm || extra

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// ganjifaTrumpStr renders the trump glyph, or a "no trump" label when none.
func ganjifaTrumpStr(suit int) string {
	if suit < domain.CardDesignSpade {
		return i18n.T("ganjifa.noTrump")
	}
	return domain.GanjifaSuitGlyph(suit) + " " + domain.GanjifaSuitName(suit)
}

// ganjifaCardStr renders one card as "<glyph> <name> <value>".
//
// **cuiCardStr は使えない。**あれは design 5..8 を "UNKNOWN" に落とすので、
// 弱いスート群の 48 枚が手札一覧で全部同じ文字列になる。
func ganjifaCardStr(card *domain.Card) string {
	if card == nil {
		return "??"
	}
	return domain.GanjifaSuitGlyph(card.GetDesign()) + " " +
		domain.GanjifaSuitName(card.GetDesign()) + " " + strconv.Itoa(card.GetValue())
}

// ganjifaIndexedHandStr renders the human hand as "[0]<card>  [1]<card>  ...".
func ganjifaIndexedHandStr(player *domain.GanjifaPlayer) string {
	parts := make([]string, 0, player.GetCardsSize())
	for i := 0; i < player.GetCardsSize(); i++ {
		parts = append(parts, "["+strconv.Itoa(i)+"]"+ganjifaCardStr(player.GetCard(i)))
	}
	return strings.Join(parts, "  ")
}

// ganjifaPlayerStr returns the display string for a single player.
func ganjifaPlayerStr(g interfaces.GanjifaGame, idx int) string {
	player := g.GetPlayer(idx)
	if player == nil {
		return ""
	}
	scores := g.GetPlayerScores()
	// **役割欄は出さない。**Ganjifa は宣言者 1 対 2 ではなく 3 人が対等なので
	// 役割が無い。切り札の群は全員に共通なので、ここに載せると 3 行とも同じ値が
	// 並ぶだけになる —— その情報は上の専用行が 1 度だけ出す。
	var b strings.Builder
	b.WriteString(i18n.Tf("ganjifa.playerLine",
		"name", cuiPlayerName(player, idx),
		"cards", strconv.Itoa(player.GetCardsSize()),
		"score", strconv.Itoa(scores[idx]),
		"tricks", strconv.Itoa(player.GetTrickCount()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(ganjifaIndexedHandStr(player) + "\n")
	}
	return b.String()
}

// GanjifaCuiPresenter renders the Ganjifa CUI view.
type GanjifaCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *GanjifaCuiPresenter) Output(g interfaces.GanjifaGame, lastErr error) string {
	return buildCuiOutput(i18n.T("ganjifa.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("ganjifa.round",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber()),
			"trump", ganjifaTrumpStr(g.GetTrumpSuit())) + "\n")

		// **切り札がどちらの群かを毎回書く。**強い群なら数字の大きい札が強く、
		// 弱い群では逆になる。この一行が無いと、手札の並びを見ても
		// どちらの向きで読むべきか分からない。
		if domain.GanjifaIsStrongSuit(g.GetTrumpSuit()) {
			b.WriteString(i18n.T("ganjifa.trumpGroupStrong") + "\n")
		} else {
			b.WriteString(i18n.T("ganjifa.trumpGroupWeak") + "\n")
		}

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(ganjifaPlayerStr(g, i))
		}

		b.WriteString("----------\n")

		cuiTrickBlock(b, g.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return ganjifaCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winner := g.GetWinnerPlayer()
			var winnerStr string
			if winner >= 0 {
				winnerStr = cuiPlayerName(g.GetPlayer(winner), winner)
			}
			banner := i18n.Tf("ganjifa.gameEnd", "name", winnerStr)
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		p.writePrompt(b, g)
	})
}

// writePrompt renders the phase-specific prompt block.
func (p *GanjifaCuiPresenter) writePrompt(b *strings.Builder, g interfaces.GanjifaGame) {
	switch g.GetPhase() {
	case domain.GanjifaPhasePlay:
		currentIdx := g.GetCurrentPlayerIdx()
		b.WriteString(i18n.Tf("ganjifa.promptPlay",
			"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
		b.WriteString(i18n.T("ganjifa.promptPlayHelp") + "\n")
	case domain.GanjifaPhaseTrickEnd:
		// **取ったのが誰かは次にリードするのが誰かでもある** (#5653)。Web は
		// TrickDisplay の winnerIdx でバッジを出しているのに、CUI は「次のトリック
		// へ」としか言っていなかった。Sedma は同じ場面で既に名前を出している。
		if winnerIdx := g.GetLeadPlayerIdx(); winnerIdx >= 0 {
			b.WriteString(i18n.Tf("ganjifa.trickWinner",
				"name", cuiPlayerName(g.GetPlayer(winnerIdx), winnerIdx)) + "\n")
		}
		b.WriteString(i18n.T("ganjifa.promptTrickEnd") + "\n")
		b.WriteString(i18n.T("ganjifa.promptTrickEndHelp") + "\n")
	case domain.GanjifaPhaseRoundEnd:
		b.WriteString(i18n.T("ganjifa.promptRoundEnd") + "\n")
		p.writeRoundEndResult(b, g)
		b.WriteString(i18n.T("ganjifa.promptRoundEndHelp") + "\n")
	}
}

// writeRoundEndResult appends a one-line trick tally for every player.
//
// **契約の達成可否は無い。**Ganjifa には宣言が無く、3 人が対等にトリック数を
// 競うだけなので、出すべきは誰が何トリック取ったかに尽きる。
func (p *GanjifaCuiPresenter) writeRoundEndResult(b *strings.Builder, g interfaces.GanjifaGame) {
	entries := make([]string, 0, g.GetPlayerCnt())
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		entries = append(entries, i18n.Tf("ganjifa.roundEndTrickEntry",
			"name", cuiPlayerName(player, i),
			"tricks", strconv.Itoa(player.GetTrickCount())))
	}
	b.WriteString(i18n.Tf("ganjifa.roundEndTricks", "list", strings.Join(entries, ", ")) + "\n")
}

// HintOutput emits the current Ganjifa hint.
func (p *GanjifaCuiPresenter) HintOutput(g interfaces.GanjifaGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("ganjifa.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, ganjifaHintReasonKeys)
	if len(hint.CardIndices) > 0 {
		playerIdx := g.GetCurrentPlayerIdx()
		player := g.GetPlayer(playerIdx)
		cards := make([]string, len(hint.CardIndices))
		for i, idx := range hint.CardIndices {
			if player != nil {
				cards[i] = "[" + strconv.Itoa(idx) + "]" + ganjifaCardStr(player.GetCard(idx))
			} else {
				cards[i] = strconv.Itoa(idx)
			}
		}
		return color.Yellow(i18n.Tf("ganjifa.hintCard",
			"cards", strings.Join(cards, ", "),
			"reason", reason)) + "\n"
	}
	return color.Yellow(i18n.Tf("ganjifa.hintCard",
		"cards", "-",
		"reason", reason)) + "\n"
}

// ganjifaHintReasonKeys maps Ganjifa-specific hint-reason identifiers to i18n keys.
var ganjifaHintReasonKeys = map[string]string{
	"lead_low":         "ganjifa.hintReasonLeadLow",
	"lead_high":        "ganjifa.hintReasonLeadHigh",
	"lead_weak_suit":   "ganjifa.hintReasonLeadWeakSuit",
	"follow_trump":     "ganjifa.hintReasonFollowTrump",
	"follow_weak_suit": "ganjifa.hintReasonFollowWeakSuit",
	"follow_win":       "ganjifa.hintReasonFollowWin",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *GanjifaCuiPresenter) ActionLogOutput(g interfaces.GanjifaGame) string {
	return actionLogOutputTextWithNames(g, func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) })
}
