//go:build !js || !wasm || extra2

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// RikkenWebPresenter リッケンWebプレゼンタークラス
type RikkenWebPresenter struct{}

// Output ゲーム状態を出力
//
// **配列は必ず配列で返します。** 空のトリックを素の変換に通すと JSON が `null` に
// なり、TS 側が非 optional な配列を約束しているのでページが落ちます。
func (rp *RikkenWebPresenter) Output(r interfaces.RikkenGame, lastErr error) string {
	resObj := new(controller.RikkenWebOutput)

	resObj.Players = rikkenPlayersToOutput(r)
	resObj.Phase = r.GetPhase()
	resObj.ValidPlays = intSliceOrEmpty(r.GetValidPlayIndices(0))
	resObj.DealerIdx = r.GetDealerIdx()
	resObj.Contract = r.GetContract()
	resObj.DeclarerIdx = r.GetDeclarerIdx()
	resObj.PartnerIdx = r.GetPartnerIdx()
	resObj.CalledCard = cardToOutput(r.GetCalledCard())
	resObj.TrumpSuit = r.GetTrumpSuit()
	resObj.CurrentTurn = r.GetCurrentTurn()
	resObj.IsHumanTurn = r.IsHumanTurn()
	resObj.CurrentTrick = rikkenTrickOrEmpty(r.GetTrick())
	resObj.LastTrick = rikkenTrickOrEmpty(r.GetLastTrick())
	resObj.LastTrickWinner = r.GetLastTrickWinner()
	resObj.TrickCount = r.GetTrickCount()
	resObj.DeclarerTricks = r.GetDeclarerTricks()
	resObj.RoundNumber = r.GetRoundNumber()
	resObj.GameEndFlag = r.GetGameEndFlag()
	resObj.WinnerIdx = r.GetWinnerIdx()
	resObj.Config = &controller.RikkenWebOutConfig{Rounds: r.GetConfig().Rounds}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if r.GetGameEndFlag() {
		resObj.MessageCode = rikkenEndMessageCode(r)
	}

	return marshalOrError(resObj)
}

// rikkenTrickOrEmpty はトリックを必ず配列にして返す。
func rikkenTrickOrEmpty(trick []*domain.TrickCard) []*controller.WebOutputTrickCard {
	out := trickCardsToOutput(trick)
	if out == nil {
		return make([]*controller.WebOutputTrickCard, 0)
	}
	return out
}

// rikkenPlayersToOutput は席の情報を組み立てる。
//
// **手札の中身は人間の席だけ。** 相手の手札まで返すとそのまま覗けてしまいます。
func rikkenPlayersToOutput(r interfaces.RikkenGame) []*controller.RikkenWebOutputPlayer {
	out := make([]*controller.RikkenWebOutputPlayer, 0, r.GetPlayerCnt())
	for i := range r.GetPlayerCnt() {
		p := r.GetPlayer(i)
		if p == nil {
			continue
		}
		entry := &controller.RikkenWebOutputPlayer{
			ID:             i,
			IsHuman:        p.GetIsHuman(),
			CardCount:      p.GetCardsSize(),
			Cards:          make([]*controller.WebOutputCard, 0),
			TrickCount:     p.GetTrickCount(),
			Score:          p.GetScore(),
			IsDeclarerSide: r.IsDeclarerSide(i),
			HasPassed:      r.HasPassed(i),
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

// rikkenEndMessageCode は終局時の i18n キーを返す。
func rikkenEndMessageCode(r interfaces.RikkenGame) string {
	if r.GetWinnerIdx() == 0 {
		return "rikken.result.win"
	}
	return "rikken.result.lose"
}

// ActionLogOutput 棋譜をJSON出力
func (rp *RikkenWebPresenter) ActionLogOutput(r interfaces.RikkenGame) string {
	return actionLogOutputJSON(r)
}

// HintOutput ヒントをJSON出力
func (rp *RikkenWebPresenter) HintOutput(r interfaces.RikkenGame) string {
	h := r.GetHint()
	if h == nil {
		return marshalOrError(map[string]any{"hint": nil})
	}
	return marshalOrError(map[string]any{
		"contract":  h.Contract,
		"cardIndex": h.CardIndex,
		"reason":    h.Reason,
	})
}
