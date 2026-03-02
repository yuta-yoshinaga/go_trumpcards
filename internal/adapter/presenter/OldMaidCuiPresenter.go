package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// OldMaidCuiPresenter ババ抜きCUIプレゼンタークラス
type OldMaidCuiPresenter struct{}

// NewOldMaidCuiPresenter コンストラクタ
func NewOldMaidCuiPresenter() *OldMaidCuiPresenter {
	return &OldMaidCuiPresenter{}
}

// Output ゲーム状態を文字列出力
func (p *OldMaidCuiPresenter) Output(om interfaces.OldMaidGame, lastErr error) string {
	var b strings.Builder

	b.WriteString("==========\n")
	if om.GetConfig().Mode == domain.OldMaidModeJijiNuki {
		b.WriteString("Old Maid (ジジ抜き)\n")
	} else {
		b.WriteString("Old Maid (ババ抜き)\n")
	}
	b.WriteString("==========\n")

	for i := 0; i < om.GetPlayerCnt(); i++ {
		player := om.GetPlayer(i)
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
				for j := 0; j < player.GetCardsSize(); j++ {
					if j != 0 {
						b.WriteString("  ")
					}
					fmt.Fprintf(&b, "[%d]%s", j, cuiCardStr(player.GetCard(j)))
				}
				b.WriteString("\n")
			}
		}
	}

	b.WriteString("----------\n")

	if om.GetHasDrawn() {
		drawPlayerIdx := om.GetLastDrawPlayerIdx()
		drawFromIdx := om.GetLastDrawFromIdx()
		discarded := om.GetLastDiscardedPairs()
		drawPlayerName := p.getPlayerName(om, drawPlayerIdx)
		drawFromName := p.getPlayerName(om, drawFromIdx)
		drawnCard := om.GetLastDrawCard()
		drawPlayer := om.GetPlayer(drawPlayerIdx)
		fmt.Fprintf(&b, "%sが%sから1枚引きました", drawPlayerName, drawFromName)
		// Only reveal drawn card for human players to preserve CPU game fairness
		if drawnCard != nil && drawPlayer != nil && drawPlayer.GetIsHuman() {
			fmt.Fprintf(&b, " (%s)", cuiCardStr(drawnCard))
		}
		if discarded > 0 {
			fmt.Fprintf(&b, "。%d組捨てました", discarded)
		}
		b.WriteString("\n")
	}

	// CPUの行動履歴を表示
	cpuActions := om.GetCpuActions()
	if len(cpuActions) > 0 {
		b.WriteString("[CPUの行動]\n")
		for _, action := range cpuActions {
			actPlayerName := p.getPlayerName(om, action.DrawPlayerIdx)
			actFromName := p.getPlayerName(om, action.DrawFromIdx)
			fmt.Fprintf(&b, "%sが%sから1枚引きました", actPlayerName, actFromName)
			// CPU drawn card is intentionally hidden to preserve game fairness
			if action.DiscardedPairs > 0 {
				fmt.Fprintf(&b, "。%d組捨てました", action.DiscardedPairs)
			}
			b.WriteString("\n")
		}
	}

	// エラーメッセージ
	if lastErr != nil {
		fmt.Fprintf(&b, "%s\n", lastErr.Error())
	}

	if om.GetGameEndFlag() {
		loserIdx := om.GetLoserIdx()
		if loserIdx >= 0 {
			loserName := p.getPlayerName(om, loserIdx)
			gameEndLine := fmt.Sprintf("ゲーム終了！ %sの負け！", loserName)
			if om.GetConfig().Mode == domain.OldMaidModeJijiNuki && om.GetRemovedCard() != nil {
				gameEndLine += fmt.Sprintf("（除外カード: %s）", cuiCardStr(om.GetRemovedCard()))
			}
			fmt.Fprintf(&b, "%s\n", gameEndLine)
		}
	} else {
		currentTurn := om.GetCurrentTurn()
		currentName := p.getPlayerName(om, currentTurn)
		targetIdx := om.GetNextDrawTargetIdx()
		if targetIdx >= 0 {
			targetName := p.getPlayerName(om, targetIdx)
			fmt.Fprintf(&b, "手番: %s → %sから引きます\n", currentName, targetName)
		} else {
			fmt.Fprintf(&b, "手番: %s\n", currentName)
		}
	}

	b.WriteString("==========\n")
	return b.String()
}

// getPlayerName プレイヤー名取得
func (p *OldMaidCuiPresenter) getPlayerName(om interfaces.OldMaidGame, idx int) string {
	player := om.GetPlayer(idx)
	if player == nil {
		return "不明"
	}
	return cuiPlayerName(player, idx)
}
