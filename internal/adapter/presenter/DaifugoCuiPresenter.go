package presenter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// DaifugoCuiPresenter 大富豪CUIプレゼンタークラス
type DaifugoCuiPresenter struct{}

// NewDaifugoCuiPresenter コンストラクタ
func NewDaifugoCuiPresenter() *DaifugoCuiPresenter {
	return &DaifugoCuiPresenter{}
}

// Output ゲーム状態を文字列出力
func (p *DaifugoCuiPresenter) Output(dg *domain.Daifugo) string {
	res := "==========\n"
	res += "Daifugo (大富豪)\n"
	res += "==========\n"

	for i := 0; i < dg.GetPlayerCnt(); i++ {
		player := dg.GetPlayer(i)
		if player.GetIsHuman() {
			res += "[You]"
		} else {
			res += fmt.Sprintf("CPU %d", i)
		}
		if player.GetIsFinished() {
			res += fmt.Sprintf(": 上がり (ランク: %s)\n", p.rankName(player.GetRank()))
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

	// ローカルルール状態
	if dg.GetRevolutionActive() {
		res += "【革命中】2が最弱、3が最強\n"
	}
	if dg.GetElevenBackActive() {
		res += "【11バック】強さが逆転中\n"
	}
	if dg.GetSuitLocked() {
		res += fmt.Sprintf("【スート縛り】%s\n", p.getSuitName(dg.GetLockedSuit()))
	}
	if dg.GetTableIsSequence() {
		res += "【階段】\n"
	}

	// カード交換記録
	exchangeActions := dg.GetExchangeActions()
	if len(exchangeActions) > 0 {
		res += "[カード交換]\n"
		for _, ex := range exchangeActions {
			cardStrs := make([]string, len(ex.Cards))
			for i, c := range ex.Cards {
				cardStrs[i] = p.getCardStr(c)
			}
			res += fmt.Sprintf("%s → %s: %s\n",
				p.getPlayerName(dg, ex.FromPlayerIdx),
				p.getPlayerName(dg, ex.ToPlayerIdx),
				strings.Join(cardStrs, ", "))
		}
	}

	// 場のカード
	tableCards := dg.GetTableCards()
	if len(tableCards) > 0 {
		cardStrs := make([]string, len(tableCards))
		for i, c := range tableCards {
			cardStrs[i] = p.getCardStr(c)
		}
		res += fmt.Sprintf("場: %s (出したプレイヤー: %s)\n",
			strings.Join(cardStrs, ", "),
			p.getPlayerName(dg, dg.GetLastPlayPlayerIdx()))
	} else {
		res += "場: なし (誰でも出せます)\n"
	}

	// 人間の前の行動
	humanAction := dg.GetHumanAction()
	if humanAction != nil {
		if len(humanAction.PlayedCards) == 0 {
			res += fmt.Sprintf("%sがパスしました\n", p.getPlayerName(dg, humanAction.PlayerIdx))
		} else {
			cardStrs := make([]string, len(humanAction.PlayedCards))
			for i, c := range humanAction.PlayedCards {
				cardStrs[i] = p.getCardStr(c)
			}
			res += fmt.Sprintf("%sが %s を出しました\n",
				p.getPlayerName(dg, humanAction.PlayerIdx),
				strings.Join(cardStrs, ", "))
		}
	}

	// CPUの行動履歴を表示
	cpuActions := dg.GetCpuActions()
	if len(cpuActions) > 0 {
		res += "[CPUの行動]\n"
		for _, action := range cpuActions {
			actPlayerName := p.getPlayerName(dg, action.PlayerIdx)
			if len(action.PlayedCards) == 0 {
				res += fmt.Sprintf("%sがパスしました\n", actPlayerName)
			} else {
				cardStrs := make([]string, len(action.PlayedCards))
				for i, c := range action.PlayedCards {
					cardStrs[i] = p.getCardStr(c)
				}
				res += fmt.Sprintf("%sが %s を出しました\n", actPlayerName, strings.Join(cardStrs, ", "))
			}
		}
	}

	if dg.GetGameEndFlag() {
		res += "ゲーム終了！\n"
		for i := 0; i < dg.GetPlayerCnt(); i++ {
			player := dg.GetPlayer(i)
			res += fmt.Sprintf("  %s: %s\n", p.getPlayerName(dg, i), p.rankName(player.GetRank()))
		}
	} else {
		currentTurn := dg.GetCurrentTurn()
		currentName := p.getPlayerName(dg, currentTurn)
		res += fmt.Sprintf("手番: %s\n", currentName)
		res += "p [インデックス...] でカードを出す / p でパス\n"
	}

	res += "==========\n"
	return res
}

// getPlayerName プレイヤー名取得
func (p *DaifugoCuiPresenter) getPlayerName(dg *domain.Daifugo, idx int) string {
	player := dg.GetPlayer(idx)
	if player == nil {
		return "不明"
	}
	if player.GetIsHuman() {
		return "あなた"
	}
	return fmt.Sprintf("CPU %d", idx)
}

// getCardStr カード情報文字列取得
func (p *DaifugoCuiPresenter) getCardStr(card *domain.Card) string {
	if card == nil {
		return "??"
	}
	switch card.GetDesign() {
	case domain.CardDesignSpade:
		return "SPADE " + strconv.Itoa(card.GetValue())
	case domain.CardDesignClover:
		return "CLOVER " + strconv.Itoa(card.GetValue())
	case domain.CardDesignHeart:
		return "HEART " + strconv.Itoa(card.GetValue())
	case domain.CardDesignDiamond:
		return "DIAMOND " + strconv.Itoa(card.GetValue())
	case domain.CardDesignJoker:
		return "JOKER"
	default:
		return "UNKNOWN"
	}
}

// getSuitName スート名取得
func (p *DaifugoCuiPresenter) getSuitName(suit int) string {
	switch suit {
	case domain.CardDesignSpade:
		return "SPADE"
	case domain.CardDesignClover:
		return "CLOVER"
	case domain.CardDesignHeart:
		return "HEART"
	case domain.CardDesignDiamond:
		return "DIAMOND"
	default:
		return "不明"
	}
}

// rankName ランク名取得
func (p *DaifugoCuiPresenter) rankName(rank int) string {
	switch rank {
	case 1:
		return "大富豪"
	case 2:
		return "富豪"
	case 3:
		return "平民"
	case 4:
		return "大貧民"
	default:
		return "不明"
	}
}
