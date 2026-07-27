//go:build !js || !wasm || classic

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// PreferenceWebPresenter プレフェランスのWebプレゼンタークラス
type PreferenceWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *PreferenceWebPresenter) Output(g interfaces.PreferenceGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *PreferenceWebPresenter) buildBase(g interfaces.PreferenceGame) *controller.PreferenceWebOutput {
	resObj := new(controller.PreferenceWebOutput)
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
	resObj.WinnerPlayer = g.GetWinnerPlayer()
	resObj.PlayerScores = g.GetPlayerScores()
	resObj.RoundTricks = g.GetRoundTricks()
	resObj.IsHumanTurn = g.IsHumanTurn()
	resObj.IsHumanBidTurn = g.IsHumanBidTurn()

	resObj.PlayableIndices = p.playableIndices(g)

	cfg := g.GetConfig()
	resObj.Config = controller.PreferenceWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetPoints:  cfg.TargetPoints,
	}

	resObj.CurrentTrick = p.buildTrickOutput(g.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// bidsOutput 各プレイヤーの入札を int 配列に変換する
func (p *PreferenceWebPresenter) bidsOutput(g interfaces.PreferenceGame) [domain.PreferencePlayerCnt]int {
	bids := g.GetBids()
	var out [domain.PreferencePlayerCnt]int
	for i := range bids {
		out[i] = int(bids[i])
	}
	return out
}

// playableIndices 人間プレイヤーがプレイできるカードのインデックスを返す
func (p *PreferenceWebPresenter) playableIndices(g interfaces.PreferenceGame) []int {
	if g.GetPhase() != domain.PreferencePhasePlay || !g.IsHumanTurn() {
		return make([]int, 0)
	}
	idx := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
	if idx == nil {
		return make([]int, 0)
	}
	return idx
}

// buildTrickOutput 現在のトリック情報を構築
func (p *PreferenceWebPresenter) buildTrickOutput(trick []*domain.TrickCard) []*controller.PreferenceWebOutputTrickCard {
	return buildTrickCards(trick, func(tc *domain.TrickCard) *controller.PreferenceWebOutputTrickCard {
		return &controller.PreferenceWebOutputTrickCard{PlayerIdx: tc.PlayerIdx, Card: cardToOutput(tc.Card)}
	})
}

// buildPlayersOutput プレイヤー情報を構築
func (p *PreferenceWebPresenter) buildPlayersOutput(g interfaces.PreferenceGame) []*controller.PreferenceWebOutputPlayer {
	scores := g.GetPlayerScores()
	declarer := g.GetDeclarerIdx()
	out := make([]*controller.PreferenceWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		out = append(out, &controller.PreferenceWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, player.GetIsHuman()),
			TrickCount: player.GetTrickCount(),
			Score:      scores[i],
			IsDeclarer: i == declarer,
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *PreferenceWebPresenter) buildMessage(g interfaces.PreferenceGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		return p.winnerMessage(g)
	}
	switch g.GetPhase() {
	case domain.PreferencePhaseBid:
		return "", "preference.bidPhase", nil
	case domain.PreferencePhasePlay:
		if len(g.GetCurrentTrick()) == 0 {
			return "", "preference.playPhase.lead", nil
		}
		return "", "preference.playPhase.follow", nil
	case domain.PreferencePhaseTrickEnd:
		return "", "preference.trickEnd", nil
	case domain.PreferencePhaseRoundEnd:
		return "", "preference.roundEnd", nil
	}
	return "", "", nil
}

// winnerMessage 勝者プレイヤーメッセージを構築する
func (p *PreferenceWebPresenter) winnerMessage(g interfaces.PreferenceGame) (string, string, map[string]string) {
	winner := g.GetWinnerPlayer()
	humanIdx := -1
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if player := g.GetPlayer(i); player != nil && player.GetIsHuman() {
			humanIdx = i
			break
		}
	}
	if humanIdx >= 0 && winner == humanIdx {
		return "ゲーム終了！ あなたの勝ち！", "preference.result.humanWin", nil
	}
	params := map[string]string{"player": fmt.Sprintf("%d", winner)}
	return fmt.Sprintf("ゲーム終了！ プレイヤー%dの勝ち！", winner), "preference.result.cpuWin", params
}

// HintOutput ヒント情報をJSON出力する
func (p *PreferenceWebPresenter) HintOutput(g interfaces.PreferenceGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = &controller.PreferenceWebOutputHint{
			CardIndices: hint.CardIndices,
			Reason:      hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *PreferenceWebPresenter) ActionLogOutput(g interfaces.PreferenceGame) string {
	return actionLogOutputJSON(g)
}
