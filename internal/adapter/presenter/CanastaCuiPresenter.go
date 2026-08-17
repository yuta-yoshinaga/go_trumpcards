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

// canastaMinMeld returns the minimum points required for a player's initial
// meld, tiered by their cumulative score (mirrors the web canastaMinMeld).
func canastaMinMeld(cumulativeScore int) int {
	switch {
	case cumulativeScore < 0:
		return 15
	case cumulativeScore < 1500:
		return 50
	case cumulativeScore < 3000:
		return 90
	default:
		return 120
	}
}

// canastaPlayerStr returns the display string for a single Canasta player.
func canastaPlayerStr(player *domain.CanastaPlayer, i int, showCards bool) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("canasta.playerLine",
		"name", cuiPlayerName(player, i),
		"cum", strconv.Itoa(player.GetCumulativeScore()),
		"round", strconv.Itoa(player.GetRoundScore()),
		"cards", strconv.Itoa(player.GetCardsSize())))
	if len(player.GetRed3s()) > 0 {
		b.WriteString(i18n.Tf("canasta.playerRed3s",
			"count", strconv.Itoa(len(player.GetRed3s()))))
	}
	if player.HasCanasta() {
		b.WriteString(i18n.T("canasta.playerCanastaTag"))
	}
	b.WriteString("\n")

	// Melds
	for _, m := range player.GetMelds() {
		meldType := i18n.T("canasta.meldTypeMixed")
		if m.IsNatural {
			meldType = i18n.T("canasta.meldTypeNatural")
		}
		if m.IsCanasta() {
			meldType += i18n.T("canasta.meldTypeCanastaSuffix")
		}
		cardStrs := make([]string, len(m.Cards))
		for j, c := range m.Cards {
			cardStrs[j] = cuiCardStr(c)
		}
		b.WriteString(i18n.Tf("canasta.meldLine",
			"type", meldType,
			"cards", strings.Join(cardStrs, ", ")) + "\n")
	}

	if showCards && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// CanastaCuiPresenter renders the Canasta CUI view.
type CanastaCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *CanastaCuiPresenter) Output(g interfaces.CanastaGame, lastErr error) string {
	return buildCuiOutput(i18n.T("canasta.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("canasta.header",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"stock", strconv.Itoa(g.GetDrawPileCount()),
			"discard", strconv.Itoa(g.GetDiscardPileCount())))
		if g.GetIsFrozen() {
			b.WriteString(i18n.T("canasta.frozenTag"))
		}
		b.WriteString("\n")

		// Top of discard
		if top := g.GetDiscardTop(); top != nil {
			b.WriteString(i18n.Tf("canasta.discardLine", "card", cuiCardStr(top)) + "\n")
			// 山ごと取るゲームなので中身は公開情報 (#5043)。Burraco と同じ形。
			b.WriteString(cuiDiscardPileLines(g.GetDiscardPile(), "canasta.discardPileLine"))
		}

		// Players
		phase := g.GetPhase()
		showAllCards := phase == domain.CanastaPhaseRoundEnd || phase == domain.CanastaPhaseGameEnd
		for i := 0; i < g.GetPlayerCnt(); i++ {
			player := g.GetPlayer(i)
			showCards := player.GetIsHuman() || showAllCards
			b.WriteString(canastaPlayerStr(player, i, showCards))
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winnerIdx := g.GetWinnerIdx()
			banner := i18n.Tf("canasta.gameEnd",
				"name", cuiPlayerName(g.GetPlayer(winnerIdx), winnerIdx))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch phase {
		case domain.CanastaPhaseDraw:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("canasta.promptDraw",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			if g.GetIsFrozen() {
				b.WriteString(color.Yellow(i18n.T("canasta.frozenPickupNote")) + "\n")
			}
			b.WriteString(i18n.T("canasta.promptDrawHelpStock") + "\n")
			b.WriteString(i18n.T("canasta.promptDrawHelpDiscard") + "\n")
			// **取れない理由は、インデックスを打つ前に言う。** Web は選択中の札を
			// 事前検証して理由を出すのに、CUI は dd を送ってサーバに弾かれて初めて
			// 1行返る形だった (#5502)。選んだ2枚に依存しない条件だけを見るので、
			// 「選択」という概念の無い CUI でも同じ案内ができる。
			if p := g.GetPlayer(currentIdx); p != nil && p.GetIsHuman() {
				if blocker := g.GetDrawFromDiscardBlocker(); blocker != "" {
					b.WriteString(color.Yellow(i18n.T("canasta.drawBlocker"+blocker)) + "\n")
				}
			}
		case domain.CanastaPhaseMeld:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("canasta.promptMeld",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			// Surface the score-tiered initial-meld minimum so the human need not
			// memorize the 15/50/90/120 bands (mirrors the web ca-meld-points).
			if cur := g.GetPlayer(currentIdx); cur != nil && cur.GetIsHuman() && !cur.GetHasInitMeld() {
				b.WriteString(color.Yellow(i18n.Tf("canasta.initialMeldRequirement",
					"min", strconv.Itoa(canastaMinMeld(cur.GetCumulativeScore())),
					"score", strconv.Itoa(cur.GetCumulativeScore()))) + "\n")
			}
			b.WriteString(i18n.T("canasta.promptMeldHelp") + "\n")
			b.WriteString(i18n.T("canasta.promptSkipMeld") + "\n")
		case domain.CanastaPhaseDiscard:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("canasta.promptDiscard",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("canasta.promptDiscardHelp") + "\n")
			b.WriteString(i18n.T("canasta.promptGoOutHelp") + "\n")
		case domain.CanastaPhaseRoundEnd:
			b.WriteString(i18n.T("canasta.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("canasta.promptRoundEndHelp") + "\n")
		}
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *CanastaCuiPresenter) ActionLogOutput(g interfaces.CanastaGame) string {
	return actionLogOutputText(g)
}
