package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// canfieldColumnStr タブロー1列の表示文字列
func canfieldColumnStr(colCards []*domain.CanfieldTableauCard) string {
	parts := make([]string, len(colCards))
	for j, tc := range colCards {
		parts[j] = fmt.Sprintf(" [%d]%s", j, cuiCardStr(tc.Card))
	}
	return strings.Join(parts, " ")
}

// CanfieldCuiPresenter キャンフィールドCUIプレゼンタークラス
type CanfieldCuiPresenter struct{}

// Output ゲーム状態を文字列出力
func (p *CanfieldCuiPresenter) Output(c interfaces.CanfieldGame, lastErr error) string {
	return buildCuiOutput("Canfield (キャンフィールド)", func(b *strings.Builder) {
		// ベースランク
		fmt.Fprintf(b, "Base rank: %d\n", c.GetBaseRank())

		// ファンデーション
		b.WriteString("Foundation: ")
		foundation := c.GetFoundation()
		for i := 0; i < domain.CanfieldFoundationCnt; i++ {
			if i != 0 {
				b.WriteString(" | ")
			}
			pile := foundation[i]
			if len(pile) == 0 {
				b.WriteString("[空]")
			} else {
				b.WriteString(cuiCardStr(pile[len(pile)-1]))
			}
		}
		b.WriteString("\n")

		// リザーブ/ストック/ウェイスト
		reserve := c.GetReserve()
		if len(reserve) > 0 {
			fmt.Fprintf(b, "Reserve: %d枚 (top: %s)", len(reserve), cuiCardStr(reserve[len(reserve)-1]))
		} else {
			b.WriteString("Reserve: [空]")
		}
		b.WriteString("\n")
		fmt.Fprintf(b, "Stock: %d枚", c.GetStockCount())
		waste := c.GetWaste()
		if len(waste) > 0 {
			fmt.Fprintf(b, " | Waste: %s", cuiCardStr(waste[len(waste)-1]))
		} else {
			b.WriteString(" | Waste: [空]")
		}
		b.WriteString("\n----------\n")

		// タブロー
		tableau := c.GetTableau()
		for col := 0; col < domain.CanfieldTableauCnt; col++ {
			colCards := tableau[col]
			fmt.Fprintf(b, "列%d:", col)
			if len(colCards) == 0 {
				b.WriteString(" [空]")
			} else {
				b.WriteString(canfieldColumnStr(colCards))
			}
			b.WriteString("\n")
		}
		b.WriteString("----------\n")

		if lastErr != nil {
			fmt.Fprintf(b, "%s\n", color.Red(lastErr.Error()))
		}
		switch c.GetPhase() {
		case domain.CanfieldPhasePlaying:
			fmt.Fprintf(b, "手数: %d\n", c.GetMoveCount())
		case domain.CanfieldPhaseGameClear:
			fmt.Fprintf(b, "%s 手数: %d\n", color.Green("ゲームクリア！"), c.GetMoveCount())
		case domain.CanfieldPhaseGameOver:
			b.WriteString(color.Red("ゲームオーバー") + "\n")
		}
	})
}

// HintOutput ヒントを文字列出力
func (p *CanfieldCuiPresenter) HintOutput(c interfaces.CanfieldGame) string {
	hint := c.GetHint()
	if hint == nil {
		return "ヒントはありません。\n"
	}
	return fmt.Sprintf("ヒント: %s", canfieldHintStr(hint)) + "\n"
}

// ActionLogOutput 棋譜をテキスト出力
func (p *CanfieldCuiPresenter) ActionLogOutput(c interfaces.CanfieldGame) string {
	if c.GetPhase() == domain.CanfieldPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(c.GetActionLog())
}

func canfieldHintStr(hint *domain.CanfieldHint) string {
	var from string
	switch hint.FromZone {
	case "tableau":
		from = fmt.Sprintf("タブロー列%d[%d]", hint.FromCol, hint.CardIndex)
	case "reserve":
		from = "リザーブ"
	default:
		from = "ウェイスト"
	}
	var to string
	if hint.ToZone == "foundation" {
		to = "ファンデーション"
	} else {
		to = fmt.Sprintf("タブロー列%d", hint.ToCol)
	}
	return fmt.Sprintf("%s → %s", from, to)
}
