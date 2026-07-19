package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// GongZhuWebPresenter 拱猪Webプレゼンタークラス
type GongZhuWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *GongZhuWebPresenter) Output(g interfaces.GongZhuGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, g.GetCurrentTrick(), lastErr)
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *GongZhuWebPresenter) buildBase(g interfaces.GongZhuGame) *controller.GongZhuWebOutput {
	resObj := new(controller.GongZhuWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.HeartsBroken = g.GetHeartsBroken()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()
	resObj.LeadPlayerIdx = g.GetLeadPlayerIdx()

	ex := g.GetExposure()
	resObj.Exposed = controller.GongZhuWebOutputExposure{
		Pig:     ex.Pig,
		Sheep:   ex.Sheep,
		Ace:     ex.Ace,
		Doubler: ex.Doubler,
	}
	resObj.ExposableIndices = p.exposableIndices(g)

	cfg := g.GetConfig()
	resObj.Config = controller.GongZhuWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		PointLimit:    cfg.PointLimit,
	}

	resObj.CurrentTrick = p.buildTrickOutput(g.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// exposableIndices 人間プレイヤーが公開できるカードのインデックスを返す
func (p *GongZhuWebPresenter) exposableIndices(g interfaces.GongZhuGame) []int {
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if player := g.GetPlayer(i); player != nil && player.GetIsHuman() {
			idx := g.GetExposableIndices(i)
			if idx == nil {
				return make([]int, 0)
			}
			return idx
		}
	}
	return make([]int, 0)
}

// buildTrickOutput 現在のトリック情報を構築
func (p *GongZhuWebPresenter) buildTrickOutput(trick []*domain.GongZhuTrickCard) []*controller.GongZhuWebOutputTrickCard {
	return buildTrickCards(trick, func(tc *domain.GongZhuTrickCard) *controller.GongZhuWebOutputTrickCard {
		return &controller.GongZhuWebOutputTrickCard{PlayerIdx: tc.PlayerIdx, Card: cardToOutput(tc.Card)}
	})
}

// buildPlayersOutput プレイヤー情報を構築
func (p *GongZhuWebPresenter) buildPlayersOutput(g interfaces.GongZhuGame) []*controller.GongZhuWebOutputPlayer {
	out := make([]*controller.GongZhuWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		pObj := &controller.GongZhuWebOutputPlayer{
			ID:                 i,
			IsHuman:            player.GetIsHuman(),
			CardCount:          player.GetCardsSize(),
			Cards:              playerCardsToOutput(player, player.GetIsHuman()),
			CapturedPointCards: gongZhuCapturedPointCards(player),
			RoundScore:         player.GetRoundScore(),
			CumulativeScore:    player.GetCumulativeScore(),
			TrickCount:         player.GetTrickCount(),
		}
		out = append(out, pObj)
	}
	return out
}

// gongZhuCapturedPointCards はプレイヤーが獲得した得点札（ハート各札・♠Q=豚・
// ♦J=羊・♣10=倍化）を WebOutputCard スライスに変換する。得点札はトリック獲得時に
// 公開される公開情報のため、全プレイヤー分をそのまま送出してよい。nil の場合は
// 空スライスを返す。
func gongZhuCapturedPointCards(player *domain.GongZhuPlayer) []*controller.WebOutputCard {
	return cardsToOutputOrEmpty(gongZhuCapturedPoints(player))
}

// buildMessage ゲーム結果メッセージを構築
func (p *GongZhuWebPresenter) buildMessage(g interfaces.GongZhuGame, trick []*domain.GongZhuTrickCard, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		winnerIdx := g.GetWinnerIdx()
		player := g.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		return buildWinnerWebMessage("gongzhu", winnerIdx, isHuman)
	}
	switch g.GetPhase() {
	case domain.GongZhuPhaseExpose:
		return "", "gongzhu.exposePhase", nil
	case domain.GongZhuPhasePlay:
		if len(trick) == 0 {
			return "", "gongzhu.playPhase.lead", nil
		}
		return "", "gongzhu.playPhase.follow", nil
	case domain.GongZhuPhaseTrickEnd:
		return "", "gongzhu.trickEnd", nil
	case domain.GongZhuPhaseRoundEnd:
		return "", "gongzhu.roundEnd", nil
	}
	return "", "", nil
}

// HintOutput ヒント情報をJSON出力する
func (p *GongZhuWebPresenter) HintOutput(g interfaces.GongZhuGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = &controller.GongZhuWebOutputHint{
			CardIndices: hint.CardIndices,
			Reason:      hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *GongZhuWebPresenter) ActionLogOutput(g interfaces.GongZhuGame) string {
	return actionLogOutputJSON(g)
}
