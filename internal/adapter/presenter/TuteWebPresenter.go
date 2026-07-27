//go:build !js || !wasm || casino

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// TuteWebPresenter トゥーテのWebプレゼンタークラス
type TuteWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *TuteWebPresenter) Output(g interfaces.TuteGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *TuteWebPresenter) buildBase(g interfaces.TuteGame) *controller.TuteWebOutput {
	resObj := new(controller.TuteWebOutput)
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
	resObj.RoundTeamPoints = g.GetRoundTeamPoints()
	resObj.CanDeclareMarriage = g.CanHumanDeclareMarriage()
	resObj.CanDeclareTute = g.CanHumanDeclareTute()
	resObj.IsHumanTurn = g.IsHumanTurn()

	resObj.DeclaredSuits = p.buildDeclaredSuits(g)
	resObj.PlayableIndices = p.playableIndices(g)

	cfg := g.GetConfig()
	resObj.Config = controller.TuteWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetPoints:  cfg.TargetPoints,
	}

	resObj.CurrentTrick = p.buildTrickOutput(g.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// buildDeclaredSuits 結婚宣言済みスートのスライスを構築 (インデックス0-4; 0は未使用)
func (p *TuteWebPresenter) buildDeclaredSuits(g interfaces.TuteGame) []bool {
	out := make([]bool, domain.CardDesignMax+1)
	for suit := 1; suit <= domain.CardDesignMax; suit++ {
		out[suit] = g.IsSuitDeclared(suit)
	}
	return out
}

// playableIndices 人間プレイヤーがプレイできるカードのインデックスを返す
func (p *TuteWebPresenter) playableIndices(g interfaces.TuteGame) []int {
	if g.GetPhase() != domain.TutePhasePlay || !g.IsHumanTurn() {
		return make([]int, 0)
	}
	idx := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
	if idx == nil {
		return make([]int, 0)
	}
	return idx
}

// buildTrickOutput 現在のトリック情報を構築
func (p *TuteWebPresenter) buildTrickOutput(trick []*domain.TrickCard) []*controller.TuteWebOutputTrickCard {
	return buildTrickCards(trick, func(tc *domain.TrickCard) *controller.TuteWebOutputTrickCard {
		return &controller.TuteWebOutputTrickCard{PlayerIdx: tc.PlayerIdx, Card: cardToOutput(tc.Card)}
	})
}

// buildPlayersOutput プレイヤー情報を構築
func (p *TuteWebPresenter) buildPlayersOutput(g interfaces.TuteGame) []*controller.TuteWebOutputPlayer {
	teamScores := g.GetTeamScores()
	out := make([]*controller.TuteWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		team := domain.TuteTeamOf(i)
		out = append(out, &controller.TuteWebOutputPlayer{
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
func (p *TuteWebPresenter) buildMessage(g interfaces.TuteGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		return p.winnerMessage(g)
	}
	switch g.GetPhase() {
	case domain.TutePhasePlay:
		if len(g.GetCurrentTrick()) == 0 {
			return "", "tute.playPhase.lead", nil
		}
		return "", "tute.playPhase.follow", nil
	case domain.TutePhaseTrickEnd:
		return "", "tute.trickEnd", nil
	case domain.TutePhaseRoundEnd:
		return "", "tute.roundEnd", nil
	}
	return "", "", nil
}

// winnerMessage 勝者チームメッセージを構築する
func (p *TuteWebPresenter) winnerMessage(g interfaces.TuteGame) (string, string, map[string]string) {
	winnerTeam := g.GetWinnerTeam()
	humanTeam := -1
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if player := g.GetPlayer(i); player != nil && player.GetIsHuman() {
			humanTeam = domain.TuteTeamOf(i)
			break
		}
	}
	if humanTeam >= 0 && winnerTeam == humanTeam {
		return "ゲーム終了！ あなたのチームの勝ち！", "tute.result.humanWin", nil
	}
	params := map[string]string{"team": fmt.Sprintf("%d", winnerTeam)}
	return fmt.Sprintf("ゲーム終了！ チーム%dの勝ち！", winnerTeam), "tute.result.cpuWin", params
}

// HintOutput ヒント情報をJSON出力する
func (p *TuteWebPresenter) HintOutput(g interfaces.TuteGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = &controller.TuteWebOutputHint{
			CardIndices: hint.CardIndices,
			Marriage:    hint.Marriage,
			Reason:      hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *TuteWebPresenter) ActionLogOutput(g interfaces.TuteGame) string {
	return actionLogOutputJSON(g)
}
