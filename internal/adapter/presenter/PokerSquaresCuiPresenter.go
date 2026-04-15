package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// PokerSquaresCuiPresenter はポーカー・スクエアズの CUI プレゼンター。
type PokerSquaresCuiPresenter struct{}

// Output はゲーム状態を文字列で出力する。
func (pr *PokerSquaresCuiPresenter) Output(p interfaces.PokerSquaresGame, lastErr error) string {
	return buildCuiOutput("Poker Squares (ポーカー・スクエアズ)", func(b *strings.Builder) {
		board := p.GetBoard()
		for r := range domain.PokerSquaresGridSize {
			for c := range domain.PokerSquaresGridSize {
				if c > 0 {
					b.WriteString(" | ")
				}
				card := board[r][c]
				if card == nil {
					fmt.Fprintf(b, "(%d,%d) .....", r, c)
				} else {
					fmt.Fprintf(b, "(%d,%d) %s", r, c, cuiCardStr(card))
				}
			}
			fmt.Fprintf(b, "   => row score: %d\n", p.RowScore(r))
		}
		b.WriteString("----------\n")
		colParts := make([]string, domain.PokerSquaresGridSize)
		for i := range domain.PokerSquaresGridSize {
			colParts[i] = fmt.Sprintf("col%d=%d", i, p.ColScore(i))
		}
		fmt.Fprintf(b, "%s\n", strings.Join(colParts, " "))
		b.WriteString("----------\n")

		if cc := p.GetCurrentCard(); cc != nil {
			fmt.Fprintf(b, "Current card: %s\n", cuiCardStr(cc))
		} else {
			b.WriteString("Current card: (none)\n")
		}
		fmt.Fprintf(b, "Placed: %d/%d  Total score: %d\n",
			p.GetPlacedCount(), domain.PokerSquaresTotalCells, p.TotalScore())

		if lastErr != nil {
			fmt.Fprintf(b, "%s\n", color.Red(lastErr.Error()))
		}
		switch p.GetPhase() {
		case domain.PokerSquaresPhasePlaying:
			// nothing extra
		case domain.PokerSquaresPhaseComplete:
			fmt.Fprintf(b, "%s 合計得点: %d\n", color.Green("ゲーム終了"), p.TotalScore())
		}
	})
}

// ActionLogOutput は棋譜をテキスト出力する。
func (pr *PokerSquaresCuiPresenter) ActionLogOutput(p interfaces.PokerSquaresGame) string {
	if p.GetPhase() == domain.PokerSquaresPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(p.GetActionLog())
}
