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
// A face-up cell is highlighted so the two flipped cards stand out — green when
// the current result is a match, yellow otherwise. Colour codes are added around
// the already-padded label, so the visible column width is unchanged.
func memoryCellStr(bc *domain.MemoryBoardCard, pos int, resultMatch bool, knownMatch bool) string {
	if bc.Taken {
		return fmt.Sprintf("[%2d]%-10s", pos, "")
	}
	if bc.FaceUp {
		label := fmt.Sprintf("%-10s", cuiCardStr(bc.Card))
		if resultMatch {
			label = color.Green(label)
		} else {
			label = color.Yellow(label)
		}
		return fmt.Sprintf("[%2d]%s", pos, label)
	}
	// Face-down: a seen cell that matches the one card face up is the play the
	// board is offering right now, so it is marked apart from the other seen
	// cells (*?) rather than left for the player to re-derive from memory.
	if knownMatch {
		return fmt.Sprintf("[%2d]%s", pos, color.Green(fmt.Sprintf("%-10s", "!?")))
	}
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
		board := m.GetBoard()

		// Progress: how many of the 26 pairs remain unmatched (taken pairs are the
		// sum of every player's captured pairs).
		matched := 0
		for i := 0; i < m.GetPlayerCnt(); i++ {
			matched += m.GetPlayer(i).GetPairCount()
		}
		// 盤面の長さはペア数設定で変わる (ADR-0035)。
		totalPairs := len(m.GetBoard()) / 2
		b.WriteString(i18n.Tf("memory.progressLine",
			"remaining", strconv.Itoa(totalPairs-matched),
			"total", strconv.Itoa(totalPairs)) + "\n")

		// Players
		for i := 0; i < m.GetPlayerCnt(); i++ {
			player := m.GetPlayer(i)
			b.WriteString(i18n.Tf("memory.playerLine",
				"name", cuiPlayerName(player, i),
				"pairs", strconv.Itoa(player.GetPairCount())) + "\n")
		}

		b.WriteString("----------\n")

		// 13 列で折り返して描画する。行数は盤面の長さから決めること: ペア数設定で
		// 52 枚未満になりうるため (ADR-0035)、4 行固定だと index out of range になる。
		// Web プレゼンターで同じ誤りを踏んでおり、こちらは CUI にペア数変更コマンドが
		// 無いので現状は到達しないが、同じ地雷を残す理由はない。
		const memoryCuiCols = 13
		resultMatch := m.GetPhase() == domain.MemoryPhaseResult && m.GetLastMatchResult()
		knownMatchIdx, hasKnownMatch := domain.MemoryKnownMatchIdx(board)
		for start := 0; start < len(board); start += memoryCuiCols {
			end := start + memoryCuiCols
			if end > len(board) {
				end = len(board)
			}
			rowParts := make([]string, 0, end-start)
			for pos := start; pos < end; pos++ {
				rowParts = append(rowParts, memoryCellStr(board[pos], pos, resultMatch, hasKnownMatch && pos == knownMatchIdx))
			}
			b.WriteString(strings.Join(rowParts, " "))
			b.WriteString("\n")
		}

		visited := 0
		for _, cell := range board {
			if cell.Visited {
				visited++
			}
		}
		b.WriteString(i18n.T("memory.visitedLegend") + "\n")
		b.WriteString(i18n.Tf("memory.visitedCountLegend", "count", strconv.Itoa(visited)) + "\n")
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
	return actionLogOutputTextForSeats[*domain.MemoryPlayer](m)
}
