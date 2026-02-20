package presenters

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/entities"
)

// SevensCuiPresenter 7並べCUIプレゼンタークラス
type SevensCuiPresenter struct{}

// NewSevensCuiPresenter コンストラクタ
func NewSevensCuiPresenter() *SevensCuiPresenter {
	return &SevensCuiPresenter{}
}

// Output ゲーム状態を文字列出力
func (p *SevensCuiPresenter) Output(s *entities.Sevens) string {
	res := "==========\n"
	res += "Sevens (7並べ)\n"
	res += "==========\n"

	for i := 0; i < s.GetPlayerCnt(); i++ {
		player := s.GetPlayer(i)
		if player.GetIsHuman() {
			res += "[You]"
		} else {
			res += fmt.Sprintf("CPU %d", i)
		}
		if player.GetIsFinished() {
			res += fmt.Sprintf(": 上がり/失格 (ランク: %d位)\n", player.GetRank())
		} else {
			res += fmt.Sprintf(": %d枚 (パス: %d/%d)\n",
				player.GetCardsSize(), player.GetPassesUsed(), player.GetMaxPasses())
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

	// ボード状態
	res += "ボード:\n"
	suits := []int{entities.CardDesignSpade, entities.CardDesignClover, entities.CardDesignHeart, entities.CardDesignDiamond}
	suitNames := []string{"SPADE", "CLOVER", "HEART", "DIAMOND"}
	mins := s.GetTableMinVals()
	maxs := s.GetTableMaxVals()
	for i, suit := range suits {
		res += fmt.Sprintf("  %s: %d〜%d\n", suitNames[i], mins[suit], maxs[suit])
	}

	// 人間の前の行動
	humanAction := s.GetHumanAction()
	if humanAction != nil {
		if humanAction.PlayedCard == nil {
			res += fmt.Sprintf("%sがパスしました\n", p.getPlayerName(s, humanAction.PlayerIdx))
		} else {
			res += fmt.Sprintf("%sが %s を出しました\n",
				p.getPlayerName(s, humanAction.PlayerIdx),
				p.getCardStr(humanAction.PlayedCard))
		}
	}

	// CPUの行動履歴を表示
	cpuActions := s.GetCpuActions()
	if len(cpuActions) > 0 {
		res += "[CPUの行動]\n"
		for _, action := range cpuActions {
			actPlayerName := p.getPlayerName(s, action.PlayerIdx)
			if action.PlayedCard == nil {
				res += fmt.Sprintf("%sがパスしました\n", actPlayerName)
			} else {
				res += fmt.Sprintf("%sが %s を出しました\n", actPlayerName, p.getCardStr(action.PlayedCard))
			}
		}
	}

	if s.GetGameEndFlag() {
		res += "ゲーム終了！\n"
		for i := 0; i < s.GetPlayerCnt(); i++ {
			player := s.GetPlayer(i)
			res += fmt.Sprintf("  %s: %d位\n", p.getPlayerName(s, i), player.GetRank())
		}
	} else {
		currentTurn := s.GetCurrentTurn()
		currentName := p.getPlayerName(s, currentTurn)
		res += fmt.Sprintf("手番: %s\n", currentName)
		res += "p [インデックス] でカードを出す / p でパス\n"
	}

	res += "==========\n"
	return res
}

// getPlayerName プレイヤー名取得
func (p *SevensCuiPresenter) getPlayerName(s *entities.Sevens, idx int) string {
	player := s.GetPlayer(idx)
	if player == nil {
		return "不明"
	}
	if player.GetIsHuman() {
		return "あなた"
	}
	return fmt.Sprintf("CPU %d", idx)
}

// getCardStr カード情報文字列取得
func (p *SevensCuiPresenter) getCardStr(card *entities.Card) string {
	if card == nil {
		return "??"
	}
	switch card.GetDesign() {
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
