//go:build !js || !wasm || casino

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// chemindeFerHumanSeat は人間の席。
const chemindeFerHumanSeat = 0

// ChemindeFerWebPresenter シュマン・ド・フェールWebプレゼンタークラス
type ChemindeFerWebPresenter struct{}

// Output ゲーム状態を出力
//
// **配列は必ず配列で返します。** 配っていない手札を素の変換に通すと JSON が `null`
// になり、TS 側が非 optional な配列を約束しているのでページが落ちます。
func (cp *ChemindeFerWebPresenter) Output(c interfaces.ChemindeFerGame, lastErr error) string {
	resObj := new(controller.ChemindeFerWebOutput)

	resObj.Players = chemindeFerPlayersToOutput(c)
	resObj.Phase = int(c.GetPhase())
	resObj.BankerIdx = c.GetBankerIdx()
	resObj.BetTurn = c.GetBetTurn()
	resObj.Stake = c.GetStake()
	resObj.RemainingStake = c.GetRemainingStake()
	resObj.TotalBet = c.GetTotalBet()
	resObj.StakeMin, resObj.StakeMax = c.StakeRangeFor(c.GetBankerIdx())
	resObj.BetMin, resObj.BetMax = chemindeFerBetRange(c)
	resObj.RepresentativeIdx = c.GetRepresentativeIdx()
	resObj.PunterMayChoose = c.PunterMayChoose()
	resObj.BankerHand = cardsToOutputOrEmpty(c.GetBankerHand())
	resObj.PunterHand = cardsToOutputOrEmpty(c.GetPunterHand())
	resObj.BankerTotal = c.GetBankerTotal()
	resObj.PunterTotal = c.GetPunterTotal()
	resObj.PunterDrew = c.GetPunterDrew()
	resObj.Result = int(c.GetResult())
	resObj.RoundNumber = c.GetRoundNumber()
	resObj.RemainingCards = c.GetRemainingCards()
	resObj.IsHumanTurn = c.IsHumanTurn()
	resObj.GameEndFlag = c.GetGameEndFlag()
	resObj.Config = &controller.ChemindeFerWebOutCfg{
		Rounds:       c.GetConfig().Rounds,
		InitialChips: c.GetConfig().InitialChips,
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if c.GetGameEndFlag() {
		resObj.MessageCode = chemindeFerEndMessageCode(c)
	}

	return marshalOrError(resObj)
}

// chemindeFerBetRange は手番の子が賭けられる額の範囲を返す。
//
// 賭けが終わっていれば (0, 0)。**手番の居ない範囲を返すと、UI が締め切った後の卓に
// 入力欄を出してしまう。**
func chemindeFerBetRange(c interfaces.ChemindeFerGame) (int, int) {
	turn := c.GetBetTurn()
	if turn < 0 {
		return 0, 0
	}
	return c.BetRangeFor(turn)
}

// chemindeFerPlayersToOutput は席の情報を組み立てる。
func chemindeFerPlayersToOutput(c interfaces.ChemindeFerGame) []*controller.ChemindeFerWebOutputPlayer {
	out := make([]*controller.ChemindeFerWebOutputPlayer, 0, domain.ChemindeFerSeatCnt)
	for i := range domain.ChemindeFerSeatCnt {
		p := c.GetPlayer(i)
		if p == nil {
			continue
		}
		out = append(out, &controller.ChemindeFerWebOutputPlayer{
			ID:               i,
			Name:             p.GetName(),
			IsHuman:          p.GetIsHuman(),
			Chips:            p.GetChips(),
			Bet:              p.GetBet(),
			IsBanker:         i == c.GetBankerIdx(),
			IsRepresentative: i == c.GetRepresentativeIdx(),
		})
	}
	return out
}

// chemindeFerEndMessageCode は終局時の i18n キーを返す。
//
// **勝敗はチップで決まる。** 親を何度取ったかではない。
func chemindeFerEndMessageCode(c interfaces.ChemindeFerGame) string {
	me := c.GetPlayer(chemindeFerHumanSeat)
	if me == nil {
		return "chemindefer.result.lose"
	}
	best, tied := me.GetChips(), false
	for i := range domain.ChemindeFerSeatCnt {
		p := c.GetPlayer(i)
		if p == nil || i == chemindeFerHumanSeat {
			continue
		}
		switch {
		case p.GetChips() > best:
			return "chemindefer.result.lose"
		case p.GetChips() == best:
			tied = true
		}
	}
	if tied {
		return "chemindefer.result.draw"
	}
	return "chemindefer.result.win"
}

// ActionLogOutput 棋譜をJSON出力
func (cp *ChemindeFerWebPresenter) ActionLogOutput(c interfaces.ChemindeFerGame) string {
	return actionLogOutputJSON(c)
}

// HintOutput ヒントをJSON出力
func (cp *ChemindeFerWebPresenter) HintOutput(c interfaces.ChemindeFerGame) string {
	h := c.GetHint()
	if h == nil {
		return marshalOrError(map[string]any{"hint": nil})
	}
	return marshalOrError(map[string]any{
		"draw":   h.Draw,
		"reason": h.Reason,
	})
}
