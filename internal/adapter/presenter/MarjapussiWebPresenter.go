//go:build !js || !wasm || extra5

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MarjapussiWebPresenter マルヤプッシ (Marjapussi) のWebプレゼンタークラス
type MarjapussiWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *MarjapussiWebPresenter) Output(g interfaces.MarjapussiGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **フェーズと手番はここでは見ない。**Marjapussi.GetHint() が自分で
	// 「人間の手番で、かつ行動を選べる状態か」を確かめて nil を返す。
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = cardHint(hint.CardIndices, hint.Reason)
	}

	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *MarjapussiWebPresenter) buildBase(g interfaces.MarjapussiGame) *controller.MarjapussiWebOutput {
	resObj := new(controller.MarjapussiWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.LeadPlayerIdx = g.GetLeadPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.TrumpSuit = g.GetTrumpSuit()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerPlayer = g.GetWinnerPlayer()
	resObj.WinnerTeam = g.GetWinnerTeam()
	resObj.TeamScores = g.GetTeamScores()
	resObj.PlayerScores = g.GetPlayerScores()
	resObj.RoundCardPoints = g.GetRoundCardPoints()
	resObj.RoundMarriage = g.GetRoundMarriage()
	resObj.IsHumanTurn = g.IsHumanTurn()

	pussiCards := g.GetPussi()
	resObj.PussiCount = len(pussiCards)
	if g.GetPhase() == domain.MarjapussiPhaseRoundEnd || g.GetPhase() == domain.MarjapussiPhaseGameEnd {
		resObj.Pussi = cardsToOutput(pussiCards)
	}

	lastTrickWinner := -1
	pussiWinnerTeam := -1
	for _, log := range g.GetActionLog() {
		if log.ActionType == "pussi_win" {
			lastTrickWinner = log.PlayerIdx
			pussiWinnerTeam = log.PlayerIdx % domain.MarjapussiTeamCnt
		}
	}
	resObj.LastTrickWinner = lastTrickWinner
	resObj.PussiWinnerTeam = pussiWinnerTeam

	resObj.PlayableIndices = p.playableIndices(g)

	cfg := g.GetConfig()
	resObj.Config = controller.MarjapussiWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetPoints:  cfg.TargetPoints,
	}

	resObj.CurrentTrick = trickCardsToOutput(g.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// playableIndices 人間プレイヤーがプレイできるカードのインデックスを返す
func (p *MarjapussiWebPresenter) playableIndices(g interfaces.MarjapussiGame) []int {
	if g.GetPhase() != domain.MarjapussiPhasePlay || !g.IsHumanTurn() {
		return make([]int, 0)
	}
	idx := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
	if idx == nil {
		return make([]int, 0)
	}
	return idx
}

// buildPlayersOutput プレイヤー情報を構築
func (p *MarjapussiWebPresenter) buildPlayersOutput(g interfaces.MarjapussiGame) []*controller.MarjapussiWebOutputPlayer {
	scores := g.GetPlayerScores()
	out := make([]*controller.MarjapussiWebOutputPlayer, 0, g.GetPlayerCnt())
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		out = append(out, &controller.MarjapussiWebOutputPlayer{
			ID:         i,
			TeamID:     i % domain.MarjapussiTeamCnt,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, player.GetIsHuman()),
			TrickCount: player.GetTrickCount(),
			Score:      scores[i],
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *MarjapussiWebPresenter) buildMessage(g interfaces.MarjapussiGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		return p.winnerMessage(g)
	}
	switch g.GetPhase() {
	case domain.MarjapussiPhasePlay:
		if len(g.GetCurrentTrick()) == 0 {
			return "", "marjapussi.playPhase.lead", nil
		}
		return "", "marjapussi.playPhase.follow", nil
	case domain.MarjapussiPhaseTrickEnd:
		return "", "marjapussi.trickEnd", nil
	case domain.MarjapussiPhaseRoundEnd:
		return "", "marjapussi.roundEnd", nil
	}
	return "", "", nil
}

// winnerMessage 勝者プレイヤーメッセージを構築する
func (p *MarjapussiWebPresenter) winnerMessage(g interfaces.MarjapussiGame) (string, string, map[string]string) {
	winnerTeam := g.GetWinnerTeam()
	if winnerTeam == 0 {
		return "ゲーム終了！ あなたのチームの勝ち！", "marjapussi.result.humanWin", nil
	}
	params := map[string]string{"team": fmt.Sprintf("%d", winnerTeam)}
	return fmt.Sprintf("ゲーム終了！ チーム%dの勝ち！", winnerTeam), "marjapussi.result.cpuWin", params
}

// HintOutput ヒント情報をJSON出力する
func (p *MarjapussiWebPresenter) HintOutput(g interfaces.MarjapussiGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = cardHint(hint.CardIndices, hint.Reason)
	}
	// **「頼んだヒントか」を CLI が見分けられるようにする。**このゲーム群の
	// `hintAvailable` は画面のラベルとして既に使われているので、別キーを出す (#4483)。
	if hint != nil {
		resObj.MessageCode = "marjapussi.hintRequested"
	} else {
		resObj.MessageCode = "marjapussi.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *MarjapussiWebPresenter) ActionLogOutput(g interfaces.MarjapussiGame) string {
	return actionLogOutputJSON(g)
}
