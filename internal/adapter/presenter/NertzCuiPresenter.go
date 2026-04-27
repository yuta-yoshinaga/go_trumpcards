package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// NertzCuiPresenter Nertz / Pounce CUI プレゼンター
type NertzCuiPresenter struct{}

// Output ゲーム状態を文字列出力
func (p *NertzCuiPresenter) Output(g interfaces.NertzGame, lastErr error) string {
	return buildCuiOutput("Nertz / Pounce (ナーツ / パウンス)", func(b *strings.Builder) {
		// ファウンデーション (共有)
		fmt.Fprintf(b, "ラウンド %d / 手数 %d\n", g.GetRoundNo(), g.GetMoveCount())
		b.WriteString("[Foundations]\n")
		founds := g.GetFoundations()
		for i, f := range founds {
			fmt.Fprintf(b, "  F%d ", i)
			if f == nil || f.IsEmpty() {
				b.WriteString("(empty)")
			} else {
				fmt.Fprintf(b, "%s (%d/%d)", cuiCardStr(f.Top()), f.Size(), domain.NertzFoundationMax)
			}
			b.WriteString("\n")
		}
		b.WriteString("----------\n")

		// 各プレイヤー
		for i, pl := range g.GetPlayers() {
			if pl == nil {
				continue
			}
			label := i18n.T("nertz.labelHuman")
			if pl.GetIsCpu() {
				label = i18n.T("nertz.labelCpu")
			}
			fmt.Fprintf(b, "[P%d %s %s] スコア: %d\n", i, label, pl.GetName(), pl.GetScore())
			// ナッツパイル
			top := pl.NertzTop()
			if top != nil {
				fmt.Fprintf(b, "  ナッツ: %s (残 %d)\n", cuiCardStr(top), pl.NertzSize())
			} else {
				b.WriteString("  ナッツ: (empty)\n")
			}
			// タブロー (4列)
			for c := range domain.NertzTableauCnt {
				col := pl.GetTableauColumn(c)
				fmt.Fprintf(b, "  T%d: ", c)
				if len(col) == 0 {
					b.WriteString("(empty)")
				} else {
					parts := make([]string, len(col))
					for k, tc := range col {
						parts[k] = cuiCardStr(tc.Card)
					}
					b.WriteString(strings.Join(parts, " "))
				}
				b.WriteString("\n")
			}
			// ウェイスト / ストック
			if w := pl.WasteTop(); w != nil {
				fmt.Fprintf(b, "  ウェイスト: %s (%d枚)  ストック: %d枚\n", cuiCardStr(w), pl.WasteSize(), pl.StockSize())
			} else {
				fmt.Fprintf(b, "  ウェイスト: (empty)  ストック: %d枚\n", pl.StockSize())
			}
		}
		b.WriteString("----------\n")

		if lastErr != nil {
			fmt.Fprintf(b, "%s\n", color.Red(lastErr.Error()))
		}

		switch g.GetPhase() {
		case domain.NertzPhasePlaying:
			b.WriteString("プレイ中\n")
		case domain.NertzPhaseRoundEnd:
			fmt.Fprintf(b, "%s\n", color.Yellow(fmt.Sprintf("ラウンド終了 (勝者: P%d)", g.GetWinnerIdx())))
		case domain.NertzPhaseGameEnd:
			if g.GetMatchWinner() == 0 {
				fmt.Fprintf(b, "%s\n", color.Green("あなたの勝ち！"))
			} else {
				fmt.Fprintf(b, "%s\n", color.Red(fmt.Sprintf("プレイヤー%dの勝ち", g.GetMatchWinner())))
			}
		}
	})
}

// HintOutput ヒントを文字列出力
func (p *NertzCuiPresenter) HintOutput(g interfaces.NertzGame) string {
	hint := g.GetHint()
	if hint == nil {
		return "ヒントはありません。\n"
	}
	from := nertzHintZoneLabel(hint.FromZone, hint.FromCol, hint.CardIndex)
	to := nertzHintZoneLabel(hint.ToZone, hint.ToCol, -1)
	return fmt.Sprintf("ヒント: %s → %s\n", from, to)
}

// ActionLogOutput 棋譜をテキスト出力。
// プレイ中はログを空にする — リアルタイム進行中に CPU の狙いを露出させない
// ため、ラウンド/マッチ終了後にのみ完全ログを返す (Web 側と同じ運用 / PR
// #1528 レビュー指摘)。
func (p *NertzCuiPresenter) ActionLogOutput(g interfaces.NertzGame) string {
	if g.GetPhase() == domain.NertzPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(g.GetActionLog())
}

// nertzHintZoneLabel ヒントゾーンを日本語ラベルに変換する。
func nertzHintZoneLabel(zone string, col, idx int) string {
	switch zone {
	case "nertz":
		return "ナッツ"
	case "waste":
		return "ウェイスト"
	case "tableau":
		if idx >= 0 {
			return fmt.Sprintf("タブロー%d(idx=%d)", col, idx)
		}
		return fmt.Sprintf("タブロー%d", col)
	case "foundation":
		return fmt.Sprintf("ファウンデーション%d", col)
	}
	return zone
}
