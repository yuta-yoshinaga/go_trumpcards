//go:build !js || !wasm || extra

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// SambaWebPresenter サンバWebプレゼンタークラス
type SambaWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *SambaWebPresenter) Output(g interfaces.SambaGame, lastErr error) string {
	resObj := new(controller.SambaWebOutput)
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
	resObj.Config = controller.SambaWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		PointLimit:    cfg.PointLimit,
	}

	resObj.Players = p.buildPlayersOutput(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)

	return marshalOrError(resObj)
}

// buildPlayersOutput プレイヤー情報を構築
func (p *SambaWebPresenter) buildPlayersOutput(g interfaces.SambaGame) []*controller.SambaWebOutputPlayer {
	out := make([]*controller.SambaWebOutputPlayer, 0)
	phase := g.GetPhase()
	showAllCards := phase == domain.SambaPhaseRoundEnd || phase == domain.SambaPhaseGameEnd

	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		showCards := player.GetIsHuman() || showAllCards

		melds := make([]*controller.SambaWebOutputMeld, 0, len(player.GetMelds()))
		for _, m := range player.GetMelds() {
			meldOut := &controller.SambaWebOutputMeld{
				Cards:     make([]*controller.WebOutputCard, 0, len(m.Cards)),
				Kind:      int(m.Kind),
				IsNatural: m.IsNatural,
				IsCanasta: m.IsCanasta(),
				IsSamba:   m.IsSamba(),
				Rank:      m.GetRank(),
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

		pObj := &controller.SambaWebOutputPlayer{
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
			HasSamba:        player.HasSamba(),
			HasInitMeld:     player.GetHasInitMeld(),
		}
		out = append(out, pObj)
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *SambaWebPresenter) buildMessage(g interfaces.SambaGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		winnerIdx := g.GetWinnerIdx()
		player := g.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		return buildWinnerWebMessage("samba", winnerIdx, isHuman)
	}
	switch g.GetPhase() {
	case domain.SambaPhaseDraw:
		return "", "samba.drawPhase", nil
	case domain.SambaPhaseMeld:
		return "", "samba.meldPhase", nil
	case domain.SambaPhaseDiscard:
		return "", "samba.discardPhase", nil
	case domain.SambaPhaseRoundEnd:
		return "", "samba.roundEnd", nil
	}
	return "", "", nil
}

// ActionLogOutput 棋譜をJSON出力
func (p *SambaWebPresenter) ActionLogOutput(g interfaces.SambaGame) string {
	return actionLogOutputJSON(g)
}
