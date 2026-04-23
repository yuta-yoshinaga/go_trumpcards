package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// CassinoCuiPresenter カシノ CUI プレゼンタークラス。
type CassinoCuiPresenter struct{}

// Output ゲーム状態を文字列出力。
func (p *CassinoCuiPresenter) Output(cg interfaces.CassinoGame, lastErr error) string {
	return buildCuiOutput("Cassino (カッシーノ)", func(b *strings.Builder) {
		for i := 0; i < cg.GetPlayerCnt(); i++ {
			b.WriteString(cassinoPlayerStr(cg.GetPlayer(i), i))
		}
		b.WriteString("----------\n")

		tableCards := cg.GetTableCards()
		if len(tableCards) > 0 {
			fmt.Fprintf(b, "場: %s\n", cuiCardSliceStr(tableCards))
		} else {
			b.WriteString("場: なし\n")
		}

		// ビルド
		builds := cg.GetBuilds()
		if len(builds) > 0 {
			b.WriteString(color.Bold("[ビルド]") + "\n")
			for i, build := range builds {
				fmt.Fprintf(b, "  #%d (値%d, owner:%s, multi:%v): ", i, build.Value, cuiPlayerName(cg.GetPlayer(build.OwnerIdx), build.OwnerIdx), build.IsMulti)
				for gi, g := range build.Groups {
					if gi > 0 {
						b.WriteString(" | ")
					}
					b.WriteString(cuiCardSliceStr(g))
				}
				b.WriteString("\n")
			}
		}

		// 人間アクション
		if ha := cg.GetHumanAction(); ha != nil {
			fmt.Fprintf(b, "あなたの行動: %s\n", cassinoActionStr(ha))
		}
		// CPU アクション履歴
		cpu := cg.GetCpuActions()
		if len(cpu) > 0 {
			b.WriteString(color.Bold("[CPUの行動]") + "\n")
			for _, a := range cpu {
				fmt.Fprintf(b, "  %s: %s\n", cuiPlayerName(cg.GetPlayer(a.PlayerIdx), a.PlayerIdx), cassinoActionStr(a))
			}
		}

		if lastErr != nil {
			fmt.Fprintf(b, "%s\n", color.Red(lastErr.Error()))
		}

		if cg.GetGameEndFlag() {
			b.WriteString("ゲーム終了！\n")
			for i := 0; i < cg.GetPlayerCnt(); i++ {
				pl := cg.GetPlayer(i)
				if pl == nil {
					continue
				}
				fmt.Fprintf(b, "  %s: %d点\n", cuiPlayerName(pl, i), pl.GetTotalScore())
			}
		} else {
			currentTurn := cg.GetCurrentTurn()
			currentName := cuiPlayerName(cg.GetPlayer(currentTurn), currentTurn)
			fmt.Fprintf(b, "手番: %s\n", currentName)
			b.WriteString("t <hand> <tbl...> で捕獲 / b <hand> <value> <tbl...> でビルド / tr <hand> で場に置く\n")
		}
	})
}

// cassinoPlayerStr returns the display string for a single Cassino player.
func cassinoPlayerStr(player *domain.CassinoPlayer, i int) string {
	var b strings.Builder
	b.WriteString(cuiPlayerName(player, i))
	fmt.Fprintf(&b, ": 手札%d枚 / 捕獲%d枚 / スイープ%d / 累計%d点\n",
		player.GetCardsSize(), player.CapturedCount(), player.GetSweepCount(), player.GetTotalScore())
	if player.GetIsHuman() {
		b.WriteString(cuiIndexedCardListStr(player))
		b.WriteString("\n")
	}
	return b.String()
}

// cassinoActionStr renders an action as a short readable line.
func cassinoActionStr(a *domain.CassinoAction) string {
	if a == nil {
		return ""
	}
	switch a.Type {
	case domain.CassinoActionTake:
		suffix := ""
		if a.IsSweep {
			suffix = " (スイープ!)"
		}
		return fmt.Sprintf("捕獲 played=%s captured=%d枚%s",
			cassinoCardShort(a.PlayedCard), len(a.CapturedCards), suffix)
	case domain.CassinoActionBuild:
		return fmt.Sprintf("ビルド値%d (played=%s)", a.BuildValue, cassinoCardShort(a.PlayedCard))
	case domain.CassinoActionTrail:
		return fmt.Sprintf("場に置く %s", cassinoCardShort(a.PlayedCard))
	default:
		return string(a.Type)
	}
}

// cassinoCardShort renders a single card as a short text representation.
func cassinoCardShort(c *domain.Card) string {
	if c == nil {
		return "-"
	}
	return cuiCardSliceStr([]*domain.Card{c})
}

// ActionLogOutput 棋譜をテキスト出力。
func (p *CassinoCuiPresenter) ActionLogOutput(cg interfaces.CassinoGame) string {
	return actionLogOutputText(cg)
}
