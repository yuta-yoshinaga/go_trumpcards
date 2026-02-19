package presenters

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/entities"
)

// OldMaidCuiPresenter ババ抜きCUIプレゼンタークラス
type OldMaidCuiPresenter struct{}

// NewOldMaidCuiPresenter コンストラクタ
func NewOldMaidCuiPresenter() *OldMaidCuiPresenter {
	return &OldMaidCuiPresenter{}
}

// Output ゲーム状態を文字列出力
func (p *OldMaidCuiPresenter) Output(om *entities.OldMaid) string {
	res := "==========\n"
	res += "Old Maid (ババ抜き)\n"
	res += "==========\n"

	for i := 0; i < om.GetPlayerCnt(); i++ {
		player := om.GetPlayer(i)
		if player.GetIsHuman() {
			res += "[You]"
		} else {
			res += fmt.Sprintf("CPU %d", i)
		}
		if player.GetIsFinished() {
			res += ": 上がり\n"
		} else {
			res += fmt.Sprintf(": %d枚\n", player.GetCardsSize())
			if player.GetIsHuman() {
				for j := 0; j < player.GetCardsSize(); j++ {
					if j != 0 {
						res += "  "
					}
					res += fmt.Sprintf("[%d]%s", j, p.getCardStr(player.GetCard(j)))
				}
				res += "\n"
			}
		}
	}

	res += "----------\n"

	if om.GetHasDrawn() {
		drawPlayerIdx := om.GetLastDrawPlayerIdx()
		drawFromIdx := om.GetLastDrawFromIdx()
		discarded := om.GetLastDiscardedPairs()
		drawPlayerName := p.getPlayerName(om, drawPlayerIdx)
		drawFromName := p.getPlayerName(om, drawFromIdx)
		drawnCard := om.GetLastDrawCard()
		res += fmt.Sprintf("%sが%sから1枚引きました", drawPlayerName, drawFromName)
		if drawnCard != nil {
			res += fmt.Sprintf(" (%s)", p.getCardStr(drawnCard))
		}
		if discarded > 0 {
			res += fmt.Sprintf("。%d組捨てました", discarded)
		}
		res += "\n"
	}

	// CPUの行動履歴を表示
	cpuActions := om.GetCpuActions()
	if len(cpuActions) > 0 {
		res += "[CPUの行動]\n"
		for _, action := range cpuActions {
			actPlayerName := p.getPlayerName(om, action.DrawPlayerIdx)
			actFromName := p.getPlayerName(om, action.DrawFromIdx)
			res += fmt.Sprintf("%sが%sから1枚引きました", actPlayerName, actFromName)
			if action.DrawnCard != nil {
				res += fmt.Sprintf(" (%s)", p.getCardStr(action.DrawnCard))
			}
			if action.DiscardedPairs > 0 {
				res += fmt.Sprintf("。%d組捨てました", action.DiscardedPairs)
			}
			res += "\n"
		}
	}

	if om.GetGameEndFlag() {
		loserIdx := om.GetLoserIdx()
		if loserIdx >= 0 {
			loserName := p.getPlayerName(om, loserIdx)
			res += fmt.Sprintf("ゲーム終了！ %sの負け！\n", loserName)
		}
	} else {
		currentTurn := om.GetCurrentTurn()
		currentName := p.getPlayerName(om, currentTurn)
		targetIdx := om.GetNextDrawTargetIdx()
		if targetIdx >= 0 {
			targetName := p.getPlayerName(om, targetIdx)
			res += fmt.Sprintf("手番: %s → %sから引きます\n", currentName, targetName)
		} else {
			res += fmt.Sprintf("手番: %s\n", currentName)
		}
	}

	res += "==========\n"
	return res
}

// getPlayerName プレイヤー名取得
func (p *OldMaidCuiPresenter) getPlayerName(om *entities.OldMaid, idx int) string {
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
func (p *OldMaidCuiPresenter) getCardStr(card *entities.Card) string {
	if card == nil {
		return "??"
	}
	switch card.GetDesign() {
	case entities.CardDesignJoker:
		return "JOKER"
	case entities.CardDesignSpade:
		return "SPADE " + strconv.Itoa(card.GetValue())
	case entities.CardDesignClover:
		return "CLOVER " + strconv.Itoa(card.GetValue())
	case entities.CardDesignHeart:
		return "HEART " + strconv.Itoa(card.GetValue())
	case entities.CardDesignDiamond:
		return "DIAMOND " + strconv.Itoa(card.GetValue())
	default:
		return "UNKNOWN"
	}
}
