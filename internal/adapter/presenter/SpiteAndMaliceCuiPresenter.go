package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// SpiteAndMaliceCuiPresenter Spite & Malice CUI プレゼンター
type SpiteAndMaliceCuiPresenter struct{}

// Output ゲーム状態を文字列出力
func (p *SpiteAndMaliceCuiPresenter) Output(g interfaces.SpiteAndMaliceGame, lastErr error) string {
	return buildCuiOutput("Spite and Malice (スパイト・アンド・マリス)", func(b *strings.Builder) {
		// ファウンデーション
		foundations := g.GetFoundations()
		for i := range domain.SpiteAndMaliceFoundationCnt {
			pile := foundations[i]
			fmt.Fprintf(b, "[F%d] ", i)
			if len(pile) == 0 {
				b.WriteString("(empty)")
			} else {
				top := pile[len(pile)-1]
				fmt.Fprintf(b, "%s (%d/%d)", cuiCardStr(top), len(pile), domain.SpiteAndMaliceFoundationMax)
			}
			b.WriteString("\n")
		}
		b.WriteString("----------\n")

		// 各プレイヤーの状態
		for i := range domain.SpiteAndMalicePlayerCnt {
			pl := g.GetPlayer(i)
			if pl == nil {
				continue
			}
			label := "人間"
			if pl.GetIsCpu() {
				label = "CPU"
			}
			fmt.Fprintf(b, "[P%d %s] ", i, label)
			// goal top と残枚数
			if top := pl.GoalTop(); top != nil {
				fmt.Fprintf(b, "ゴール: %s (残 %d)", cuiCardStr(top), pl.GoalSize())
			} else {
				b.WriteString("ゴール: (empty)")
			}
			b.WriteString("\n")

			// 手札 (自分のターン or 相手が CPU なら表示)
			hand := pl.GetHand()
			if i == g.GetCurrent() || pl.GetIsCpu() {
				b.WriteString("  手札: ")
				if len(hand) == 0 {
					b.WriteString("(empty)")
				} else {
					parts := make([]string, len(hand))
					for k, c := range hand {
						parts[k] = fmt.Sprintf("%d:%s", k, cuiCardStr(c))
					}
					b.WriteString(strings.Join(parts, " "))
				}
				b.WriteString("\n")
			} else {
				fmt.Fprintf(b, "  手札: %d枚\n", len(hand))
			}

			// サイドパイル
			for s := range domain.SpiteAndMaliceSideCnt {
				side := pl.GetSide(s)
				fmt.Fprintf(b, "  [S%d] ", s)
				if len(side) == 0 {
					b.WriteString("(empty)")
				} else {
					top := side[len(side)-1]
					fmt.Fprintf(b, "%s (%d枚)", cuiCardStr(top), len(side))
				}
				b.WriteString("\n")
			}
		}
		b.WriteString("----------\n")

		// ストック
		fmt.Fprintf(b, "ストック: %d枚 / 完成: %d枚\n", g.GetStockSize(), g.GetCompletedSize())

		if lastErr != nil {
			fmt.Fprintf(b, "%s\n", color.Red(lastErr.Error()))
		}

		switch g.GetPhase() {
		case domain.SpiteAndMalicePhasePlaying:
			fmt.Fprintf(b, "ターン: %d (手数 %d)\n", g.GetCurrent(), g.GetMoveCount())
		case domain.SpiteAndMalicePhaseGameOver:
			if g.GetWinner() == domain.SpiteAndMaliceHumanIdx {
				fmt.Fprintf(b, "%s 手数: %d\n", color.Green("あなたの勝ち！"), g.GetMoveCount())
			} else {
				fmt.Fprintf(b, "%s 手数: %d\n", color.Red("CPU の勝ち"), g.GetMoveCount())
			}
		}
	})
}

// HintOutput ヒントを文字列出力
func (p *SpiteAndMaliceCuiPresenter) HintOutput(g interfaces.SpiteAndMaliceGame) string {
	hint := g.GetHint()
	if hint == nil {
		return "ヒントはありません。\n"
	}
	if hint.Discard {
		return fmt.Sprintf("ヒント: 手札%d をサイド%d にディスカード\n", hint.Index, hint.FoundationIdx)
	}
	switch hint.Source {
	case domain.SpiteAndMaliceSourceGoal:
		return fmt.Sprintf("ヒント: ゴール → ファウンデーション%d\n", hint.FoundationIdx)
	case domain.SpiteAndMaliceSourceHand:
		return fmt.Sprintf("ヒント: 手札%d → ファウンデーション%d\n", hint.Index, hint.FoundationIdx)
	case domain.SpiteAndMaliceSourceSide:
		return fmt.Sprintf("ヒント: サイド%d → ファウンデーション%d\n", hint.Index, hint.FoundationIdx)
	}
	return "ヒントはありません。\n"
}

// ActionLogOutput 棋譜をテキスト出力
func (p *SpiteAndMaliceCuiPresenter) ActionLogOutput(g interfaces.SpiteAndMaliceGame) string {
	if g.GetPhase() == domain.SpiteAndMalicePhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(g.GetActionLog())
}
