//go:build !js || !wasm || classic

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// BotifarraWebPresenter ボティファラWebプレゼンタークラス
type BotifarraWebPresenter struct{}

// Output ゲーム状態を出力
//
// **配列は必ず配列で返します。** 空のトリックや手札を素の変換に通すと JSON が
// `null` になり、TS 側が非 optional な配列を約束しているのでページが落ちます。
func (bp *BotifarraWebPresenter) Output(b interfaces.BotifarraGame, lastErr error) string {
	resObj := new(controller.BotifarraWebOutput)

	resObj.Players = botifarraPlayersToOutput(b)
	resObj.Phase = b.GetPhase()
	resObj.ValidPlays = intSliceOrEmpty(b.GetValidPlayIndices(0))
	resObj.DealerIdx = b.GetDealerIdx()
	resObj.DeclarerIdx = b.GetDeclarerIdx()
	resObj.TrumpSuit = b.GetTrumpSuit()
	resObj.Multiplier = b.GetMultiplier()
	resObj.CurrentTurn = b.GetCurrentTurn()
	resObj.IsHumanTurn = b.IsHumanTurn()
	resObj.CurrentTrick = botifarraTrickOrEmpty(b.GetTrick())
	resObj.LastTrick = botifarraTrickOrEmpty(b.GetLastTrick())
	resObj.LastTrickWinner = b.GetLastTrickWinner()
	resObj.TrickCount = b.GetTrickCount()
	resObj.RoundPoints = []int{b.GetRoundPoints(0), b.GetRoundPoints(1)}
	resObj.Scores = []int{b.GetScore(0), b.GetScore(1)}
	resObj.GameEndFlag = b.GetGameEndFlag()
	resObj.WinnerTeam = b.GetWinnerTeam()

	cfg := b.GetConfig()
	resObj.Config = &controller.BotifarraWebOutConfig{
		TargetScore:   cfg.TargetScore,
		AllowDoubling: cfg.AllowDoubling,
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if b.GetGameEndFlag() {
		resObj.MessageCode = botifarraEndMessageCode(b)
	}

	return marshalOrError(resObj)
}

// botifarraTrickOrEmpty はトリックを必ず配列にして返す。
func botifarraTrickOrEmpty(trick []*domain.TrickCard) []*controller.WebOutputTrickCard {
	out := trickCardsToOutput(trick)
	if out == nil {
		return make([]*controller.WebOutputTrickCard, 0)
	}
	return out
}

// botifarraPlayersToOutput は席の情報を組み立てる。
//
// **手札の中身は人間の席だけ。** 相手の手札まで返すと、そのまま覗けてしまいます。
func botifarraPlayersToOutput(b interfaces.BotifarraGame) []*controller.BotifarraWebOutputPlayer {
	out := make([]*controller.BotifarraWebOutputPlayer, 0, b.GetPlayerCnt())
	for i := range b.GetPlayerCnt() {
		p := b.GetPlayer(i)
		if p == nil {
			continue
		}
		entry := &controller.BotifarraWebOutputPlayer{
			ID:         i,
			IsHuman:    p.GetIsHuman(),
			Team:       domain.BotifarraTeamOf(i),
			CardCount:  p.GetCardsSize(),
			Cards:      make([]*controller.WebOutputCard, 0),
			TrickCount: p.GetTrickCount(),
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

// botifarraEndMessageCode は終局時の i18n キーを返す。
func botifarraEndMessageCode(b interfaces.BotifarraGame) string {
	if b.GetWinnerTeam() == domain.BotifarraTeamOf(0) {
		return "botifarra.result.win"
	}
	return "botifarra.result.lose"
}

// ActionLogOutput 棋譜をJSON出力
func (bp *BotifarraWebPresenter) ActionLogOutput(b interfaces.BotifarraGame) string {
	return actionLogOutputJSON(b)
}

// HintOutput ヒントをJSON出力
func (bp *BotifarraWebPresenter) HintOutput(b interfaces.BotifarraGame) string {
	h := b.GetHint()
	if h == nil {
		return marshalOrError(map[string]any{"hint": nil})
	}
	return marshalOrError(map[string]any{
		"cardIndex": h.CardIndex,
		"suit":      h.Suit,
		"reason":    h.Reason,
	})
}
