package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// shitheadPlayerStr returns the display string for a single Shithead player.
func shitheadPlayerStr(player *domain.ShitheadPlayer, idx int, currentTurn int) string {
	var b strings.Builder
	name := cuiPlayerName(player, idx)
	b.WriteString(name)
	if idx == currentTurn {
		b.WriteString(" ← turn")
	}
	if player.GetIsFinished() {
		fmt.Fprintf(&b, " [上がり rank=%d]", player.GetRank())
		b.WriteString("\n")
		return b.String()
	}
	fmt.Fprintf(&b, ": 手札%d / 表%d / 裏%d\n",
		player.GetCardsSize(), player.GetFaceUpSize(), player.GetFaceDownSize())
	if player.GetIsHuman() {
		if player.GetCardsSize() > 0 {
			b.WriteString("  hand:    ")
			b.WriteString(cuiIndexedCardListStr(player))
			b.WriteString("\n")
		}
		if player.GetFaceUpSize() > 0 {
			b.WriteString("  faceup:  ")
			b.WriteString(cuiCardSliceStr(player.GetFaceUpCards()))
			b.WriteString("\n")
		}
	} else if player.GetFaceUpSize() > 0 {
		b.WriteString("  faceup:  ")
		b.WriteString(cuiCardSliceStr(player.GetFaceUpCards()))
		b.WriteString("\n")
	}
	return b.String()
}

// ShitheadCuiPresenter シットヘッドCUIプレゼンタークラス
type ShitheadCuiPresenter struct{}

// Output ゲーム状態を文字列出力
func (p *ShitheadCuiPresenter) Output(sg interfaces.ShitheadGame, lastErr error) string {
	return buildCuiOutput("Shithead (シットヘッド / カーマ)", func(b *strings.Builder) {
		currentTurn := sg.GetCurrentTurn()
		for i := 0; i < sg.GetPlayerCnt(); i++ {
			b.WriteString(shitheadPlayerStr(sg.GetPlayer(i), i, currentTurn))
		}

		b.WriteString("----------\n")

		// 場札
		discard := sg.GetDiscardPile()
		if len(discard) > 0 {
			fmt.Fprintf(b, "場札: %s\n", cuiCardSliceStr(discard))
		} else {
			b.WriteString("場札: なし\n")
		}
		fmt.Fprintf(b, "山札: %d 枚\n", sg.GetStockSize())
		if sg.GetSevenActive() {
			b.WriteString(color.BoldYellow("【7発動中】次は7以下のカードしか出せません\n"))
		}
		if sg.GetSkipNext() {
			b.WriteString(color.BoldYellow("【次プレイヤースキップ】\n"))
		}

		// 人間の前の行動
		if humanAction := sg.GetHumanAction(); humanAction != nil {
			b.WriteString(formatShitheadAction(sg, humanAction))
		}

		// CPUの行動履歴
		cpuActions := sg.GetCpuActions()
		if len(cpuActions) > 0 {
			b.WriteString(color.Bold("[CPUの行動]\n"))
			for _, action := range cpuActions {
				b.WriteString(formatShitheadAction(sg, action))
			}
		}

		if lastErr != nil {
			fmt.Fprintf(b, "%s\n", color.Red(lastErr.Error()))
		}

		if sg.GetGameEndFlag() {
			b.WriteString("ゲーム終了！\n")
			for i := 0; i < sg.GetPlayerCnt(); i++ {
				player := sg.GetPlayer(i)
				rank := player.GetRank()
				suffix := ""
				if rank == sg.GetPlayerCnt() {
					suffix = " (Shithead!)"
				}
				fmt.Fprintf(b, "  %s: rank=%d%s\n", cuiPlayerName(player, i), rank, suffix)
			}
		} else {
			source := sg.CurrentSource()
			currentName := cuiPlayerName(sg.GetPlayer(currentTurn), currentTurn)
			fmt.Fprintf(b, "手番: %s (出すソース: %s)\n", currentName, source)
			b.WriteString("p [インデックス...] でカードを出す / p で場札を引き取る\n")
		}
	})
}

// formatShitheadAction returns one line describing a player action.
func formatShitheadAction(sg interfaces.ShitheadGame, action *domain.ShitheadCpuAction) string {
	name := cuiPlayerName(sg.GetPlayer(action.PlayerIdx), action.PlayerIdx)
	if action.Pickup {
		return fmt.Sprintf("%s が場札を引き取りました\n", name)
	}
	cards := cuiCardSliceStr(action.PlayedCards)
	suffix := ""
	if action.Burned {
		suffix += " (場札焼却!)"
	}
	if action.Skipped {
		suffix += " (次プレイヤーをスキップ)"
	}
	return fmt.Sprintf("%s が %s を出しました [%s]%s\n", name, cards, action.Source, suffix)
}

// ActionLogOutput 棋譜をテキスト出力
func (p *ShitheadCuiPresenter) ActionLogOutput(sg interfaces.ShitheadGame) string {
	return actionLogOutputText(sg)
}
