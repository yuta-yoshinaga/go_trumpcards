//go:build !js || !wasm || extra

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// HandAndFootWebPresenter ハンドアンドフットWebプレゼンタークラス
type HandAndFootWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *HandAndFootWebPresenter) Output(g interfaces.HandAndFootGame, lastErr error) string {
	resObj := new(controller.HandAndFootWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.DrawPileCount = g.GetDrawPileCount()
	resObj.DiscardPileCount = g.GetDiscardPileCount()
	resObj.IsFrozen = g.GetIsFrozen()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerTeam = g.GetWinnerTeam()

	top := g.GetDiscardTop()
	if top != nil {
		resObj.DiscardTop = cardToOutput(top)
	}

	cfg := g.GetConfig()
	resObj.Config = controller.HandAndFootWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		PointLimit:    cfg.PointLimit,
	}

	resObj.Players = p.buildPlayersOutput(g)
	resObj.Teams = p.buildTeamsOutput(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)

	return marshalOrError(resObj)
}

// buildPlayersOutput プレイヤー情報を構築
func (p *HandAndFootWebPresenter) buildPlayersOutput(g interfaces.HandAndFootGame) []*controller.HandAndFootWebOutputPlayer {
	out := make([]*controller.HandAndFootWebOutputPlayer, 0)
	phase := g.GetPhase()
	showAllCards := phase == domain.HandAndFootPhaseRoundEnd || phase == domain.HandAndFootPhaseGameEnd

	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		showCards := player.GetIsHuman() || showAllCards

		pObj := &controller.HandAndFootWebOutputPlayer{
			ID:              i,
			Team:            domain.HandAndFootTeamOf(i),
			IsHuman:         player.GetIsHuman(),
			CardCount:       player.GetCardsSize(),
			Cards:           playerCardsToOutput(player, showCards),
			FootCount:       player.GetFootSize(),
			InFoot:          player.GetInFoot(),
			RoundScore:      player.GetRoundScore(),
			CumulativeScore: player.GetCumulativeScore(),
		}
		out = append(out, pObj)
	}
	return out
}

// buildTeamsOutput チーム（メルド・赤3）情報を構築
func (p *HandAndFootWebPresenter) buildTeamsOutput(g interfaces.HandAndFootGame) []*controller.HandAndFootWebOutputTeam {
	out := make([]*controller.HandAndFootWebOutputTeam, 0, domain.HandAndFootTeamCnt)
	for t := 0; t < domain.HandAndFootTeamCnt; t++ {
		melds := make([]*controller.HandAndFootWebOutputMeld, 0, len(g.GetTeamMelds(t)))
		for _, m := range g.GetTeamMelds(t) {
			meldOut := &controller.HandAndFootWebOutputMeld{
				Cards:     make([]*controller.WebOutputCard, 0, len(m.Cards)),
				IsNatural: m.IsNatural,
				IsCanasta: m.IsCanasta(),
				Rank:      m.GetRank(),
			}
			for _, card := range m.Cards {
				meldOut.Cards = append(meldOut.Cards, cardToOutput(card))
			}
			melds = append(melds, meldOut)
		}
		red3s := make([]*controller.WebOutputCard, 0, len(g.GetTeamRed3s(t)))
		for _, card := range g.GetTeamRed3s(t) {
			red3s = append(red3s, cardToOutput(card))
		}
		out = append(out, &controller.HandAndFootWebOutputTeam{
			Team:      t,
			Melds:     melds,
			Red3Count: len(g.GetTeamRed3s(t)),
			Red3s:     red3s,
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *HandAndFootWebPresenter) buildMessage(g interfaces.HandAndFootGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		winnerTeam := g.GetWinnerTeam()
		// チーム0が人間 (シート0) のチーム。
		isHuman := winnerTeam == domain.HandAndFootTeamOf(0)
		return buildWinnerWebMessage("handandfoot", winnerTeam, isHuman)
	}
	switch g.GetPhase() {
	case domain.HandAndFootPhaseDraw:
		return "", "handandfoot.drawPhase", nil
	case domain.HandAndFootPhaseMeld:
		return "", "handandfoot.meldPhase", nil
	case domain.HandAndFootPhaseDiscard:
		return "", "handandfoot.discardPhase", nil
	case domain.HandAndFootPhaseRoundEnd:
		return "", "handandfoot.roundEnd", nil
	}
	return "", "", nil
}

// ActionLogOutput 棋譜をJSON出力
func (p *HandAndFootWebPresenter) ActionLogOutput(g interfaces.HandAndFootGame) string {
	return actionLogOutputJSON(g)
}

// HintOutput ヒントを出力する。Web ではヒントはクライアント側 (useGameHint) で
// 算出するため、通常の状態出力を返す。HandAndFootPresenter インタフェースを満たすための実装。
func (p *HandAndFootWebPresenter) HintOutput(g interfaces.HandAndFootGame) string {
	return p.Output(g, nil)
}
