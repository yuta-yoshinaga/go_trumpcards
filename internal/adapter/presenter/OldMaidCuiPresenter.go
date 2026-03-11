package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// oldMaidPlayerStr returns the display string for a single OldMaid player.
func oldMaidPlayerStr(player *domain.OldMaidPlayer, i int) string {
	var b strings.Builder
	if player.GetIsHuman() {
		b.WriteString("[You]")
	} else {
		fmt.Fprintf(&b, "CPU %d", i)
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

// OldMaidCuiPresenter ババ抜きCUIプレゼンタークラス
type OldMaidCuiPresenter struct{}

// NewOldMaidCuiPresenter コンストラクタ
func NewOldMaidCuiPresenter() *OldMaidCuiPresenter {
	return &OldMaidCuiPresenter{}
}

// Output ゲーム状態を文字列出力
func (p *OldMaidCuiPresenter) Output(om interfaces.OldMaidGame, lastErr error) string {
	title := "Old Maid (ババ抜き)"
	if om.GetConfig().Mode == domain.OldMaidModeJijiNuki {
		title = "Old Maid (ジジ抜き)"
	}
	return buildCuiOutput(title, func(b *strings.Builder) {
		for i := 0; i < om.GetPlayerCnt(); i++ {
			b.WriteString(oldMaidPlayerStr(om.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		if om.GetHasDrawn() {
			drawPlayerIdx := om.GetLastDrawPlayerIdx()
			drawFromIdx := om.GetLastDrawFromIdx()
			discarded := om.GetLastDiscardedPairs()
			drawPlayerName := cuiPlayerName(om.GetPlayer(drawPlayerIdx), drawPlayerIdx)
			drawFromName := cuiPlayerName(om.GetPlayer(drawFromIdx), drawFromIdx)
			drawnCard := om.GetLastDrawCard()
			drawPlayer := om.GetPlayer(drawPlayerIdx)
			fmt.Fprintf(b, "%sが%sから1枚引きました", drawPlayerName, drawFromName)
			// Only reveal drawn card for human players to preserve CPU game fairness
			if drawnCard != nil && drawPlayer != nil && drawPlayer.GetIsHuman() {
				fmt.Fprintf(b, " (%s)", cuiCardStr(drawnCard))
			}
			if discarded > 0 {
				fmt.Fprintf(b, "。%d組捨てました", discarded)
			}
			b.WriteString("\n")
		}

		// CPUの行動履歴を表示
		cpuActions := om.GetCpuActions()
		if len(cpuActions) > 0 {
			b.WriteString("[CPUの行動]\n")
			for _, action := range cpuActions {
				actPlayerName := cuiPlayerName(om.GetPlayer(action.DrawPlayerIdx), action.DrawPlayerIdx)
				actFromName := cuiPlayerName(om.GetPlayer(action.DrawFromIdx), action.DrawFromIdx)
				fmt.Fprintf(b, "%sが%sから1枚引きました", actPlayerName, actFromName)
				// CPU drawn card is intentionally hidden to preserve game fairness
				if action.DiscardedPairs > 0 {
					fmt.Fprintf(b, "。%d組捨てました", action.DiscardedPairs)
				}
				b.WriteString("\n")
			}
		}

		// 引き履歴
		drawHistory := om.GetDrawHistory()
		if len(drawHistory) > 0 {
			b.WriteString("[引き履歴]\n")
			for i, entry := range drawHistory {
				drawerName := cuiPlayerName(om.GetPlayer(entry.DrawPlayerIdx), entry.DrawPlayerIdx)
				fromName := cuiPlayerName(om.GetPlayer(entry.DrawFromIdx), entry.DrawFromIdx)
				fmt.Fprintf(b, "%d. %sが%sから引いた", i+1, drawerName, fromName)
				if entry.DiscardedPairs > 0 {
					fmt.Fprintf(b, " (%d組捨て)", entry.DiscardedPairs)
				}
				if entry.DrawerFinished {
					fmt.Fprintf(b, " [%s上がり]", drawerName)
				}
				if entry.TargetFinished {
					fmt.Fprintf(b, " [%s上がり]", fromName)
				}
				b.WriteString("\n")
			}
		}

		// メタAI状態
		if profile := om.GetHumanProfile(); profile != nil {
			fmt.Fprintf(b, "[メタAI] 適応中 (ゲーム数: %d, 端ピック率: %.0f%%)\n",
				profile.GamesPlayed,
				(profile.PickRate(0)+profile.PickRate(2))*100,
			)
		}

		// エラーメッセージ
		if lastErr != nil {
			fmt.Fprintf(b, "%s\n", lastErr.Error())
		}

		if om.GetGameEndFlag() {
			loserIdx := om.GetLoserIdx()
			if loserIdx >= 0 {
				loserName := cuiPlayerName(om.GetPlayer(loserIdx), loserIdx)
				gameEndLine := fmt.Sprintf("ゲーム終了！ %sの負け！", loserName)
				if om.GetConfig().Mode == domain.OldMaidModeJijiNuki && om.GetRemovedCard() != nil {
					gameEndLine += fmt.Sprintf("（除外カード: %s）", cuiCardStr(om.GetRemovedCard()))
				}
				fmt.Fprintf(b, "%s\n", gameEndLine)
			}
		} else {
			currentTurn := om.GetCurrentTurn()
			currentName := cuiPlayerName(om.GetPlayer(currentTurn), currentTurn)
			targetIdx := om.GetNextDrawTargetIdx()
			if targetIdx >= 0 {
				targetName := cuiPlayerName(om.GetPlayer(targetIdx), targetIdx)
				fmt.Fprintf(b, "手番: %s → %sから引きます\n", currentName, targetName)
			} else {
				fmt.Fprintf(b, "手番: %s\n", currentName)
			}
		}
	})
}

// ActionLogOutput 棋譜をテキスト出力
func (p *OldMaidCuiPresenter) ActionLogOutput(om interfaces.OldMaidGame) string {
	return actionLogOutputText(om)
}
