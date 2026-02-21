package presenters

import (
	"encoding/json"
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/entities"
	"github.com/yuta-yoshinaga/go_trumpcards/interface_adapters/controllers"
)

// DaifugoWebPresenter 大富豪Webプレゼンタークラス
type DaifugoWebPresenter struct{}

// NewDaifugoWebPresenter コンストラクタ
func NewDaifugoWebPresenter() *DaifugoWebPresenter {
	return &DaifugoWebPresenter{}
}

// Output ゲーム状態をJSON出力
func (dwp *DaifugoWebPresenter) Output(dg *entities.Daifugo) string {
	resObj := new(controllers.DaifugoWebOutput)
	resObj.Players = make([]*controllers.DaifugoWebOutputPlayer, 0)
	resObj.TableCards = make([]*controllers.DaifugoWebOutputCard, 0)
	resObj.CurrentTurn = dg.GetCurrentTurn()
	resObj.LastPlayPlayerIdx = dg.GetLastPlayPlayerIdx()
	resObj.GameEndFlag = dg.GetGameEndFlag()
	resObj.RevolutionActive = dg.GetRevolutionActive()

	// 場のカード
	for _, c := range dg.GetTableCards() {
		resObj.TableCards = append(resObj.TableCards, dwp.getCardObj(c))
	}

	// CPU行動履歴
	resObj.CpuActions = make([]*controllers.DaifugoWebOutputAction, 0)
	for _, action := range dg.GetCpuActions() {
		a := &controllers.DaifugoWebOutputAction{
			PlayerIdx:   action.PlayerIdx,
			PlayedCards: dwp.getCardObjs(action.PlayedCards),
		}
		resObj.CpuActions = append(resObj.CpuActions, a)
	}

	// 人間の最後の行動
	humanAction := dg.GetHumanAction()
	if humanAction != nil {
		resObj.HumanAction = &controllers.DaifugoWebOutputAction{
			PlayerIdx:   humanAction.PlayerIdx,
			PlayedCards: dwp.getCardObjs(humanAction.PlayedCards),
		}
	}

	// プレイヤー情報
	for i := 0; i < dg.GetPlayerCnt(); i++ {
		player := dg.GetPlayer(i)
		pObj := new(controllers.DaifugoWebOutputPlayer)
		pObj.ID = i
		pObj.IsHuman = player.GetIsHuman()
		pObj.IsFinished = player.GetIsFinished()
		pObj.Rank = player.GetRank()
		pObj.CardCount = player.GetCardsSize()
		pObj.Cards = make([]*controllers.DaifugoWebOutputCard, 0)
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
func (dwp *DaifugoWebPresenter) buildResultMessage(dg *entities.Daifugo) string {
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

// getCardObjs カードオブジェクトの配列取得 (nil → nil)
func (dwp *DaifugoWebPresenter) getCardObjs(cards []*entities.Card) []*controllers.DaifugoWebOutputCard {
	if cards == nil {
		return nil
	}
	result := make([]*controllers.DaifugoWebOutputCard, len(cards))
	for i, c := range cards {
		result[i] = dwp.getCardObj(c)
	}
	return result
}

// getCardObj カード情報オブジェクト取得
func (dwp *DaifugoWebPresenter) getCardObj(card *entities.Card) *controllers.DaifugoWebOutputCard {
	if card == nil {
		return nil
	}
	res := new(controllers.DaifugoWebOutputCard)
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
