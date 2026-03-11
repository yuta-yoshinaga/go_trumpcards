package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// BlackJackWebPresenter ブラックジャックWebプレゼンタークラス
type BlackJackWebPresenter struct {
}

// Output ゲーム状態を出力
func (bjp *BlackJackWebPresenter) Output(bj interfaces.BlackJackGame, lastErr error) string {
	resObj := new(controller.BlackJackWebOutput)

	resObj.Dealer = bjp.buildDealerOutput(bj)

	player := bj.GetPlayer()
	resObj.Player = new(controller.BlackJackWebOutputPlayer)
	resObj.Player.Chips = player.GetChips()

	resObj.Phase = bj.GetPhase()
	resObj.CurrentHandIdx = bj.GetCurrentHandIdx()
	resObj.InsuranceBet = bj.GetInsuranceBet()
	resObj.InsuranceAvailable = bj.IsInsuranceAvailable()
	resObj.HintEnabled = bj.IsHintEnabled()
	resObj.SuggestedAction = int(bj.GetBasicStrategySuggestion())
	resObj.DeckCount = bj.GetDeckCount()
	config := bj.GetConfig()
	resObj.DealerHitsSoft17 = config.DealerHitsSoft17
	resObj.CountingEnabled = config.CountingEnabled
	resObj.CpuPlayerCount = config.CpuPlayerCount
	resObj.DoubleAfterSplit = config.DoubleAfterSplit
	resObj.CountingSystem = config.CountingSystem
	resObj.DeckPenetration = bj.GetDeckPenetration()
	resObj.RunningCount = bj.GetRunningCount()
	resObj.TrueCount = bj.GetTrueCount()
	resObj.MultiHandCount = bj.GetMultiHandCount()
	resObj.SurrenderRule = config.SurrenderRule

	resObj.Hands = bjp.buildHandsOutput(bj)
	resObj.CpuPlayers = bjp.buildCpuPlayersOutput(bj)
	resObj.SideBetResults = bjp.buildSideBetsOutput(bj)

	resObj.PerfectPairsBet = bj.GetPerfectPairsBet()
	resObj.TwentyOnePlus3Bet = bj.Get21Plus3Bet()

	resObj.Message, resObj.MessageCode = bjp.buildMessage(bj, lastErr)

	return marshalOrError(resObj)
}

// buildDealerOutput ディーラー情報を構築
func (bjp *BlackJackWebPresenter) buildDealerOutput(bj interfaces.BlackJackGame) *controller.BlackJackWebOutputPlayer {
	dealer := bj.GetDealer()
	out := new(controller.BlackJackWebOutputPlayer)
	out.Cards = make([]*controller.WebOutputCard, 0)
	out.Chips = dealer.GetChips()
	if bj.GetGameEndFlag() {
		out.Score = dealer.GetScore()
		for i := 0; i < dealer.GetCardsSize(); i++ {
			out.Cards = append(out.Cards, cardToOutput(dealer.GetCard(i)))
		}
	} else if dealer.GetCardsSize() > 0 {
		out.Cards = append(out.Cards, cardToOutput(dealer.GetCard(0)))
	}
	return out
}

// buildHandsOutput プレイヤーハンド情報を構築
func (bjp *BlackJackWebPresenter) buildHandsOutput(bj interfaces.BlackJackGame) []*controller.BlackJackWebOutputHand {
	hands := bj.GetPlayerHands()
	out := make([]*controller.BlackJackWebOutputHand, len(hands))
	for i, hand := range hands {
		h := new(controller.BlackJackWebOutputHand)
		h.Score = hand.GetScore()
		h.Cards = make([]*controller.WebOutputCard, 0)
		for j := 0; j < hand.GetCardsSize(); j++ {
			h.Cards = append(h.Cards, cardToOutput(hand.GetCard(j)))
		}
		h.Bet = hand.GetBet()
		h.Stood = hand.IsStood()
		h.Doubled = hand.IsDoubled()
		h.Busted = hand.IsBusted()
		h.IsBlackJack = hand.IsBlackJack()
		h.CanSplit = hand.CanSplit()
		h.Surrendered = hand.IsSurrendered()
		h.CanSurrender = bj.CanSurrenderHand(i)
		out[i] = h
	}
	return out
}

// buildCpuPlayersOutput CPUプレイヤー情報を構築
func (bjp *BlackJackWebPresenter) buildCpuPlayersOutput(bj interfaces.BlackJackGame) []*controller.BlackJackWebOutputCpuSeat {
	cpuPlayers := bj.GetCpuPlayers()
	if len(cpuPlayers) == 0 {
		return nil
	}
	out := make([]*controller.BlackJackWebOutputCpuSeat, len(cpuPlayers))
	for i, cpu := range cpuPlayers {
		seat := new(controller.BlackJackWebOutputCpuSeat)
		seat.Chips = cpu.GetPlayer().GetChips()
		seat.InsuranceBet = cpu.GetInsuranceBet()
		seat.Hands = make([]*controller.BlackJackWebOutputHand, len(cpu.GetHands()))
		for j, hand := range cpu.GetHands() {
			h := new(controller.BlackJackWebOutputHand)
			h.Score = hand.GetScore()
			h.Cards = make([]*controller.WebOutputCard, 0)
			if bj.GetGameEndFlag() || bj.GetPhase() != domain.BJPhaseBet {
				for k := 0; k < hand.GetCardsSize(); k++ {
					h.Cards = append(h.Cards, cardToOutput(hand.GetCard(k)))
				}
			}
			h.Bet = hand.GetBet()
			h.Stood = hand.IsStood()
			h.Doubled = hand.IsDoubled()
			h.Busted = hand.IsBusted()
			h.IsBlackJack = hand.IsBlackJack()
			h.CanSplit = hand.CanSplit()
			h.Surrendered = hand.IsSurrendered()
			h.CanSurrender = bj.CanSurrenderCpuHand(i, j)
			seat.Hands[j] = h
		}
		out[i] = seat
	}
	return out
}

// buildSideBetsOutput サイドベット情報を構築
func (bjp *BlackJackWebPresenter) buildSideBetsOutput(bj interfaces.BlackJackGame) []*controller.BlackJackWebOutputSideBetResult {
	sideBetResults := bj.GetSideBetResults()
	if len(sideBetResults) == 0 {
		return nil
	}
	out := make([]*controller.BlackJackWebOutputSideBetResult, len(sideBetResults))
	for i, r := range sideBetResults {
		out[i] = &controller.BlackJackWebOutputSideBetResult{
			BetType:    r.BetType,
			ResultType: r.ResultType,
			ResultName: r.ResultName,
			BetAmount:  r.BetAmount,
			Payout:     r.Payout,
		}
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (bjp *BlackJackWebPresenter) buildMessage(bj interfaces.BlackJackGame, lastErr error) (string, string) {
	if lastErr != nil {
		return lastErr.Error(), ""
	}
	if bj.GetGameEndFlag() {
		switch bj.GameJudgment() {
		case domain.GameResultDraw:
			return "It is a draw.", "blackjack.result.draw"
		case domain.GameResultWin:
			return "You are the winner.", "blackjack.result.win"
		case domain.GameResultLose:
			return "It is your loss.", "blackjack.result.lose"
		}
	}
	return "", ""
}

// ActionLogOutput 棋譜をJSON出力
func (bjp *BlackJackWebPresenter) ActionLogOutput(bj interfaces.BlackJackGame) string {
	return actionLogOutputJSON(bj)
}
