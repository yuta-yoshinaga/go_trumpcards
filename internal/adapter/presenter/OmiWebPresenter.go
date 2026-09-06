//go:build !js || !wasm || extra5

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// OmiWebPresenter オミWebプレゼンタークラス
type OmiWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *OmiWebPresenter) Output(e interfaces.OmiGame, lastErr error) string {
	resObj := p.buildBase(e)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(e, e.GetCurrentTrick(), lastErr)
	if hint := e.GetHint(); hint != nil {
		resObj.Hint = &controller.OmiWebOutputHint{
			CardIndex: hint.CardIndex,
			Suit:      hint.Suit,
			Reason:    hint.Reason,
		}
	}

	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *OmiWebPresenter) buildBase(e interfaces.OmiGame) *controller.OmiWebOutput {
	resObj := new(controller.OmiWebOutput)
	resObj.Phase = int(e.GetPhase())
	resObj.RoundNumber = e.GetRoundNumber()
	resObj.TrickNumber = e.GetTrickNumber()
	resObj.CurrentPlayerIdx = e.GetCurrentPlayerIdx()
	resObj.TrumpCallerIdx = e.GetTrumpCallerIdx()
	resObj.BidPlayerIdx = e.GetTrumpCallerIdx()
	resObj.DealerIdx = e.GetDealerIdx()
	resObj.TrumpSuit = e.GetTrumpSuit()
	if e.GetPhase() == domain.OmiPhaseCallTrump {
		resObj.DealStage = 1
	} else {
		resObj.DealStage = 2
	}
	resObj.FaceUpCard = nil
	resObj.MakerTeam = e.GetMakerTeam()
	resObj.GoingAlone = false
	resObj.GoingAlonePlayerIdx = -1
	resObj.TeamScores = [2]int{e.GetTeamScore(0), e.GetTeamScore(1)}

	tricks0 := 0
	tricks1 := 0
	for i := 0; i < e.GetPlayerCnt(); i++ {
		if pl := e.GetPlayer(i); pl != nil {
			if pl.GetTeam() == 0 {
				tricks0 += pl.GetTrickCount()
			} else if pl.GetTeam() == 1 {
				tricks1 += pl.GetTrickCount()
			}
		}
	}
	resObj.TeamTricks = [2]int{tricks0, tricks1}

	resObj.GameEndFlag = e.GetGameEndFlag()
	resObj.WinnerTeam = e.GetWinnerTeam()
	resObj.LeadPlayerIdx = e.GetLeadPlayerIdx()

	// 設定
	cfg := e.GetConfig()
	resObj.Config = controller.OmiWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		PointLimit:    cfg.PointLimit,
	}

	resObj.CurrentTrick = trickCardsToOutput(e.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(e)
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築
func (p *OmiWebPresenter) buildPlayersOutput(e interfaces.OmiGame) []*controller.OmiWebOutputPlayer {
	out := make([]*controller.OmiWebOutputPlayer, 0)
	for i := 0; i < e.GetPlayerCnt(); i++ {
		player := e.GetPlayer(i)
		pObj := &controller.OmiWebOutputPlayer{
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

// buildMessage ゲーム結果メッセージを構築
func (p *OmiWebPresenter) buildMessage(e interfaces.OmiGame, trick []*domain.TrickCard, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if e.GetGameEndFlag() {
		winnerTeam := e.GetWinnerTeam()
		msg := fmt.Sprintf("ゲーム終了！ チーム%dの勝ち！", winnerTeam)
		code := fmt.Sprintf("omi.result.team%dWin", winnerTeam)
		params := map[string]string{"team": fmt.Sprintf("%d", winnerTeam)}
		return msg, code, params
	}
	switch e.GetPhase() {
	case domain.OmiPhaseCallTrump:
		return "", "omi.callTrumpPhase", nil
	case domain.OmiPhasePlay:
		if len(trick) == 0 {
			return "", "omi.playPhase.lead", nil
		}
		return "", "omi.playPhase.follow", nil
	case domain.OmiPhaseTrickEnd:
		return "", "omi.trickEnd", nil
	case domain.OmiPhaseRoundEnd:
		return "", "omi.roundEnd", nil
	}
	return "", "", nil
}

// HintOutput ヒント情報をJSON出力する
func (p *OmiWebPresenter) HintOutput(e interfaces.OmiGame) string {
	hint := e.GetHint()
	resObj := p.buildBase(e)
	if hint != nil {
		resObj.Hint = &controller.OmiWebOutputHint{
			CardIndex: hint.CardIndex,
			Suit:      hint.Suit,
			Reason:    hint.Reason,
		}
		resObj.MessageCode = "omi.hintRequested"
	} else {
		resObj.MessageCode = "omi.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *OmiWebPresenter) ActionLogOutput(e interfaces.OmiGame) string {
	return actionLogOutputJSON(e)
}
