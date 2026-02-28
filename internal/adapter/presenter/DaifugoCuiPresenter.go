package presenter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// DaifugoCuiPresenter 大富豪CUIプレゼンタークラス
type DaifugoCuiPresenter struct{}

// NewDaifugoCuiPresenter コンストラクタ
func NewDaifugoCuiPresenter() *DaifugoCuiPresenter {
	return &DaifugoCuiPresenter{}
}

// Output ゲーム状態を文字列出力
func (p *DaifugoCuiPresenter) Output(dg interfaces.DaifugoGame, lastErr error) string {
	var b strings.Builder

	b.WriteString("==========\n")
	b.WriteString("Daifugo (大富豪)\n")
	b.WriteString("==========\n")

	for i := 0; i < dg.GetPlayerCnt(); i++ {
		player := dg.GetPlayer(i)
		if player.GetIsHuman() {
			b.WriteString("[You]")
		} else {
			fmt.Fprintf(&b, "CPU %d", i)
		}
		if player.GetIsFinished() {
			fmt.Fprintf(&b, ": 上がり (ランク: %s)\n", p.rankName(player.GetRank()))
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

	// ローカルルール状態
	if dg.GetRevolutionActive() {
		b.WriteString("【革命中】2が最弱、3が最強\n")
	}
	if dg.GetElevenBackActive() {
		b.WriteString("【11バック】強さが逆転中\n")
	}
	if dg.GetSuitLocked() {
		fmt.Fprintf(&b, "【スート縛り】%s\n", p.getSuitName(dg.GetLockedSuit()))
	}
	if dg.GetTableIsSequence() {
		b.WriteString("【階段】\n")
	}

	// カード交換記録
	exchangeActions := dg.GetExchangeActions()
	if len(exchangeActions) > 0 {
		b.WriteString("[カード交換]\n")
		for _, ex := range exchangeActions {
			cardStrs := make([]string, len(ex.Cards))
			for i, c := range ex.Cards {
				cardStrs[i] = p.getCardStr(c)
			}
			fmt.Fprintf(&b, "%s → %s: %s\n",
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
		fmt.Fprintf(&b, "場: %s (出したプレイヤー: %s)\n",
			strings.Join(cardStrs, ", "),
			p.getPlayerName(dg, dg.GetLastPlayPlayerIdx()))
	} else {
		b.WriteString("場: なし (誰でも出せます)\n")
	}

	// 人間の前の行動
	humanAction := dg.GetHumanAction()
	if humanAction != nil {
		if len(humanAction.PlayedCards) == 0 {
			fmt.Fprintf(&b, "%sがパスしました\n", p.getPlayerName(dg, humanAction.PlayerIdx))
		} else {
			cardStrs := make([]string, len(humanAction.PlayedCards))
			for i, c := range humanAction.PlayedCards {
				cardStrs[i] = p.getCardStr(c)
			}
			fmt.Fprintf(&b, "%sが %s を出しました\n",
				p.getPlayerName(dg, humanAction.PlayerIdx),
				strings.Join(cardStrs, ", "))
		}
	}

	// CPUの行動履歴を表示
	cpuActions := dg.GetCpuActions()
	if len(cpuActions) > 0 {
		b.WriteString("[CPUの行動]\n")
		for _, action := range cpuActions {
			actPlayerName := p.getPlayerName(dg, action.PlayerIdx)
			if len(action.PlayedCards) == 0 {
				fmt.Fprintf(&b, "%sがパスしました\n", actPlayerName)
			} else {
				cardStrs := make([]string, len(action.PlayedCards))
				for i, c := range action.PlayedCards {
					cardStrs[i] = p.getCardStr(c)
				}
				fmt.Fprintf(&b, "%sが %s を出しました\n", actPlayerName, strings.Join(cardStrs, ", "))
			}
		}
	}

	// エラーメッセージ
	if lastErr != nil {
		fmt.Fprintf(&b, "%s\n", lastErr.Error())
	}

	if dg.GetGameEndFlag() {
		b.WriteString("ゲーム終了！\n")
		for i := 0; i < dg.GetPlayerCnt(); i++ {
			player := dg.GetPlayer(i)
			fmt.Fprintf(&b, "  %s: %s\n", p.getPlayerName(dg, i), p.rankName(player.GetRank()))
		}
	} else {
		currentTurn := dg.GetCurrentTurn()
		currentName := p.getPlayerName(dg, currentTurn)
		fmt.Fprintf(&b, "手番: %s\n", currentName)
		switch dg.GetPendingActionType() {
		case domain.DaifugoPendingSevenPass:
			b.WriteString("【7渡し】渡すカードを選択してください (p [インデックス])\n")
		case domain.DaifugoPendingTenDiscard:
			b.WriteString("【10捨て】捨てるカードを選択してください (p [インデックス])\n")
		default:
			b.WriteString("p [インデックス...] でカードを出す / p でパス\n")
		}
	}

	b.WriteString("==========\n")
	return b.String()
}

// getPlayerName プレイヤー名取得
func (p *DaifugoCuiPresenter) getPlayerName(dg interfaces.DaifugoGame, idx int) string {
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
