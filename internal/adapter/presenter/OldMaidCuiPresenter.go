package presenter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// OldMaidCuiPresenter ババ抜きCUIプレゼンタークラス
type OldMaidCuiPresenter struct{}

// NewOldMaidCuiPresenter コンストラクタ
func NewOldMaidCuiPresenter() *OldMaidCuiPresenter {
	return &OldMaidCuiPresenter{}
}

// Output ゲーム状態を文字列出力
func (p *OldMaidCuiPresenter) Output(om *domain.OldMaid) string {
	var b strings.Builder

	b.WriteString("==========\n")
	b.WriteString("Old Maid (ババ抜き)\n")
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
					fmt.Fprintf(&b, "[%d]%s", j, p.getCardStr(player.GetCard(j)))
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
			fmt.Fprintf(&b, " (%s)", p.getCardStr(drawnCard))
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

	if om.GetGameEndFlag() {
		loserIdx := om.GetLoserIdx()
		if loserIdx >= 0 {
			loserName := p.getPlayerName(om, loserIdx)
			fmt.Fprintf(&b, "ゲーム終了！ %sの負け！\n", loserName)
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
func (p *OldMaidCuiPresenter) getPlayerName(om *domain.OldMaid, idx int) string {
	player := om.GetPlayer(idx)
	if player == nil {
		return "不明"
	}
	if player.GetIsHuman() {
		return "あなた"
	}
	return fmt.Sprintf("CPU %d", idx)
}

// getCardStr カード情報文字列取得
func (p *OldMaidCuiPresenter) getCardStr(card *domain.Card) string {
	if card == nil {
		return "??"
	}
	switch card.GetDesign() {
	case domain.CardDesignJoker:
		return "JOKER"
	case domain.CardDesignSpade:
		return "SPADE " + strconv.Itoa(card.GetValue())
	case domain.CardDesignClover:
		return "CLOVER " + strconv.Itoa(card.GetValue())
	case domain.CardDesignHeart:
		return "HEART " + strconv.Itoa(card.GetValue())
	case domain.CardDesignDiamond:
		return "DIAMOND " + strconv.Itoa(card.GetValue())
	default:
		return "UNKNOWN"
	}
}
