//go:build !js || !wasm || casino

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// SheepsheadWebPresenter シープスヘッドのWebプレゼンタークラス
type SheepsheadWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *SheepsheadWebPresenter) Output(g interfaces.SheepsheadGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *SheepsheadWebPresenter) buildBase(g interfaces.SheepsheadGame) *controller.SheepsheadWebOutput {
	resObj := new(controller.SheepsheadWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.LeadPlayerIdx = g.GetLeadPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()
	resObj.PickerIdx = g.GetPickerIdx()
	resObj.PassCount = g.GetPassCount()
	resObj.CalledSuit = g.GetCalledSuit()
	resObj.PartnerRevealed = g.IsPartnerRevealed()
	resObj.RoundPickerPoints = g.GetRoundPickerPoints()
	resObj.RoundMultiplier = g.GetRoundMultiplier()
	resObj.RoundPickerWon = g.GetRoundPickerWon()

	// Blind count (number of cards face-down; content hidden until picker takes them).
	blind := g.GetBlind()
	resObj.BlindCount = len(blind)

	// Buried cards: reveal to frontend only at round end (or game end).
	phase := g.GetPhase()
	if phase == domain.SheepsheadPhaseRoundEnd || phase == domain.SheepsheadPhaseGameEnd {
		resObj.Buried = cardsToOutputOrEmpty(g.GetBuried())
	} else {
		resObj.Buried = make([]*controller.WebOutputCard, 0)
	}

	// PartnerIdx: hide (send -1) until the partner has been revealed during play.
	if g.IsPartnerRevealed() || phase == domain.SheepsheadPhaseRoundEnd || phase == domain.SheepsheadPhaseGameEnd {
		resObj.PartnerIdx = g.GetPartnerIdx()
	} else {
		resObj.PartnerIdx = -1
	}

	resObj.CallableSuits = p.callableSuits(g)
	resObj.PlayableIndices = p.playableIndices(g)

	cfg := g.GetConfig()
	resObj.Config = controller.SheepsheadWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		BaseChips:     cfg.BaseChips,
		StartChips:    cfg.StartChips,
		TargetChips:   cfg.TargetChips,
	}

	resObj.CurrentTrick = trickCardsToOutput(g.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// callableSuits 呼び可能スートを返す (Call フェーズ以外は空)
func (p *SheepsheadWebPresenter) callableSuits(g interfaces.SheepsheadGame) []int {
	if g.GetPhase() != domain.SheepsheadPhaseCall {
		return make([]int, 0)
	}
	suits := g.GetCallableSuits()
	if suits == nil {
		return make([]int, 0)
	}
	return suits
}

// playableIndices 人間プレイヤーがプレイできるカードのインデックスを返す
func (p *SheepsheadWebPresenter) playableIndices(g interfaces.SheepsheadGame) []int {
	if g.GetPhase() != domain.SheepsheadPhasePlay || !g.IsHumanTurn() {
		return make([]int, 0)
	}
	idx := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
	if idx == nil {
		return make([]int, 0)
	}
	return idx
}

// buildPlayersOutput プレイヤー情報を構築
func (p *SheepsheadWebPresenter) buildPlayersOutput(g interfaces.SheepsheadGame) []*controller.SheepsheadWebOutputPlayer {
	out := make([]*controller.SheepsheadWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		out = append(out, &controller.SheepsheadWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, player.GetIsHuman()),
			TrickCount: player.GetTrickCount(),
			Chips:      player.GetChips(),
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *SheepsheadWebPresenter) buildMessage(g interfaces.SheepsheadGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		return p.winnerMessage(g)
	}
	switch g.GetPhase() {
	case domain.SheepsheadPhasePick:
		return "", "sheepshead.pickPhase", nil
	case domain.SheepsheadPhaseBury:
		return "", "sheepshead.buryPhase", nil
	case domain.SheepsheadPhaseCall:
		return "", "sheepshead.callPhase", nil
	case domain.SheepsheadPhasePlay:
		if len(g.GetCurrentTrick()) == 0 {
			return "", "sheepshead.playPhase.lead", nil
		}
		return "", "sheepshead.playPhase.follow", nil
	case domain.SheepsheadPhaseTrickEnd:
		return "", "sheepshead.trickEnd", nil
	case domain.SheepsheadPhaseRoundEnd:
		return "", "sheepshead.roundEnd", nil
	}
	return "", "", nil
}

// winnerMessage 勝者メッセージを構築する
func (p *SheepsheadWebPresenter) winnerMessage(g interfaces.SheepsheadGame) (string, string, map[string]string) {
	winnerIdx := g.GetWinnerIdx()
	isHuman := false
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if player := g.GetPlayer(i); player != nil && player.GetIsHuman() && i == winnerIdx {
			isHuman = true
			break
		}
	}
	if isHuman {
		return "ゲーム終了！ あなたの勝ち！", "sheepshead.result.humanWin", nil
	}
	params := map[string]string{"cpuId": fmt.Sprintf("%d", winnerIdx)}
	return fmt.Sprintf("ゲーム終了！ CPU %dの勝ち！", winnerIdx), "sheepshead.result.cpuWin", params
}

// HintOutput ヒント情報をJSON出力する
func (p *SheepsheadWebPresenter) HintOutput(g interfaces.SheepsheadGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = &controller.SheepsheadWebOutputHint{
			CardIndices: hint.CardIndices,
			Suit:        hint.Suit,
			Pick:        hint.Pick,
			Reason:      hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *SheepsheadWebPresenter) ActionLogOutput(g interfaces.SheepsheadGame) string {
	return actionLogOutputJSON(g)
}
