//go:build !js || !wasm || extra

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// IndianRummyWebPresenter インドラミー Web プレゼンター
type IndianRummyWebPresenter struct{}

// Output ゲーム状態を JSON 出力
func (p *IndianRummyWebPresenter) Output(g interfaces.IndianRummyGame, lastErr error) string {
	resObj := new(controller.IndianRummyWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TargetRounds = g.GetTargetRounds()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.DrawPileCount = g.GetDrawPileCount()
	resObj.WildRank = g.GetWildRank()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()
	resObj.DeclarerIdx = g.GetDeclarerIdx()
	resObj.DeclarationValid = g.GetDeclarationValid()

	if top := g.GetDiscardTop(); top != nil {
		resObj.DiscardTop = cardToOutput(top)
	}
	if wj := g.GetWildJoker(); wj != nil {
		resObj.WildJoker = cardToOutput(wj)
	}

	cfg := g.GetConfig()
	resObj.Config = controller.IndianRummyWebOutputConfig{
		PlayerCount:   cfg.PlayerCount,
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetRounds:  cfg.TargetRounds,
	}

	resObj.Players = p.buildPlayersOutput(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)

	return marshalOrError(resObj)
}

// buildPlayersOutput プレイヤー情報を構築
func (p *IndianRummyWebPresenter) buildPlayersOutput(g interfaces.IndianRummyGame) []*controller.IndianRummyWebOutputPlayer {
	out := make([]*controller.IndianRummyWebOutputPlayer, 0)
	phase := g.GetPhase()
	revealAll := phase == domain.IndianRummyPhaseRoundEnd || phase == domain.IndianRummyPhaseGameEnd
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		showCards := player.GetIsHuman() || revealAll
		deadwood := 0
		hasPure := false
		if revealAll {
			deadwood = g.PlayerDeadwoodValue(i)
			hasPure = g.PlayerHasPureSequence(i)
		}
		pObj := &controller.IndianRummyWebOutputPlayer{
			ID:              i,
			IsHuman:         player.GetIsHuman(),
			CardCount:       player.GetCardsSize(),
			Cards:           playerCardsToOutput(player, showCards),
			RoundScore:      player.GetRoundScore(),
			CumulativeScore: player.GetCumulativeScore(),
			Deadwood:        deadwood,
			HasPureSequence: hasPure,
		}
		out = append(out, pObj)
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *IndianRummyWebPresenter) buildMessage(g interfaces.IndianRummyGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		winnerIdx := g.GetWinnerIdx()
		player := g.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		return buildWinnerWebMessage("indianrummy", winnerIdx, isHuman)
	}
	switch g.GetPhase() {
	case domain.IndianRummyPhaseDraw:
		return "", "indianrummy.drawPhase", nil
	case domain.IndianRummyPhaseDiscard:
		return "", "indianrummy.discardPhase", nil
	case domain.IndianRummyPhaseRoundEnd:
		return "", "indianrummy.roundEnd", nil
	}
	return "", "", nil
}

// ActionLogOutput 棋譜を JSON 出力
func (p *IndianRummyWebPresenter) ActionLogOutput(g interfaces.IndianRummyGame) string {
	return actionLogOutputJSON(g)
}
