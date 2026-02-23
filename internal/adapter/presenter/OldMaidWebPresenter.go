package presenter

import (
	"encoding/json"
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
)

// OldMaidWebPresenter ババ抜きWebプレゼンタークラス
type OldMaidWebPresenter struct{}

// NewOldMaidWebPresenter コンストラクタ
func NewOldMaidWebPresenter() *OldMaidWebPresenter {
	return &OldMaidWebPresenter{}
}

// Output ゲーム状態をJSON出力
func (owp *OldMaidWebPresenter) Output(om *domain.OldMaid, lastErr error) string {
	resObj := new(controller.OldMaidWebOutput)
	resObj.Players = make([]*controller.OldMaidWebOutputPlayer, 0)
	resObj.CurrentTurn = om.GetCurrentTurn()
	resObj.NextDrawTargetIdx = om.GetNextDrawTargetIdx()
	resObj.GameEndFlag = om.GetGameEndFlag()
	resObj.LoserIdx = om.GetLoserIdx()
	resObj.LastDrawPlayerIdx = om.GetLastDrawPlayerIdx()
	resObj.LastDrawFromIdx = om.GetLastDrawFromIdx()
	// Only reveal drawn card for human players to preserve CPU game fairness
	lastDrawPlayerIdx := om.GetLastDrawPlayerIdx()
	lastDrawPlayer := om.GetPlayer(lastDrawPlayerIdx)
	if lastDrawPlayer != nil && lastDrawPlayer.GetIsHuman() {
		resObj.LastDrawCard = owp.getCardObj(om.GetLastDrawCard())
	}
	resObj.LastDiscardedPairs = om.GetLastDiscardedPairs()
	resObj.LastDiscardedCards = make([]*controller.OldMaidWebOutputCard, 0)
	for _, card := range om.GetLastDiscardedCards() {
		resObj.LastDiscardedCards = append(resObj.LastDiscardedCards, owp.getCardObj(card))
	}
	resObj.HasDrawn = om.GetHasDrawn()

	// CPU行動履歴
	resObj.CpuActions = make([]*controller.OldMaidWebOutputCpuAction, 0)
	for _, action := range om.GetCpuActions() {
		a := &controller.OldMaidWebOutputCpuAction{
			DrawPlayerIdx:  action.DrawPlayerIdx,
			DrawFromIdx:    action.DrawFromIdx,
			DrawnCard:      nil, // CPU drawn card is hidden to preserve game fairness
			DiscardedPairs: action.DiscardedPairs,
			DiscardedCards: make([]*controller.OldMaidWebOutputCard, 0),
		}
		for _, card := range action.DiscardedCards {
			a.DiscardedCards = append(a.DiscardedCards, owp.getCardObj(card))
		}
		resObj.CpuActions = append(resObj.CpuActions, a)
	}

	// 人間プレイヤーの行動記録
	if ha := om.GetHumanAction(); ha != nil {
		haObj := &controller.OldMaidWebOutputCpuAction{
			DrawPlayerIdx:  ha.DrawPlayerIdx,
			DrawFromIdx:    ha.DrawFromIdx,
			DrawnCard:      owp.getCardObj(ha.DrawnCard),
			DiscardedPairs: ha.DiscardedPairs,
			DiscardedCards: make([]*controller.OldMaidWebOutputCard, 0),
		}
		for _, card := range ha.DiscardedCards {
			haObj.DiscardedCards = append(haObj.DiscardedCards, owp.getCardObj(card))
		}
		resObj.HumanAction = haObj
	}

	for i := 0; i < om.GetPlayerCnt(); i++ {
		player := om.GetPlayer(i)
		pObj := new(controller.OldMaidWebOutputPlayer)
		pObj.ID = i
		pObj.IsHuman = player.GetIsHuman()
		pObj.IsFinished = player.GetIsFinished()
		pObj.CardCount = player.GetCardsSize()
		pObj.Cards = make([]*controller.OldMaidWebOutputCard, 0)
		if player.GetIsHuman() {
			for j := 0; j < player.GetCardsSize(); j++ {
				pObj.Cards = append(pObj.Cards, owp.getCardObj(player.GetCard(j)))
			}
		}
		resObj.Players = append(resObj.Players, pObj)
	}

	// エラーメッセージ
	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if om.GetGameEndFlag() {
		loserIdx := om.GetLoserIdx()
		if loserIdx >= 0 {
			loser := om.GetPlayer(loserIdx)
			if loser != nil && loser.GetIsHuman() {
				resObj.Message = "ゲーム終了！ あなたの負け！"
			} else {
				resObj.Message = fmt.Sprintf("ゲーム終了！ CPU %dの負け！", loserIdx)
			}
		}
	}

	res, err := json.Marshal(resObj)
	if err != nil {
		return `{"error":"internal server error"}`
	}
	return string(res)
}

// getCardObj カード情報オブジェクト取得 (nil の場合は nil を返す)
func (owp *OldMaidWebPresenter) getCardObj(card *domain.Card) *controller.OldMaidWebOutputCard {
	if card == nil {
		return nil
	}
	res := new(controller.OldMaidWebOutputCard)
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
