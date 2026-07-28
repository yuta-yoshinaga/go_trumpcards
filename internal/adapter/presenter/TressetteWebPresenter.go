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

	resObj.CurrentTrick = trickCardsToOutput(g.GetCurrentTrick())
	resObj.LastTrick, resObj.LastTrickWinner = p.buildLastTrickOutput(g)
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// buildLastTrickOutput は直前に解決されたトリック（誰が何を出し誰が取ったか）を
// アクションログから再構築する。ドメインは専用の lastTrick フィールドを持たないが、
// 各トリックの "play" ログ（プレイヤーと札）と "trick_win" ログ（勝者）から
// 現ラウンドの直近トリックを復元できる。ラウンド開始直後（プレイフェーズのトリック 1
// で、この局のトリックがまだ確定していない）は空スライスと -1 を返す。
func (p *TressetteWebPresenter) buildLastTrickOutput(g interfaces.TressetteGame) ([]*controller.WebOutputTrickCard, int) {
	empty := make([]*controller.WebOutputTrickCard, 0)
	// ラウンド最初のトリックがプレイ中は、この局に確定済みトリックが無いため空を返す。
	if g.GetPhase() == domain.TressettePhasePlay && g.GetTrickNumber() <= 1 {
		return empty, -1
	}

	log := g.GetActionLog()
	winIdx := -1
	for i := len(log) - 1; i >= 0; i-- {
		if log[i] != nil && log[i].ActionType == "trick_win" {
			winIdx = i
			break
		}
	}
	if winIdx < 0 {
		return empty, -1
	}

	// trick_win 直前の "play" ログ（プレイ順）が、そのトリックの各札に対応する。
	var plays []*domain.ActionLogEntry
	for i := 0; i < winIdx; i++ {
		if e := log[i]; e != nil && e.ActionType == "play" && len(e.Cards) > 0 {
			plays = append(plays, e)
		}
	}
	if len(plays) < domain.TressettePlayerCnt {
		return empty, -1
	}
	plays = plays[len(plays)-domain.TressettePlayerCnt:]

	out := make([]*controller.WebOutputTrickCard, 0, len(plays))
	for _, e := range plays {
		out = append(out, &controller.WebOutputTrickCard{
			PlayerIdx: e.PlayerIdx,
			Card:      cardToOutput(e.Cards[0]),
		})
	}
	return out, log[winIdx].PlayerIdx
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
		resObj.Hint = &controller.WebOutputCardHint{
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
