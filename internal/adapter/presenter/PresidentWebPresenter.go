package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// PresidentWebPresenter プレジデントWebプレゼンタークラス
type PresidentWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (pwp *PresidentWebPresenter) Output(pg interfaces.PresidentGame, lastErr error) string {
	resObj := new(controller.PresidentWebOutput)
	resObj.Players = make([]*controller.PresidentWebOutputPlayer, 0)
	resObj.CurrentTurn = pg.GetCurrentTurn()
	resObj.LastPlayPlayerIdx = pg.GetLastPlayPlayerIdx()
	resObj.GameEndFlag = pg.GetGameEndFlag()
	resObj.RevolutionActive = pg.GetRevolutionActive()

	config := pg.GetConfig()
	resObj.Config = controller.PresidentWebConfig{
		RevolutionEnabled:     config.RevolutionEnabled,
		CardExchangeEnabled:   config.CardExchangeEnabled,
		PassFieldFlushEnabled: config.PassFieldFlushEnabled,
		CpuDifficulty:         int(config.CpuDifficulty),
	}

	// カード交換記録
	resObj.ExchangeActions = make([]*controller.PresidentWebOutputExchangeAction, 0)
	for _, ex := range pg.GetExchangeActions() {
		resObj.ExchangeActions = append(resObj.ExchangeActions, &controller.PresidentWebOutputExchangeAction{
			FromPlayerIdx: ex.FromPlayerIdx,
			ToPlayerIdx:   ex.ToPlayerIdx,
			Cards:         cardsToOutput(ex.Cards),
		})
	}

	// 場のカード
	resObj.TableCards = cardsToOutputOrEmpty(pg.GetTableCards())

	// CPU行動履歴
	resObj.CpuActions = make([]*controller.PresidentWebOutputAction, 0)
	for _, action := range pg.GetCpuActions() {
		resObj.CpuActions = append(resObj.CpuActions, &controller.PresidentWebOutputAction{
			PlayerIdx:   action.PlayerIdx,
			PlayedCards: cardsToOutput(action.PlayedCards),
		})
	}

	// 人間の最後の行動
	if humanAction := pg.GetHumanAction(); humanAction != nil {
		resObj.HumanAction = &controller.PresidentWebOutputAction{
			PlayerIdx:   humanAction.PlayerIdx,
			PlayedCards: cardsToOutput(humanAction.PlayedCards),
		}
	}

	// プレイヤー情報
	for i := 0; i < pg.GetPlayerCnt(); i++ {
		player := pg.GetPlayer(i)
		if player == nil {
			continue
		}
		pObj := &controller.PresidentWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			IsFinished: player.GetIsFinished(),
			Rank:       player.GetRank(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, player.GetIsHuman()),
		}
		resObj.Players = append(resObj.Players, pObj)
	}

	// メッセージ
	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if pg.GetGameEndFlag() {
		resObj.Message = pwp.buildResultMessage(pg)
		resObj.MessageCode = "president.result.rankings"
		resObj.MessageParams = map[string]string{"rankings": resObj.Message}
	}

	return marshalOrError(resObj)
}

// buildResultMessage ゲーム終了メッセージ
func (pwp *PresidentWebPresenter) buildResultMessage(pg interfaces.PresidentGame) string {
	msg := "ゲーム終了！ "
	for i := 0; i < pg.GetPlayerCnt(); i++ {
		player := pg.GetPlayer(i)
		if player == nil {
			continue
		}
		rank := player.GetRank()
		if rank < 1 || rank > 4 {
			continue
		}
		var name string
		if player.GetIsHuman() {
			name = "あなた"
		} else {
			name = fmt.Sprintf("CPU %d", i)
		}
		msg += fmt.Sprintf("%s:%s ", name, presidentRankName(rank))
	}
	return msg
}

// ActionLogOutput 棋譜をJSON出力
func (pwp *PresidentWebPresenter) ActionLogOutput(pg interfaces.PresidentGame) string {
	return actionLogOutputJSON(pg)
}

// HintOutput ヒントを出力する。Web ではヒントはクライアント側 (useGameHint) で
// 算出するため、通常の状態出力を返す。PresidentPresenter インタフェースを満たすための実装。
func (pwp *PresidentWebPresenter) HintOutput(pg interfaces.PresidentGame) string {
	return pwp.Output(pg, nil)
}
