package presenters

import (
	"encoding/json"
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/entities"
	"github.com/yuta-yoshinaga/go_trumpcards/interface_adapters/controllers"
)

// OldMaidWebPresenter ババ抜きWebプレゼンタークラス
type OldMaidWebPresenter struct{}

// NewOldMaidWebPresenter コンストラクタ
func NewOldMaidWebPresenter() *OldMaidWebPresenter {
	return &OldMaidWebPresenter{}
}

// Output ゲーム状態をJSON出力
func (owp *OldMaidWebPresenter) Output(om *entities.OldMaid) string {
	resObj := new(controllers.OldMaidWebOutput)
	resObj.Players = make([]*controllers.OldMaidWebOutputPlayer, 0)
	resObj.CurrentTurn = om.GetCurrentTurn()
	resObj.NextDrawTargetIdx = om.GetNextDrawTargetIdx()
	resObj.GameEndFlag = om.GetGameEndFlag()
	resObj.LoserIdx = om.GetLoserIdx()
	resObj.LastDrawPlayerIdx = om.GetLastDrawPlayerIdx()
	resObj.LastDrawFromIdx = om.GetLastDrawFromIdx()
	resObj.LastDrawCard = owp.getCardObj(om.GetLastDrawCard())
	resObj.LastDiscardedPairs = om.GetLastDiscardedPairs()
	resObj.HasDrawn = om.GetHasDrawn()

	// CPU行動履歴
	resObj.CpuActions = make([]*controllers.OldMaidWebOutputCpuAction, 0)
	for _, action := range om.GetCpuActions() {
		a := &controllers.OldMaidWebOutputCpuAction{
			DrawPlayerIdx:  action.DrawPlayerIdx,
			DrawFromIdx:    action.DrawFromIdx,
			DrawnCard:      owp.getCardObj(action.DrawnCard),
			DiscardedPairs: action.DiscardedPairs,
		}
		resObj.CpuActions = append(resObj.CpuActions, a)
	}

	for i := 0; i < om.GetPlayerCnt(); i++ {
		player := om.GetPlayer(i)
		pObj := new(controllers.OldMaidWebOutputPlayer)
		pObj.ID = i
		pObj.IsHuman = player.GetIsHuman()
		pObj.IsFinished = player.GetIsFinished()
		pObj.CardCount = player.GetCardsSize()
		pObj.Cards = make([]*controllers.OldMaidWebOutputCard, 0)
		if player.GetIsHuman() {
			for j := 0; j < player.GetCardsSize(); j++ {
				pObj.Cards = append(pObj.Cards, owp.getCardObj(player.GetCard(j)))
			}
		}
		resObj.Players = append(resObj.Players, pObj)
	}

	if om.GetGameEndFlag() {
		loserIdx := om.GetLoserIdx()
		if loserIdx == 0 {
			resObj.Message = "ゲーム終了！ あなたの負け！"
		} else if loserIdx > 0 {
			resObj.Message = fmt.Sprintf("ゲーム終了！ CPU %dの負け！", loserIdx)
		}
	}

	res, _ := json.Marshal(resObj)
	return string(res)
}

// getCardObj カード情報オブジェクト取得 (nil の場合は nil を返す)
func (owp *OldMaidWebPresenter) getCardObj(card *entities.Card) *controllers.OldMaidWebOutputCard {
	if card == nil {
		return nil
	}
	res := new(controllers.OldMaidWebOutputCard)
	switch card.GetDesign() {
	case entities.CardDesignSpade:
		res.Design = "SPADE"
	case entities.CardDesignClover:
		res.Design = "CLOVER"
	case entities.CardDesignHeart:
		res.Design = "HEART"
	case entities.CardDesignDiamond:
		res.Design = "DIAMOND"
	default:
		res.Design = "JOKER"
	}
	res.Value = card.GetValue()
	return res
}
