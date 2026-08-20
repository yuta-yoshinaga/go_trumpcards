package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// tonkBestDeadwood returns the lowest deadwood value the player can reach by
// discarding one card — the value that gates knocking (<= TonkKnockThreshold).

// tonkMinOpponentCards は相手のうち最も手札が少ない枚数を返す。相手がいなければ false。
//
// 人間を除くのは、自分の枚数はアンダーカットのリスクと関係ないから。
func tonkMinOpponentCards(g interfaces.TonkGame) (int, bool) {
	minCards, found := 0, false
	for i := range domain.TonkPlayerCnt {
		p := g.GetPlayer(i)
		if p == nil || p.GetIsHuman() {
			continue
		}
		if !found || p.GetCardsSize() < minCards {
			minCards, found = p.GetCardsSize(), true
		}
	}
	return minCards, found
}

// tonkPlayerStr returns the display string for a single Tonk player.
func tonkPlayerStr(player *domain.TonkPlayer, i int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("tonk.playerLine",
		"name", cuiPlayerName(player, i),
		"cum", strconv.Itoa(player.GetCumulativeScore()),
		"round", strconv.Itoa(player.GetRoundScore()),
		"cards", strconv.Itoa(player.GetCardsSize())) + "\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// tonkMeldIsSet reports whether a meld is a set (same rank) rather than a run.
func tonkMeldIsSet(meld []*domain.Card) bool {
	if len(meld) == 0 {
		return false
	}
	rank := meld[0].GetValue()
	for _, c := range meld {
		if c.GetValue() != rank {
			return false
		}
	}
	return true
}

// writeTonkKnockerMelds lists the knocker's melds with a set/run label.
func writeTonkKnockerMelds(b *strings.Builder, melds [][]*domain.Card) {
	if len(melds) == 0 {
		return
	}
	b.WriteString(color.Bold(i18n.T("tonk.knockerMeldsHeader")) + "\n")
	for i, meld := range melds {
		typeLabel := i18n.T("tonk.meldRun")
		if tonkMeldIsSet(meld) {
			typeLabel = i18n.T("tonk.meldSet")
		}
		b.WriteString(i18n.Tf("tonk.knockerMeldLine",
			"idx", strconv.Itoa(i+1),
			"type", typeLabel,
			"cards", cuiCardSliceStr(meld)) + "\n")
	}
}

// tonkHandCards returns a player's remaining cards as a slice.
func tonkHandCards(player *domain.TonkPlayer) []*domain.Card {
	cards := make([]*domain.Card, player.GetCardsSize())
	for i := range cards {
		cards[i] = player.GetCard(i)
	}
	return cards
}

// TonkCuiPresenter renders the Tonk CUI view.
type TonkCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *TonkCuiPresenter) Output(g interfaces.TonkGame, lastErr error) string {
	return buildCuiOutput(i18n.T("tonk.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("tonk.header",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"stock", strconv.Itoa(g.GetDrawPileCount())) + "\n")

		if top := g.GetDiscardTop(); top != nil {
			b.WriteString(i18n.Tf("tonk.discardLine", "card", cuiCardStr(top)) + "\n")
		}

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(tonkPlayerStr(g.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winnerIdx := g.GetWinnerIdx()
			banner := i18n.Tf("tonk.gameEnd",
				"name", cuiPlayerName(g.GetPlayer(winnerIdx), winnerIdx))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.TonkPhaseDraw:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("tonk.promptDraw",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("tonk.promptDrawHelpStock") + "\n")
			b.WriteString(i18n.T("tonk.promptDrawHelpDiscard") + "\n")
		case domain.TonkPhaseDiscard:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("tonk.promptDiscard",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			if cur := g.GetPlayer(currentIdx); cur.GetIsHuman() {
				best, _ := g.GetBestDeadwood(currentIdx)
				if best <= domain.TonkKnockThreshold {
					b.WriteString(color.Yellow(i18n.Tf("tonk.currentDeadwood", "value", strconv.Itoa(best))) +
						" " + color.Yellow(i18n.T("tonk.knockable")) + "\n")
				} else {
					b.WriteString(i18n.Tf("tonk.currentDeadwood", "value", strconv.Itoa(best)) +
						" " + i18n.T("tonk.knockUnable") + "\n")
				}
			}
			b.WriteString(i18n.T("tonk.promptDiscardHelp") + "\n")
			b.WriteString(i18n.T("tonk.promptKnockHelp") + "\n")
			// **相手の残りが少ないほどノックは裏目。**Web はボタンに警告リングと
			// ⚠️ を出しているのに、CUI は各行の枚数を見比べさせるだけだった (#5582)。
			// 人間の手番だけに出す。上のデッドウッド表示と同じ条件 ── ノックを
			// 決めるのは人間なので、CPU の捨て札中に警告しても行動できない。
			if cur := g.GetPlayer(currentIdx); cur.GetIsHuman() {
				if n, ok := tonkMinOpponentCards(g); ok && n <= domain.TonkUndercutRiskMax {
					b.WriteString(color.Yellow(i18n.Tf("tonk.knockUndercutWarning", "count", strconv.Itoa(n))) + "\n")
				}
			}
		case domain.TonkPhaseRoundEnd:
			if g.GetIsTonk() {
				b.WriteString(i18n.T("tonk.promptDealtTonk") + "\n")
			}
			// Reveal the knocker's melds and each CPU's remaining hand so the
			// round score has visible justification (parity with the web panel).
			writeTonkKnockerMelds(b, g.GetKnockerMelds())
			for i := 0; i < g.GetPlayerCnt(); i++ {
				cp := g.GetPlayer(i)
				if !cp.GetIsHuman() && cp.GetCardsSize() > 0 {
					b.WriteString(i18n.Tf("tonk.revealedHand",
						"name", cuiPlayerName(cp, i),
						"cards", cuiCardSliceStr(tonkHandCards(cp))) + "\n")
				}
			}
			b.WriteString(i18n.T("tonk.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("tonk.promptRoundEndHelp") + "\n")
		}
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *TonkCuiPresenter) ActionLogOutput(g interfaces.TonkGame) string {
	return actionLogOutputTextForSeats[*domain.TonkPlayer](g)
}
