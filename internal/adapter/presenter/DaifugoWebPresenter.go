package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// DaifugoWebPresenter 大富豪Webプレゼンタークラス
type DaifugoWebPresenter struct{}

// NewDaifugoWebPresenter コンストラクタ
func NewDaifugoWebPresenter() *DaifugoWebPresenter {
	return &DaifugoWebPresenter{}
}

// Output ゲーム状態をJSON出力
func (dwp *DaifugoWebPresenter) Output(dg interfaces.DaifugoGame, lastErr error) string {
	resObj := new(controller.DaifugoWebOutput)
	resObj.Players = make([]*controller.DaifugoWebOutputPlayer, 0)
	resObj.TableCards = make([]*controller.WebOutputCard, 0)
	resObj.CurrentTurn = dg.GetCurrentTurn()
	resObj.LastPlayPlayerIdx = dg.GetLastPlayPlayerIdx()
	resObj.GameEndFlag = dg.GetGameEndFlag()
	resObj.RevolutionActive = dg.GetRevolutionActive()
	resObj.ElevenBackActive = dg.GetElevenBackActive()
	resObj.SuitLocked = dg.GetSuitLocked()
	resObj.LockedSuit = daifugoSuitName(dg.GetLockedSuit())
	resObj.TableIsSequence = dg.GetTableIsSequence()

	// ローカルルール設定
	config := dg.GetConfig()
	resObj.Config = controller.DaifugoWebConfig{
		JokerCount:                config.JokerCount,
		EightCutEnabled:           config.EightCutEnabled,
		SuitLockMode:              int(config.SuitLockMode),
		ElevenBackEnabled:         config.ElevenBackEnabled,
		SequenceEnabled:           config.SequenceEnabled,
		CardExchangeEnabled:       config.CardExchangeEnabled,
		FiveSkipEnabled:           config.FiveSkipEnabled,
		FiveSkipCount:             config.FiveSkipCount,
		SevenPassEnabled:          config.SevenPassEnabled,
		TenDiscardEnabled:         config.TenDiscardEnabled,
		SpadeThreeEnabled:         config.SpadeThreeEnabled,
		CapitalFallEnabled:        config.CapitalFallEnabled,
		NineReverseEnabled:        config.NineReverseEnabled,
		CoupDetatEnabled:          config.CoupDetatEnabled,
		NumberLockEnabled:         config.NumberLockEnabled,
		SandstormEnabled:          config.SandstormEnabled,
		EmperorEnabled:            config.EmperorEnabled,
		SequenceRevolutionEnabled: config.SequenceRevolutionEnabled,
		IllegalFinishEnabled:      config.IllegalFinishEnabled,
		QueenBomberEnabled:        config.QueenBomberEnabled,
		CpuDifficulty:             int(config.CpuDifficulty),
	}

	resObj.ReverseDirection = dg.GetReverseDirection()
	resObj.NumberLocked = dg.GetNumberLocked()
	resObj.SortMode = int(dg.GetSortMode())

	// ペンディングアクション
	switch dg.GetPendingActionType() {
	case domain.DaifugoPendingSevenPass:
		resObj.PendingAction = "sevenPass"
	case domain.DaifugoPendingTenDiscard:
		resObj.PendingAction = "tenDiscard"
	case domain.DaifugoPendingQueenBomber:
		resObj.PendingAction = "queenBomber"
	default:
		resObj.PendingAction = "none"
	}
	resObj.PendingActionTarget = dg.GetPendingActionTarget()

	// カード交換記録
	resObj.ExchangeActions = make([]*controller.DaifugoWebOutputExchangeAction, 0)
	for _, ex := range dg.GetExchangeActions() {
		exObj := &controller.DaifugoWebOutputExchangeAction{
			FromPlayerIdx: ex.FromPlayerIdx,
			ToPlayerIdx:   ex.ToPlayerIdx,
			Cards:         cardsToOutput(ex.Cards),
		}
		resObj.ExchangeActions = append(resObj.ExchangeActions, exObj)
	}

	// 場のカード
	for _, c := range dg.GetTableCards() {
		resObj.TableCards = append(resObj.TableCards, cardToOutput(c))
	}

	// CPU行動履歴
	resObj.CpuActions = make([]*controller.DaifugoWebOutputAction, 0)
	for _, action := range dg.GetCpuActions() {
		a := &controller.DaifugoWebOutputAction{
			PlayerIdx:   action.PlayerIdx,
			PlayedCards: cardsToOutput(action.PlayedCards),
		}
		resObj.CpuActions = append(resObj.CpuActions, a)
	}

	// 人間の最後の行動
	humanAction := dg.GetHumanAction()
	if humanAction != nil {
		resObj.HumanAction = &controller.DaifugoWebOutputAction{
			PlayerIdx:   humanAction.PlayerIdx,
			PlayedCards: cardsToOutput(humanAction.PlayedCards),
		}
	}

	// プレイヤー情報
	for i := 0; i < dg.GetPlayerCnt(); i++ {
		player := dg.GetPlayer(i)
		if player == nil {
			continue
		}
		pObj := new(controller.DaifugoWebOutputPlayer)
		pObj.ID = i
		pObj.IsHuman = player.GetIsHuman()
		pObj.IsFinished = player.GetIsFinished()
		pObj.Rank = player.GetRank()
		pObj.CardCount = player.GetCardsSize()
		pObj.IllegalFinishPenalty = player.GetIllegalFinishPenalty()
		pObj.Cards = make([]*controller.WebOutputCard, 0)
		if player.GetIsHuman() {
			for j := 0; j < player.GetCardsSize(); j++ {
				pObj.Cards = append(pObj.Cards, cardToOutput(player.GetCard(j)))
			}
		}
		resObj.Players = append(resObj.Players, pObj)
	}

	// エラーメッセージ
	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if dg.GetGameEndFlag() {
		resObj.Message = dwp.buildResultMessage(dg)
		resObj.MessageCode = "daifugo.result.rankings"
		resObj.MessageParams = map[string]string{"rankings": resObj.Message}
	}

	return marshalOrError(resObj)
}

// buildResultMessage ゲーム終了メッセージを生成
func (dwp *DaifugoWebPresenter) buildResultMessage(dg interfaces.DaifugoGame) string {
	msg := "ゲーム終了！ "
	for i := 0; i < dg.GetPlayerCnt(); i++ {
		player := dg.GetPlayer(i)
		if player == nil {
			continue
		}
		rank := player.GetRank()
		if rank < daifugoRankMin || rank > daifugoRankMax {
			continue
		}
		var name string
		if player.GetIsHuman() {
			name = "あなた"
		} else {
			name = fmt.Sprintf("CPU %d", i)
		}
		msg += fmt.Sprintf("%s:%s ", name, daifugoRankName(rank))
	}
	return msg
}

// ActionLogOutput 棋譜をJSON出力
func (dwp *DaifugoWebPresenter) ActionLogOutput(dg interfaces.DaifugoGame) string {
	return actionLogOutputJSON(dg)
}
