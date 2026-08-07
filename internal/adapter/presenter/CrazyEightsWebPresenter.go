package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// CrazyEightsWebPresenter クレイジーエイトWebプレゼンタークラス
type CrazyEightsWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *CrazyEightsWebPresenter) Output(g interfaces.CrazyEightsGame, lastErr error) string {
	resObj := new(controller.CrazyEightsWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.DrawPileCount = g.GetDrawPileCount()
	resObj.ChosenSuit = g.GetChosenSuit()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()

	top := g.GetDiscardTop()
	if top != nil {
		resObj.DiscardTop = cardToOutput(top)
	}

	cfg := g.GetConfig()
	resObj.Config = controller.CrazyEightsWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		PointLimit:    cfg.PointLimit,
	}

	resObj.Players = p.buildPlayersOutput(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)

	return marshalOrError(resObj)
}

// buildPlayersOutput プレイヤー情報を構築
func (p *CrazyEightsWebPresenter) buildPlayersOutput(g interfaces.CrazyEightsGame) []*controller.CrazyEightsWebOutputPlayer {
	out := make([]*controller.CrazyEightsWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		pObj := &controller.CrazyEightsWebOutputPlayer{
			ID:              i,
			IsHuman:         player.GetIsHuman(),
			CardCount:       player.GetCardsSize(),
			Cards:           playerCardsToOutput(player, player.GetIsHuman()),
			RoundScore:      player.GetRoundScore(),
			CumulativeScore: player.GetCumulativeScore(),
		}
		out = append(out, pObj)
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *CrazyEightsWebPresenter) buildMessage(g interfaces.CrazyEightsGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		winnerIdx := g.GetWinnerIdx()
		player := g.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		return buildWinnerWebMessage("crazyeights", winnerIdx, isHuman)
	}
	switch g.GetPhase() {
	case domain.CrazyEightsPhasePlay:
		return "", "crazyeights.playPhase", nil
	case domain.CrazyEightsPhaseChooseSuit:
		return "", "crazyeights.chooseSuitPhase", nil
	case domain.CrazyEightsPhaseRoundEnd:
		return "", "crazyeights.roundEnd", nil
	}
	return "", "", nil
}

// ActionLogOutput 棋譜をJSON出力
func (p *CrazyEightsWebPresenter) ActionLogOutput(g interfaces.CrazyEightsGame) string {
	return actionLogOutputJSON(g)
}

// HintOutput はヒントを返す。Web ではクライアント側でも簡易ヒントを出すが、
// これはサーバー計算の理由付きヒント (`command: "hint"` 専用のレスポンス)。
func (p *CrazyEightsWebPresenter) HintOutput(g interfaces.CrazyEightsGame) string {
	return p.Output(g, nil)
}
