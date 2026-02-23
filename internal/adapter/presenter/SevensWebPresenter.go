package presenter

import (
	"encoding/json"
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
)

// SevensWebPresenter 7並べWebプレゼンタークラス
type SevensWebPresenter struct{}

// NewSevensWebPresenter コンストラクタ
func NewSevensWebPresenter() *SevensWebPresenter {
	return &SevensWebPresenter{}
}

// Output ゲーム状態をJSON出力
func (swp *SevensWebPresenter) Output(s *domain.Sevens) string {
	resObj := new(controller.SevensWebOutput)
	resObj.Players = make([]*controller.SevensWebOutputPlayer, 0)
	resObj.CurrentTurn = s.GetCurrentTurn()
	resObj.TableMinVals = s.GetTableMinVals()
	resObj.TableMaxVals = s.GetTableMaxVals()
	resObj.GameEndFlag = s.GetGameEndFlag()

	// ボードビットマスク (uint16 → int に変換)
	placed := s.GetTablePlaced()
	for i := 0; i < 5; i++ {
		resObj.TablePlaced[i] = int(placed[i])
	}

	// ゲーム設定
	cfg := s.GetConfig()
	resObj.Config = controller.SevensWebOutputConfig{
		TunnelEnabled: cfg.TunnelEnabled,
		JokerCount:    cfg.JokerCount,
		CpuStrategy:   cfg.CpuStrategy,
	}

	// CPU行動履歴
	resObj.CpuActions = make([]*controller.SevensWebOutputAction, 0)
	for _, action := range s.GetCpuActions() {
		a := &controller.SevensWebOutputAction{
			PlayerIdx:   action.PlayerIdx,
			PlayedCard:  swp.getCardObj(action.PlayedCard),
			TargetSuit:  action.TargetSuit,
			TargetValue: action.TargetValue,
		}
		resObj.CpuActions = append(resObj.CpuActions, a)
	}

	// 人間の最後の行動
	humanAction := s.GetHumanAction()
	if humanAction != nil {
		resObj.HumanAction = &controller.SevensWebOutputAction{
			PlayerIdx:   humanAction.PlayerIdx,
			PlayedCard:  swp.getCardObj(humanAction.PlayedCard),
			TargetSuit:  humanAction.TargetSuit,
			TargetValue: humanAction.TargetValue,
		}
	}

	// プレイヤー情報
	for i := 0; i < s.GetPlayerCnt(); i++ {
		player := s.GetPlayer(i)
		pObj := new(controller.SevensWebOutputPlayer)
		pObj.ID = i
		pObj.IsHuman = player.GetIsHuman()
		pObj.IsFinished = player.GetIsFinished()
		pObj.Rank = player.GetRank()
		pObj.CardCount = player.GetCardsSize()
		pObj.PassesUsed = player.GetPassesUsed()
		pObj.MaxPasses = player.GetMaxPasses()
		pObj.Cards = make([]*controller.SevensWebOutputCard, 0)
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

	res, err := json.Marshal(resObj)
	if err != nil {
		return `{"error":"internal server error"}`
	}
	return string(res)
}

// buildResultMessage ゲーム終了メッセージを生成
func (swp *SevensWebPresenter) buildResultMessage(s *domain.Sevens) string {
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
func (swp *SevensWebPresenter) getCardObj(card *domain.Card) *controller.SevensWebOutputCard {
	if card == nil {
		return nil
	}
	res := new(controller.SevensWebOutputCard)
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
