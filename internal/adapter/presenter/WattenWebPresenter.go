//go:build !js || !wasm || extra

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// WattenWebPresenter ヴァッテンWebプレゼンタークラス
type WattenWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *WattenWebPresenter) Output(g interfaces.WattenGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	return marshalOrError(resObj)
}

func (p *WattenWebPresenter) buildBase(g interfaces.WattenGame) *controller.WattenWebOutput {
	resObj := new(controller.WattenWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.LeadPlayerIdx = g.GetLeadPlayerIdx()
	resObj.SchlagRank = g.GetSchlagRank()
	resObj.CriticalSuit = g.GetCriticalSuit()
	resObj.Stake = g.GetStake()
	resObj.PendingStake = g.GetPendingStake()
	resObj.RaiseCount = g.GetRaiseCount()
	resObj.RaiserTeam = g.GetRaiserTeam()
	resObj.ResponderIdx = g.GetResponderIdx()
	resObj.CanRaise = g.CanHumanRaise()
	resObj.TeamScores = [2]int{g.GetTeamScore(0), g.GetTeamScore(1)}
	resObj.TeamTricks = [2]int{g.GetTeamTricks(0), g.GetTeamTricks(1)}
	resObj.DealWinnerTeam = g.GetDealWinnerTeam()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerTeam = g.GetWinnerTeam()
	resObj.Result = int(g.GetResult())

	cfg := g.GetConfig()
	resObj.Config = controller.WattenWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetScore:   cfg.TargetScore,
		MaxRaises:     cfg.MaxRaises,
	}

	resObj.CurrentTrick = p.buildTrickOutput(g.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

func (p *WattenWebPresenter) buildTrickOutput(trick []*domain.TrickCard) []*controller.WattenWebOutputTrickCard {
	return buildTrickCards(trick, func(tc *domain.TrickCard) *controller.WattenWebOutputTrickCard {
		return &controller.WattenWebOutputTrickCard{PlayerIdx: tc.PlayerIdx, Card: cardToOutput(tc.Card)}
	})
}

func (p *WattenWebPresenter) buildPlayersOutput(g interfaces.WattenGame) []*controller.WattenWebOutputPlayer {
	out := make([]*controller.WattenWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		pObj := &controller.WattenWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, player.GetIsHuman()),
			Team:       player.GetTeam(),
			TrickCount: player.GetTrickCount(),
		}
		out = append(out, pObj)
	}
	return out
}

func (p *WattenWebPresenter) buildMessage(g interfaces.WattenGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		winnerTeam := g.GetWinnerTeam()
		msg := fmt.Sprintf("ゲーム終了！ チーム%dの勝ち！", winnerTeam)
		code := fmt.Sprintf("watten.result.team%dWin", winnerTeam)
		params := map[string]string{"team": fmt.Sprintf("%d", winnerTeam)}
		return msg, code, params
	}
	switch g.GetPhase() {
	case domain.WattenPhaseDeclare:
		return "", "watten.declarePhase", nil
	case domain.WattenPhasePlay:
		if len(g.GetCurrentTrick()) == 0 {
			return "", "watten.playPhase.lead", nil
		}
		return "", "watten.playPhase.follow", nil
	case domain.WattenPhaseRespond:
		return "", "watten.respondPhase", nil
	case domain.WattenPhaseRoundEnd:
		return "", "watten.roundEnd", nil
	}
	return "", "", nil
}

// HintOutput ヒント情報をJSON出力する
func (p *WattenWebPresenter) HintOutput(g interfaces.WattenGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = &controller.WattenWebOutputHint{
			Action:    hint.Action,
			CardIndex: hint.CardIndex,
			Rank:      hint.Rank,
			Suit:      hint.Suit,
			Reason:    hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *WattenWebPresenter) ActionLogOutput(g interfaces.WattenGame) string {
	return actionLogOutputJSON(g)
}
