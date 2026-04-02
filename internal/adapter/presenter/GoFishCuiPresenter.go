package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// goFishPlayerStr returns the display string for a single GoFish player.
func goFishPlayerStr(player *domain.GoFishPlayer, i int) string {
	var b strings.Builder
	b.WriteString(cuiPlayerName(player, i))
	fmt.Fprintf(&b, ": %d枚, ブック: %d\n", player.GetCardsSize(), player.GetBookCount())
	if player.GetIsHuman() {
		b.WriteString(cuiIndexedCardListStr(player))
		b.WriteString("\n")
	}
	return b.String()
}

// GoFishCuiPresenter Go FishCUIプレゼンタークラス
type GoFishCuiPresenter struct{}

// Output ゲーム状態を文字列出力
func (p *GoFishCuiPresenter) Output(gf interfaces.GoFishGame, lastErr error) string {
	return buildCuiOutput("Go Fish (ゴーフィッシュ)", func(b *strings.Builder) {
		for i := 0; i < gf.GetPlayerCnt(); i++ {
			b.WriteString(goFishPlayerStr(gf.GetPlayer(i), i))
		}

		b.WriteString("----------\n")
		fmt.Fprintf(b, "山札: %d枚\n", gf.GetDeckRemaining())

		// 最後の要求結果
		if gf.GetLastAskPlayerIdx() >= 0 {
			askerName := cuiPlayerName(gf.GetPlayer(gf.GetLastAskPlayerIdx()), gf.GetLastAskPlayerIdx())
			targetName := cuiPlayerName(gf.GetPlayer(gf.GetLastAskTargetIdx()), gf.GetLastAskTargetIdx())
			if gf.GetLastAskSuccess() {
				fmt.Fprintf(b, "%s が %s にランク %d を要求 → %d枚もらった！\n",
					askerName, targetName, gf.GetLastAskRank(), len(gf.GetLastCardsReceived()))
			} else {
				fmt.Fprintf(b, "%s が %s にランク %d を要求 → Go Fish!\n",
					askerName, targetName, gf.GetLastAskRank())
			}
			if gf.GetLastBookFormed() {
				fmt.Fprintf(b, "ブック完成！ ランク %d\n", gf.GetLastBookRank())
			}
		}

		// CPU行動履歴
		for _, action := range gf.GetCpuActions() {
			askerName := cuiPlayerName(gf.GetPlayer(action.AskPlayerIdx), action.AskPlayerIdx)
			targetName := cuiPlayerName(gf.GetPlayer(action.AskTargetIdx), action.AskTargetIdx)
			if action.Success {
				fmt.Fprintf(b, "[CPU] %s → %s: ランク %d → %d枚もらった\n",
					askerName, targetName, action.AskRank, action.CardsReceived)
			} else {
				fmt.Fprintf(b, "[CPU] %s → %s: ランク %d → Go Fish!\n",
					askerName, targetName, action.AskRank)
			}
			if action.BookFormed {
				fmt.Fprintf(b, "[CPU] %s: ブック完成！ ランク %d\n", askerName, action.BookRank)
			}
		}

		b.WriteString("----------\n")

		if lastErr != nil {
			b.WriteString("Error: " + lastErr.Error() + "\n")
		}

		if gf.GetGameEndFlag() {
			winnerIdx := gf.GetWinnerIdx()
			winner := gf.GetPlayer(winnerIdx)
			if winner != nil && winner.GetIsHuman() {
				b.WriteString("ゲーム終了！ あなたの勝ち！\n")
			} else {
				fmt.Fprintf(b, "ゲーム終了！ CPU %dの勝ち！\n", winnerIdx)
			}
			for i := 0; i < gf.GetPlayerCnt(); i++ {
				name := cuiPlayerName(gf.GetPlayer(i), i)
				fmt.Fprintf(b, "  %s: %dブック\n", name, gf.GetPlayer(i).GetBookCount())
			}
		} else {
			turnName := cuiPlayerName(gf.GetPlayer(gf.GetCurrentTurn()), gf.GetCurrentTurn())
			fmt.Fprintf(b, "%sのターン\n", turnName)
			if gf.IsHumanTurn() {
				b.WriteString("ask <相手番号> <ランク> で要求\n")
			}
		}
	})
}

// ActionLogOutput 棋譜を文字列出力
func (p *GoFishCuiPresenter) ActionLogOutput(gf interfaces.GoFishGame) string {
	return actionLogOutputText(gf)
}
