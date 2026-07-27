//go:build !js || !wasm || classic

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// NapWebPresenter ナップのWebプレゼンタークラス
type NapWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *NapWebPresenter) Output(g interfaces.NapGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *NapWebPresenter) buildBase(g interfaces.NapGame) *controller.NapWebOutput {
	resObj := new(controller.NapWebOutput)
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
	resObj.Config = controller.NapWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetPoints:  cfg.TargetPoints,
	}

	resObj.CurrentTrick = trickCardsToOutput(g.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// bidsOutput 各プレイヤーの入札を int 配列に変換する
func (p *NapWebPresenter) bidsOutput(g interfaces.NapGame) [domain.NapPlayerCnt]int {
	bids := g.GetBids()
	var out [domain.NapPlayerCnt]int
	for i := range bids {
		out[i] = int(bids[i])
	}
	return out
}

// playableIndices 人間プレイヤーがプレイできるカードのインデックスを返す
func (p *NapWebPresenter) playableIndices(g interfaces.NapGame) []int {
	if g.GetPhase() != domain.NapPhasePlay || !g.IsHumanTurn() {
		return make([]int, 0)
	}
	idx := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
	if idx == nil {
		return make([]int, 0)
	}
	return idx
}

// buildPlayersOutput プレイヤー情報を構築
func (p *NapWebPresenter) buildPlayersOutput(g interfaces.NapGame) []*controller.NapWebOutputPlayer {
	scores := g.GetPlayerScores()
	declarer := g.GetDeclarerIdx()
	out := make([]*controller.NapWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		out = append(out, &controller.NapWebOutputPlayer{
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
func (p *NapWebPresenter) buildMessage(g interfaces.NapGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		return p.winnerMessage(g)
	}
	switch g.GetPhase() {
	case domain.NapPhaseBid:
		return "", "nap.bidPhase", nil
	case domain.NapPhasePlay:
		if len(g.GetCurrentTrick()) == 0 {
			return "", "nap.playPhase.lead", nil
		}
		return "", "nap.playPhase.follow", nil
	case domain.NapPhaseTrickEnd:
		return "", "nap.trickEnd", nil
	case domain.NapPhaseRoundEnd:
		return "", "nap.roundEnd", nil
	}
	return "", "", nil
}

// winnerMessage 勝者プレイヤーメッセージを構築する
func (p *NapWebPresenter) winnerMessage(g interfaces.NapGame) (string, string, map[string]string) {
	winner := g.GetWinnerPlayer()
	humanIdx := -1
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if player := g.GetPlayer(i); player != nil && player.GetIsHuman() {
			humanIdx = i
			break
		}
	}
	if humanIdx >= 0 && winner == humanIdx {
		return "ゲーム終了！ あなたの勝ち！", "nap.result.humanWin", nil
	}
	params := map[string]string{"player": fmt.Sprintf("%d", winner)}
	return fmt.Sprintf("ゲーム終了！ プレイヤー%dの勝ち！", winner), "nap.result.cpuWin", params
}

// HintOutput ヒント情報をJSON出力する
func (p *NapWebPresenter) HintOutput(g interfaces.NapGame) string {
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
func (p *NapWebPresenter) ActionLogOutput(g interfaces.NapGame) string {
	return actionLogOutputJSON(g)
}
