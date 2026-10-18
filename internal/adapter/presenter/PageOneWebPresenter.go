package presenter

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// PageOneWebPresenter ページワンWebプレゼンタークラス
type PageOneWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *PageOneWebPresenter) Output(g interfaces.PageOneGame, lastErr error) string {
	resObj := new(controller.PageOneWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.RoundHistory = g.GetRoundHistory()
	if resObj.RoundHistory == nil {
		resObj.RoundHistory = make([]domain.PageOneRoundScore, 0)
	}
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.DrawPileCount = g.GetDrawPileCount()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()

	top := g.GetDiscardTop()
	if top != nil {
		resObj.DiscardTop = cardToOutput(top)
	}

	cfg := g.GetConfig()
	resObj.Config = controller.PageOneWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		PointLimit:    cfg.PointLimit,
	}

	resObj.Players = p.buildPlayersOutput(g)
	for _, penalty := range g.GetRecentPenalties() {
		resObj.PenaltySeats = append(resObj.PenaltySeats, penalty.PlayerIdx)
	}
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)

	return marshalOrError(resObj)
}

// buildPlayersOutput プレイヤー情報を構築
func (p *PageOneWebPresenter) buildPlayersOutput(g interfaces.PageOneGame) []*controller.PageOneWebOutputPlayer {
	out := make([]*controller.PageOneWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		pObj := &controller.PageOneWebOutputPlayer{
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
func (p *PageOneWebPresenter) buildMessage(g interfaces.PageOneGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		if code, params := domain.ErrorMessageCode(lastErr); code != "" {
			return "", code, params
		}
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		winnerIdx := g.GetWinnerIdx()
		player := g.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		return buildWinnerWebMessage("pageone", winnerIdx, isHuman)
	}
	// **ペナルティはフェーズ名より先に出す。**棋譜は既定で畳まれているので、
	// ここで譲ると手札が 2 枚増えたことに次のターンまで気づけない (#6333)。
	if penalties := g.GetRecentPenalties(); len(penalties) > 0 {
		return "", "pageone.penaltyApplied", map[string]string{
			"count": strconv.Itoa(domain.PageOnePenaltyDraw),
		}
	}
	switch g.GetPhase() {
	case domain.PageOnePhasePlay:
		return "", "pageone.playPhase", nil
	case domain.PageOnePhaseMustDeclare:
		return "", "pageone.mustDeclarePhase", nil
	case domain.PageOnePhaseRoundEnd:
		return "", "pageone.roundEnd", nil
	}
	return "", "", nil
}

// ActionLogOutput 棋譜をJSON出力
func (p *PageOneWebPresenter) ActionLogOutput(g interfaces.PageOneGame) string {
	return actionLogOutputJSON(g)
}

// HintOutput はヒントを返す。Web ではクライアント側でヒントを算出するため、
// 状態出力にフォールバックする (CUI プレゼンターのみが専用ヒントを返す)。
func (p *PageOneWebPresenter) HintOutput(g interfaces.PageOneGame) string {
	return p.Output(g, nil)
}
