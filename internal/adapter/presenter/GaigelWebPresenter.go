//go:build !js || !wasm || extra

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// GaigelWebPresenter ガイゲルWebプレゼンタークラス
type GaigelWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *GaigelWebPresenter) Output(g interfaces.GaigelGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, g.GetCurrentTrick(), lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **フェーズと手番はここでは見ない。**Gaigel.GetHint() が自分で
	// 「人間の手番で、かつ行動を選べる状態か」を確かめて nil を返す。
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.GaigelWebOutputHint{
			CardIndex:  hint.CardIndex,
			Reason:     hint.Reason,
			IsMarriage: hint.IsMarriage,
		}
	}

	return marshalOrError(resObj)
}

func (p *GaigelWebPresenter) buildBase(g interfaces.GaigelGame) *controller.GaigelWebOutput {
	resObj := new(controller.GaigelWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.TrumpSuit = g.GetTrumpSuit()
	resObj.TrumpCard = cardToOutput(g.GetTrumpCard())
	resObj.StockRemaining = g.GetStockRemaining()
	resObj.TeamScores = [2]int{g.GetTeamScore(0), g.GetTeamScore(1)}
	resObj.RoundPoints = [2]int{g.GetRoundPoints(0), g.GetRoundPoints(1)}
	resObj.RoundMarriage = [2]int{g.GetRoundMarriagePoints(0), g.GetRoundMarriagePoints(1)}
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerTeam = g.GetWinnerTeam()
	resObj.LeadPlayerIdx = g.GetLeadPlayerIdx()

	resObj.MarriageIndices = g.GetMarriageIndices(g.GetCurrentPlayerIdx())
	if resObj.MarriageIndices == nil {
		resObj.MarriageIndices = make([]int, 0)
	}

	cfg := g.GetConfig()
	resObj.Config = controller.GaigelWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetScore:   cfg.TargetScore,
	}

	resObj.CurrentTrick = trickCardsToOutput(g.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

func (p *GaigelWebPresenter) buildPlayersOutput(g interfaces.GaigelGame) []*controller.GaigelWebOutputPlayer {
	out := make([]*controller.GaigelWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		pObj := &controller.GaigelWebOutputPlayer{
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

func (p *GaigelWebPresenter) buildMessage(g interfaces.GaigelGame, trick []*domain.TrickCard, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		winnerTeam := g.GetWinnerTeam()
		msg := fmt.Sprintf("ゲーム終了！ チーム%dの勝ち！", winnerTeam)
		code := fmt.Sprintf("gaigel.result.team%dWin", winnerTeam)
		params := map[string]string{"team": fmt.Sprintf("%d", winnerTeam)}
		return msg, code, params
	}
	switch g.GetPhase() {
	case domain.GaigelPhasePlay:
		if len(trick) == 0 {
			return "", "gaigel.playPhase.lead", nil
		}
		return "", "gaigel.playPhase.follow", nil
	case domain.GaigelPhaseTrickEnd:
		return "", "gaigel.trickEnd", nil
	case domain.GaigelPhaseRoundEnd:
		return "", "gaigel.roundEnd", nil
	}
	return "", "", nil
}

// HintOutput ヒント情報をJSON出力する
func (p *GaigelWebPresenter) HintOutput(g interfaces.GaigelGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = &controller.GaigelWebOutputHint{
			CardIndex:  hint.CardIndex,
			Reason:     hint.Reason,
			IsMarriage: hint.IsMarriage,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *GaigelWebPresenter) ActionLogOutput(g interfaces.GaigelGame) string {
	return actionLogOutputJSON(g)
}
