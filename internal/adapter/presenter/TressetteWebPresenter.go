//go:build !js || !wasm || casino

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// TressetteWebPresenter トレセッテのWebプレゼンタークラス
type TressetteWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *TressetteWebPresenter) Output(g interfaces.TressetteGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *TressetteWebPresenter) buildBase(g interfaces.TressetteGame) *controller.TressetteWebOutput {
	resObj := new(controller.TressetteWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.LeadPlayerIdx = g.GetLeadPlayerIdx()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerTeam = g.GetWinnerTeam()

	scores := g.GetTeamScores()
	resObj.TeamScores = scores[:]
	thirds := g.GetTeamRoundThirds()
	resObj.TeamRoundThirds = thirds[:]
	resObj.PlayableIndices = p.playableIndices(g)

	cfg := g.GetConfig()
	resObj.Config = controller.TressetteWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetPoints:  cfg.TargetPoints,
	}

	resObj.CurrentTrick = p.buildTrickOutput(g.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// playableIndices 人間プレイヤーがプレイできるカードのインデックスを返す
func (p *TressetteWebPresenter) playableIndices(g interfaces.TressetteGame) []int {
	if g.GetPhase() != domain.TressettePhasePlay || !g.IsHumanTurn() {
		return make([]int, 0)
	}
	idx := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
	if idx == nil {
		return make([]int, 0)
	}
	return idx
}

// buildTrickOutput 現在のトリック情報を構築
func (p *TressetteWebPresenter) buildTrickOutput(trick []*domain.TressetteTrickCard) []*controller.TressetteWebOutputTrickCard {
	return buildTrickCards(trick, func(tc *domain.TressetteTrickCard) *controller.TressetteWebOutputTrickCard {
		return &controller.TressetteWebOutputTrickCard{PlayerIdx: tc.PlayerIdx, Card: cardToOutput(tc.Card)}
	})
}

// buildPlayersOutput プレイヤー情報を構築
func (p *TressetteWebPresenter) buildPlayersOutput(g interfaces.TressetteGame) []*controller.TressetteWebOutputPlayer {
	out := make([]*controller.TressetteWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		out = append(out, &controller.TressetteWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, player.GetIsHuman()),
			TrickCount: player.GetTrickCount(),
			TeamID:     domain.TressetteTeamOf(i),
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *TressetteWebPresenter) buildMessage(g interfaces.TressetteGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		return p.winnerMessage(g)
	}
	switch g.GetPhase() {
	case domain.TressettePhasePlay:
		if len(g.GetCurrentTrick()) == 0 {
			return "", "tressette.playPhase.lead", nil
		}
		return "", "tressette.playPhase.follow", nil
	case domain.TressettePhaseTrickEnd:
		return "", "tressette.trickEnd", nil
	case domain.TressettePhaseRoundEnd:
		return "", "tressette.roundEnd", nil
	}
	return "", "", nil
}

// winnerMessage チーム勝利メッセージを構築する。人間は常にチーム0に属する。
func (p *TressetteWebPresenter) winnerMessage(g interfaces.TressetteGame) (string, string, map[string]string) {
	winnerTeam := g.GetWinnerTeam()
	humanTeam := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if player := g.GetPlayer(i); player != nil && player.GetIsHuman() {
			humanTeam = domain.TressetteTeamOf(i)
			break
		}
	}
	teamLabel := "A"
	if winnerTeam == 1 {
		teamLabel = "B"
	}
	params := map[string]string{"team": teamLabel}
	if winnerTeam == humanTeam {
		return fmt.Sprintf("ゲーム終了！ チーム%sの勝ち！", teamLabel), "tressette.result.humanTeamWin", params
	}
	return fmt.Sprintf("ゲーム終了！ チーム%sの勝ち！", teamLabel), "tressette.result.cpuTeamWin", params
}

// HintOutput ヒント情報をJSON出力する
func (p *TressetteWebPresenter) HintOutput(g interfaces.TressetteGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = &controller.TressetteWebOutputHint{
			CardIndices: hint.CardIndices,
			Reason:      hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *TressetteWebPresenter) ActionLogOutput(g interfaces.TressetteGame) string {
	return actionLogOutputJSON(g)
}
