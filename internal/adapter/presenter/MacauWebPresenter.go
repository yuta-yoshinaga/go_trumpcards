//go:build !js || !wasm || solo

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MacauWebPresenter マカオWebプレゼンタークラス
type MacauWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *MacauWebPresenter) Output(g interfaces.MacauGame, lastErr error) string {
	resObj := new(controller.MacauWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.DrawPileCount = g.GetDrawPileCount()
	resObj.ChosenSuit = g.GetChosenSuit()
	resObj.PenaltyDrawCount = g.GetPenaltyDrawCount()
	resObj.Direction = g.GetDirection()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()

	top := g.GetDiscardTop()
	if top != nil {
		resObj.DiscardTop = cardToOutput(top)
	}

	cfg := g.GetConfig()
	resObj.Config = controller.MacauWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		PointLimit:    cfg.PointLimit,
	}

	resObj.Players = p.buildPlayersOutput(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)

	resObj.PlayableIndices = p.playableIndices(g)

	return marshalOrError(resObj)
}

// buildPlayersOutput プレイヤー情報を構築
func (p *MacauWebPresenter) buildPlayersOutput(g interfaces.MacauGame) []*controller.MacauWebOutputPlayer {
	out := make([]*controller.MacauWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		pObj := &controller.MacauWebOutputPlayer{
			ID:              i,
			IsHuman:         player.GetIsHuman(),
			CardCount:       player.GetCardsSize(),
			Cards:           playerCardsToOutput(player, player.GetIsHuman()),
			RoundScore:      player.GetRoundScore(),
			CumulativeScore: player.GetCumulativeScore(),
			HasDeclared:     player.GetHasDeclared(),
		}
		out = append(out, pObj)
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *MacauWebPresenter) buildMessage(g interfaces.MacauGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		winnerIdx := g.GetWinnerIdx()
		player := g.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		return buildWinnerWebMessage("macau", winnerIdx, isHuman)
	}
	switch g.GetPhase() {
	case domain.MacauPhasePlay:
		return "", "macau.playPhase", nil
	case domain.MacauPhaseChooseSuit:
		return "", "macau.chooseSuitPhase", nil
	case domain.MacauPhaseMustDeclare:
		return "", "macau.mustDeclarePhase", nil
	case domain.MacauPhaseRoundEnd:
		return "", "macau.roundEnd", nil
	case domain.MacauPhaseGameEnd:
		// Handled by the GetGameEndFlag() branch above; explicit for clarity.
		return "", "", nil
	}
	return "", "", nil
}

// ActionLogOutput 棋譜をJSON出力
func (p *MacauWebPresenter) ActionLogOutput(g interfaces.MacauGame) string {
	return actionLogOutputJSON(g)
}

// HintOutput はヒントを返す。Web ではクライアント側でヒントを算出するため、
// 状態出力にフォールバックする (CUI プレゼンターのみが専用ヒントを返す)。
func (p *MacauWebPresenter) HintOutput(g interfaces.MacauGame) string {
	return p.Output(g, nil)
}

// playableIndices は人間がいま出せる手札の位置を返す。
//
// **CUI は HintOutput で全部並べているのに、Web は都度クリックしてエラーで確かめる
// しかなかった (#4805)。**判定はドメインの IsValidPlay をそのまま呼ぶ。
func (p *MacauWebPresenter) playableIndices(g interfaces.MacauGame) []int {
	if g.GetGameEndFlag() || !g.IsHumanTurn() {
		return make([]int, 0)
	}
	human := g.GetPlayer(g.GetCurrentPlayerIdx())
	if human == nil {
		return make([]int, 0)
	}
	out := make([]int, 0, human.GetCardsSize())
	for i := 0; i < human.GetCardsSize(); i++ {
		if g.IsValidPlay(human.GetCard(i)) {
			out = append(out, i)
		}
	}
	return out
}
