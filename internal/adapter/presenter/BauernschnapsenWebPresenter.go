//go:build !js || !wasm || extra

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// BauernschnapsenWebPresenter バウエルンシュナプセンWebプレゼンタークラス
type BauernschnapsenWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *BauernschnapsenWebPresenter) Output(g interfaces.BauernschnapsenGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, g.GetCurrentTrick(), lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **フェーズと手番はここでは見ない。**Bauernschnapsen.GetHint() が自分で
	// 「人間の手番で、かつ行動を選べる状態か」を確かめて nil を返す。
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.BauernschnapsenWebOutputHint{
			CardIndex:  hint.CardIndex,
			Reason:     hint.Reason,
			IsMarriage: hint.IsMarriage,
		}
	}

	return marshalOrError(resObj)
}

func (p *BauernschnapsenWebPresenter) buildBase(g interfaces.BauernschnapsenGame) *controller.BauernschnapsenWebOutput {
	resObj := new(controller.BauernschnapsenWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.TrumpSuit = g.GetTrumpSuit()
	resObj.Contract = int(g.GetContract())
	resObj.DeclarerIdx = g.GetDeclarerIdx()
	resObj.TeamScores = [2]int{g.GetTeamScore(0), g.GetTeamScore(1)}
	resObj.RoundPoints = [2]int{g.GetRoundPoints(0), g.GetRoundPoints(1)}
	resObj.RoundMarriage = [2]int{g.GetRoundMarriagePoints(0), g.GetRoundMarriagePoints(1)}
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerTeam = g.GetWinnerTeam()
	resObj.LeadPlayerIdx = g.GetLeadPlayerIdx()

	resObj.ValidPlayIndices = g.GetValidPlayIndices(g.GetCurrentPlayerIdx())
	if resObj.ValidPlayIndices == nil {
		resObj.ValidPlayIndices = make([]int, 0)
	}
	resObj.MarriageIndices = g.GetMarriageIndices(g.GetCurrentPlayerIdx())
	if resObj.MarriageIndices == nil {
		resObj.MarriageIndices = make([]int, 0)
	}

	cfg := g.GetConfig()
	resObj.Config = controller.BauernschnapsenWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetScore:   cfg.TargetScore,
	}

	resObj.CurrentTrick = trickCardsToOutput(g.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

func (p *BauernschnapsenWebPresenter) buildPlayersOutput(g interfaces.BauernschnapsenGame) []*controller.BauernschnapsenWebOutputPlayer {
	out := make([]*controller.BauernschnapsenWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		pObj := &controller.BauernschnapsenWebOutputPlayer{
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

func (p *BauernschnapsenWebPresenter) buildMessage(g interfaces.BauernschnapsenGame, trick []*domain.TrickCard, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		winnerTeam := g.GetWinnerTeam()
		msg := fmt.Sprintf("ゲーム終了！ チーム%dの勝ち！", winnerTeam)
		code := fmt.Sprintf("bauernschnapsen.result.team%dWin", winnerTeam)
		params := map[string]string{"team": fmt.Sprintf("%d", winnerTeam)}
		return msg, code, params
	}
	switch g.GetPhase() {
	case domain.BauernschnapsenPhasePlay:
		if len(trick) == 0 {
			return "", "bauernschnapsen.playPhase.lead", nil
		}
		return "", "bauernschnapsen.playPhase.follow", nil
	case domain.BauernschnapsenPhaseTrickEnd:
		return "", "bauernschnapsen.trickEnd", nil
	case domain.BauernschnapsenPhaseRoundEnd:
		return "", "bauernschnapsen.roundEnd", nil
	}
	return "", "", nil
}

// HintOutput ヒント情報をJSON出力する
func (p *BauernschnapsenWebPresenter) HintOutput(g interfaces.BauernschnapsenGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = &controller.BauernschnapsenWebOutputHint{
			CardIndex:  hint.CardIndex,
			Reason:     hint.Reason,
			IsMarriage: hint.IsMarriage,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *BauernschnapsenWebPresenter) ActionLogOutput(g interfaces.BauernschnapsenGame) string {
	return actionLogOutputJSON(g)
}
