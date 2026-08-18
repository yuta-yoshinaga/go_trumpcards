//go:build !js || !wasm || extra2

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

// TichuCuiPresenter renders the Tichu CUI view.
type TichuCuiPresenter struct{}

// Output renders the current game state.
func (p *TichuCuiPresenter) Output(tg interfaces.TichuGame, lastErr error) string {
	return buildCuiOutput(i18n.T("tichu.helpTitle"), func(b *strings.Builder) {
		phase := tg.GetPhase()

		if phase == domain.TichuPhaseDeclare {
			b.WriteString(color.BoldYellow(i18n.T("tichu.phaseDeclare")) + "\n")
		}

		for idx := 0; idx < tg.GetPlayerCnt(); idx++ {
			// 印は自分の手札にだけ。他家の手札は伏せられている。
			var bomb []int
			if p := tg.GetPlayer(idx); p != nil && p.GetIsHuman() {
				cards := make([]*domain.Card, p.GetCardsSize())
				for i := range cards {
					cards[i] = p.GetCard(i)
				}
				bomb = domain.TichuBombIndices(cards)
			}
			b.WriteString(tichuPlayerStr(tg.GetPlayer(idx), idx, bomb))
		}

		b.WriteString("----------\n")

		if phase == domain.TichuPhasePlay || phase == domain.TichuPhaseEnd {
			if combo := tg.GetTableCombo(); combo != nil {
				fmt.Fprintf(b, "%s: %s (%s)\n", i18n.T("tichu.tableCards"),
					cuiCardListStr(tichuCardSlice(combo.Cards)), tichuComboName(combo.Type))
			} else {
				fmt.Fprintf(b, "%s: ---\n", i18n.T("tichu.tableCards"))
			}
		}

		// **得点差もボムの使用状況も終局まで分からなかった。**Web は常時
		// スコアバーに出している (#4888)。終局時は下でもう一度出すので、
		// ここは進行中だけ。
		if phase != domain.TichuPhaseEnd {
			scores := tg.GetScores()
			fmt.Fprintf(b, "%s (P0/P2): %d  %s (P1/P3): %d\n",
				i18n.T("tichu.teamA"), scores[0], i18n.T("tichu.teamB"), scores[1])
			if n := tg.GetBombCount(); n > 0 {
				b.WriteString(i18n.Tf("tichu.bombCount", "count", strconv.Itoa(n)) + "\n")
			}
			if tg.GetIsOneTwo() {
				b.WriteString(color.BoldYellow(i18n.T("tichu.oneTwo")) + "\n")
			}
		}

		if phase == domain.TichuPhaseEnd {
			scores := tg.GetScores()
			b.WriteString(color.BoldYellow(i18n.T("tichu.gameEnd")) + "\n")
			fmt.Fprintf(b, "%s (P0/P2): %d\n", i18n.T("tichu.teamA"), scores[0])
			fmt.Fprintf(b, "%s (P1/P3): %d\n", i18n.T("tichu.teamB"), scores[1])
		}

		if lastErr != nil {
			b.WriteString(i18n.MarkErrorLine(color.Red(lastErr.Error())) + "\n")
		}
	})
}

// ActionLogOutput 棋譜をCUI出力
func (p *TichuCuiPresenter) ActionLogOutput(tg interfaces.TichuGame) string {
	return actionLogOutputText(tg)
}

func tichuPlayerStr(player *domain.TichuPlayer, idx int, bomb []int) string {
	if player == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString(cuiPlayerName(player, idx))
	fmt.Fprintf(&b, " [%s %d]", i18n.T("tichu.team"), domain.TichuTeamOf(idx))
	switch player.GetDeclType() {
	case domain.TichuDeclTichu:
		b.WriteString(" [" + i18n.T("tichu.tichu") + "]")
	case domain.TichuDeclGrand:
		b.WriteString(" [" + i18n.T("tichu.grandTichu") + "]")
	}
	if player.GetIsFinished() {
		b.WriteString(" " + i18n.Tf("tichu.playerRank", "rank", strconv.Itoa(player.GetRank())) + "\n")
	} else {
		b.WriteString(i18n.Tf("tichu.playerCardCount", "count", strconv.Itoa(player.GetCardsSize())) + "\n")
		if player.GetIsHuman() {
			// **ボムは得点を左右する。**Web は赤いリングと💣で示しているのに、
			// CUI は手札を目視で数えるしかなかった (#5635)。
			b.WriteString(cuiIndexMarkedCardListStr(player, bomb, CuiBombMark) + "\n")
		}
	}
	return b.String()
}

// tichuCardSlice wraps []*Card to satisfy the cuiCardList interface.
type tichuCardSlice []*domain.Card

func (s tichuCardSlice) GetCardsSize() int          { return len(s) }
func (s tichuCardSlice) GetCard(i int) *domain.Card { return s[i] }
