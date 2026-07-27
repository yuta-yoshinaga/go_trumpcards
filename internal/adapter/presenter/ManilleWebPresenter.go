//go:build !js || !wasm || classic

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// ManilleWebPresenter マニーユのWebプレゼンタークラス
type ManilleWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *ManilleWebPresenter) Output(g interfaces.ManilleGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *ManilleWebPresenter) buildBase(g interfaces.ManilleGame) *controller.ManilleWebOutput {
	resObj := new(controller.ManilleWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.LeadPlayerIdx = g.GetLeadPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.TrumpSuit = g.GetTrumpSuit()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerTeam = g.GetWinnerTeam()
	resObj.TeamScores = g.GetTeamScores()
	resObj.RoundCardPoints = g.GetRoundCardPoints()
	resObj.IsHumanTurn = g.IsHumanTurn()

	resObj.PlayableIndices = p.playableIndices(g)

	cfg := g.GetConfig()
	resObj.Config = controller.ManilleWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetPoints:  cfg.TargetPoints,
	}

	resObj.CurrentTrick = p.buildTrickOutput(g.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// playableIndices 人間プレイヤーがプレイできるカードのインデックスを返す
func (p *ManilleWebPresenter) playableIndices(g interfaces.ManilleGame) []int {
	if g.GetPhase() != domain.ManillePhasePlay || !g.IsHumanTurn() {
		return make([]int, 0)
	}
	idx := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
	if idx == nil {
		return make([]int, 0)
	}
	return idx
}

// buildTrickOutput 現在のトリック情報を構築
func (p *ManilleWebPresenter) buildTrickOutput(trick []*domain.TrickCard) []*controller.ManilleWebOutputTrickCard {
	return buildTrickCards(trick, func(tc *domain.TrickCard) *controller.ManilleWebOutputTrickCard {
		return &controller.ManilleWebOutputTrickCard{PlayerIdx: tc.PlayerIdx, Card: cardToOutput(tc.Card)}
	})
}

// buildPlayersOutput プレイヤー情報を構築
func (p *ManilleWebPresenter) buildPlayersOutput(g interfaces.ManilleGame) []*controller.ManilleWebOutputPlayer {
	teamScores := g.GetTeamScores()
	out := make([]*controller.ManilleWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		team := domain.ManilleTeamOf(i)
		out = append(out, &controller.ManilleWebOutputPlayer{
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
func (p *ManilleWebPresenter) buildMessage(g interfaces.ManilleGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		return p.winnerMessage(g)
	}
	switch g.GetPhase() {
	case domain.ManillePhasePlay:
		if len(g.GetCurrentTrick()) == 0 {
			return "", "manille.playPhase.lead", nil
		}
		return "", "manille.playPhase.follow", nil
	case domain.ManillePhaseTrickEnd:
		return "", "manille.trickEnd", nil
	case domain.ManillePhaseRoundEnd:
		return "", "manille.roundEnd", nil
	}
	return "", "", nil
}

// winnerMessage 勝者チームメッセージを構築する
func (p *ManilleWebPresenter) winnerMessage(g interfaces.ManilleGame) (string, string, map[string]string) {
	winnerTeam := g.GetWinnerTeam()
	humanTeam := -1
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if player := g.GetPlayer(i); player != nil && player.GetIsHuman() {
			humanTeam = domain.ManilleTeamOf(i)
			break
		}
	}
	if humanTeam >= 0 && winnerTeam == humanTeam {
		return "ゲーム終了！ あなたのチームの勝ち！", "manille.result.humanWin", nil
	}
	params := map[string]string{"team": fmt.Sprintf("%d", winnerTeam)}
	return fmt.Sprintf("ゲーム終了！ チーム%dの勝ち！", winnerTeam), "manille.result.cpuWin", params
}

// HintOutput ヒント情報をJSON出力する
func (p *ManilleWebPresenter) HintOutput(g interfaces.ManilleGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = &controller.ManilleWebOutputHint{
			CardIndices: hint.CardIndices,
			Reason:      hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *ManilleWebPresenter) ActionLogOutput(g interfaces.ManilleGame) string {
	return actionLogOutputJSON(g)
}
