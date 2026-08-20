//go:build !js || !wasm || extra4

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// ColourWhistWebPresenter カラーホイストWebプレゼンタークラス
type ColourWhistWebPresenter struct{}

// Output ゲーム状態を出力
//
// **配列は必ず配列で返します。** 空のトリックを素の変換に通すと JSON が `null` に
// なり、TS 側が非 optional な配列を約束しているのでページが落ちます。
func (cp *ColourWhistWebPresenter) Output(c interfaces.ColourWhistGame, lastErr error) string {
	resObj := new(controller.ColourWhistWebOutput)

	resObj.Players = colourWhistPlayersToOutput(c)
	resObj.Phase = c.GetPhase()
	resObj.ValidPlays = intSliceOrEmpty(c.GetValidPlayIndices(0))
	resObj.DealerIdx = c.GetDealerIdx()
	resObj.Contract = c.GetContract()
	resObj.DeclarerIdx = c.GetDeclarerIdx()
	resObj.PartnerIdx = c.GetPartnerIdx()
	resObj.CalledCard = cardToOutput(c.GetCalledCard())
	resObj.TrumpSuit = c.GetTrumpSuit()
	resObj.TroelForced = c.IsTroelForced()
	resObj.CurrentTurn = c.GetCurrentTurn()
	resObj.IsHumanTurn = c.IsHumanTurn()
	resObj.CurrentTrick = colourWhistTrickOrEmpty(c.GetTrick())
	resObj.LastTrick = colourWhistTrickOrEmpty(c.GetLastTrick())
	resObj.LastTrickWinner = c.GetLastTrickWinner()
	resObj.TrickCount = c.GetTrickCount()
	resObj.DeclarerTricks = c.GetDeclarerTricks()
	resObj.RoundNumber = c.GetRoundNumber()
	resObj.GameEndFlag = c.GetGameEndFlag()
	resObj.WinnerIdx = c.GetWinnerIdx()
	resObj.Config = &controller.ColourWhistWebOutCfg{Rounds: c.GetConfig().Rounds}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if c.GetGameEndFlag() {
		resObj.MessageCode = colourWhistEndMessageCode(c)
	}

	return marshalOrError(resObj)
}

// colourWhistTrickOrEmpty はトリックを必ず配列にして返す。
func colourWhistTrickOrEmpty(trick []*domain.TrickCard) []*controller.WebOutputTrickCard {
	out := trickCardsToOutput(trick)
	if out == nil {
		return make([]*controller.WebOutputTrickCard, 0)
	}
	return out
}

// colourWhistPlayersToOutput は席の情報を組み立てる。
//
// **手札の中身は人間の席だけ。**
func colourWhistPlayersToOutput(c interfaces.ColourWhistGame) []*controller.ColourWhistWebOutputPlayer {
	out := make([]*controller.ColourWhistWebOutputPlayer, 0, c.GetPlayerCnt())
	for i := range c.GetPlayerCnt() {
		p := c.GetPlayer(i)
		if p == nil {
			continue
		}
		entry := &controller.ColourWhistWebOutputPlayer{
			ID:             i,
			IsHuman:        p.GetIsHuman(),
			CardCount:      p.GetCardsSize(),
			Cards:          make([]*controller.WebOutputCard, 0),
			TrickCount:     p.GetTrickCount(),
			Score:          p.GetScore(),
			IsDeclarerSide: c.IsDeclarerSideVisible(i),
			HasPassed:      c.HasPassed(i),
		}
		if p.GetIsHuman() {
			cards := make([]*domain.Card, 0, p.GetCardsSize())
			for k := range p.GetCardsSize() {
				cards = append(cards, p.GetCard(k))
			}
			entry.Cards = cardsToOutputOrEmpty(cards)
		}
		out = append(out, entry)
	}
	return out
}

// colourWhistEndMessageCode は終局時の i18n キーを返す。
func colourWhistEndMessageCode(c interfaces.ColourWhistGame) string {
	if c.GetWinnerIdx() == 0 {
		return "colourwhist.result.win"
	}
	return "colourwhist.result.lose"
}

// ActionLogOutput 棋譜をJSON出力
func (cp *ColourWhistWebPresenter) ActionLogOutput(c interfaces.ColourWhistGame) string {
	return actionLogOutputJSON(c)
}

// HintOutput ヒントをJSON出力
func (cp *ColourWhistWebPresenter) HintOutput(c interfaces.ColourWhistGame) string {
	h := c.GetHint()
	if h == nil {
		return marshalOrError(map[string]any{"hint": nil})
	}
	return marshalOrError(map[string]any{
		"contract":  h.Contract,
		"cardIndex": h.CardIndex,
		"reason":    h.Reason,
	})
}
