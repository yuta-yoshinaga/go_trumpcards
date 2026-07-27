//go:build !js || !wasm || classic

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// SedmaWebPresenter セドマのWebプレゼンタークラス
type SedmaWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *SedmaWebPresenter) Output(g interfaces.SedmaGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *SedmaWebPresenter) buildBase(g interfaces.SedmaGame) *controller.SedmaWebOutput {
	resObj := new(controller.SedmaWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.LeadPlayerIdx = g.GetLeadPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerTeam = g.GetWinnerTeam()
	resObj.TeamScores = g.GetTeamScores()
	resObj.RoundCardPoints = g.GetRoundCardPoints()
	resObj.IsHumanTurn = g.IsHumanTurn()

	resObj.PlayableIndices = p.playableIndices(g)

	cfg := g.GetConfig()
	resObj.Config = controller.SedmaWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetPoints:  cfg.TargetPoints,
	}

	resObj.CurrentTrick = p.buildTrickOutput(g.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// playableIndices 人間プレイヤーがプレイできるカードのインデックスを返す
func (p *SedmaWebPresenter) playableIndices(g interfaces.SedmaGame) []int {
	if g.GetPhase() != domain.SedmaPhasePlay || !g.IsHumanTurn() {
		return make([]int, 0)
	}
	idx := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
	if idx == nil {
		return make([]int, 0)
	}
	return idx
}

// buildTrickOutput 現在のトリック情報を構築
func (p *SedmaWebPresenter) buildTrickOutput(trick []*domain.TrickCard) []*controller.SedmaWebOutputTrickCard {
	return buildTrickCards(trick, func(tc *domain.TrickCard) *controller.SedmaWebOutputTrickCard {
		return &controller.SedmaWebOutputTrickCard{PlayerIdx: tc.PlayerIdx, Card: cardToOutput(tc.Card)}
	})
}

// buildPlayersOutput プレイヤー情報を構築
func (p *SedmaWebPresenter) buildPlayersOutput(g interfaces.SedmaGame) []*controller.SedmaWebOutputPlayer {
	teamScores := g.GetTeamScores()
	out := make([]*controller.SedmaWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		team := domain.SedmaTeamOf(i)
		out = append(out, &controller.SedmaWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, player.GetIsHuman()),
			TrickCount: player.GetTrickCount(),
			TeamScore:  teamScores[team],
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *SedmaWebPresenter) buildMessage(g interfaces.SedmaGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		return p.winnerMessage(g)
	}
	switch g.GetPhase() {
	case domain.SedmaPhasePlay:
		if len(g.GetCurrentTrick()) == 0 {
			return "", "sedma.playPhase.lead", nil
		}
		return "", "sedma.playPhase.follow", nil
	case domain.SedmaPhaseTrickEnd:
		return "", "sedma.trickEnd", nil
	case domain.SedmaPhaseRoundEnd:
		return "", "sedma.roundEnd", nil
	}
	return "", "", nil
}

// winnerMessage 勝者チームメッセージを構築する
func (p *SedmaWebPresenter) winnerMessage(g interfaces.SedmaGame) (string, string, map[string]string) {
	winnerTeam := g.GetWinnerTeam()
	humanTeam := -1
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if player := g.GetPlayer(i); player != nil && player.GetIsHuman() {
			humanTeam = domain.SedmaTeamOf(i)
			break
		}
	}
	if humanTeam >= 0 && winnerTeam == humanTeam {
		return "ゲーム終了！ あなたのチームの勝ち！", "sedma.result.humanWin", nil
	}
	teamName := domain.SedmaTeamName(winnerTeam)
	params := map[string]string{"team": teamName}
	return fmt.Sprintf("ゲーム終了！ チーム%sの勝ち！", teamName), "sedma.result.cpuWin", params
}

// HintOutput ヒント情報をJSON出力する
func (p *SedmaWebPresenter) HintOutput(g interfaces.SedmaGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = &controller.SedmaWebOutputHint{
			CardIndices: hint.CardIndices,
			Reason:      hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *SedmaWebPresenter) ActionLogOutput(g interfaces.SedmaGame) string {
	return actionLogOutputJSON(g)
}
