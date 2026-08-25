//go:build !js || !wasm || extra

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// BoliviaWebPresenter ボリビアWebプレゼンタークラス
type BoliviaWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *BoliviaWebPresenter) Output(g interfaces.BoliviaGame, lastErr error) string {
	resObj := new(controller.BoliviaWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.DrawPileCount = g.GetDrawPileCount()
	resObj.DiscardPileCount = g.GetDiscardPileCount()
	resObj.IsFrozen = g.GetIsFrozen()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()

	top := g.GetDiscardTop()
	if top != nil {
		resObj.DiscardTop = cardToOutput(top)
	}

	teamScores := make([]int, 0, g.GetTeamCount())
	for t := 0; t < g.GetTeamCount(); t++ {
		teamScores = append(teamScores, g.GetTeamScore(t))
	}
	resObj.TeamScores = teamScores

	cfg := g.GetConfig()
	resObj.Config = controller.BoliviaWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		PointLimit:    cfg.PointLimit,
	}

	resObj.Players = p.buildPlayersOutput(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)

	return marshalOrError(resObj)
}

// buildPlayersOutput プレイヤー情報を構築
func (p *BoliviaWebPresenter) buildPlayersOutput(g interfaces.BoliviaGame) []*controller.BoliviaWebOutputPlayer {
	out := make([]*controller.BoliviaWebOutputPlayer, 0)
	phase := g.GetPhase()
	showAllCards := phase == domain.BoliviaPhaseRoundEnd || phase == domain.BoliviaPhaseGameEnd

	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		showCards := player.GetIsHuman() || showAllCards

		melds := make([]*controller.BoliviaWebOutputMeld, 0, len(player.GetMelds()))
		for _, m := range player.GetMelds() {
			meldOut := &controller.BoliviaWebOutputMeld{
				Cards:      make([]*controller.WebOutputCard, 0, len(m.Cards)),
				Kind:       int(m.Kind),
				IsNatural:  m.IsNatural,
				IsCanasta:  m.IsCanasta(),
				IsEscalera: m.IsEscalera(),
				IsBolivia:  m.IsBoliviaCanasta(),
				Rank:       m.GetRank(),
			}
			for _, card := range m.Cards {
				meldOut.Cards = append(meldOut.Cards, cardToOutput(card))
			}
			melds = append(melds, meldOut)
		}

		red3s := make([]*controller.WebOutputCard, 0, len(player.GetRed3s()))
		for _, card := range player.GetRed3s() {
			red3s = append(red3s, cardToOutput(card))
		}

		pObj := &controller.BoliviaWebOutputPlayer{
			ID:              i,
			Team:            player.GetTeam(),
			IsHuman:         player.GetIsHuman(),
			CardCount:       player.GetCardsSize(),
			Cards:           playerCardsToOutput(player, showCards),
			Melds:           melds,
			Red3Count:       len(player.GetRed3s()),
			Red3s:           red3s,
			RoundScore:      player.GetRoundScore(),
			CumulativeScore: player.GetCumulativeScore(),
			HasCanasta:      player.HasCanasta(),
			HasEscalera:     player.HasEscalera(),
			HasBolivia:      player.HasBolivia(),
			HasInitMeld:     player.GetHasInitMeld(),
		}
		out = append(out, pObj)
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *BoliviaWebPresenter) buildMessage(g interfaces.BoliviaGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		winnerIdx := g.GetWinnerIdx()
		player := g.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		return buildWinnerWebMessage("bolivia", winnerIdx, isHuman)
	}
	switch g.GetPhase() {
	case domain.BoliviaPhaseDraw:
		return "", "bolivia.drawPhase", nil
	case domain.BoliviaPhaseMeld:
		return "", "bolivia.meldPhase", nil
	case domain.BoliviaPhaseDiscard:
		return "", "bolivia.discardPhase", nil
	case domain.BoliviaPhaseRoundEnd:
		return "", "bolivia.roundEnd", nil
	}
	return "", "", nil
}

// ActionLogOutput 棋譜をJSON出力
func (p *BoliviaWebPresenter) ActionLogOutput(g interfaces.BoliviaGame) string {
	return actionLogOutputJSON(g)
}
