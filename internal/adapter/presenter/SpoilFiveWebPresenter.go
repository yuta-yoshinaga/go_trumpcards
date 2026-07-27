//go:build !js || !wasm || classic

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// SpoilFiveWebPresenter スポイル・ファイブのWebプレゼンタークラス
type SpoilFiveWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *SpoilFiveWebPresenter) Output(g interfaces.SpoilFiveGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *SpoilFiveWebPresenter) buildBase(g interfaces.SpoilFiveGame) *controller.SpoilFiveWebOutput {
	resObj := new(controller.SpoilFiveWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.LeadPlayerIdx = g.GetLeadPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.TrumpSuit = g.GetTrumpSuit()
	resObj.Pot = g.GetPot()
	resObj.RoundWinnerIdx = g.GetRoundWinnerIdx()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerPlayer = g.GetWinnerPlayer()
	resObj.IsHumanTurn = g.IsHumanTurn()

	resObj.PlayableIndices = p.playableIndices(g)

	cfg := g.GetConfig()
	resObj.Config = controller.SpoilFiveWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetPoints:  cfg.TargetPoints,
	}

	resObj.CurrentTrick = p.buildTrickOutput(g.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// playableIndices 人間プレイヤーがプレイできるカードのインデックスを返す
func (p *SpoilFiveWebPresenter) playableIndices(g interfaces.SpoilFiveGame) []int {
	if g.GetPhase() != domain.SpoilFivePhasePlay || !g.IsHumanTurn() {
		return make([]int, 0)
	}
	idx := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
	if idx == nil {
		return make([]int, 0)
	}
	return idx
}

// buildTrickOutput 現在のトリック情報を構築
func (p *SpoilFiveWebPresenter) buildTrickOutput(trick []*domain.TrickCard) []*controller.SpoilFiveWebOutputTrickCard {
	return buildTrickCards(trick, func(tc *domain.TrickCard) *controller.SpoilFiveWebOutputTrickCard {
		return &controller.SpoilFiveWebOutputTrickCard{PlayerIdx: tc.PlayerIdx, Card: cardToOutput(tc.Card)}
	})
}

// buildPlayersOutput プレイヤー情報を構築
func (p *SpoilFiveWebPresenter) buildPlayersOutput(g interfaces.SpoilFiveGame) []*controller.SpoilFiveWebOutputPlayer {
	out := make([]*controller.SpoilFiveWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		out = append(out, &controller.SpoilFiveWebOutputPlayer{
			ID:          i,
			IsHuman:     player.GetIsHuman(),
			CardCount:   player.GetCardsSize(),
			Cards:       playerCardsToOutput(player, player.GetIsHuman()),
			TrickCount:  player.GetTrickCount(),
			Score:       player.GetScore(),
			RoundTricks: player.GetRoundTricks(),
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *SpoilFiveWebPresenter) buildMessage(g interfaces.SpoilFiveGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		return p.winnerMessage(g)
	}
	switch g.GetPhase() {
	case domain.SpoilFivePhasePlay:
		if len(g.GetCurrentTrick()) == 0 {
			return "", "spoilfive.playPhase.lead", nil
		}
		return "", "spoilfive.playPhase.follow", nil
	case domain.SpoilFivePhaseTrickEnd:
		return "", "spoilfive.trickEnd", nil
	case domain.SpoilFivePhaseRoundEnd:
		if g.GetRoundWinnerIdx() < 0 {
			return "", "spoilfive.spoil", nil
		}
		return "", "spoilfive.roundEnd", nil
	}
	return "", "", nil
}

// winnerMessage 勝者プレイヤーメッセージを構築する
func (p *SpoilFiveWebPresenter) winnerMessage(g interfaces.SpoilFiveGame) (string, string, map[string]string) {
	winner := g.GetWinnerPlayer()
	humanIdx := -1
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if player := g.GetPlayer(i); player != nil && player.GetIsHuman() {
			humanIdx = i
			break
		}
	}
	if humanIdx >= 0 && winner == humanIdx {
		return "ゲーム終了！ あなたの勝ち！", "spoilfive.result.humanWin", nil
	}
	params := map[string]string{"player": fmt.Sprintf("%d", winner)}
	return fmt.Sprintf("ゲーム終了！ プレイヤー%dの勝ち！", winner), "spoilfive.result.cpuWin", params
}

// HintOutput ヒント情報をJSON出力する
func (p *SpoilFiveWebPresenter) HintOutput(g interfaces.SpoilFiveGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = &controller.SpoilFiveWebOutputHint{
			CardIndices: hint.CardIndices,
			Reason:      hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *SpoilFiveWebPresenter) ActionLogOutput(g interfaces.SpoilFiveGame) string {
	return actionLogOutputJSON(g)
}
