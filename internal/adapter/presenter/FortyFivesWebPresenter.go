//go:build !js || !wasm || casino

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// FortyFivesWebPresenter オークション・フォーティファイブズのWebプレゼンタークラス
type FortyFivesWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *FortyFivesWebPresenter) Output(g interfaces.FortyFivesGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *FortyFivesWebPresenter) buildBase(g interfaces.FortyFivesGame) *controller.FortyFivesWebOutput {
	resObj := new(controller.FortyFivesWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.LeadPlayerIdx = g.GetLeadPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.DeclarerIdx = g.GetDeclarerIdx()
	resObj.Contract = int(g.GetContract())
	resObj.TrumpSuit = g.GetTrumpSuit()
	resObj.Bids = p.bidsOutput(g)
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerTeam = g.GetWinnerTeam()
	resObj.TeamScores = g.GetTeamScores()
	resObj.RoundTeamPoints = g.GetRoundTeamPoints()
	resObj.IsHumanTurn = g.IsHumanTurn()
	resObj.IsHumanBidTurn = g.IsHumanBidTurn()

	resObj.PlayableIndices = p.playableIndices(g)

	cfg := g.GetConfig()
	resObj.Config = controller.FortyFivesWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetPoints:  cfg.TargetPoints,
	}

	resObj.CurrentTrick = trickCardsToOutput(g.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// bidsOutput 各プレイヤーの入札を int 配列に変換する
func (p *FortyFivesWebPresenter) bidsOutput(g interfaces.FortyFivesGame) [domain.FortyFivesPlayerCnt]int {
	bids := g.GetBids()
	var out [domain.FortyFivesPlayerCnt]int
	for i := range bids {
		out[i] = int(bids[i])
	}
	return out
}

// playableIndices 人間プレイヤーがプレイできるカードのインデックスを返す
func (p *FortyFivesWebPresenter) playableIndices(g interfaces.FortyFivesGame) []int {
	if g.GetPhase() != domain.FortyFivesPhasePlay || !g.IsHumanTurn() {
		return make([]int, 0)
	}
	idx := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
	if idx == nil {
		return make([]int, 0)
	}
	return idx
}

// buildPlayersOutput プレイヤー情報を構築
func (p *FortyFivesWebPresenter) buildPlayersOutput(g interfaces.FortyFivesGame) []*controller.FortyFivesWebOutputPlayer {
	teamScores := g.GetTeamScores()
	declarer := g.GetDeclarerIdx()
	out := make([]*controller.FortyFivesWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		out = append(out, &controller.FortyFivesWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, player.GetIsHuman()),
			TrickCount: player.GetTrickCount(),
			TeamScore:  teamScores[domain.FortyFivesTeamOf(i)],
			IsDeclarer: i == declarer,
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *FortyFivesWebPresenter) buildMessage(g interfaces.FortyFivesGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		return p.winnerMessage(g)
	}
	switch g.GetPhase() {
	case domain.FortyFivesPhaseBid:
		return "", "fortyfives.bidPhase", nil
	case domain.FortyFivesPhasePlay:
		if len(g.GetCurrentTrick()) == 0 {
			return "", "fortyfives.playPhase.lead", nil
		}
		return "", "fortyfives.playPhase.follow", nil
	case domain.FortyFivesPhaseTrickEnd:
		return "", "fortyfives.trickEnd", nil
	case domain.FortyFivesPhaseRoundEnd:
		return "", "fortyfives.roundEnd", nil
	}
	return "", "", nil
}

// winnerMessage 勝者チームメッセージを構築する
func (p *FortyFivesWebPresenter) winnerMessage(g interfaces.FortyFivesGame) (string, string, map[string]string) {
	winnerTeam := g.GetWinnerTeam()
	humanTeam := -1
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if player := g.GetPlayer(i); player != nil && player.GetIsHuman() {
			humanTeam = domain.FortyFivesTeamOf(i)
			break
		}
	}
	if humanTeam >= 0 && winnerTeam == humanTeam {
		return "ゲーム終了！ あなたのチームの勝ち！", "fortyfives.result.humanWin", nil
	}
	teamName := fortyFivesTeamLabel(winnerTeam)
	params := map[string]string{"team": teamName}
	return fmt.Sprintf("ゲーム終了！ チーム%sの勝ち！", teamName), "fortyfives.result.cpuWin", params
}

// HintOutput ヒント情報をJSON出力する
func (p *FortyFivesWebPresenter) HintOutput(g interfaces.FortyFivesGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = &controller.WebOutputCardHint{
			CardIndices: hint.CardIndices,
			Reason:      hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *FortyFivesWebPresenter) ActionLogOutput(g interfaces.FortyFivesGame) string {
	return actionLogOutputJSON(g)
}
