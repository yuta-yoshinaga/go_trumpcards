//go:build !js || !wasm || extra2

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// TrappolaWebPresenter トラッポラのWebプレゼンタークラス
type TrappolaWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *TrappolaWebPresenter) Output(g interfaces.TrappolaGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **フェーズと手番はここでは見ない。**Trappola.GetHint() が自分で
	// 「人間の手番で、かつ行動を選べる状態か」を確かめて nil を返す。
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = cardHint(hint.CardIndices, hint.Reason)
	}

	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *TrappolaWebPresenter) buildBase(g interfaces.TrappolaGame) *controller.TrappolaWebOutput {
	resObj := new(controller.TrappolaWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.LeadPlayerIdx = g.GetLeadPlayerIdx()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerTeam = g.GetWinnerTeam()

	scores := g.GetTeamScores()
	resObj.TeamScores = scores[:]
	thirds := g.GetTeamRoundThirds()
	resObj.TeamRoundThirds = thirds[:]
	resObj.PlayableIndices = p.playableIndices(g)

	cfg := g.GetConfig()
	resObj.Config = controller.TrappolaWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetPoints:  cfg.TargetPoints,
	}

	resObj.CurrentTrick = trickCardsToOutput(g.GetCurrentTrick())
	resObj.LastTrick, resObj.LastTrickWinner = p.buildLastTrickOutput(g)
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// buildLastTrickOutput は直前に解決されたトリック（誰が何を出し誰が取ったか）を
// アクションログから再構築する。ドメインは専用の lastTrick フィールドを持たないが、
// 各トリックの "play" ログ（プレイヤーと札）と "trick_win" ログ（勝者）から
// 現ラウンドの直近トリックを復元できる。ラウンド開始直後（プレイフェーズのトリック 1
// で、この局のトリックがまだ確定していない）は空スライスと -1 を返す。
func (p *TrappolaWebPresenter) buildLastTrickOutput(g interfaces.TrappolaGame) ([]*controller.WebOutputTrickCard, int) {
	empty := make([]*controller.WebOutputTrickCard, 0)
	// ラウンド最初のトリックがプレイ中は、この局に確定済みトリックが無いため空を返す。
	if g.GetPhase() == domain.TrappolaPhasePlay && g.GetTrickNumber() <= 1 {
		return empty, -1
	}

	log := g.GetActionLog()
	winIdx := -1
	for i := len(log) - 1; i >= 0; i-- {
		if log[i] != nil && log[i].ActionType == "trick_win" {
			winIdx = i
			break
		}
	}
	if winIdx < 0 {
		return empty, -1
	}

	// trick_win 直前の "play" ログ（プレイ順）が、そのトリックの各札に対応する。
	var plays []*domain.ActionLogEntry
	for i := 0; i < winIdx; i++ {
		if e := log[i]; e != nil && e.ActionType == "play" && len(e.Cards) > 0 {
			plays = append(plays, e)
		}
	}
	if len(plays) < domain.TrappolaPlayerCnt {
		return empty, -1
	}
	plays = plays[len(plays)-domain.TrappolaPlayerCnt:]

	out := make([]*controller.WebOutputTrickCard, 0, len(plays))
	for _, e := range plays {
		out = append(out, &controller.WebOutputTrickCard{
			PlayerIdx: e.PlayerIdx,
			Card:      cardToOutput(e.Cards[0]),
		})
	}
	return out, log[winIdx].PlayerIdx
}

// playableIndices 人間プレイヤーがプレイできるカードのインデックスを返す
func (p *TrappolaWebPresenter) playableIndices(g interfaces.TrappolaGame) []int {
	if g.GetPhase() != domain.TrappolaPhasePlay || !g.IsHumanTurn() {
		return make([]int, 0)
	}
	idx := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
	if idx == nil {
		return make([]int, 0)
	}
	return idx
}

// buildPlayersOutput プレイヤー情報を構築
func (p *TrappolaWebPresenter) buildPlayersOutput(g interfaces.TrappolaGame) []*controller.TrappolaWebOutputPlayer {
	out := make([]*controller.TrappolaWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		out = append(out, &controller.TrappolaWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, player.GetIsHuman()),
			TrickCount: player.GetTrickCount(),
			TeamID:     domain.TrappolaTeamOf(i),
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *TrappolaWebPresenter) buildMessage(g interfaces.TrappolaGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		return p.winnerMessage(g)
	}
	switch g.GetPhase() {
	case domain.TrappolaPhasePlay:
		if len(g.GetCurrentTrick()) == 0 {
			return "", "trappola.playPhase.lead", nil
		}
		return "", "trappola.playPhase.follow", nil
	case domain.TrappolaPhaseTrickEnd:
		return "", "trappola.trickEnd", nil
	case domain.TrappolaPhaseRoundEnd:
		return "", "trappola.roundEnd", nil
	}
	return "", "", nil
}

// winnerMessage チーム勝利メッセージを構築する。人間は常にチーム0に属する。
func (p *TrappolaWebPresenter) winnerMessage(g interfaces.TrappolaGame) (string, string, map[string]string) {
	winnerTeam := g.GetWinnerTeam()
	humanTeam := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if player := g.GetPlayer(i); player != nil && player.GetIsHuman() {
			humanTeam = domain.TrappolaTeamOf(i)
			break
		}
	}
	teamLabel := "A"
	if winnerTeam == 1 {
		teamLabel = "B"
	}
	params := map[string]string{"team": teamLabel}
	if winnerTeam == humanTeam {
		return fmt.Sprintf("ゲーム終了！ チーム%sの勝ち！", teamLabel), "trappola.result.humanTeamWin", params
	}
	return fmt.Sprintf("ゲーム終了！ チーム%sの勝ち！", teamLabel), "trappola.result.cpuTeamWin", params
}

// HintOutput ヒント情報をJSON出力する
func (p *TrappolaWebPresenter) HintOutput(g interfaces.TrappolaGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = cardHint(hint.CardIndices, hint.Reason)
	}
	// **「頼んだヒントか」を CLI が見分けられるようにする。**このゲーム群の
	// `hintAvailable` は画面のラベルとして既に使われているので、別キーを出す (#4483)。
	if hint != nil {
		resObj.MessageCode = "trappola.hintRequested"
	} else {
		resObj.MessageCode = "trappola.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *TrappolaWebPresenter) ActionLogOutput(g interfaces.TrappolaGame) string {
	return actionLogOutputJSON(g)
}
