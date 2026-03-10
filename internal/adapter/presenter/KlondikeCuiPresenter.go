package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// KlondikeCuiPresenter クロンダイクCUIプレゼンタークラス
type KlondikeCuiPresenter struct{}

// NewKlondikeCuiPresenter コンストラクタ
func NewKlondikeCuiPresenter() *KlondikeCuiPresenter {
	return &KlondikeCuiPresenter{}
}

// Output ゲーム状態を文字列出力
func (p *KlondikeCuiPresenter) Output(k interfaces.KlondikeGame, lastErr error) string {
	return buildCuiOutput("Klondike (ソリティア)", func(b *strings.Builder) {
		// ファンデーション
		b.WriteString("Foundation: ")
		foundation := k.GetFoundation()
		for i := 0; i < domain.KlondikeFoundationCnt; i++ {
			if i != 0 {
				b.WriteString(" | ")
			}
			pile := foundation[i]
			if len(pile) == 0 {
				b.WriteString("[空]")
			} else {
				topCard := pile[len(pile)-1]
				b.WriteString(cuiCardStr(topCard))
			}
		}
		b.WriteString("\n")

		// ストックとウェイスト
		fmt.Fprintf(b, "Stock: %d枚", k.GetStockCount())
		waste := k.GetWaste()
		if len(waste) > 0 {
			fmt.Fprintf(b, " | Waste: %s", cuiCardStr(waste[len(waste)-1]))
		} else {
			b.WriteString(" | Waste: [空]")
		}
		b.WriteString("\n")

		b.WriteString("----------\n")

		// タブロー
		tableau := k.GetTableau()
		for col := 0; col < domain.KlondikeTableauCnt; col++ {
			fmt.Fprintf(b, "列%d:", col)
			colCards := tableau[col]
			if len(colCards) == 0 {
				b.WriteString(" [空]")
			} else {
				for j, tc := range colCards {
					if j != 0 {
						b.WriteString(" ")
					}
					if tc.FaceUp {
						fmt.Fprintf(b, " [%d]%s", j, cuiCardStr(tc.Card))
					} else {
						fmt.Fprintf(b, " [%d]??", j)
					}
				}
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		// エラーメッセージ
		if lastErr != nil {
			fmt.Fprintf(b, "%s\n", lastErr.Error())
		}

		// ゲーム状態
		phase := k.GetPhase()
		switch phase {
		case domain.KlondikePhasePlaying:
			fmt.Fprintf(b, "手数: %d\n", k.GetMoveCount())
		case domain.KlondikePhaseGameClear:
			fmt.Fprintf(b, "ゲームクリア！ 手数: %d\n", k.GetMoveCount())
		case domain.KlondikePhaseGameOver:
			b.WriteString("ゲームオーバー\n")
		}
	})
}

// HintOutput ヒントを文字列出力
func (p *KlondikeCuiPresenter) HintOutput(k interfaces.KlondikeGame) string {
	hint := k.GetHint()
	if hint == nil {
		return "ヒントはありません。\n"
	}
	return fmt.Sprintf("ヒント: %s", klondikeHintStr(hint)) + "\n"
}

// ActionLogOutput 棋譜をテキスト出力
func (p *KlondikeCuiPresenter) ActionLogOutput(k interfaces.KlondikeGame) string {
	phase := k.GetPhase()
	if phase == domain.KlondikePhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(k.GetActionLog())
}

// klondikeHintStr ヒントを文字列に変換
func klondikeHintStr(hint *domain.KlondikeHint) string {
	var from string
	if hint.FromZone == "tableau" {
		from = fmt.Sprintf("タブロー列%d[%d]", hint.FromCol, hint.CardIndex)
	} else {
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
