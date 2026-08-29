package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// PageOneWebPresenter ページワンWebプレゼンタークラス
type PageOneWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *PageOneWebPresenter) Output(g interfaces.PageOneGame, lastErr error) string {
	resObj := new(controller.PageOneWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
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
		names := make([]string, 0, len(penalties))
		for _, pen := range penalties {
			names = append(names, pageOnePenaltyPlayerName(g.GetPlayer(pen.PlayerIdx), pen.PlayerIdx))
		}
		return "", "pageone.penaltyApplied", map[string]string{
			"name":  strings.Join(names, ", "),
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

// pageOnePenaltyPlayerName は messageParams に載せる表示名を返す。
//
// 名前は {{name}} としてそのまま画面に出るので、ここで訳しておかないと
// 英語でプレイしていても日本語の名前が出る。共通の cuiPlayerYou /
// cuiPlayerCpu を使うのは、同じ相手が画面ごとに別の名前で出ないため。
// CUI 側の cuiPlayerName は ANSI の装飾を混ぜるので JSON には渡せない。
func pageOnePenaltyPlayerName(player *domain.PageOnePlayer, idx int) string {
	if player != nil && player.GetIsHuman() {
		return i18n.T("cuiPlayerYou")
	}
	return i18n.Tf("cuiPlayerCpu", "idx", strconv.Itoa(idx))
}
