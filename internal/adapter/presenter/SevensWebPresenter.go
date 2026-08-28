package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// SevensWebPresenter 7並べWebプレゼンタークラス
type SevensWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (swp *SevensWebPresenter) Output(s interfaces.SevensGame, lastErr error) string {
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
		TunnelEnabled:          cfg.TunnelEnabled,
		TunnelSkipWidth:        cfg.TunnelSkipWidth,
		JokerCount:             cfg.JokerCount,
		CpuStrategy:            cfg.CpuStrategy,
		MaxPasses:              cfg.MaxPasses,
		NoJokerFinish:          cfg.NoJokerFinish,
		JokerReclaimEnabled:    cfg.JokerReclaimEnabled,
		EndStopEnabled:         cfg.EndStopEnabled,
		JokerConsecutiveBanned: cfg.JokerConsecutiveBanned,
	}

	// CPU行動履歴
	resObj.CpuActions = make([]*controller.SevensWebOutputAction, 0)
	for _, action := range s.GetCpuActions() {
		a := &controller.SevensWebOutputAction{
			PlayerIdx:      action.PlayerIdx,
			PlayedCard:     cardToOutput(action.PlayedCard),
			TargetSuit:     action.TargetSuit,
			TargetValue:    action.TargetValue,
			ForcedPass:     action.ForcedPass,
			JokerReclaimed: action.JokerReclaimed,
		}
		resObj.CpuActions = append(resObj.CpuActions, a)
	}

	// 人間の最後の行動
	humanAction := s.GetHumanAction()
	if humanAction != nil {
		resObj.HumanAction = &controller.SevensWebOutputAction{
			PlayerIdx:      humanAction.PlayerIdx,
			PlayedCard:     cardToOutput(humanAction.PlayedCard),
			TargetSuit:     humanAction.TargetSuit,
			TargetValue:    humanAction.TargetValue,
			ForcedPass:     humanAction.ForcedPass,
			JokerReclaimed: humanAction.JokerReclaimed,
		}
	}

	// プレイヤー情報
	for i := 0; i < s.GetPlayerCnt(); i++ {
		player := s.GetPlayer(i)
		if player == nil {
			continue
		}
		pObj := new(controller.SevensWebOutputPlayer)
		pObj.ID = i
		pObj.IsHuman = player.GetIsHuman()
		pObj.IsFinished = player.GetIsFinished()
		pObj.Rank = player.GetRank()
		pObj.CardCount = player.GetCardsSize()
		pObj.PassesUsed = player.GetPassesUsed()
		pObj.MaxPasses = player.GetMaxPasses()
		pObj.Cards = playerCardsToOutput(player, player.GetIsHuman())
		if player.GetIsHuman() {
			pObj.LastPlayedJoker = player.GetLastPlayedJoker()
		}
		resObj.Players = append(resObj.Players, pObj)
	}

	// エラーメッセージ
	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if s.GetGameEndFlag() {
		resObj.Message = swp.buildResultMessage(s)
		resObj.MessageCode = "sevens.result.rankings"
		resObj.MessageParams = map[string]string{"rankings": resObj.Message}
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (swp *SevensWebPresenter) ActionLogOutput(s interfaces.SevensGame) string {
	return actionLogOutputJSON(s)
}

// buildResultMessage ゲーム終了メッセージを生成
func (swp *SevensWebPresenter) buildResultMessage(s interfaces.SevensGame) string {
	msg := "ゲーム終了！ "
	for i := 0; i < s.GetPlayerCnt(); i++ {
		player := s.GetPlayer(i)
		if player == nil {
			continue
		}
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
