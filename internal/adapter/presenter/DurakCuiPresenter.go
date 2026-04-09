package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// durakPlayerStr returns the display string for a single Durak player.
func durakPlayerStr(player *domain.DurakPlayer, i int, isAttacker, isDefender bool) string {
	var b strings.Builder
	b.WriteString(cuiPlayerName(player, i))
	if isAttacker {
		b.WriteString(color.BoldYellow(" [攻撃]"))
	}
	if isDefender {
		b.WriteString(color.Bold(" [防御]"))
	}
	if player.GetIsFinished() {
		b.WriteString(": 上がり\n")
	} else {
		fmt.Fprintf(&b, ": %d枚\n", player.GetCardsSize())
		if player.GetIsHuman() {
			b.WriteString(cuiIndexedCardListStr(player))
			b.WriteString("\n")
		}
	}
	return b.String()
}

// DurakCuiPresenter ドゥラークCUIプレゼンタークラス
type DurakCuiPresenter struct{}

// Output ゲーム状態を文字列出力
func (p *DurakCuiPresenter) Output(dg interfaces.DurakGame, lastErr error) string {
	return buildCuiOutput("Durak (ドゥラーク)", func(b *strings.Builder) {
		// 切り札情報
		fmt.Fprintf(b, "切り札: %s", cuiSuitName(dg.GetTrumpSuit()))
		if dg.GetTrumpCard() != nil {
			fmt.Fprintf(b, " (底: %s)", cuiCardStr(dg.GetTrumpCard()))
		}
		fmt.Fprintf(b, "  山札: %d枚\n", dg.GetStockCount())

		b.WriteString("----------\n")

		// プレイヤー情報
		for i := 0; i < dg.GetPlayerCnt(); i++ {
			b.WriteString(durakPlayerStr(dg.GetPlayer(i), i, i == dg.GetAttackerIdx(), i == dg.GetDefenderIdx()))
		}

		b.WriteString("----------\n")

		// テーブル
		pairs := dg.GetTablePairs()
		if len(pairs) > 0 {
			b.WriteString("テーブル:\n")
			for i, pair := range pairs {
				if pair.Defense != nil {
					fmt.Fprintf(b, "  %d: %s ← %s\n", i, cuiCardStr(pair.Attack), cuiCardStr(pair.Defense))
				} else {
					fmt.Fprintf(b, "  %d: %s ← ?\n", i, cuiCardStr(pair.Attack))
				}
			}
		} else {
			b.WriteString("テーブル: (空)\n")
		}

		// フェーズ
		switch dg.GetPhase() {
		case domain.DurakPhaseAttack:
			b.WriteString("フェーズ: 攻撃\n")
		case domain.DurakPhaseDefend:
			b.WriteString("フェーズ: 防御\n")
		case domain.DurakPhaseGameEnd:
			b.WriteString("フェーズ: ゲーム終了\n")
		}

		// ゲーム終了
		if dg.GetGameEndFlag() {
			loserIdx := dg.GetLoserIdx()
			if loserIdx < 0 {
				b.WriteString(color.Green("引き分け！\n"))
			} else {
				player := dg.GetPlayer(loserIdx)
				if player.GetIsHuman() {
					b.WriteString(color.Red("あなたがドゥラーク（負け）です！\n"))
				} else {
					fmt.Fprintf(b, "%s\n", color.Green(fmt.Sprintf("CPU %d がドゥラーク（負け）です！", loserIdx)))
				}
			}
		}

		// エラー
		if lastErr != nil {
			fmt.Fprintf(b, "%s\n", color.Red(lastErr.Error()))
		}
	})
}

// ActionLogOutput 棋譜を文字列出力
func (p *DurakCuiPresenter) ActionLogOutput(dg interfaces.DurakGame) string {
	return actionLogOutputText(dg)
}
