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

// biribaPlayerStr returns the display string for a single Biriba player.
func biribaPlayerStr(player *domain.BiribaPlayer, i int, showCards bool) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("biriba.playerLine",
		"name", cuiPlayerName(player, i),
		"cum", strconv.Itoa(player.GetCumulativeScore()),
		"round", strconv.Itoa(player.GetRoundScore()),
		"cards", strconv.Itoa(player.GetCardsSize())))
	if len(player.GetRed3s()) > 0 {
		b.WriteString(i18n.Tf("biriba.playerRed3s",
			"count", strconv.Itoa(len(player.GetRed3s()))))
	}
	if player.HasBiriba() {
		b.WriteString(i18n.T("biriba.playerBiribaTag"))
	}
	if player.GetTookPozzetto() {
		b.WriteString(i18n.T("biriba.playerPozzettoTag"))
	}
	b.WriteString("\n")

	// Melds
	for _, m := range player.GetMelds() {
		meldType := i18n.T("biriba.meldTypeMixed")
		if m.IsNatural {
			meldType = i18n.T("biriba.meldTypeNatural")
		}
		if m.IsBiriba() {
			meldType += i18n.T("biriba.meldTypeBiribaSuffix")
		}
		cardStrs := make([]string, len(m.Cards))
		for j, c := range m.Cards {
			cardStrs[j] = cuiCardStr(c)
		}
		b.WriteString(i18n.Tf("biriba.meldLine",
			"type", meldType,
			"cards", strings.Join(cardStrs, ", ")) + "\n")
	}

	if showCards && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// BiribaCuiPresenter renders the Biriba CUI view.
type BiribaCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *BiribaCuiPresenter) Output(g interfaces.BiribaGame, lastErr error) string {
	return buildCuiOutput(i18n.T("biriba.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("biriba.header",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"stock", strconv.Itoa(g.GetDrawPileCount()),
			"discard", strconv.Itoa(g.GetDiscardPileCount())))
		if g.GetIsFrozen() {
			b.WriteString(i18n.T("biriba.frozenTag"))
		}
		b.WriteString("\n")
		b.WriteString(i18n.Tf("biriba.pozzettoLine",
			"count", strconv.Itoa(g.GetPozzettoCount())) + "\n")

		// Top of discard
		if top := g.GetDiscardTop(); top != nil {
			b.WriteString(i18n.Tf("biriba.discardLine", "card", cuiCardStr(top)) + "\n")
			// **山ごと取れるゲームなので中身は公開情報。**Web は details で全部
			// 見せているのに、CUI は一番上の 1 枚しか出しておらず、「山全体を取る」
			// 判断を一番上だけで迫っていた (#4833)。
			b.WriteString(cuiDiscardPileLines(g.GetDiscardPile(), "biriba.discardPileLine"))
		}

		// Players
		phase := g.GetPhase()
		showAllCards := phase == domain.BiribaPhaseRoundEnd || phase == domain.BiribaPhaseGameEnd
		for i := 0; i < g.GetPlayerCnt(); i++ {
			player := g.GetPlayer(i)
			showCards := player.GetIsHuman() || showAllCards
			b.WriteString(biribaPlayerStr(player, i, showCards))
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winnerIdx := g.GetWinnerIdx()
			banner := i18n.Tf("biriba.gameEnd",
				"name", cuiPlayerName(g.GetPlayer(winnerIdx), winnerIdx))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch phase {
		case domain.BiribaPhaseDraw:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("biriba.promptDraw",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("biriba.promptDrawHelpStock") + "\n")
			b.WriteString(i18n.T("biriba.promptDrawHelpDiscard") + "\n")
			if g.GetIsFrozen() {
				b.WriteString(i18n.T("biriba.promptDrawFrozen") + "\n")
			}
		case domain.BiribaPhaseMeld:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("biriba.promptMeld",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("biriba.promptMeldHelp") + "\n")
			b.WriteString(i18n.T("biriba.promptSkipMeld") + "\n")
		case domain.BiribaPhaseDiscard:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("biriba.promptDiscard",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("biriba.promptDiscardHelp") + "\n")
			b.WriteString(i18n.T("biriba.promptGoOutHelp") + "\n")
		case domain.BiribaPhaseRoundEnd:
			b.WriteString(i18n.T("biriba.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("biriba.promptRoundEndHelp") + "\n")
		}
	})
}

// biribaHintReasonKeys maps Biriba-specific hint reasons to i18n keys.
var biribaHintReasonKeys = map[string]string{
	"draw_discard_pair": "biriba.hintReasonDrawDiscard",
	"draw_stock_safe":   "biriba.hintReasonDrawStock",
	"meld_available":    "biriba.hintReasonMeld",
	"no_meld":           "biriba.hintReasonNoMeld",
	"discard_safe":      "biriba.hintReasonDiscard",
}

// HintOutput emits the current recommended action for the human player.
func (p *BiribaCuiPresenter) HintOutput(g interfaces.BiribaGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("biriba.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, biribaHintReasonKeys)
	switch hint.Action {
	case "draw_stock":
		return color.Yellow(i18n.Tf("biriba.hintDrawStock", "reason", reason)) + "\n"
	case "draw_discard":
		return color.Yellow(i18n.Tf("biriba.hintDrawDiscard",
			"indices", biribaJoinInts(hint.Indices), "reason", reason)) + "\n"
	case "meld":
		return color.Yellow(i18n.Tf("biriba.hintMeld",
			"indices", biribaJoinInts(hint.Indices), "reason", reason)) + "\n"
	case "skip_meld":
		return color.Yellow(i18n.Tf("biriba.hintSkipMeld", "reason", reason)) + "\n"
	case "discard":
		idx := 0
		if len(hint.Indices) > 0 {
			idx = hint.Indices[0]
		}
		card := ""
		if c := g.GetPlayer(g.GetCurrentPlayerIdx()).GetCard(idx); c != nil {
			card = cuiCardStr(c)
		}
		return color.Yellow(i18n.Tf("biriba.hintDiscard",
			"idx", strconv.Itoa(idx), "card", card, "reason", reason)) + "\n"
	}
	return i18n.T("biriba.hintNone") + "\n"
}

// biribaJoinInts formats a slice of indices as a comma-separated string.
func biribaJoinInts(xs []int) string {
	parts := make([]string, len(xs))
	for i, x := range xs {
		parts[i] = strconv.Itoa(x)
	}
	return strings.Join(parts, ",")
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *BiribaCuiPresenter) ActionLogOutput(g interfaces.BiribaGame) string {
	return actionLogOutputTextForSeats[*domain.CanastaPlayer](g)
}
