//go:build !js || !wasm || extra2

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// CuckooWebPresenter Cuckoo Webプレゼンタークラス
type CuckooWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *CuckooWebPresenter) Output(g interfaces.CuckooGame, lastErr error) string {
	resObj := new(controller.CuckooWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.StockCount = g.GetStockCount()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()
	resObj.PendingSwapFrom = g.GetPendingSwapFrom()
	resObj.PendingSwapTo = g.GetPendingSwapTo()
	resObj.RoundLowest = g.GetRoundLowest()
	resObj.RoundLosers = append([]int{}, g.GetRoundLosers()...)

	cfg := g.GetConfig()
	resObj.Config = controller.CuckooWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		InitialLives:  cfg.InitialLives,
	}

	resObj.Players = p.buildPlayersOutput(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)

	return marshalOrError(resObj)
}

// cuckooWebReveal ラウンド終了/ゲーム終了時に全員の手札を公開するか
func cuckooWebReveal(g interfaces.CuckooGame) bool {
	phase := g.GetPhase()
	return phase == domain.CuckooPhaseRoundEnd || phase == domain.CuckooPhaseGameEnd
}

// buildPlayersOutput プレイヤー情報を構築
func (p *CuckooWebPresenter) buildPlayersOutput(g interfaces.CuckooGame) []*controller.CuckooWebOutputPlayer {
	out := make([]*controller.CuckooWebOutputPlayer, 0, g.GetPlayerCnt())
	reveal := cuckooWebReveal(g)
	current := g.GetCurrentPlayerIdx()
	turnPhase := g.GetPhase() == domain.CuckooPhaseTurn
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		showCard := player.GetIsHuman() || reveal || g.IsKingRevealed(i)
		var card *controller.WebOutputCard
		if showCard && !player.IsEliminated() {
			card = cardToOutput(player.Card())
		}
		out = append(out, &controller.CuckooWebOutputPlayer{
			ID:            i,
			IsHuman:       player.GetIsHuman(),
			Card:          card,
			Lives:         maxInt0(player.GetLives()),
			IsEliminated:  player.IsEliminated(),
			KingRevealed:  g.IsKingRevealed(i),
			IsCurrentTurn: turnPhase && i == current,
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *CuckooWebPresenter) buildMessage(g interfaces.CuckooGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		winnerIdx := g.GetWinnerIdx()
		player := g.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		return buildWinnerWebMessage("cuckoo", winnerIdx, isHuman)
	}
	switch g.GetPhase() {
	case domain.CuckooPhaseTurn:
		return "", "cuckoo.turnPhase", nil
	case domain.CuckooPhaseRefuse:
		return "", "cuckoo.refusePhase", nil
	case domain.CuckooPhaseRoundEnd:
		return "", "cuckoo.roundEnd", nil
	}
	return "", "", nil
}

// ActionLogOutput 棋譜をJSON出力
func (p *CuckooWebPresenter) ActionLogOutput(g interfaces.CuckooGame) string {
	return actionLogOutputJSON(g)
}
