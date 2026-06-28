package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// PrsiWebPresenter プルシーWebプレゼンタークラス
type PrsiWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *PrsiWebPresenter) Output(g interfaces.PrsiGame, lastErr error) string {
	resObj := new(controller.PrsiWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.DrawPileCount = g.GetDrawPileCount()
	resObj.PenaltyDrawCount = g.GetPenaltyDrawCount()
	resObj.PendingSkips = g.GetPendingSkips()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()

	top := g.GetDiscardTop()
	if top != nil {
		resObj.DiscardTop = cardToOutput(top)
	}

	cfg := g.GetConfig()
	resObj.Config = controller.PrsiWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
	}

	resObj.Players = p.buildPlayersOutput(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)

	return marshalOrError(resObj)
}

// buildPlayersOutput プレイヤー情報を構築
func (p *PrsiWebPresenter) buildPlayersOutput(g interfaces.PrsiGame) []*controller.PrsiWebOutputPlayer {
	out := make([]*controller.PrsiWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		pObj := &controller.PrsiWebOutputPlayer{
			ID:        i,
			IsHuman:   player.GetIsHuman(),
			CardCount: player.GetCardsSize(),
			Cards:     playerCardsToOutput(player, player.GetIsHuman()),
		}
		out = append(out, pObj)
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *PrsiWebPresenter) buildMessage(g interfaces.PrsiGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		winnerIdx := g.GetWinnerIdx()
		player := g.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		return buildWinnerWebMessage("prsi", winnerIdx, isHuman)
	}
	if g.GetPhase() == domain.PrsiPhasePlay {
		return "", "prsi.playPhase", nil
	}
	return "", "", nil
}

// ActionLogOutput 棋譜をJSON出力
func (p *PrsiWebPresenter) ActionLogOutput(g interfaces.PrsiGame) string {
	return actionLogOutputJSON(g)
}
