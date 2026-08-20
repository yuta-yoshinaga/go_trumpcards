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

// burracoPlayerStr returns the display string for a single Burraco player.
func burracoPlayerStr(player *domain.BurracoPlayer, i int, showCards bool) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("burraco.playerLine",
		"name", cuiPlayerName(player, i),
		"cum", strconv.Itoa(player.GetCumulativeScore()),
		"round", strconv.Itoa(player.GetRoundScore()),
		"cards", strconv.Itoa(player.GetCardsSize())))
	if len(player.GetRed3s()) > 0 {
		b.WriteString(i18n.Tf("burraco.playerRed3s",
			"count", strconv.Itoa(len(player.GetRed3s()))))
	}
	if player.HasBurraco() {
		b.WriteString(i18n.T("burraco.playerBurracoTag"))
	}
	if player.GetTookPozzetto() {
		b.WriteString(i18n.T("burraco.playerPozzettoTag"))
	}
	b.WriteString("\n")

	// Melds
	for _, m := range player.GetMelds() {
		meldType := i18n.T("burraco.meldTypeMixed")
		if m.IsNatural {
			meldType = i18n.T("burraco.meldTypeNatural")
		}
		if m.IsBurraco() {
			meldType += i18n.T("burraco.meldTypeBurracoSuffix")
		}
		cardStrs := make([]string, len(m.Cards))
		for j, c := range m.Cards {
			cardStrs[j] = cuiCardStr(c)
		}
		b.WriteString(i18n.Tf("burraco.meldLine",
			"type", meldType,
			"cards", strings.Join(cardStrs, ", ")) + "\n")
	}

	if showCards && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// BurracoCuiPresenter renders the Burraco CUI view.
type BurracoCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *BurracoCuiPresenter) Output(g interfaces.BurracoGame, lastErr error) string {
	return buildCuiOutput(i18n.T("burraco.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("burraco.header",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"stock", strconv.Itoa(g.GetDrawPileCount()),
			"discard", strconv.Itoa(g.GetDiscardPileCount())))
		if g.GetIsFrozen() {
			b.WriteString(i18n.T("burraco.frozenTag"))
		}
		b.WriteString("\n")
		b.WriteString(i18n.Tf("burraco.pozzettoLine",
			"count", strconv.Itoa(g.GetPozzettoCount())) + "\n")

		// Top of discard
		if top := g.GetDiscardTop(); top != nil {
			b.WriteString(i18n.Tf("burraco.discardLine", "card", cuiCardStr(top)) + "\n")
			// **山ごと取れるゲームなので中身は公開情報。**Web は details で全部
			// 見せているのに、CUI は一番上の 1 枚しか出しておらず、「山全体を取る」
			// 判断を一番上だけで迫っていた (#4833)。
			b.WriteString(cuiDiscardPileLines(g.GetDiscardPile(), "burraco.discardPileLine"))
		}

		// Players
		phase := g.GetPhase()
		showAllCards := phase == domain.BurracoPhaseRoundEnd || phase == domain.BurracoPhaseGameEnd
		for i := 0; i < g.GetPlayerCnt(); i++ {
			player := g.GetPlayer(i)
			showCards := player.GetIsHuman() || showAllCards
			b.WriteString(burracoPlayerStr(player, i, showCards))
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winnerIdx := g.GetWinnerIdx()
			banner := i18n.Tf("burraco.gameEnd",
				"name", cuiPlayerName(g.GetPlayer(winnerIdx), winnerIdx))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch phase {
		case domain.BurracoPhaseDraw:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("burraco.promptDraw",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("burraco.promptDrawHelpStock") + "\n")
			b.WriteString(i18n.T("burraco.promptDrawHelpDiscard") + "\n")
			if g.GetIsFrozen() {
				b.WriteString(i18n.T("burraco.promptDrawFrozen") + "\n")
			}
		case domain.BurracoPhaseMeld:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("burraco.promptMeld",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("burraco.promptMeldHelp") + "\n")
			b.WriteString(i18n.T("burraco.promptSkipMeld") + "\n")
		case domain.BurracoPhaseDiscard:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("burraco.promptDiscard",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("burraco.promptDiscardHelp") + "\n")
			b.WriteString(i18n.T("burraco.promptGoOutHelp") + "\n")
		case domain.BurracoPhaseRoundEnd:
			b.WriteString(i18n.T("burraco.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("burraco.promptRoundEndHelp") + "\n")
		}
	})
}

// burracoHintReasonKeys maps Burraco-specific hint reasons to i18n keys.
var burracoHintReasonKeys = map[string]string{
	"draw_discard_pair": "burraco.hintReasonDrawDiscard",
	"draw_stock_safe":   "burraco.hintReasonDrawStock",
	"meld_available":    "burraco.hintReasonMeld",
	"no_meld":           "burraco.hintReasonNoMeld",
	"discard_safe":      "burraco.hintReasonDiscard",
}

// HintOutput emits the current recommended action for the human player.
func (p *BurracoCuiPresenter) HintOutput(g interfaces.BurracoGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("burraco.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, burracoHintReasonKeys)
	switch hint.Action {
	case "draw_stock":
		return color.Yellow(i18n.Tf("burraco.hintDrawStock", "reason", reason)) + "\n"
	case "draw_discard":
		return color.Yellow(i18n.Tf("burraco.hintDrawDiscard",
			"indices", burracoJoinInts(hint.Indices), "reason", reason)) + "\n"
	case "meld":
		return color.Yellow(i18n.Tf("burraco.hintMeld",
			"indices", burracoJoinInts(hint.Indices), "reason", reason)) + "\n"
	case "skip_meld":
		return color.Yellow(i18n.Tf("burraco.hintSkipMeld", "reason", reason)) + "\n"
	case "discard":
		idx := 0
		if len(hint.Indices) > 0 {
			idx = hint.Indices[0]
		}
		card := ""
		if c := g.GetPlayer(g.GetCurrentPlayerIdx()).GetCard(idx); c != nil {
			card = cuiCardStr(c)
		}
		return color.Yellow(i18n.Tf("burraco.hintDiscard",
			"idx", strconv.Itoa(idx), "card", card, "reason", reason)) + "\n"
	}
	return i18n.T("burraco.hintNone") + "\n"
}

// burracoJoinInts formats a slice of indices as a comma-separated string.
func burracoJoinInts(xs []int) string {
	parts := make([]string, len(xs))
	for i, x := range xs {
		parts[i] = strconv.Itoa(x)
	}
	return strings.Join(parts, ",")
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *BurracoCuiPresenter) ActionLogOutput(g interfaces.BurracoGame) string {
	return actionLogOutputTextForSeats[*domain.CanastaPlayer](g)
}
