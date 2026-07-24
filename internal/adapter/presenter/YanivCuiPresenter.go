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

// yanivReveal ラウンド終了/ゲーム終了時に全員の手札を公開するか
func yanivReveal(g interfaces.YanivGame) bool {
	phase := g.GetPhase()
	return phase == domain.YanivPhaseRoundEnd || phase == domain.YanivPhaseGameEnd
}

// yanivPlayerStr returns the display string for a single Yaniv player.
func yanivPlayerStr(g interfaces.YanivGame, player *domain.YanivPlayer, i int) string {
	var b strings.Builder
	reveal := yanivReveal(g)
	total := "?"
	if player.GetIsHuman() || reveal {
		total = strconv.Itoa(player.HandTotal())
	}
	status := strconv.Itoa(player.GetScore())
	if player.IsEliminated() {
		status = "OUT"
	} else if limit := g.GetConfig().ScoreLimit; limit > 0 && player.GetScore()*100 > limit*80 {
		// Within 20% of the elimination threshold → flag the impending OUT.
		status = color.Yellow(status + i18n.T("yaniv.nearOut"))
	}
	b.WriteString(i18n.Tf("yaniv.playerLine",
		"name", cuiPlayerName(player, i),
		"score", status,
		"cards", strconv.Itoa(player.GetCardsSize()),
		"total", total) + "\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// YanivCuiPresenter renders the Yaniv CUI view.
type YanivCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *YanivCuiPresenter) Output(g interfaces.YanivGame, lastErr error) string {
	return buildCuiOutput(i18n.T("yaniv.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("yaniv.header",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"stock", strconv.Itoa(g.GetDrawPileCount())) + "\n")
		b.WriteString(i18n.Tf("yaniv.limitLine",
			"limit", strconv.Itoa(g.GetConfig().ScoreLimit)) + "\n")

		if pickup := g.GetPickupCards(); len(pickup) > 0 {
			cards := make([]string, 0, len(pickup))
			for _, c := range pickup {
				cards = append(cards, cuiCardStr(c))
			}
			b.WriteString(i18n.Tf("yaniv.pickupLine", "cards", strings.Join(cards, " ")) + "\n")
		}

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(yanivPlayerStr(g, g.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			banner := i18n.Tf("yaniv.gameEnd",
				"name", cuiPlayerName(g.GetPlayer(g.GetWinnerIdx()), g.GetWinnerIdx()))
			b.WriteString(color.Green(banner) + "\n")
			return
		}

		idx := g.GetCurrentPlayerIdx()
		switch g.GetPhase() {
		case domain.YanivPhaseDiscard:
			b.WriteString(i18n.Tf("yaniv.promptDiscard", "name", cuiPlayerName(g.GetPlayer(idx), idx)) + "\n")
			b.WriteString(i18n.T("yaniv.promptDiscardHelp") + "\n")
			if g.GetPlayer(idx).HandTotal() <= domain.YanivCallThreshold {
				b.WriteString(i18n.T("yaniv.promptYanivHelp") + "\n")
			}
		case domain.YanivPhaseDraw:
			b.WriteString(i18n.Tf("yaniv.promptDraw", "name", cuiPlayerName(g.GetPlayer(idx), idx)) + "\n")
			b.WriteString(i18n.T("yaniv.promptDrawHelpStock") + "\n")
			b.WriteString(i18n.T("yaniv.promptDrawHelpPickup") + "\n")
		case domain.YanivPhaseRoundEnd:
			if caller := g.GetCallerIdx(); caller >= 0 {
				key := "yaniv.promptYanivResult"
				if g.GetIsAsaf() {
					key = "yaniv.promptAsafResult"
				}
				b.WriteString(i18n.Tf(key, "name", cuiPlayerName(g.GetPlayer(caller), caller)) + "\n")
			}
			b.WriteString(i18n.T("yaniv.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("yaniv.promptRoundEndHelp") + "\n")
		}
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *YanivCuiPresenter) ActionLogOutput(g interfaces.YanivGame) string {
	return actionLogOutputText(g)
}
