//go:build !js || !wasm || casino

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// DoppelkopfWebPresenter ドッペルコップのWebプレゼンタークラス
type DoppelkopfWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *DoppelkopfWebPresenter) Output(g interfaces.DoppelkopfGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **フェーズと手番はここでは見ない。**Doppelkopf.GetHint() が自分で
	// 「人間の手番で、かつ行動を選べる状態か」を確かめて nil を返す。
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.WebOutputCardHint{
			CardIndices: hint.CardIndices,
			Reason:      hint.Reason,
		}
	}

	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *DoppelkopfWebPresenter) buildBase(g interfaces.DoppelkopfGame) *controller.DoppelkopfWebOutput {
	resObj := new(controller.DoppelkopfWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.LeadPlayerIdx = g.GetLeadPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()
	resObj.SoloRe = g.IsSoloRe()
	resObj.TeamsRevealed = g.AreTeamsRevealed()
	resObj.ReAnnounced = g.IsReAnnounced()
	resObj.KontraAnnounced = g.IsKontraAnnounced()
	resObj.CanAnnounce = g.CanHumanAnnounce()
	resObj.RoundRePoints = g.GetRoundRePoints()
	resObj.RoundReWon = g.GetRoundReWon()
	resObj.RoundGamePoints = g.GetRoundGamePoints()

	// youAreRe: human always knows their own team.
	humanIdx := -1
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if pl := g.GetPlayer(i); pl != nil && pl.GetIsHuman() {
			humanIdx = i
			break
		}
	}
	if humanIdx >= 0 {
		resObj.YouAreRe = g.IsRe(humanIdx)
	}

	// reTeam and players[i].IsRe: only populated after teams are revealed.
	resObj.ReTeam = p.buildReTeam(g)
	resObj.PlayableIndices = p.playableIndices(g)

	cfg := g.GetConfig()
	resObj.Config = controller.DoppelkopfWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		BaseChips:     cfg.BaseChips,
		StartChips:    cfg.StartChips,
		TargetChips:   cfg.TargetChips,
	}

	resObj.CurrentTrick = trickCardsToOutput(g.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// buildReTeam Re チームの公開情報を構築 (公開前は全 false)
func (p *DoppelkopfWebPresenter) buildReTeam(g interfaces.DoppelkopfGame) []bool {
	out := make([]bool, g.GetPlayerCnt())
	if g.AreTeamsRevealed() {
		for i := range out {
			out[i] = g.IsRe(i)
		}
	}
	return out
}

// playableIndices 人間プレイヤーがプレイできるカードのインデックスを返す
func (p *DoppelkopfWebPresenter) playableIndices(g interfaces.DoppelkopfGame) []int {
	if g.GetPhase() != domain.DoppelkopfPhasePlay || !g.IsHumanTurn() {
		return make([]int, 0)
	}
	idx := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
	if idx == nil {
		return make([]int, 0)
	}
	return idx
}

// buildPlayersOutput プレイヤー情報を構築
func (p *DoppelkopfWebPresenter) buildPlayersOutput(g interfaces.DoppelkopfGame) []*controller.DoppelkopfWebOutputPlayer {
	out := make([]*controller.DoppelkopfWebOutputPlayer, 0)
	revealed := g.AreTeamsRevealed()
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		isRe := false
		if revealed {
			isRe = g.IsRe(i)
		}
		out = append(out, &controller.DoppelkopfWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, player.GetIsHuman()),
			TrickCount: player.GetTrickCount(),
			Chips:      player.GetChips(),
			IsRe:       isRe,
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *DoppelkopfWebPresenter) buildMessage(g interfaces.DoppelkopfGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		return p.winnerMessage(g)
	}
	switch g.GetPhase() {
	case domain.DoppelkopfPhasePlay:
		if len(g.GetCurrentTrick()) == 0 {
			return "", "doppelkopf.playPhase.lead", nil
		}
		return "", "doppelkopf.playPhase.follow", nil
	case domain.DoppelkopfPhaseTrickEnd:
		return "", "doppelkopf.trickEnd", nil
	case domain.DoppelkopfPhaseRoundEnd:
		return "", "doppelkopf.roundEnd", nil
	}
	return "", "", nil
}

// winnerMessage 勝者メッセージを構築する
func (p *DoppelkopfWebPresenter) winnerMessage(g interfaces.DoppelkopfGame) (string, string, map[string]string) {
	winnerIdx := g.GetWinnerIdx()
	isHuman := false
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if player := g.GetPlayer(i); player != nil && player.GetIsHuman() && i == winnerIdx {
			isHuman = true
			break
		}
	}
	if isHuman {
		return "ゲーム終了！ あなたの勝ち！", "doppelkopf.result.humanWin", nil
	}
	params := map[string]string{"cpuId": fmt.Sprintf("%d", winnerIdx)}
	return fmt.Sprintf("ゲーム終了！ CPU %dの勝ち！", winnerIdx), "doppelkopf.result.cpuWin", params
}

// HintOutput ヒント情報をJSON出力する
func (p *DoppelkopfWebPresenter) HintOutput(g interfaces.DoppelkopfGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = &controller.WebOutputCardHint{
			CardIndices: hint.CardIndices,
			Reason:      hint.Reason,
		}
	}
	// **「頼んだヒントか」を CLI が見分けられるようにする。**このゲーム群の
	// `hintAvailable` は画面のラベルとして既に使われているので、別キーを出す (#4483)。
	if hint != nil {
		resObj.MessageCode = "doppelkopf.hintRequested"
	} else {
		resObj.MessageCode = "doppelkopf.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *DoppelkopfWebPresenter) ActionLogOutput(g interfaces.DoppelkopfGame) string {
	return actionLogOutputJSON(g)
}
