package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// CatchTenWebPresenter Catch the Ten Webプレゼンタークラス
type CatchTenWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *CatchTenWebPresenter) Output(g interfaces.CatchTenGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, g.GetCurrentTrick(), lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **フェーズと手番はここでは見ない。**CatchTen.GetHint() が自分で
	// 「人間の手番で、かつ行動を選べる状態か」を確かめて nil を返す。
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.CatchTenWebOutputHint{
			CardIndex: hint.CardIndex,
			Reason:    hint.Reason,
		}
	}

	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *CatchTenWebPresenter) buildBase(g interfaces.CatchTenGame) *controller.CatchTenWebOutput {
	resObj := new(controller.CatchTenWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.TrumpSuit = g.GetTrumpSuit()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.TeamScores = [2]int{g.GetTeamScore(0), g.GetTeamScore(1)}
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerTeam = g.GetWinnerTeam()
	resObj.LeadPlayerIdx = g.GetLeadPlayerIdx()

	cfg := g.GetConfig()
	resObj.Config = controller.CatchTenWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		PointLimit:    cfg.PointLimit,
	}

	resObj.CurrentTrick = trickCardsToOutput(g.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築
func (p *CatchTenWebPresenter) buildPlayersOutput(g interfaces.CatchTenGame) []*controller.CatchTenWebOutputPlayer {
	out := make([]*controller.CatchTenWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		pObj := &controller.CatchTenWebOutputPlayer{
			ID:              i,
			IsHuman:         player.GetIsHuman(),
			CardCount:       player.GetCardsSize(),
			Cards:           playerCardsToOutput(player, player.GetIsHuman()),
			RoundScore:      player.GetRoundScore(),
			CumulativeScore: player.GetCumulativeScore(),
			TrickCount:      player.GetTrickCount(),
			Team:            player.GetTeam(),
		}
		out = append(out, pObj)
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
// 終了時は humanWin(チーム0=人間側) / cpuWin(チーム1) / draw のいずれかを返す。
func (p *CatchTenWebPresenter) buildMessage(g interfaces.CatchTenGame, trick []*domain.TrickCard, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		winnerTeam := g.GetWinnerTeam()
		if winnerTeam == domain.CatchTenDrawTeam {
			return "ゲーム終了！ 引き分けです。", "catchten.result.draw", nil
		}
		// チーム0は人間プレイヤーのチーム → humanWin、チーム1 → cpuWin
		return buildWinnerWebMessage("catchten", winnerTeam, winnerTeam == 0)
	}
	switch g.GetPhase() {
	case domain.CatchTenPhasePlay:
		if len(trick) == 0 {
			return "", "catchten.playPhase.lead", nil
		}
		return "", "catchten.playPhase.follow", nil
	case domain.CatchTenPhaseTrickEnd:
		return "", "catchten.trickEnd", nil
	case domain.CatchTenPhaseRoundEnd:
		return "", "catchten.roundEnd", nil
	}
	return "", "", nil
}

// HintOutput ヒント情報をJSON出力する
func (p *CatchTenWebPresenter) HintOutput(g interfaces.CatchTenGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = &controller.CatchTenWebOutputHint{
			CardIndex: hint.CardIndex,
			Reason:    hint.Reason,
		}
	}
	// **「頼んだヒントか」を CLI が見分けられるようにする。**このゲーム群の
	// `hintAvailable` は画面のラベルとして既に使われているので、別キーを出す (#4483)。
	if hint != nil {
		resObj.MessageCode = "catchten.hintRequested"
	} else {
		resObj.MessageCode = "catchten.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *CatchTenWebPresenter) ActionLogOutput(g interfaces.CatchTenGame) string {
	return actionLogOutputJSON(g)
}
