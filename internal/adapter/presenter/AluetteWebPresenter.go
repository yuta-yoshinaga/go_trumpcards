//go:build !js || !wasm || extra2

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// AluetteWebPresenter アリュエットのWebプレゼンタークラス
//
// **手続き描画 (ADR-0033) は使わない。**アリュエットの 48 枚は 4 スート ×
// {A-9, 従/騎/王} で、10 が抜けているだけの標準デッキ。PNG 画像がそのまま使える。
// 代わりにリュエット 6 枚の序列表を毎レスポンスに載せ、名前は画面側で札に添える。
type AluetteWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *AluetteWebPresenter) Output(g interfaces.AluetteGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"` 専用の
	// レスポンスで、ページの state にはマージされない (#4483)。
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.WebOutputCardHint{
			CardIndices: hint.CardIndices,
			Reason:      hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *AluetteWebPresenter) buildBase(g interfaces.AluetteGame) *controller.AluetteWebOutput {
	resObj := new(controller.AluetteWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.LeadPlayerIdx = g.GetLeadPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.TeamScores = g.GetTeamScores()
	resObj.RoundTricks = g.GetRoundTricks()
	resObj.LastTrickWinner = g.GetLastTrickWinner()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerTeam = g.GetWinnerTeam()
	resObj.IsHumanTurn = g.IsHumanTurn()
	resObj.PlayableIndices = p.playableIndices(g)
	resObj.Luettes = aluetteLuetteOutput()

	cfg := g.GetConfig()
	resObj.Config = controller.AluetteWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetPoints:  cfg.TargetPoints,
	}

	resObj.CurrentTrick = trickCardsToOutput(g.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// aluetteLuetteOutput はドメインの序列表を送出形へ写す。
func aluetteLuetteOutput() []controller.AluetteWebOutputLuette {
	table := domain.AluetteLuetteTable()
	out := make([]controller.AluetteWebOutputLuette, 0, len(table))
	for _, l := range table {
		out = append(out, controller.AluetteWebOutputLuette{
			Design: cardDesignToString(l.Design),
			Value:  l.Value,
			Name:   l.Name,
		})
	}
	return out
}

// playableIndices 人間プレイヤーがプレイできるカードのインデックスを返す
func (p *AluetteWebPresenter) playableIndices(g interfaces.AluetteGame) []int {
	if g.GetPhase() != domain.AluettePhasePlay || !g.IsHumanTurn() {
		return make([]int, 0)
	}
	idx := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
	if idx == nil {
		return make([]int, 0)
	}
	return idx
}

// buildPlayersOutput プレイヤー情報を構築
func (p *AluetteWebPresenter) buildPlayersOutput(g interfaces.AluetteGame) []*controller.AluetteWebOutputPlayer {
	dealer := g.GetDealerIdx()
	out := make([]*controller.AluetteWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		out = append(out, &controller.AluetteWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, player.GetIsHuman()),
			TrickCount: player.GetTrickCount(),
			Team:       domain.AluetteTeamOf(i),
			IsDealer:   i == dealer,
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *AluetteWebPresenter) buildMessage(g interfaces.AluetteGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		return p.winnerMessage(g)
	}
	switch g.GetPhase() {
	case domain.AluettePhasePlay:
		if len(g.GetCurrentTrick()) == 0 {
			return "", "aluette.playPhase.lead", nil
		}
		return "", "aluette.playPhase.follow", nil
	case domain.AluettePhaseTrickEnd:
		return "", "aluette.trickEnd", nil
	case domain.AluettePhaseRoundEnd:
		return "", "aluette.roundEnd", nil
	}
	return "", "", nil
}

// winnerMessage 勝利チームのメッセージを構築する。
//
// **勝敗はチーム単位。**人間の席そのものではなく、人間が属するチームが勝ったかを見る。
func (p *AluetteWebPresenter) winnerMessage(g interfaces.AluetteGame) (string, string, map[string]string) {
	winner := g.GetWinnerTeam()
	if winner < 0 {
		return "ゲーム終了！ 引き分け。", "aluette.result.draw", nil
	}
	humanTeam := -1
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if player := g.GetPlayer(i); player != nil && player.GetIsHuman() {
			humanTeam = domain.AluetteTeamOf(i)
			break
		}
	}
	if humanTeam == winner {
		return "ゲーム終了！ あなたのチームの勝ち！", "aluette.result.humanWin", nil
	}
	params := map[string]string{"team": fmt.Sprintf("%d", winner)}
	return fmt.Sprintf("ゲーム終了！ チーム%dの勝ち！", winner), "aluette.result.cpuWin", params
}

// HintOutput ヒント情報をJSON出力する
func (p *AluetteWebPresenter) HintOutput(g interfaces.AluetteGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = &controller.WebOutputCardHint{
			CardIndices: hint.CardIndices,
			Reason:      hint.Reason,
		}
		resObj.MessageCode = "aluette.hintRequested"
	} else {
		resObj.MessageCode = "aluette.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *AluetteWebPresenter) ActionLogOutput(g interfaces.AluetteGame) string {
	return actionLogOutputJSON(g)
}
