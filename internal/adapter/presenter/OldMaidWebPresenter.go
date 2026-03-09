package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// OldMaidWebPresenter ババ抜きWebプレゼンタークラス
type OldMaidWebPresenter struct{}

// NewOldMaidWebPresenter コンストラクタ
func NewOldMaidWebPresenter() *OldMaidWebPresenter {
	return &OldMaidWebPresenter{}
}

// Output ゲーム状態をJSON出力
func (owp *OldMaidWebPresenter) Output(om interfaces.OldMaidGame, lastErr error) string {
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
		resObj.LastDrawCard = cardToOutput(om.GetLastDrawCard())
	}
	resObj.LastDiscardedPairs = om.GetLastDiscardedPairs()
	resObj.LastDiscardedCards = make([]*controller.WebOutputCard, 0)
	for _, card := range om.GetLastDiscardedCards() {
		resObj.LastDiscardedCards = append(resObj.LastDiscardedCards, cardToOutput(card))
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
			DiscardedCards: make([]*controller.WebOutputCard, 0),
			HesitationMs:   action.HesitationMs,
		}
		for _, card := range action.DiscardedCards {
			a.DiscardedCards = append(a.DiscardedCards, cardToOutput(card))
		}
		resObj.CpuActions = append(resObj.CpuActions, a)
	}

	// 人間プレイヤーの行動記録
	if ha := om.GetHumanAction(); ha != nil {
		haObj := &controller.OldMaidWebOutputCpuAction{
			DrawPlayerIdx:  ha.DrawPlayerIdx,
			DrawFromIdx:    ha.DrawFromIdx,
			DrawnCard:      cardToOutput(ha.DrawnCard),
			DiscardedPairs: ha.DiscardedPairs,
			DiscardedCards: make([]*controller.WebOutputCard, 0),
		}
		for _, card := range ha.DiscardedCards {
			haObj.DiscardedCards = append(haObj.DiscardedCards, cardToOutput(card))
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
		pObj.Cards = make([]*controller.WebOutputCard, 0)
		if player.GetIsHuman() {
			for j := 0; j < player.GetCardsSize(); j++ {
				pObj.Cards = append(pObj.Cards, cardToOutput(player.GetCard(j)))
			}
		}
		resObj.Players = append(resObj.Players, pObj)
	}

	// ゲーム全体の引き履歴
	resObj.DrawHistory = make([]*controller.OldMaidWebOutputDrawHistoryEntry, 0)
	for _, entry := range om.GetDrawHistory() {
		resObj.DrawHistory = append(resObj.DrawHistory, &controller.OldMaidWebOutputDrawHistoryEntry{
			DrawPlayerIdx:  entry.DrawPlayerIdx,
			DrawFromIdx:    entry.DrawFromIdx,
			DiscardedPairs: entry.DiscardedPairs,
			DrawerFinished: entry.DrawerFinished,
			TargetFinished: entry.TargetFinished,
		})
	}

	// CPU心理戦: 強調カードインデックスとモード
	resObj.CpuHighlightedCardIdx = om.GetCpuHighlightedCardIdx()
	resObj.Mode = int(om.GetConfig().Mode)

	// ジジ抜き: ゲーム終了時に除外カードを公開
	if om.GetGameEndFlag() && om.GetConfig().Mode == domain.OldMaidModeJijiNuki {
		resObj.RemovedCard = cardToOutput(om.GetRemovedCard())
	}

	// メタAI情報
	if profile := om.GetHumanProfile(); profile != nil {
		resObj.MetaAI = &controller.OldMaidWebOutputMetaAI{
			Enabled:      true,
			GamesPlayed:  profile.GamesPlayed,
			EdgePickRate: profile.PickRate(0) + profile.PickRate(2),
		}
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
				resObj.MessageCode = "oldmaid.result.humanLose"
			} else {
				resObj.Message = fmt.Sprintf("ゲーム終了！ CPU %dの負け！", loserIdx)
				resObj.MessageCode = "oldmaid.result.cpuLose"
				resObj.MessageParams = map[string]string{"cpuId": fmt.Sprintf("%d", loserIdx)}
			}
		}
	}

	return marshalOrError(resObj)
}
