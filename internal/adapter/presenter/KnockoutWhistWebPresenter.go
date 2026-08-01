//go:build !js || !wasm || classic

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// KnockoutWhistWebPresenter ノックアウト・ホイストのWebプレゼンタークラス
type KnockoutWhistWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *KnockoutWhistWebPresenter) Output(g interfaces.KnockoutWhistGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **フェーズと手番はここでは見ない。**KnockoutWhist.GetHint() が自分で
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
func (p *KnockoutWhistWebPresenter) buildBase(g interfaces.KnockoutWhistGame) *controller.KnockoutWhistWebOutput {
	resObj := new(controller.KnockoutWhistWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.HandSize = g.GetHandSize()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.LeadPlayerIdx = g.GetLeadPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.TrumpSuit = g.GetTrumpSuit()
	resObj.RoundWinnerIdx = g.GetRoundWinnerIdx()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerPlayer = g.GetWinnerPlayer()
	resObj.ActiveCount = g.GetActiveCount()
	resObj.IsHumanTurn = g.IsHumanTurn()

	resObj.PlayableIndices = p.playableIndices(g)

	cfg := g.GetConfig()
	resObj.Config = controller.KnockoutWhistWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
	}

	resObj.CurrentTrick = trickCardsToOutput(g.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// playableIndices 人間プレイヤーがプレイできるカードのインデックスを返す
func (p *KnockoutWhistWebPresenter) playableIndices(g interfaces.KnockoutWhistGame) []int {
	if g.GetPhase() != domain.KnockoutWhistPhasePlay || !g.IsHumanTurn() {
		return make([]int, 0)
	}
	idx := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
	if idx == nil {
		return make([]int, 0)
	}
	return idx
}

// buildPlayersOutput プレイヤー情報を構築
func (p *KnockoutWhistWebPresenter) buildPlayersOutput(g interfaces.KnockoutWhistGame) []*controller.KnockoutWhistWebOutputPlayer {
	out := make([]*controller.KnockoutWhistWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		out = append(out, &controller.KnockoutWhistWebOutputPlayer{
			ID:          i,
			IsHuman:     player.GetIsHuman(),
			CardCount:   player.GetCardsSize(),
			Cards:       playerCardsToOutput(player, player.GetIsHuman()),
			TrickCount:  player.GetTrickCount(),
			Eliminated:  player.GetEliminated(),
			Dogbones:    player.GetDogbones(),
			RoundTricks: player.GetRoundTricks(),
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *KnockoutWhistWebPresenter) buildMessage(g interfaces.KnockoutWhistGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		return p.winnerMessage(g)
	}
	switch g.GetPhase() {
	case domain.KnockoutWhistPhasePlay:
		if len(g.GetCurrentTrick()) == 0 {
			return "", "knockoutwhist.playPhase.lead", nil
		}
		return "", "knockoutwhist.playPhase.follow", nil
	case domain.KnockoutWhistPhaseTrickEnd:
		return "", "knockoutwhist.trickEnd", nil
	case domain.KnockoutWhistPhaseRoundEnd:
		return "", "knockoutwhist.roundEnd", nil
	}
	return "", "", nil
}

// winnerMessage 勝者プレイヤーメッセージを構築する
func (p *KnockoutWhistWebPresenter) winnerMessage(g interfaces.KnockoutWhistGame) (string, string, map[string]string) {
	winner := g.GetWinnerPlayer()
	humanIdx := -1
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if player := g.GetPlayer(i); player != nil && player.GetIsHuman() {
			humanIdx = i
			break
		}
	}
	if humanIdx >= 0 && winner == humanIdx {
		return "ゲーム終了！ あなたの勝ち！", "knockoutwhist.result.humanWin", nil
	}
	params := map[string]string{"player": fmt.Sprintf("%d", winner)}
	return fmt.Sprintf("ゲーム終了！ プレイヤー%dの勝ち！", winner), "knockoutwhist.result.cpuWin", params
}

// HintOutput ヒント情報をJSON出力する
func (p *KnockoutWhistWebPresenter) HintOutput(g interfaces.KnockoutWhistGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = &controller.WebOutputCardHint{
			CardIndices: hint.CardIndices,
			Reason:      hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *KnockoutWhistWebPresenter) ActionLogOutput(g interfaces.KnockoutWhistGame) string {
	return actionLogOutputJSON(g)
}
