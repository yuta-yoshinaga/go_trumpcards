package presenter

import (
	"encoding/json"
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
)

// DaifugoWebPresenter 大富豪Webプレゼンタークラス
type DaifugoWebPresenter struct{}

// NewDaifugoWebPresenter コンストラクタ
func NewDaifugoWebPresenter() *DaifugoWebPresenter {
	return &DaifugoWebPresenter{}
}

// Output ゲーム状態をJSON出力
func (dwp *DaifugoWebPresenter) Output(dg *domain.Daifugo) string {
	resObj := new(controller.DaifugoWebOutput)
	resObj.Players = make([]*controller.DaifugoWebOutputPlayer, 0)
	resObj.TableCards = make([]*controller.DaifugoWebOutputCard, 0)
	resObj.CurrentTurn = dg.GetCurrentTurn()
	resObj.LastPlayPlayerIdx = dg.GetLastPlayPlayerIdx()
	resObj.GameEndFlag = dg.GetGameEndFlag()
	resObj.RevolutionActive = dg.GetRevolutionActive()
	resObj.ElevenBackActive = dg.GetElevenBackActive()
	resObj.SuitLocked = dg.GetSuitLocked()
	resObj.LockedSuit = dwp.getSuitName(dg.GetLockedSuit())
	resObj.TableIsSequence = dg.GetTableIsSequence()

	// ローカルルール設定
	config := dg.GetConfig()
	resObj.Config = controller.DaifugoWebOutputConfig{
		JokerCount:          config.JokerCount,
		EightCutEnabled:     config.EightCutEnabled,
		SuitLockEnabled:     config.SuitLockEnabled,
		ElevenBackEnabled:   config.ElevenBackEnabled,
		SequenceEnabled:     config.SequenceEnabled,
		CardExchangeEnabled: config.CardExchangeEnabled,
	}

	// カード交換記録
	resObj.ExchangeActions = make([]*controller.DaifugoWebOutputExchangeAction, 0)
	for _, ex := range dg.GetExchangeActions() {
		exObj := &controller.DaifugoWebOutputExchangeAction{
			FromPlayerIdx: ex.FromPlayerIdx,
			ToPlayerIdx:   ex.ToPlayerIdx,
			Cards:         dwp.getCardObjs(ex.Cards),
		}
		resObj.ExchangeActions = append(resObj.ExchangeActions, exObj)
	}

	// 場のカード
	for _, c := range dg.GetTableCards() {
		resObj.TableCards = append(resObj.TableCards, dwp.getCardObj(c))
	}

	// CPU行動履歴
	resObj.CpuActions = make([]*controller.DaifugoWebOutputAction, 0)
	for _, action := range dg.GetCpuActions() {
		a := &controller.DaifugoWebOutputAction{
			PlayerIdx:   action.PlayerIdx,
			PlayedCards: dwp.getCardObjs(action.PlayedCards),
		}
		resObj.CpuActions = append(resObj.CpuActions, a)
	}

	// 人間の最後の行動
	humanAction := dg.GetHumanAction()
	if humanAction != nil {
		resObj.HumanAction = &controller.DaifugoWebOutputAction{
			PlayerIdx:   humanAction.PlayerIdx,
			PlayedCards: dwp.getCardObjs(humanAction.PlayedCards),
		}
	}

	// プレイヤー情報
	for i := 0; i < dg.GetPlayerCnt(); i++ {
		player := dg.GetPlayer(i)
		pObj := new(controller.DaifugoWebOutputPlayer)
		pObj.ID = i
		pObj.IsHuman = player.GetIsHuman()
		pObj.IsFinished = player.GetIsFinished()
		pObj.Rank = player.GetRank()
		pObj.CardCount = player.GetCardsSize()
		pObj.Cards = make([]*controller.DaifugoWebOutputCard, 0)
		if player.GetIsHuman() {
			for j := 0; j < player.GetCardsSize(); j++ {
				pObj.Cards = append(pObj.Cards, dwp.getCardObj(player.GetCard(j)))
			}
		}
		resObj.Players = append(resObj.Players, pObj)
	}

	// メッセージ (ゲーム終了時)
	if dg.GetGameEndFlag() {
		resObj.Message = dwp.buildResultMessage(dg)
	}

	res, _ := json.Marshal(resObj)
	return string(res)
}

// buildResultMessage ゲーム終了メッセージを生成
func (dwp *DaifugoWebPresenter) buildResultMessage(dg *domain.Daifugo) string {
	rankNames := []string{"大富豪", "富豪", "平民", "大貧民"}
	msg := "ゲーム終了！ "
	for i := 0; i < dg.GetPlayerCnt(); i++ {
		player := dg.GetPlayer(i)
		rank := player.GetRank()
		if rank < 1 || rank > len(rankNames) {
			continue
		}
		var name string
		if player.GetIsHuman() {
			name = "あなた"
		} else {
			name = fmt.Sprintf("CPU %d", i)
		}
		msg += fmt.Sprintf("%s:%s ", name, rankNames[rank-1])
	}
	return msg
}

// getSuitName スート名取得
func (dwp *DaifugoWebPresenter) getSuitName(suit int) string {
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
		return ""
	}
}

// getCardObjs カードオブジェクトの配列取得 (nil → nil)
func (dwp *DaifugoWebPresenter) getCardObjs(cards []*domain.Card) []*controller.DaifugoWebOutputCard {
	if cards == nil {
		return nil
	}
	result := make([]*controller.DaifugoWebOutputCard, len(cards))
	for i, c := range cards {
		result[i] = dwp.getCardObj(c)
	}
	return result
}

// getCardObj カード情報オブジェクト取得
func (dwp *DaifugoWebPresenter) getCardObj(card *domain.Card) *controller.DaifugoWebOutputCard {
	if card == nil {
		return nil
	}
	res := new(controller.DaifugoWebOutputCard)
	switch card.GetDesign() {
	case domain.CardDesignSpade:
		res.Design = "SPADE"
	case domain.CardDesignClover:
		res.Design = "CLOVER"
	case domain.CardDesignHeart:
		res.Design = "HEART"
	case domain.CardDesignDiamond:
		res.Design = "DIAMOND"
	default:
		res.Design = "JOKER"
	}
	res.Value = card.GetValue()
	return res
}
