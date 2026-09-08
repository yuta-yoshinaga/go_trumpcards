//go:build !js || !wasm || extra5

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// tongitsBestDeadwood returns the lowest deadwood value the player can reach by
// discarding one card — the value that gates knocking (<= TongitsKnockThreshold).

// tongitsMinOpponentCards は相手のうち最も手札が少ない枚数を返す。相手がいなければ false。
//
// 人間を除くのは、自分の枚数はアンダーカットのリスクと関係ないから。
func tongitsMinOpponentCards(g interfaces.TongitsGame) (int, bool) {
	minCards, found := 0, false
	for i := range domain.TongitsPlayerCnt {
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

// tongitsPlayerStr returns the display string for a single Tongits player.
func tongitsPlayerStr(player *domain.TongitsPlayer, i int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("tongits.playerLine",
		"name", cuiPlayerName(player, i),
		"cum", strconv.Itoa(player.GetCumulativeScore()),
		"round", strconv.Itoa(player.GetRoundScore()),
		"cards", strconv.Itoa(player.GetCardsSize())) + "\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// tongitsMeldIsSet reports whether a meld is a set (same rank) rather than a run.
func tongitsMeldIsSet(meld []*domain.Card) bool {
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

// writeTongitsKnockerMelds lists the knocker's melds with a set/run label.
func writeTongitsKnockerMelds(b *strings.Builder, melds [][]*domain.Card) {
	if len(melds) == 0 {
		return
	}
	b.WriteString(color.Bold(i18n.T("tongits.knockerMeldsHeader")) + "\n")
	for i, meld := range melds {
		typeLabel := i18n.T("tongits.meldRun")
		if tongitsMeldIsSet(meld) {
			typeLabel = i18n.T("tongits.meldSet")
		}
		b.WriteString(i18n.Tf("tongits.knockerMeldLine",
			"idx", strconv.Itoa(i+1),
			"type", typeLabel,
			"cards", cuiCardSliceStr(meld)) + "\n")
	}
}

// writeTongitsUndercutDetail lists the opponent's melds and both sides' deadwood.
//
// ラウンドの点差はアンダーカット判定（ノッカーと相手のデッドウッド比較）から
// 来るのに、出ているのはノッカーのメルドだけで、比較の相手側が見えなかった。
func writeTongitsUndercutDetail(b *strings.Builder, knockerDeadwood []*domain.Card, opponentMelds [][]*domain.Card, opponentDeadwood []*domain.Card) {
	if len(knockerDeadwood) > 0 {
		b.WriteString(i18n.Tf("tongits.knockerDeadwoodLine",
			"cards", cuiCardSliceStr(knockerDeadwood),
			"points", strconv.Itoa(domain.CalcDeadwoodValue(knockerDeadwood))) + "\n")
	}
	if len(opponentMelds) > 0 {
		b.WriteString(color.Bold(i18n.T("tongits.opponentMeldsHeader")) + "\n")
		for i, meld := range opponentMelds {
			typeLabel := i18n.T("tongits.meldRun")
			if tongitsMeldIsSet(meld) {
				typeLabel = i18n.T("tongits.meldSet")
			}
			b.WriteString(i18n.Tf("tongits.opponentMeldLine",
				"idx", strconv.Itoa(i+1),
				"type", typeLabel,
				"cards", cuiCardSliceStr(meld)) + "\n")
		}
	}
	if len(opponentDeadwood) > 0 {
		b.WriteString(i18n.Tf("tongits.opponentDeadwoodLine",
			"cards", cuiCardSliceStr(opponentDeadwood),
			"points", strconv.Itoa(domain.CalcDeadwoodValue(opponentDeadwood))) + "\n")
	}
}

// tongitsHandCards returns a player's remaining cards as a slice.
func tongitsHandCards(player *domain.TongitsPlayer) []*domain.Card {
	cards := make([]*domain.Card, player.GetCardsSize())
	for i := range cards {
		cards[i] = player.GetCard(i)
	}
	return cards
}

// TongitsCuiPresenter renders the Tongits CUI view.
type TongitsCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *TongitsCuiPresenter) Output(g interfaces.TongitsGame, lastErr error) string {
	return buildCuiOutput(i18n.T("tongits.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("tongits.header",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"stock", strconv.Itoa(g.GetDrawPileCount())) + "\n")

		if top := g.GetDiscardTop(); top != nil {
			b.WriteString(i18n.Tf("tongits.discardLine", "card", cuiCardStr(top)) + "\n")
		}

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(tongitsPlayerStr(g.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winnerIdx := g.GetWinnerIdx()
			banner := i18n.Tf("tongits.gameEnd",
				"name", cuiPlayerName(g.GetPlayer(winnerIdx), winnerIdx))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.TongitsPhaseDraw:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("tongits.promptDraw",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("tongits.promptDrawHelpStock") + "\n")
			b.WriteString(i18n.T("tongits.promptDrawHelpDiscard") + "\n")
		case domain.TongitsPhaseDiscard:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("tongits.promptDiscard",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			if cur := g.GetPlayer(currentIdx); cur.GetIsHuman() {
				best, _ := g.GetBestDeadwood(currentIdx)
				if best <= domain.TongitsKnockThreshold {
					b.WriteString(color.Yellow(i18n.Tf("tongits.currentDeadwood", "value", strconv.Itoa(best))) +
						" " + color.Yellow(i18n.T("tongits.knockable")) + "\n")
				} else {
					b.WriteString(i18n.Tf("tongits.currentDeadwood", "value", strconv.Itoa(best)) +
						" " + i18n.T("tongits.knockUnable") + "\n")
				}
			}
			b.WriteString(i18n.T("tongits.promptDiscardHelp") + "\n")
			b.WriteString(i18n.T("tongits.promptKnockHelp") + "\n")
			// **相手の残りが少ないほどノックは裏目。**Web はボタンに警告リングと
			// ⚠️ を出しているのに、CUI は各行の枚数を見比べさせるだけだった (#5582)。
			// 人間の手番だけに出す。上のデッドウッド表示と同じ条件 ── ノックを
			// 決めるのは人間なので、CPU の捨て札中に警告しても行動できない。
			if cur := g.GetPlayer(currentIdx); cur.GetIsHuman() {
				if n, ok := tongitsMinOpponentCards(g); ok && n <= domain.TongitsUndercutRiskMax {
					b.WriteString(color.Yellow(i18n.Tf("tongits.knockUndercutWarning", "count", strconv.Itoa(n))) + "\n")
				}
			}
		case domain.TongitsPhaseRoundEnd:
			if g.GetIsTongits() {
				b.WriteString(i18n.T("tongits.promptDealtTongits") + "\n")
			}
			// Reveal the knocker's melds and each CPU's remaining hand so the
			// round score has visible justification (parity with the web panel).
			writeTongitsKnockerMelds(b, g.GetKnockerMelds())
			writeTongitsUndercutDetail(b, g.GetKnockerDeadwood(), g.GetOpponentMelds(), g.GetOpponentDeadwood())
			for i := 0; i < g.GetPlayerCnt(); i++ {
				cp := g.GetPlayer(i)
				if !cp.GetIsHuman() && cp.GetCardsSize() > 0 {
					b.WriteString(i18n.Tf("tongits.revealedHand",
						"name", cuiPlayerName(cp, i),
						"cards", cuiCardSliceStr(tongitsHandCards(cp))) + "\n")
				}
			}
			b.WriteString(i18n.T("tongits.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("tongits.promptRoundEndHelp") + "\n")
		}
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *TongitsCuiPresenter) ActionLogOutput(g interfaces.TongitsGame) string {
	return actionLogOutputTextForSeats[*domain.TongitsPlayer](g)
}
