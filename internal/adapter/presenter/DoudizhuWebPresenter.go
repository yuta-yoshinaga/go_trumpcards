//go:build !js || !wasm || extra4

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// DoudizhuWebPresenter 斗地主Webプレゼンタークラス
type DoudizhuWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *DoudizhuWebPresenter) Output(dg interfaces.DoudizhuGame, lastErr error) string {
	resObj := new(controller.DoudizhuWebOutput)

	resObj.Phase = doudizhuPhaseName(dg.GetPhase())
	resObj.CurrentTurn = dg.GetCurrentTurn()
	resObj.LandlordIdx = dg.GetLandlordIdx()
	resObj.BaseBid = dg.GetBaseBid()
	resObj.HighestBid = dg.GetHighestBid()
	resObj.BombCount = dg.GetBombCount()
	resObj.Scores = dg.GetScores()
	resObj.GameEndFlag = dg.GetGameEndFlag()

	config := dg.GetConfig()
	resObj.Config = controller.DoudizhuWebConfig{
		CpuDifficulty: int(config.CpuDifficulty),
	}

	resObj.KittyCards = cardsToOutputOrEmpty(dg.GetKittyCards())

	if combo := dg.GetTableCombo(); combo != nil {
		resObj.TableCards = cardsToOutputOrEmpty(combo.Cards)
		resObj.TableCombo = doudizhuComboName(combo.Type)
	} else {
		resObj.TableCards = make([]*controller.WebOutputCard, 0)
	}

	resObj.CpuActions = make([]*controller.DoudizhuWebOutputAction, 0)
	for _, action := range dg.GetCpuActions() {
		a := &controller.DoudizhuWebOutputAction{
			PlayerIdx:   action.PlayerIdx,
			PlayedCards: cardsToOutput(action.PlayedCards),
			BidValue:    action.BidValue,
		}
		resObj.CpuActions = append(resObj.CpuActions, a)
	}

	humanAction := dg.GetHumanAction()
	if humanAction != nil {
		resObj.HumanAction = &controller.DoudizhuWebOutputAction{
			PlayerIdx:   humanAction.PlayerIdx,
			PlayedCards: cardsToOutput(humanAction.PlayedCards),
			BidValue:    humanAction.BidValue,
		}
	}

	resObj.Players = make([]*controller.DoudizhuWebOutputPlayer, 0)
	for i := 0; i < dg.GetPlayerCnt(); i++ {
		player := dg.GetPlayer(i)
		if player == nil {
			continue
		}
		pObj := &controller.DoudizhuWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			IsFinished: player.GetIsFinished(),
			IsLandlord: player.GetIsLandlord(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, player.GetIsHuman()),
		}
		resObj.Players = append(resObj.Players, pObj)
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if dg.GetGameEndFlag() {
		resObj.Message = p.buildResultMessage(dg)
		resObj.MessageCode = "doudizhu.result.summary"
		resObj.MessageParams = map[string]string{"summary": resObj.Message}
	}

	return marshalOrError(resObj)
}

// buildResultMessage ゲーム終了メッセージを生成
func (p *DoudizhuWebPresenter) buildResultMessage(dg interfaces.DoudizhuGame) string {
	scores := dg.GetScores()
	landlordIdx := dg.GetLandlordIdx()
	landlordWon := scores[landlordIdx] > 0

	var winner string
	if landlordWon {
		winner = "地主"
	} else {
		winner = "農民"
	}

	return fmt.Sprintf("%s の勝利！ スコア: %d", winner, scores[landlordIdx])
}

// ActionLogOutput 棋譜をJSON出力
func (p *DoudizhuWebPresenter) ActionLogOutput(dg interfaces.DoudizhuGame) string {
	return actionLogOutputJSON(dg)
}

func doudizhuPhaseName(phase domain.DoudizhuPhase) string {
	switch phase {
	case domain.DoudizhuPhaseBid:
		return "bid"
	case domain.DoudizhuPhasePlay:
		return "play"
	case domain.DoudizhuPhaseEnd:
		return "end"
	default:
		return "unknown"
	}
}

func doudizhuComboName(t domain.DoudizhuComboType) string {
	switch t {
	case domain.DoudizhuComboSingle:
		return "single"
	case domain.DoudizhuComboPair:
		return "pair"
	case domain.DoudizhuComboTrio:
		return "trio"
	case domain.DoudizhuComboTrioSingle:
		return "trioSingle"
	case domain.DoudizhuComboTrioPair:
		return "trioPair"
	case domain.DoudizhuComboStraight:
		return "straight"
	case domain.DoudizhuComboConsecutivePair:
		return "consecutivePair"
	case domain.DoudizhuComboAirplane:
		return "airplane"
	case domain.DoudizhuComboAirplaneSingle:
		return "airplaneSingle"
	case domain.DoudizhuComboAirplanePair:
		return "airplanePair"
	case domain.DoudizhuComboBomb:
		return "bomb"
	case domain.DoudizhuComboRocket:
		return "rocket"
	default:
		return "pass"
	}
}
