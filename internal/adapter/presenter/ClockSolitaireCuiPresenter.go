package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// ClockSolitaireCuiPresenter クロックソリティアCUIプレゼンタークラス
type ClockSolitaireCuiPresenter struct{}

// Output ゲーム状態を文字列出力
func (pr *ClockSolitaireCuiPresenter) Output(g interfaces.ClockSolitaireGame, lastErr error) string {
	return buildCuiOutput("Clock Solitaire (クロックソリティア)", func(b *strings.Builder) {
		piles := g.GetPiles()
		fuc := g.GetFaceUpCount()

		// 時計の位置: 12時、1時、2時、... 11時（表示順）
		displayOrder := []int{11, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		labels := []string{"12", " 1", " 2", " 3", " 4", " 5", " 6", " 7", " 8", " 9", "10", "11"}

		for i, pileIdx := range displayOrder {
			pile := piles[pileIdx]
			label := labels[i]
			fmt.Fprintf(b, "[%s時] ", label)
			for _, pc := range pile[:domain.ClockSolitaireCardsPerPile] {
				if pc.FaceUp {
					fmt.Fprintf(b, "%s ", cuiCardStr(pc.Card))
				} else {
					b.WriteString("?? ")
				}
			}
			fmt.Fprintf(b, "(%d/%d)\n", fuc[pileIdx], domain.ClockSolitaireCardsPerPile)
		}

		// 中央パイル（K）
		b.WriteString("----------\n")
		centerPile := piles[domain.ClockSolitaireKingPileIdx]
		b.WriteString("[中央K] ")
		for _, pc := range centerPile[:domain.ClockSolitaireCardsPerPile] {
			if pc.FaceUp {
				fmt.Fprintf(b, "%s ", cuiCardStr(pc.Card))
			} else {
				b.WriteString("?? ")
			}
		}
		fmt.Fprintf(b, "(%d/%d)\n", fuc[domain.ClockSolitaireKingPileIdx], domain.ClockSolitaireCardsPerPile)

		b.WriteString("----------\n")

		// 現在のカード
		if cc := g.GetCurrentCard(); cc != nil {
			fmt.Fprintf(b, "手持ち: %s\n", cuiCardStr(cc))
		}

		// エラーメッセージ
		if lastErr != nil {
			fmt.Fprintf(b, "%s\n", color.Red(lastErr.Error()))
		}

		// ゲーム状態
		phase := g.GetPhase()
		switch phase {
		case domain.ClockSolitairePhasePlaying:
			fmt.Fprintf(b, "ステップ: %d\n", g.GetStepCount())
		case domain.ClockSolitairePhaseGameClear:
			fmt.Fprintf(b, "ゲームクリア！ ステップ数: %d\n", g.GetStepCount())
		case domain.ClockSolitairePhaseGameOver:
			fmt.Fprintf(b, "ゲームオーバー ステップ数: %d\n", g.GetStepCount())
		}
	})
}

// ActionLogOutput 棋譜を文字列出力
func (pr *ClockSolitaireCuiPresenter) ActionLogOutput(g interfaces.ClockSolitaireGame) string {
	return buildCuiOutput("Clock Solitaire Action Log", func(b *strings.Builder) {
		for _, entry := range g.GetActionLog() {
			fmt.Fprintf(b, "[%d] %s\n", entry.TurnNumber, entry.Detail)
		}
	})
}
