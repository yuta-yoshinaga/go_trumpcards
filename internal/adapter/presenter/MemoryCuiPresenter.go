//go:build !js || !wasm || solo

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

// memoryCellStr returns the display string for a single Memory board cell.
func memoryCellStr(bc *domain.MemoryBoardCard, pos int) string {
	if bc.Taken {
		return fmt.Sprintf("[%2d]%-10s", pos, "")
	}
	if bc.FaceUp {
		return fmt.Sprintf("[%2d]%-10s", pos, cuiCardStr(bc.Card))
	}
	// Face-down: distinguish previously-seen cells (*?) from unseen ones (??).
	if bc.Visited {
		return fmt.Sprintf("[%2d]%-10s", pos, "*?")
	}
	return fmt.Sprintf("[%2d]%-10s", pos, "??")
}

// MemoryCuiPresenter renders the Memory (Concentration) CUI view.
type MemoryCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *MemoryCuiPresenter) Output(m interfaces.MemoryGame, lastErr error) string {
	return buildCuiOutput(i18n.T("memory.helpTitle"), func(b *strings.Builder) {
		// Players
		for i := 0; i < m.GetPlayerCnt(); i++ {
			player := m.GetPlayer(i)
			b.WriteString(i18n.Tf("memory.playerLine",
				"name", cuiPlayerName(player, i),
				"pairs", strconv.Itoa(player.GetPairCount())) + "\n")
		}

		b.WriteString("----------\n")

		// Render the 4×13 board
		board := m.GetBoard()
		for row := 0; row < 4; row++ {
			rowParts := make([]string, 13)
			for col := 0; col < 13; col++ {
				pos := row*13 + col
				rowParts[col] = memoryCellStr(board[pos], pos)
			}
			b.WriteString(strings.Join(rowParts, " "))
			b.WriteString("\n")
		}

		b.WriteString(i18n.T("memory.visitedLegend") + "\n")
		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if m.GetGameEndFlag() {
			winnerIdx := m.GetWinnerIdx()
			banner := i18n.Tf("memory.gameEnd",
				"name", cuiPlayerName(m.GetPlayer(winnerIdx), winnerIdx))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		currentIdx := m.GetCurrentPlayerIdx()
		name := cuiPlayerName(m.GetPlayer(currentIdx), currentIdx)
		switch m.GetPhase() {
		case domain.MemoryPhaseFlip1:
			b.WriteString(i18n.Tf("memory.promptFlip1", "name", name) + "\n")
			b.WriteString(i18n.T("memory.promptFlipHelp") + "\n")
		case domain.MemoryPhaseFlip2:
			b.WriteString(i18n.Tf("memory.promptFlip2", "name", name) + "\n")
			b.WriteString(i18n.T("memory.promptFlipHelp") + "\n")
		case domain.MemoryPhaseResult:
			if m.GetLastMatchResult() {
				b.WriteString(color.Green(i18n.T("memory.resultMatch")) + "\n")
			} else {
				b.WriteString(color.Yellow(i18n.T("memory.resultMismatch")) + "\n")
			}
			b.WriteString(i18n.T("memory.promptNextHelp") + "\n")
		}
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *MemoryCuiPresenter) ActionLogOutput(m interfaces.MemoryGame) string {
	return actionLogOutputText(m)
}
