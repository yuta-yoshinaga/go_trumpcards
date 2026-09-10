//go:build !js || !wasm || extra5

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MarriageWebPresenter マリッジ Web プレゼンター
type MarriageWebPresenter struct{}

// Output ゲーム状態を JSON 出力
func (p *MarriageWebPresenter) Output(g interfaces.MarriageGame, lastErr error) string {
	resObj := new(controller.MarriageWebOutput)
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
	resObj.Config = controller.MarriageWebOutputConfig{
		PlayerCount:   cfg.PlayerCount,
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetRounds:  cfg.TargetRounds,
	}

	resObj.Players = p.buildPlayersOutput(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)

	return marshalOrError(resObj)
}

// buildPlayersOutput プレイヤー情報を構築
func (p *MarriageWebPresenter) buildPlayersOutput(g interfaces.MarriageGame) []*controller.MarriageWebOutputPlayer {
	out := make([]*controller.MarriageWebOutputPlayer, 0)
	phase := g.GetPhase()
	revealAll := phase == domain.MarriagePhaseRoundEnd || phase == domain.MarriagePhaseGameEnd
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		showCards := player.GetIsHuman() || revealAll
		// **自分の手札の情報は隠す理由がない。**CUI は DISCARD の手番で毎ターン
		// デッドウッドとピュアシーケンス充足を出しているのに、Web は公開時にしか
		// 載せておらず、盤面から確認できなかった (#4824)。
		deadwood := 0
		maal := 0
		hasPure := false
		if revealAll || player.GetIsHuman() {
			deadwood = g.PlayerDeadwoodValue(i)
			maal = g.PlayerMaalValue(i)
			hasPure = g.PlayerHasPureSequence(i)
		}
		pObj := &controller.MarriageWebOutputPlayer{
			ID:              i,
			IsHuman:         player.GetIsHuman(),
			CardCount:       player.GetCardsSize(),
			Cards:           playerCardsToOutput(player, showCards),
			RoundScore:      player.GetRoundScore(),
			CumulativeScore: player.GetCumulativeScore(),
			Deadwood:        deadwood,
			Maal:            maal,
			HasPureSequence: hasPure,
		}
		out = append(out, pObj)
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *MarriageWebPresenter) buildMessage(g interfaces.MarriageGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		winnerIdx := g.GetWinnerIdx()
		player := g.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		return buildWinnerWebMessage("marriage", winnerIdx, isHuman)
	}
	switch g.GetPhase() {
	case domain.MarriagePhaseDraw:
		return "", "marriage.drawPhase", nil
	case domain.MarriagePhaseDiscard:
		return "", "marriage.discardPhase", nil
	case domain.MarriagePhaseRoundEnd:
		return "", "marriage.roundEnd", nil
	}
	return "", "", nil
}

// ActionLogOutput 棋譜を JSON 出力
func (p *MarriageWebPresenter) ActionLogOutput(g interfaces.MarriageGame) string {
	return actionLogOutputJSON(g)
}
