//go:build !js || !wasm || extra2

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// ContinentalRummyWebPresenter はコンチネンタル・ラミーの Web プレゼンター。
type ContinentalRummyWebPresenter struct{}

// Output はゲーム状態を JSON 出力する。
func (p *ContinentalRummyWebPresenter) Output(g interfaces.ContinentalRummyGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	if hint := g.GetHint(); hint != nil {
		resObj.HintDiscardIdx = hint.DiscardIdx
		resObj.HintReason = hint.Reason
	}
	return marshalOrError(resObj)
}

// buildBase は共通フィールドを構築する。
func (p *ContinentalRummyWebPresenter) buildBase(g interfaces.ContinentalRummyGame) *controller.ContinentalRummyWebOutput {
	resObj := new(controller.ContinentalRummyWebOutput)
	resObj.Phase = g.GetPhase()
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.StockCount = g.GetStockCount()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()
	resObj.IsHumanTurn = g.IsHumanTurn()
	// **認められた形は毎回ドメインから返す。** ページ側に書き写すと、
	// 規則が 2 か所に増えてどこかで食い違う。
	resObj.Layouts = domain.ContinentalRummyLayouts()
	resObj.HintDiscardIdx = -1
	if top := g.GetDiscardTop(); top != nil {
		resObj.DiscardTop = cardToOutput(top)
	}
	// **上がれるかはサーバが解く。** 15 枚の分割問題をページ側で解き直さない。
	resObj.GoOutIdx = -1
	if idx, ok := g.CanGoOut(); ok {
		resObj.GoOutIdx = idx
	}
	resObj.LastResult = p.lastResult(g)

	cfg := g.GetConfig()
	resObj.TotalRounds = cfg.TotalRounds
	resObj.Config = controller.ContinentalRummyWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TotalRounds:   cfg.TotalRounds,
	}
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// lastResult は直前のラウンドの決着を構築する。
func (p *ContinentalRummyWebPresenter) lastResult(g interfaces.ContinentalRummyGame) *controller.ContinentalRummyWebOutputResult {
	res := g.GetLastResult()
	if res == nil {
		return nil
	}
	out := &controller.ContinentalRummyWebOutputResult{
		WinnerIdx:   res.WinnerIdx,
		PerOpponent: res.PerOpponent,
		Total:       res.Total,
		Bonuses:     make([]controller.ContinentalRummyWebOutputBonus, 0, len(res.Bonuses)),
	}
	for _, b := range res.Bonuses {
		out.Bonuses = append(out.Bonuses,
			controller.ContinentalRummyWebOutputBonus{Key: b.Key, Points: b.Points})
	}
	return out
}

// buildPlayersOutput は席の情報を構築する。
//
// **相手の手札は枚数だけ。** 中身まで返すと、画面に出していなくても
// レスポンスを覗けば全部見えてしまう。
func (p *ContinentalRummyWebPresenter) buildPlayersOutput(g interfaces.ContinentalRummyGame) []*controller.ContinentalRummyWebOutputPlayer {
	out := make([]*controller.ContinentalRummyWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		entry := &controller.ContinentalRummyWebOutputPlayer{
			ID:        i,
			IsHuman:   player.GetIsHuman(),
			Cards:     make([]*controller.WebOutputCard, 0),
			CardCount: player.GetCardsSize(),
			Melds:     make([][]*controller.WebOutputCard, 0),
			Score:     player.GetScore(),
			IsDealer:  i == g.GetDealerIdx(),
		}
		if player.GetIsHuman() {
			entry.Cards = cardsToOutput(player.GetHand())
		}
		// 並べたシーケンスは全員ぶん見せる。上がった手は公開情報。
		for _, run := range player.GetMelds() {
			entry.Melds = append(entry.Melds, cardsToOutput(run))
		}
		out = append(out, entry)
	}
	return out
}

// buildMessage はフェーズ / 結果メッセージを構築する。
func (p *ContinentalRummyWebPresenter) buildMessage(g interfaces.ContinentalRummyGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		code, params := domain.ErrorMessageCode(lastErr)
		return lastErr.Error(), code, params
	}
	if g.GetGameEndFlag() {
		switch g.GetWinnerIdx() {
		case domain.ContinentalRummyHumanIdx:
			return "", "continentalrummy.result.humanWin", nil
		case -1:
			return "", "continentalrummy.result.draw", nil
		default:
			return "", "continentalrummy.result.cpuWin", nil
		}
	}
	switch g.GetPhase() {
	case domain.ContinentalRummyPhaseDraw:
		return "", "continentalrummy.drawPhase", nil
	case domain.ContinentalRummyPhaseDiscard:
		if _, ok := g.CanGoOut(); ok {
			return "", "continentalrummy.discardPhase.canGoOut", nil
		}
		return "", "continentalrummy.discardPhase", nil
	case domain.ContinentalRummyPhaseRoundEnd:
		if res := g.GetLastResult(); res != nil && res.WinnerIdx < 0 {
			return "", "continentalrummy.roundEnd.washout", nil
		}
		return "", "continentalrummy.roundEnd", nil
	}
	return "", "continentalrummy.drawPhase", nil
}

// HintOutput はヒント情報を JSON 出力する。
func (p *ContinentalRummyWebPresenter) HintOutput(g interfaces.ContinentalRummyGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.HintDiscardIdx = hint.DiscardIdx
		resObj.HintReason = hint.Reason
		resObj.MessageCode = "continentalrummy.hintRequested"
	} else {
		resObj.MessageCode = "continentalrummy.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput は棋譜を JSON 出力する。
func (p *ContinentalRummyWebPresenter) ActionLogOutput(g interfaces.ContinentalRummyGame) string {
	return actionLogOutputJSON(g)
}
