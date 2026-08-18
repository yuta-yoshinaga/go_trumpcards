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

// DoudizhuCuiPresenter renders the Dou Dizhu CUI view.
type DoudizhuCuiPresenter struct{}

// Output renders the current game state.
func (p *DoudizhuCuiPresenter) Output(dg interfaces.DoudizhuGame, lastErr error) string {
	return buildCuiOutput(i18n.T("doudizhu.helpTitle"), func(b *strings.Builder) {
		phase := dg.GetPhase()

		if phase == domain.DoudizhuPhaseBid {
			b.WriteString(color.BoldYellow(i18n.T("doudizhu.phaseBid")) + "\n")
			fmt.Fprintf(b, "%s: %d\n", i18n.T("doudizhu.highestBid"), dg.GetHighestBid())
			// Name whose bid it is and, on the human's turn, how to respond —
			// play/end phases already carry this much context.
			curIdx := dg.GetCurrentTurn()
			b.WriteString(i18n.Tf("doudizhu.currentBidder",
				"name", cuiPlayerName(dg.GetPlayer(curIdx), curIdx)) + "\n")
			if dg.IsHumanTurn() {
				b.WriteString(i18n.T("doudizhu.promptBid") + "\n")
			}
		}

		for idx := 0; idx < dg.GetPlayerCnt(); idx++ {
			player := dg.GetPlayer(idx)
			b.WriteString(doudizhuPlayerStr(player, idx))
		}

		b.WriteString("----------\n")

		if phase == domain.DoudizhuPhasePlay || phase == domain.DoudizhuPhaseEnd {
			landlordIdx := dg.GetLandlordIdx()
			if landlordIdx >= 0 {
				// **名前は cuiPlayerName に通す。**ここだけ自前で組むと、
				// 日本語ロケールでも "Player 2" と出て他の行と食い違う (#5617)。
				fmt.Fprintf(b, "%s: %s\n", i18n.T("doudizhu.landlord"), cuiPlayerName(dg.GetPlayer(landlordIdx), landlordIdx))
				fmt.Fprintf(b, "%s: %s\n", i18n.T("doudizhu.kittyCards"), cuiCardListStr(doudizhuCardSlice(dg.GetKittyCards())))
			}

			if combo := dg.GetTableCombo(); combo != nil {
				fmt.Fprintf(b, "%s: %s (%s)\n", i18n.T("doudizhu.tableCards"),
					cuiCardListStr(doudizhuCardSlice(combo.Cards)),
					doudizhuComboName(combo.Type))
			} else {
				fmt.Fprintf(b, "%s: ---\n", i18n.T("doudizhu.tableCards"))
			}
		}

		if phase == domain.DoudizhuPhaseEnd {
			scores := dg.GetScores()
			b.WriteString(color.BoldYellow(i18n.T("doudizhu.gameEnd")) + "\n")
			for idx := 0; idx < dg.GetPlayerCnt(); idx++ {
				// Use the shared name helper so the end-score names match every
				// other phase instead of a locally hardcoded "CPU %d".
				name := cuiPlayerName(dg.GetPlayer(idx), idx)
				fmt.Fprintf(b, "%s: %d\n", name, scores[idx])
			}
		}

		if lastErr != nil {
			b.WriteString(i18n.MarkErrorLine(color.Red(lastErr.Error())) + "\n")
		}
	})
}

// ActionLogOutput 棋譜をCUI出力
func (p *DoudizhuCuiPresenter) ActionLogOutput(dg interfaces.DoudizhuGame) string {
	return actionLogOutputText(dg)
}

func doudizhuPlayerStr(player *domain.DoudizhuPlayer, idx int) string {
	var b strings.Builder
	b.WriteString(cuiPlayerName(player, idx))
	if player.GetIsLandlord() {
		b.WriteString(" [" + i18n.T("doudizhu.landlord") + "]")
	}
	if player.GetIsFinished() {
		b.WriteString(i18n.T("doudizhu.playerFinished") + "\n")
	} else {
		b.WriteString(i18n.Tf("doudizhu.playerCardCount", "count", strconv.Itoa(player.GetCardsSize())) + "\n")
		if player.GetIsHuman() {
			b.WriteString(cuiIndexedCardListStr(player) + "\n")
		}
	}
	return b.String()
}

// doudizhuCardSlice wraps []*Card to satisfy cuiCardList interface.
type doudizhuCardSlice []*domain.Card

func (s doudizhuCardSlice) GetCardsSize() int          { return len(s) }
func (s doudizhuCardSlice) GetCard(i int) *domain.Card { return s[i] }
