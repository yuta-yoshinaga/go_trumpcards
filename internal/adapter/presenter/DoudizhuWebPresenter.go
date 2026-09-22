//go:build !js || !wasm || extra4

package presenter

import (
	"strconv"

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
		resObj.MessageCode = "doudizhu.result.summary"
		scores := dg.GetScores()
		winnerKey := "doudizhu.peasant"
		if scores[dg.GetLandlordIdx()] > 0 {
			winnerKey = "doudizhu.landlord"
		}
		resObj.MessageParams = map[string]string{
			"winnerKey": winnerKey,
			"score":     strconv.Itoa(scores[dg.GetLandlordIdx()]),
		}
	}

	return marshalOrError(resObj)
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
