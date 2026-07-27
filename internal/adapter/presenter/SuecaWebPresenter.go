//go:build !js || !wasm || casino

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// SuecaWebPresenter スエカのWebプレゼンタークラス
type SuecaWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *SuecaWebPresenter) Output(g interfaces.SuecaGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *SuecaWebPresenter) buildBase(g interfaces.SuecaGame) *controller.SuecaWebOutput {
	resObj := new(controller.SuecaWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.LeadPlayerIdx = g.GetLeadPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.TrumpSuit = g.GetTrumpSuit()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerTeam = g.GetWinnerTeam()
	resObj.TeamGamePoints = g.GetTeamGamePoints()
	resObj.RoundCardPoints = g.GetRoundCardPoints()
	resObj.RoundWinnerTeam = g.GetRoundWinnerTeam()
	resObj.RoundGamePoints = g.GetRoundGamePoints()
	resObj.IsHumanTurn = g.IsHumanTurn()

	resObj.PlayableIndices = p.playableIndices(g)

	cfg := g.GetConfig()
	resObj.Config = controller.SuecaWebOutputConfig{
		CpuDifficulty:    int(cfg.CpuDifficulty),
		TargetGamePoints: cfg.TargetGamePoints,
	}

	resObj.CurrentTrick = p.buildTrickOutput(g.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// playableIndices 人間プレイヤーがプレイできるカードのインデックスを返す
func (p *SuecaWebPresenter) playableIndices(g interfaces.SuecaGame) []int {
	if g.GetPhase() != domain.SuecaPhasePlay || !g.IsHumanTurn() {
		return make([]int, 0)
	}
	idx := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
	if idx == nil {
		return make([]int, 0)
	}
	return idx
}

// buildTrickOutput 現在のトリック情報を構築
func (p *SuecaWebPresenter) buildTrickOutput(trick []*domain.TrickCard) []*controller.SuecaWebOutputTrickCard {
	return buildTrickCards(trick, func(tc *domain.TrickCard) *controller.SuecaWebOutputTrickCard {
		return &controller.SuecaWebOutputTrickCard{PlayerIdx: tc.PlayerIdx, Card: cardToOutput(tc.Card)}
	})
}

// buildPlayersOutput プレイヤー情報を構築
func (p *SuecaWebPresenter) buildPlayersOutput(g interfaces.SuecaGame) []*controller.SuecaWebOutputPlayer {
	teamPts := g.GetTeamGamePoints()
	out := make([]*controller.SuecaWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		team := domain.SuecaTeamOf(i)
		out = append(out, &controller.SuecaWebOutputPlayer{
			ID:             i,
			IsHuman:        player.GetIsHuman(),
			CardCount:      player.GetCardsSize(),
			Cards:          playerCardsToOutput(player, player.GetIsHuman()),
			TrickCount:     player.GetTrickCount(),
			TeamGamePoints: teamPts[team],
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *SuecaWebPresenter) buildMessage(g interfaces.SuecaGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		return p.winnerMessage(g)
	}
	switch g.GetPhase() {
	case domain.SuecaPhasePlay:
		if len(g.GetCurrentTrick()) == 0 {
			return "", "sueca.playPhase.lead", nil
		}
		return "", "sueca.playPhase.follow", nil
	case domain.SuecaPhaseTrickEnd:
		return "", "sueca.trickEnd", nil
	case domain.SuecaPhaseRoundEnd:
		return "", "sueca.roundEnd", nil
	}
	return "", "", nil
}

// winnerMessage 勝者チームメッセージを構築する
func (p *SuecaWebPresenter) winnerMessage(g interfaces.SuecaGame) (string, string, map[string]string) {
	winnerTeam := g.GetWinnerTeam()
	humanTeam := -1
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if player := g.GetPlayer(i); player != nil && player.GetIsHuman() {
			humanTeam = domain.SuecaTeamOf(i)
			break
		}
	}
	if humanTeam >= 0 && winnerTeam == humanTeam {
		return "ゲーム終了！ あなたのチームの勝ち！", "sueca.result.humanWin", nil
	}
	params := map[string]string{"team": fmt.Sprintf("%d", winnerTeam)}
	return fmt.Sprintf("ゲーム終了！ チーム%dの勝ち！", winnerTeam), "sueca.result.cpuWin", params
}

// HintOutput ヒント情報をJSON出力する
func (p *SuecaWebPresenter) HintOutput(g interfaces.SuecaGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = &controller.SuecaWebOutputHint{
			CardIndices: hint.CardIndices,
			Reason:      hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *SuecaWebPresenter) ActionLogOutput(g interfaces.SuecaGame) string {
	return actionLogOutputJSON(g)
}
