package presenters

import (
	"encoding/json"
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/entities"
	"github.com/yuta-yoshinaga/go_trumpcards/interface_adapters/controllers"
)

// SevensWebPresenter 7並べWebプレゼンタークラス
type SevensWebPresenter struct{}

// NewSevensWebPresenter コンストラクタ
func NewSevensWebPresenter() *SevensWebPresenter {
	return &SevensWebPresenter{}
}

// Output ゲーム状態をJSON出力
func (swp *SevensWebPresenter) Output(s *entities.Sevens) string {
	resObj := new(controllers.SevensWebOutput)
	resObj.Players = make([]*controllers.SevensWebOutputPlayer, 0)
	resObj.CurrentTurn = s.GetCurrentTurn()
	resObj.TableMinVals = s.GetTableMinVals()
	resObj.TableMaxVals = s.GetTableMaxVals()
	resObj.GameEndFlag = s.GetGameEndFlag()

	// CPU行動履歴
	resObj.CpuActions = make([]*controllers.SevensWebOutputAction, 0)
	for _, action := range s.GetCpuActions() {
		a := &controllers.SevensWebOutputAction{
			PlayerIdx:  action.PlayerIdx,
			PlayedCard: swp.getCardObj(action.PlayedCard),
		}
		resObj.CpuActions = append(resObj.CpuActions, a)
	}

	// 人間の最後の行動
	humanAction := s.GetHumanAction()
	if humanAction != nil {
		resObj.HumanAction = &controllers.SevensWebOutputAction{
			PlayerIdx:  humanAction.PlayerIdx,
			PlayedCard: swp.getCardObj(humanAction.PlayedCard),
		}
	}

	// プレイヤー情報
	for i := 0; i < s.GetPlayerCnt(); i++ {
		player := s.GetPlayer(i)
		pObj := new(controllers.SevensWebOutputPlayer)
		pObj.ID = i
		pObj.IsHuman = player.GetIsHuman()
		pObj.IsFinished = player.GetIsFinished()
		pObj.Rank = player.GetRank()
		pObj.CardCount = player.GetCardsSize()
		pObj.PassesUsed = player.GetPassesUsed()
		pObj.MaxPasses = player.GetMaxPasses()
		pObj.Cards = make([]*controllers.SevensWebOutputCard, 0)
		if player.GetIsHuman() {
			for j := 0; j < player.GetCardsSize(); j++ {
				pObj.Cards = append(pObj.Cards, swp.getCardObj(player.GetCard(j)))
			}
		}
		resObj.Players = append(resObj.Players, pObj)
	}

	// メッセージ (ゲーム終了時)
	if s.GetGameEndFlag() {
		resObj.Message = swp.buildResultMessage(s)
	}

	res, _ := json.Marshal(resObj)
	return string(res)
}

// buildResultMessage ゲーム終了メッセージを生成
func (swp *SevensWebPresenter) buildResultMessage(s *entities.Sevens) string {
	msg := "ゲーム終了！ "
	for i := 0; i < s.GetPlayerCnt(); i++ {
		player := s.GetPlayer(i)
		rank := player.GetRank()
		if rank < 1 {
			continue
		}
		var name string
		if player.GetIsHuman() {
			name = "あなた"
		} else {
			name = fmt.Sprintf("CPU %d", i)
		}
		msg += fmt.Sprintf("%s:%d位 ", name, rank)
	}
	return msg
}

// getCardObj カード情報オブジェクト取得 (nil → nil)
func (swp *SevensWebPresenter) getCardObj(card *entities.Card) *controllers.SevensWebOutputCard {
	if card == nil {
		return nil
	}
	res := new(controllers.SevensWebOutputCard)
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
