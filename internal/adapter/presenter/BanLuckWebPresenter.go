//go:build !js || !wasm || casino

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// BanLuckWebPresenter バンラックWebプレゼンタークラス
type BanLuckWebPresenter struct{}

// Output ゲーム状態を出力
//
// **配列は必ず配列で返します。**
func (cp *BanLuckWebPresenter) Output(c interfaces.BanLuckGame, lastErr error) string {
	resObj := new(controller.BanLuckWebOutput)

	resObj.Phase = int(c.GetPhase())
	resObj.Seats = banLuckSeatsToOutput(c)
	resObj.BankerSeat = c.GetBankerSeat()
	resObj.TurnSeat = c.GetTurnSeat()
	resObj.HumanSeat = c.GetHumanSeat()
	resObj.IsHumanTurn = c.IsHumanTurn()
	resObj.MustHit = c.MustHit()
	resObj.RoundNumber = c.GetRoundNumber()
	resObj.RemainingCards = c.GetRemainingCards()
	resObj.WinnerSeat = c.WinnerSeat()
	resObj.GameEndFlag = c.GetGameEndFlag()
	cfg := c.GetConfig()
	resObj.Config = &controller.BanLuckWebOutCfg{
		Seats: cfg.Seats, InitialChips: cfg.InitialChips,
		Rounds: cfg.Rounds, DefaultBet: cfg.DefaultBet,
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if c.GetGameEndFlag() {
		resObj.MessageCode = "banluck.result.gameEnd"
	}

	return marshalOrError(resObj)
}

// banLuckSeatsToOutput は席ごとの状態を組み立てる。
//
// **どの席の札も伏せない。** 全員が親と比べるだけで、席同士は競らないので、
// 隠す情報が無い ── 隠すと「なぜ自分が負けたか」が読めなくなるだけ。
func banLuckSeatsToOutput(c interfaces.BanLuckGame) []*controller.BanLuckWebOutputSeat {
	players := c.GetPlayers()
	hands := c.GetHands()
	results := c.GetResults()
	out := make([]*controller.BanLuckWebOutputSeat, 0, len(players))
	for i, p := range players {
		if p == nil {
			continue
		}
		seat := &controller.BanLuckWebOutputSeat{
			Name:     p.GetName(),
			IsHuman:  p.GetIsHuman(),
			Chips:    p.GetChips(),
			Bet:      p.GetBet(),
			Cards:    make([]*controller.WebOutputCard, 0),
			IsBanker: i == c.GetBankerSeat(),
			IsTurn:   i == c.GetTurnSeat() && c.GetPhase() == domain.BanLuckPhasePlay,
		}
		if i < len(hands) && hands[i] != nil {
			seat.Cards = cardsToOutputOrEmpty(hands[i].GetCards())
			seat.Score = hands[i].GetScore()
			seat.Busted = hands[i].IsBusted()
			seat.Stood = hands[i].IsStood()
		}
		if i < len(results) {
			seat.Rank = int(results[i].Rank)
			seat.Outcome = int(results[i].Outcome)
			seat.RoundBet = results[i].Bet
			seat.Delta = results[i].Delta
		}
		out = append(out, seat)
	}
	return out
}

// ActionLogOutput 棋譜をJSON出力
func (cp *BanLuckWebPresenter) ActionLogOutput(c interfaces.BanLuckGame) string {
	return actionLogOutputJSON(c)
}

// HintOutput ヒントをJSON出力
func (cp *BanLuckWebPresenter) HintOutput(c interfaces.BanLuckGame) string {
	h := c.GetHint()
	if h == nil {
		return marshalOrError(map[string]any{"hint": nil})
	}
	return marshalOrError(map[string]any{"action": h.Action, "reason": h.Reason})
}
