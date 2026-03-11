package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// memoryCellStr returns the display string for a single Memory board cell.
func memoryCellStr(bc *domain.MemoryBoardCard, pos int) string {
	if bc.Taken {
		return fmt.Sprintf("[%2d]%-10s", pos, "")
	}
	if bc.FaceUp {
		return fmt.Sprintf("[%2d]%-10s", pos, cuiCardStr(bc.Card))
	}
	return fmt.Sprintf("[%2d]%-10s", pos, "??")
}

// MemoryCuiPresenter 神経衰弱CUIプレゼンタークラス
type MemoryCuiPresenter struct{}

// Output ゲーム状態を文字列出力
func (p *MemoryCuiPresenter) Output(m interfaces.MemoryGame, lastErr error) string {
	return buildCuiOutput("Memory (神経衰弱)", func(b *strings.Builder) {
		// プレイヤー情報
		for i := 0; i < m.GetPlayerCnt(); i++ {
			player := m.GetPlayer(i)
			name := cuiPlayerName(player, i)
			fmt.Fprintf(b, "%s: %dペア\n", name, player.GetPairCount())
		}

		b.WriteString("----------\n")

		// ボードを4×13グリッドで表示
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

		b.WriteString("----------\n")

		// エラーメッセージ
		if lastErr != nil {
			fmt.Fprintf(b, "%s\n", lastErr.Error())
		}

		// ゲーム状態
		if m.GetGameEndFlag() {
			winnerIdx := m.GetWinnerIdx()
			player := m.GetPlayer(winnerIdx)
			fmt.Fprintf(b, "ゲーム終了！ %sの勝利です！\n", cuiPlayerName(player, winnerIdx))
		} else {
			phase := m.GetPhase()
			currentIdx := m.GetCurrentPlayerIdx()
			player := m.GetPlayer(currentIdx)
			playerStr := cuiPlayerName(player, currentIdx)
			switch phase {
			case domain.MemoryPhaseFlip1:
				fmt.Fprintf(b, "手番: %s — 1枚目を選んでください\n", playerStr)
				b.WriteString("f <pos>・・・カードをめくる\n")
			case domain.MemoryPhaseFlip2:
				fmt.Fprintf(b, "手番: %s — 2枚目を選んでください\n", playerStr)
				b.WriteString("f <pos>・・・カードをめくる\n")
			case domain.MemoryPhaseResult:
				if m.GetLastMatchResult() {
					b.WriteString("ペアが揃いました！\n")
				} else {
					b.WriteString("残念、不一致です。\n")
				}
				b.WriteString("n・・・次へ\n")
			}
		}
	})
}

// ActionLogOutput 棋譜をテキスト出力
func (p *MemoryCuiPresenter) ActionLogOutput(m interfaces.MemoryGame) string {
	return actionLogOutputText(m)
}
