//go:build !js || !wasm || casino

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// BeloteWebPresenter ベロートWebプレゼンタークラス
type BeloteWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *BeloteWebPresenter) Output(b interfaces.BeloteGame, lastErr error) string {
	resObj := p.buildBase(b)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(b, b.GetCurrentTrick(), lastErr)
	return marshalOrError(resObj)
}

func (p *BeloteWebPresenter) buildBase(b interfaces.BeloteGame) *controller.BeloteWebOutput {
	resObj := new(controller.BeloteWebOutput)
	resObj.Phase = int(b.GetPhase())
	resObj.RoundNumber = b.GetRoundNumber()
	resObj.TrickNumber = b.GetTrickNumber()
	resObj.CurrentPlayerIdx = b.GetCurrentPlayerIdx()
	resObj.BidPlayerIdx = b.GetBidPlayerIdx()
	resObj.DealerIdx = b.GetDealerIdx()
	resObj.TrumpSuit = b.GetTrumpSuit()
	resObj.FaceUpCard = cardToOutput(b.GetFaceUpCard())
	resObj.MakerTeam = b.GetMakerTeam()
	resObj.MakerPlayerIdx = b.GetMakerPlayerIdx()
	resObj.TeamScores = [2]int{b.GetTeamScore(0), b.GetTeamScore(1)}
	resObj.RoundPoints = [2]int{b.GetRoundPoints(0), b.GetRoundPoints(1)}
	resObj.RoundBeloteBonus = [2]int{b.GetRoundBeloteBonus(0), b.GetRoundBeloteBonus(1)}
	resObj.GameEndFlag = b.GetGameEndFlag()
	resObj.WinnerTeam = b.GetWinnerTeam()
	resObj.LeadPlayerIdx = b.GetLeadPlayerIdx()

	cfg := b.GetConfig()
	resObj.Config = controller.BeloteWebOutputConfig{
		CpuDifficulty:        int(cfg.CpuDifficulty),
		TargetScore:          cfg.TargetScore,
		DixDeDer:             cfg.DixDeDer,
		EnableBeloteRebelote: cfg.EnableBeloteRebelote,
	}

	resObj.CurrentTrick = trickCardsToOutput(b.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(b)
	return resObj
}

func (p *BeloteWebPresenter) buildPlayersOutput(b interfaces.BeloteGame) []*controller.BeloteWebOutputPlayer {
	out := make([]*controller.BeloteWebOutputPlayer, 0)
	for i := 0; i < b.GetPlayerCnt(); i++ {
		player := b.GetPlayer(i)
		pObj := &controller.BeloteWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, player.GetIsHuman()),
			Team:       player.GetTeam(),
			TrickCount: player.GetTrickCount(),
		}
		out = append(out, pObj)
	}
	return out
}

func (p *BeloteWebPresenter) buildMessage(b interfaces.BeloteGame, trick []*domain.TrickCard, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if b.GetGameEndFlag() {
		winnerTeam := b.GetWinnerTeam()
		msg := fmt.Sprintf("ゲーム終了！ チーム%dの勝ち！", winnerTeam)
		code := fmt.Sprintf("belote.result.team%dWin", winnerTeam)
		params := map[string]string{"team": fmt.Sprintf("%d", winnerTeam)}
		return msg, code, params
	}
	switch b.GetPhase() {
	case domain.BelotePhaseBidPickUp:
		return "", "belote.pickUpPhase", nil
	case domain.BelotePhaseBidCallTrump:
		return "", "belote.callTrumpPhase", nil
	case domain.BelotePhasePlay:
		if len(trick) == 0 {
			return "", "belote.playPhase.lead", nil
		}
		return "", "belote.playPhase.follow", nil
	case domain.BelotePhaseTrickEnd:
		return "", "belote.trickEnd", nil
	case domain.BelotePhaseRoundEnd:
		return "", "belote.roundEnd", nil
	}
	return "", "", nil
}

// HintOutput ヒント情報をJSON出力する
func (p *BeloteWebPresenter) HintOutput(b interfaces.BeloteGame) string {
	hint := b.GetHint()
	resObj := p.buildBase(b)
	if hint != nil {
		resObj.Hint = &controller.BeloteWebOutputHint{
			CardIndex: hint.CardIndex,
			OrderUp:   hint.OrderUp,
			Suit:      hint.Suit,
			Reason:    hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *BeloteWebPresenter) ActionLogOutput(b interfaces.BeloteGame) string {
	return actionLogOutputJSON(b)
}
