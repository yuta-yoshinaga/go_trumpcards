//go:build !js || !wasm || extra

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// SpeculationWebPresenter スペキュレーションWebプレゼンタークラス
type SpeculationWebPresenter struct {
}

// Output ゲーム状態を出力
func (cp *SpeculationWebPresenter) Output(c interfaces.SpeculationGame, lastErr error) string {
	resObj := new(controller.SpeculationWebOutput)

	resObj.Phase = int(c.GetPhase())
	resObj.Seats = speculationSeatsToOutput(c)
	resObj.TrumpSuit = c.GetTrumpSuit()
	if tc := c.GetTrumpCard(); tc != nil {
		resObj.TrumpCard = cardToOutput(tc)
	}
	resObj.Pot = c.GetPot()
	resObj.TurnSeat = c.GetTurnSeat()
	resObj.BestSeat = c.GetBestSeat()
	resObj.OfferFrom = c.GetOfferFrom()
	resObj.OfferTo = c.GetOfferTo()
	resObj.OfferAmount = c.GetOfferAmount()
	resObj.RoundNo = c.GetRoundNo()
	resObj.WinnerSeat = c.GetWinnerSeat()
	resObj.GameEndFlag = c.GetGameEndFlag()

	cfg := c.GetConfig()
	resObj.Config = &controller.SpeculationWebOutCfg{
		Players:      cfg.Players,
		InitialChips: cfg.InitialChips,
		Stake:        cfg.Stake,
		Rounds:       cfg.Rounds,
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	}
	return marshalOrError(resObj)
}

// speculationSeatsToOutput は各席の公開情報を組み立てる。
//
// **伏せ札の中身は載せない。** 枚数だけを返す —— 中身を送ると、競りでいくら
// 出すべきかがネットワーク越しに丸見えになり、賭けが賭けでなくなる。
func speculationSeatsToOutput(c interfaces.SpeculationGame) []*controller.SpeculationWebOutputSeat {
	players := c.GetPlayers()
	seats := make([]*controller.SpeculationWebOutputSeat, len(players))
	for i, p := range players {
		s := &controller.SpeculationWebOutputSeat{
			Name:        p.GetName(),
			Chips:       p.GetChips(),
			HiddenCount: p.GetHiddenCount(),
		}
		if b := p.GetBest(); b != nil {
			s.Best = cardToOutput(b)
		}
		seats[i] = s
	}
	return seats
}

// ActionLogOutput 棋譜をJSON出力
func (cp *SpeculationWebPresenter) ActionLogOutput(c interfaces.SpeculationGame) string {
	return actionLogOutputJSON(c)
}

// HintOutput はヒントを返す。Web GUI はクライアント側で助言を組み立てるので
// 盤面をそのまま返す。
func (cp *SpeculationWebPresenter) HintOutput(c interfaces.SpeculationGame) string {
	return cp.Output(c, nil)
}
