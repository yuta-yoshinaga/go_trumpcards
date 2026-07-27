//go:build !js || !wasm || solo

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// SchnapsenWebPresenter シュナプセンWebプレゼンタークラス
type SchnapsenWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *SchnapsenWebPresenter) Output(s interfaces.SchnapsenGame, lastErr error) string {
	resObj := p.buildBase(s)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(s, lastErr)
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *SchnapsenWebPresenter) buildBase(s interfaces.SchnapsenGame) *controller.SchnapsenWebOutput {
	resObj := new(controller.SchnapsenWebOutput)
	resObj.Phase = int(s.GetPhase())
	resObj.TrickNumber = s.GetTrickNumber()
	resObj.CurrentPlayerIdx = s.GetCurrentPlayerIdx()
	resObj.TrumpSuit = s.GetTrumpSuit()
	if tc := s.GetTrumpCard(); tc != nil {
		resObj.TrumpCard = cardToOutput(tc)
	}
	resObj.DealerIdx = s.GetDealerIdx()
	resObj.LeadPlayerIdx = s.GetLeadPlayerIdx()
	resObj.StockRemaining = s.GetStockRemaining()
	resObj.IsEndgame = s.IsEndgame()
	resObj.ValidPlays = intSliceOrEmpty(s.GetValidPlayIndices(0))
	resObj.MarriagePlays = intSliceOrEmpty(s.GetMarriageIndices(0))
	resObj.GameEndFlag = s.GetGameEndFlag()
	resObj.WinnerIdx = s.GetWinnerIdx()

	cfg := s.GetConfig()
	resObj.Config = controller.SchnapsenWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
	}

	resObj.CurrentTrick = p.buildTrickOutput(s.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(s)
	return resObj
}

// intSliceOrEmpty nil スライスを空スライスに正規化する (JSON で null を避ける)。
func intSliceOrEmpty(in []int) []int {
	if in == nil {
		return make([]int, 0)
	}
	return in
}

// buildTrickOutput 現在のトリック情報を構築
func (p *SchnapsenWebPresenter) buildTrickOutput(trick []*domain.SchnapsenTrickCard) []*controller.SchnapsenWebOutputTrickCard {
	return buildTrickCards(trick, func(tc *domain.SchnapsenTrickCard) *controller.SchnapsenWebOutputTrickCard {
		return &controller.SchnapsenWebOutputTrickCard{PlayerIdx: tc.PlayerIdx, Card: cardToOutput(tc.Card)}
	})
}

// buildPlayersOutput プレイヤー情報を構築
func (p *SchnapsenWebPresenter) buildPlayersOutput(s interfaces.SchnapsenGame) []*controller.SchnapsenWebOutputPlayer {
	out := make([]*controller.SchnapsenWebOutputPlayer, 0)
	for i := 0; i < s.GetPlayerCnt(); i++ {
		player := s.GetPlayer(i)
		out = append(out, &controller.SchnapsenWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, player.GetIsHuman()),
			Points:     s.GetPlayerPoints(i),
			TrickCount: player.GetTrickCount(),
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *SchnapsenWebPresenter) buildMessage(s interfaces.SchnapsenGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if s.GetGameEndFlag() {
		p0 := s.GetPlayerPoints(0)
		p1 := s.GetPlayerPoints(1)
		params := map[string]string{
			"p0": fmt.Sprintf("%d", p0),
			"p1": fmt.Sprintf("%d", p1),
		}
		switch s.GetWinnerIdx() {
		case 0:
			return "", "schnapsen.result.p0Win", params
		case 1:
			return "", "schnapsen.result.p1Win", params
		default:
			return "", "schnapsen.result.tie", params
		}
	}
	switch s.GetPhase() {
	case domain.SchnapsenPhasePlay:
		if len(s.GetCurrentTrick()) == 0 {
			return "", "schnapsen.playPhase.lead", nil
		}
		return "", "schnapsen.playPhase.follow", nil
	case domain.SchnapsenPhaseTrickEnd:
		return "", "schnapsen.trickEnd", nil
	}
	return "", "", nil
}

// HintOutput ヒント情報をJSON出力する
func (p *SchnapsenWebPresenter) HintOutput(s interfaces.SchnapsenGame) string {
	hint := s.GetHint()
	resObj := p.buildBase(s)
	if hint != nil {
		resObj.Hint = &controller.SchnapsenWebOutputHint{
			CardIndex:  hint.CardIndex,
			Reason:     hint.Reason,
			IsMarriage: hint.IsMarriage,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *SchnapsenWebPresenter) ActionLogOutput(s interfaces.SchnapsenGame) string {
	return actionLogOutputJSON(s)
}
