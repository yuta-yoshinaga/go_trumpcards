//go:build !js || !wasm || extra

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// KingoWebPresenter キンゴWebプレゼンタークラス
type KingoWebPresenter struct{}

// Output ゲーム状態を出力
//
// **配列は必ず配列で返します。**
func (cp *KingoWebPresenter) Output(c interfaces.KingoGame, lastErr error) string {
	resObj := new(controller.KingoWebOutput)

	resObj.Phase = int(c.GetPhase())
	resObj.Seats = kingoSeatsToOutput(c)
	resObj.BankerSeat = c.GetBankerSeat()
	resObj.RoundNumber = c.GetRoundNumber()
	resObj.Rounds = c.GetConfig().Rounds
	resObj.HumanSeat = c.HumanSeat()
	resObj.IsHumanBanker = c.IsHumanBanker()
	resObj.IsHumanTurn = c.IsHumanTurn()
	resObj.HandSize = domain.KingoHandSize
	// **配当はサーバが載せる。** 画面に倍率を書き写させると、役の出にくさと
	// 表示が別々に育って食い違う。
	resObj.PayoutArashi = domain.KingoPayout(domain.KingoRankArashi)
	resObj.PayoutPair = domain.KingoPayout(domain.KingoRankPair)
	resObj.RemainingCards = c.GetRemainingCards()
	resObj.WinnerSeat = c.WinnerSeat()
	resObj.GameEndFlag = c.GetGameEndFlag()
	cfg := c.GetConfig()
	resObj.Config = &controller.KingoWebOutCfg{
		Seats: cfg.Seats, InitialChips: cfg.InitialChips,
		MinBet: cfg.MinBet, Rounds: cfg.Rounds,
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if c.GetGameEndFlag() {
		resObj.MessageCode = "kingo.result.gameEnd"
	}

	return marshalOrError(resObj)
}

// kingoSeatsToOutput は席ごとの状態を組み立てる。
//
// **張りの段階では誰の手札も送らない。** 配る前なので手札は存在せず、
// 決着まで進んで初めて全員ぶんが出る ── キンゴに「自分だけ見える手札」は無い。
func kingoSeatsToOutput(c interfaces.KingoGame) []*controller.KingoWebOutputSeat {
	players := c.GetPlayers()
	results := c.GetResults()
	shown := c.GetPhase() != domain.KingoPhaseBet

	out := make([]*controller.KingoWebOutputSeat, 0, len(players))
	for i, p := range players {
		if p == nil {
			continue
		}
		seat := &controller.KingoWebOutputSeat{
			Name:     p.GetName(),
			IsHuman:  p.GetIsHuman(),
			Chips:    p.GetChips(),
			Bet:      p.GetBet(),
			Cards:    make([]*controller.WebOutputCard, 0),
			IsBanker: i == c.GetBankerSeat(),
		}
		if shown && len(p.GetCards()) > 0 {
			seat.Cards = cardsToOutputOrEmpty(p.GetCards())
			seat.Rank = int(p.GetRank())
			seat.MatchedValue = domain.KingoMatchedValue(p.GetCards())
		}
		if i < len(results) {
			seat.WonAmount = results[i].WonAmount
		}
		out = append(out, seat)
	}
	return out
}

// ActionLogOutput 棋譜をJSON出力
func (cp *KingoWebPresenter) ActionLogOutput(c interfaces.KingoGame) string {
	return actionLogOutputJSON(c)
}

// HintOutput ヒントをJSON出力
func (cp *KingoWebPresenter) HintOutput(c interfaces.KingoGame) string {
	h := c.GetHint()
	if h == nil {
		return marshalOrError(map[string]any{"hint": nil})
	}
	return marshalOrError(map[string]any{"action": h.Action, "amount": h.Amount, "reason": h.Reason})
}
