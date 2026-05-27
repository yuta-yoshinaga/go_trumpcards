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

// SixCardGolfCuiPresenter SixCardGolf CUIプレゼンター
type SixCardGolfCuiPresenter struct{}

// Output ゲーム状態をCUI出力
func (p *SixCardGolfCuiPresenter) Output(g interfaces.SixCardGolfGame, lastErr error) string {
	return buildCuiOutput(i18n.T("sixcardgolf.helpTitle"), func(b *strings.Builder) {
		cfg := g.GetConfig()
		b.WriteString(i18n.Tf("sixcardgolf.header",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"total", strconv.Itoa(cfg.Rounds),
			"stock", strconv.Itoa(g.GetDrawPileCount())) + "\n")

		if top := g.GetDiscardTop(); top != nil {
			b.WriteString(i18n.Tf("sixcardgolf.discardLine", "card", cuiCardStr(top)) + "\n")
		}

		phase := g.GetPhase()
		revealAll := phase == domain.SixCardGolfPhaseRoundOver || phase == domain.SixCardGolfPhaseGameOver
		for i := 0; i < g.GetPlayerCnt(); i++ {
			player := g.GetPlayer(i)
			b.WriteString(sixCardGolfPlayerStr(player, i, g.GetCurrentPlayerIdx() == i, revealAll, g) + "\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winnerIdx := g.GetWinnerIdx()
			banner := i18n.Tf("sixcardgolf.gameEnd",
				"name", scgPlayerName(g.GetPlayer(winnerIdx), winnerIdx))
			b.WriteString(color.Green(banner) + "\n")
			return
		}

		switch phase {
		case domain.SixCardGolfPhaseSetup:
			b.WriteString(color.Green(i18n.T("sixcardgolf.promptSetup")) + "\n")
		case domain.SixCardGolfPhasePlayerTurn:
			if g.GetCanFlip() {
				b.WriteString(color.Green(i18n.T("sixcardgolf.promptFlip")) + "\n")
			} else {
				b.WriteString(i18n.T("sixcardgolf.promptDraw") + "\n")
			}
		case domain.SixCardGolfPhaseDrawPending:
			drawn := g.GetDrawnCard()
			if drawn != nil {
				b.WriteString(i18n.Tf("sixcardgolf.drawnCard", "card", cuiCardStr(drawn)) + "\n")
			}
			b.WriteString(color.Green(i18n.T("sixcardgolf.promptSwapOrDiscard")) + "\n")
		case domain.SixCardGolfPhaseRoundOver:
			b.WriteString(color.Green(i18n.T("sixcardgolf.promptNextRound")) + "\n")
		}
	})
}

// sixCardGolfPlayerStr プレイヤー表示文字列
func sixCardGolfPlayerStr(player *domain.SixCardGolfPlayer, idx int, isCurrent, revealAll bool, g interfaces.SixCardGolfGame) string {
	var b strings.Builder
	name := scgPlayerName(player, idx)
	marker := ""
	if isCurrent {
		marker = " <<"
	}
	b.WriteString(fmt.Sprintf("%s (累計=%d, R=%d)%s\n", name, player.CumulativeScore, player.RoundScore, marker))

	for row := 0; row < 2; row++ {
		b.WriteString("  ")
		for col := 0; col < 3; col++ {
			pos := row*3 + col
			slot := player.Grid[pos]
			if col > 0 {
				b.WriteString(" ")
			}
			b.WriteString(fmt.Sprintf("[%d]", pos))
			if slot.FaceUp || revealAll {
				b.WriteString(cuiCardStr(slot.Card))
			} else {
				b.WriteString("??")
			}
		}
		b.WriteString("\n")
	}

	if revealAll {
		score := g.ScorePlayer(idx)
		b.WriteString(fmt.Sprintf("  Score: %d\n", score))
	}
	return b.String()
}

// scgPlayerName プレイヤー名
func scgPlayerName(player *domain.SixCardGolfPlayer, idx int) string {
	if player == nil {
		return fmt.Sprintf("Player%d", idx)
	}
	if !player.IsCpu {
		return i18n.T("cuiPlayerNameHuman")
	}
	return i18n.Tf("cuiPlayerNameCPU", "id", strconv.Itoa(idx))
}

// ActionLogOutput 棋譜
func (p *SixCardGolfCuiPresenter) ActionLogOutput(g interfaces.SixCardGolfGame) string {
	return actionLogOutputText(g)
}
