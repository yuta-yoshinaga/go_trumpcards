package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// FiftyOneCuiPresenter フィフティワンCUIプレゼンタークラス
type FiftyOneCuiPresenter struct{}

// Output ゲーム状態を文字列出力
func (p *FiftyOneCuiPresenter) Output(fo interfaces.FiftyOneGame, lastErr error) string {
	return buildCuiOutput("Fifty-one (フィフティワン)", func(b *strings.Builder) {
		// 各プレイヤー情報
		for i := 0; i < fo.GetPlayerCnt(); i++ {
			player := fo.GetPlayer(i)
			b.WriteString(cuiPlayerName(player, i))
			score := player.BestSuitScore()
			if player.GetIsHuman() {
				fmt.Fprintf(b, " (スコア: %d)\n", score)
				b.WriteString(cuiIndexedCardListStr(player))
				b.WriteString("\n")
			} else {
				fmt.Fprintf(b, ": %d枚 (スコア: %d)\n", player.GetCardsSize(), score)
			}
		}

		// 場札
		b.WriteString("----------\n")
		b.WriteString("場札: ")
		tableCards := fo.GetTableCards()
		for i, c := range tableCards {
			if i > 0 {
				b.WriteString("  ")
			}
			fmt.Fprintf(b, "[%d]%s", i, cuiCardStr(c))
		}
		b.WriteString("\n")

		// ストップ状態
		if fo.GetStopCallerIdx() >= 0 {
			callerName := cuiPlayerName(fo.GetPlayer(fo.GetStopCallerIdx()), fo.GetStopCallerIdx())
			fmt.Fprintf(b, "⚠ %s がストップ宣言！\n", callerName)
		}

		b.WriteString("----------\n")

		if lastErr != nil {
			b.WriteString("Error: " + lastErr.Error() + "\n")
		}

		if fo.GetGameEndFlag() {
			b.WriteString("=== ゲーム終了 ===\n")
			for i := 0; i < fo.GetPlayerCnt(); i++ {
				name := cuiPlayerName(fo.GetPlayer(i), i)
				fmt.Fprintf(b, "  %s: %d点\n", name, fo.GetPlayer(i).BestSuitScore())
			}
			winnerIdx := fo.GetWinnerIdx()
			winner := fo.GetPlayer(winnerIdx)
			if winner != nil && winner.GetIsHuman() {
				b.WriteString("あなたの勝ち！\n")
			} else {
				fmt.Fprintf(b, "CPU %dの勝ち！\n", winnerIdx)
			}
		} else if fo.IsHumanTurn() {
			b.WriteString("あなたのターン: p <手札番号> <場札番号> / a (全交換) / stop\n")
		}
	})
}

// ActionLogOutput 棋譜を文字列出力
func (p *FiftyOneCuiPresenter) ActionLogOutput(fo interfaces.FiftyOneGame) string {
	return actionLogOutputText(fo)
}
