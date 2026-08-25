//go:build !js || !wasm || extra2

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// BaccaratBanqueWebPresenter はバカラ・バンクの Web プレゼンター。
type BaccaratBanqueWebPresenter struct{}

// Output はゲーム状態を JSON 出力する。
func (p *BaccaratBanqueWebPresenter) Output(g interfaces.BaccaratBanqueGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	if hint := g.GetHint(); hint != nil {
		resObj.HintDraw = hint.Draw
		resObj.HintReason = hint.Reason
	}
	return marshalOrError(resObj)
}

// baccaratBanqueRole は席の役どころを返す。
func baccaratBanqueRole(i int) string {
	switch i {
	case domain.BaccaratBanqueBankerIdx:
		return "banker"
	case domain.BaccaratBanqueRightIdx:
		return "right"
	default:
		return "left"
	}
}

// buildBase は共通フィールドを構築する。
func (p *BaccaratBanqueWebPresenter) buildBase(g interfaces.BaccaratBanqueGame) *controller.BaccaratBanqueWebOutput {
	resObj := new(controller.BaccaratBanqueWebOutput)
	resObj.Phase = g.GetPhase()
	resObj.CoupNumber = g.GetCoupNumber()
	// **バンクの継続数は見せる。** 1 回負けても途切れないことが、この形式の
	// 要なので、続いている数字が見えていないと伝わらない。
	resObj.BankHeld = g.GetBankHeld()
	resObj.ShoeRemaining = g.GetShoeRemaining()
	resObj.Retired = g.IsRetired()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()
	resObj.IsHumanTurn = g.IsHumanTurn()
	resObj.LastResult = p.lastResult(g)

	cfg := g.GetConfig()
	resObj.Config = controller.BaccaratBanqueWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		StartChips:    cfg.StartChips,
		BetAmount:     cfg.BetAmount,
	}
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// lastResult は直前のクーの結果を出力形へ直す。
func (p *BaccaratBanqueWebPresenter) lastResult(g interfaces.BaccaratBanqueGame) *controller.BaccaratBanqueWebOutputResult {
	res := g.GetLastResult()
	if res == nil {
		return nil
	}
	sides := make([]*controller.BaccaratBanqueWebOutputSide, 0, len(res.Sides))
	for _, s := range res.Sides {
		sides = append(sides, &controller.BaccaratBanqueWebOutputSide{
			SeatIdx: s.SeatIdx, Outcome: s.Outcome, Bet: s.Bet, Delta: s.Delta,
		})
	}
	return &controller.BaccaratBanqueWebOutputResult{
		BankerTotal: res.BankerTotal, Sides: sides,
		BankerDelta: res.BankerDelta, BankerNatural: res.BankerNatural,
	}
}

// buildPlayersOutput は席の情報を構築する。
//
// **バカラは全部表向き。** 伏せる札が無いので、3 席とも手札をそのまま出す。
func (p *BaccaratBanqueWebPresenter) buildPlayersOutput(g interfaces.BaccaratBanqueGame) []*controller.BaccaratBanqueWebOutputPlayer {
	out := make([]*controller.BaccaratBanqueWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		out = append(out, &controller.BaccaratBanqueWebOutputPlayer{
			ID:      i,
			IsHuman: player.GetIsHuman(),
			Role:    baccaratBanqueRole(i),
			Cards:   cardsToOutput(player.GetHand()),
			Total:   player.GetTotal(),
			Chips:   player.GetChips(),
			Bet:     player.GetBet(),
			Drawn:   player.HasDrawn(),
		})
	}
	return out
}

// buildMessage はフェーズ / 結果メッセージを構築する。
func (p *BaccaratBanqueWebPresenter) buildMessage(g interfaces.BaccaratBanqueGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		code, params := domain.ErrorMessageCode(lastErr)
		return lastErr.Error(), code, params
	}
	if g.GetGameEndFlag() {
		if g.IsRetired() {
			return "", "baccaratbanque.result.retired", nil
		}
		if g.GetWinnerIdx() == domain.BaccaratBanqueBankerIdx {
			return "", "baccaratbanque.result.bankerAhead", nil
		}
		return "", "baccaratbanque.result.bankerBehind", nil
	}
	switch g.GetPhase() {
	case domain.BaccaratBanquePhaseBanker:
		return "", "baccaratbanque.bankerPhase", nil
	case domain.BaccaratBanquePhaseResult:
		return "", "baccaratbanque.resultPhase", nil
	}
	return "", "baccaratbanque.puntersPhase", nil
}

// HintOutput はヒント情報を JSON 出力する。
func (p *BaccaratBanqueWebPresenter) HintOutput(g interfaces.BaccaratBanqueGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil && hint.Reason != "none" {
		resObj.HintDraw = hint.Draw
		resObj.HintReason = hint.Reason
		resObj.MessageCode = "baccaratbanque.hintRequested"
	} else {
		resObj.MessageCode = "baccaratbanque.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput は棋譜を JSON 出力する。
func (p *BaccaratBanqueWebPresenter) ActionLogOutput(g interfaces.BaccaratBanqueGame) string {
	return actionLogOutputJSON(g)
}
