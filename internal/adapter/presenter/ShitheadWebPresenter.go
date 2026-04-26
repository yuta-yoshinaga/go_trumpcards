package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// ShitheadWebPresenter シットヘッドWebプレゼンタークラス
type ShitheadWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (swp *ShitheadWebPresenter) Output(sg interfaces.ShitheadGame, lastErr error) string {
	resObj := new(controller.ShitheadWebOutput)
	resObj.Players = make([]*controller.ShitheadWebOutputPlayer, 0)
	resObj.CurrentTurn = sg.GetCurrentTurn()
	resObj.CurrentSource = sg.CurrentSource()
	resObj.GameEndFlag = sg.GetGameEndFlag()
	resObj.SkipNext = sg.GetSkipNext()
	resObj.SevenActive = sg.GetSevenActive()
	resObj.StockSize = sg.GetStockSize()

	cfg := sg.GetConfig()
	resObj.Config = controller.ShitheadWebConfig{
		MagicTwo:        cfg.MagicTwo,
		MagicSeven:      cfg.MagicSeven,
		MagicEight:      cfg.MagicEight,
		MagicTen:        cfg.MagicTen,
		FourOfAKindBurn: cfg.FourOfAKindBurn,
		CpuDifficulty:   int(cfg.CpuDifficulty),
	}

	resObj.DiscardPile = cardsToOutputOrEmpty(sg.GetDiscardPile())

	resObj.CpuActions = make([]*controller.ShitheadWebOutputAction, 0)
	for _, a := range sg.GetCpuActions() {
		resObj.CpuActions = append(resObj.CpuActions, shitheadActionToOutput(a))
	}
	if humanAction := sg.GetHumanAction(); humanAction != nil {
		resObj.HumanAction = shitheadActionToOutput(humanAction)
	}

	for i := 0; i < sg.GetPlayerCnt(); i++ {
		player := sg.GetPlayer(i)
		if player == nil {
			continue
		}
		out := &controller.ShitheadWebOutputPlayer{
			ID:            i,
			IsHuman:       player.GetIsHuman(),
			IsFinished:    player.GetIsFinished(),
			Rank:          player.GetRank(),
			HandCount:     player.GetCardsSize(),
			FaceDownCount: player.GetFaceDownSize(),
			FaceUpCards:   cardsToOutputOrEmpty(player.GetFaceUpCards()),
		}
		if player.GetIsHuman() {
			out.HandCards = playerCardsToOutput(player, true)
		} else {
			out.HandCards = make([]*controller.WebOutputCard, 0)
		}
		resObj.Players = append(resObj.Players, out)
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if sg.GetGameEndFlag() {
		resObj.Message = swp.buildResultMessage(sg)
		resObj.MessageCode = "shithead.result.rankings"
		resObj.MessageParams = map[string]string{"rankings": resObj.Message}
	}

	return marshalOrError(resObj)
}

func shitheadActionToOutput(a *domain.ShitheadCpuAction) *controller.ShitheadWebOutputAction {
	played := make([]*controller.WebOutputCard, 0)
	if !a.Pickup {
		played = cardsToOutput(a.PlayedCards)
	}
	return &controller.ShitheadWebOutputAction{
		PlayerIdx:   a.PlayerIdx,
		Source:      a.Source,
		PlayedCards: played,
		Pickup:      a.Pickup,
		Burned:      a.Burned,
		Skipped:     a.Skipped,
	}
}

// buildResultMessage ゲーム終了メッセージ
func (swp *ShitheadWebPresenter) buildResultMessage(sg interfaces.ShitheadGame) string {
	msg := "ゲーム終了！ "
	for i := 0; i < sg.GetPlayerCnt(); i++ {
		player := sg.GetPlayer(i)
		if player == nil {
			continue
		}
		var name string
		if player.GetIsHuman() {
			name = "あなた"
		} else {
			name = fmt.Sprintf("CPU %d", i)
		}
		rankStr := fmt.Sprintf("rank=%d", player.GetRank())
		if player.GetRank() == sg.GetPlayerCnt() {
			rankStr += "(Shithead)"
		}
		msg += fmt.Sprintf("%s:%s ", name, rankStr)
	}
	return msg
}

// ActionLogOutput 棋譜をJSON出力
func (swp *ShitheadWebPresenter) ActionLogOutput(sg interfaces.ShitheadGame) string {
	return actionLogOutputJSON(sg)
}
