//go:build !js || !wasm || casino

package presenter

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// TarneebWebPresenter Tarneeb Web プレゼンタークラス
type TarneebWebPresenter struct{}

// Output ゲーム状態を JSON 出力
func (p *TarneebWebPresenter) Output(t interfaces.TarneebGame, lastErr error) string {
	resObj := p.buildBase(t)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(t, lastErr)
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *TarneebWebPresenter) buildBase(t interfaces.TarneebGame) *controller.TarneebWebOutput {
	resObj := new(controller.TarneebWebOutput)
	resObj.Phase = int(t.GetPhase())
	resObj.RoundNumber = t.GetRoundNumber()
	resObj.TrickNumber = t.GetTrickNumber()
	resObj.CurrentPlayerIdx = t.GetCurrentPlayerIdx()
	resObj.BidPlayerIdx = t.GetBidPlayerIdx()
	resObj.BidWinnerIdx = t.GetBidWinnerIdx()
	resObj.HighestBid = t.GetHighestBid()
	resObj.TrumpSuit = t.GetTrumpSuit()
	resObj.RedealCount = t.GetRedealCount()
	resObj.DealerIdx = t.GetDealerIdx()
	resObj.GameEndFlag = t.GetGameEndFlag()
	resObj.WinnerTeam = t.GetWinnerTeam()
	resObj.LeadPlayerIdx = t.GetLeadPlayerIdx()

	cfg := t.GetConfig()
	resObj.Config = controller.TarneebWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		PointLimit:    cfg.PointLimit,
		MinBid:        cfg.MinBid,
	}

	resObj.TeamScores = []int{t.GetTeamScore(0), t.GetTeamScore(1)}
	resObj.CurrentTrick = p.buildTrickOutput(t.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(t)
	return resObj
}

// buildTrickOutput 現在のトリック情報を構築
func (p *TarneebWebPresenter) buildTrickOutput(trick []*domain.TrickCard) []*controller.TarneebWebOutputTrickCard {
	return buildTrickCards(trick, func(tc *domain.TrickCard) *controller.TarneebWebOutputTrickCard {
		return &controller.TarneebWebOutputTrickCard{PlayerIdx: tc.PlayerIdx, Card: cardToOutput(tc.Card)}
	})
}

// buildPlayersOutput プレイヤー情報を構築
func (p *TarneebWebPresenter) buildPlayersOutput(t interfaces.TarneebGame) []*controller.TarneebWebOutputPlayer {
	out := make([]*controller.TarneebWebOutputPlayer, 0)
	for i := 0; i < t.GetPlayerCnt(); i++ {
		player := t.GetPlayer(i)
		pObj := &controller.TarneebWebOutputPlayer{
			ID:              i,
			IsHuman:         player.GetIsHuman(),
			Team:            player.GetTeam(),
			CardCount:       player.GetCardsSize(),
			Cards:           playerCardsToOutput(player, player.GetIsHuman()),
			Bid:             player.GetBid(),
			RoundScore:      player.GetRoundScore(),
			CumulativeScore: player.GetCumulativeScore(),
			TrickCount:      player.GetTrickCount(),
		}
		out = append(out, pObj)
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *TarneebWebPresenter) buildMessage(t interfaces.TarneebGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if t.GetGameEndFlag() {
		winnerTeam := t.GetWinnerTeam()
		humanTeam := -1
		for i := 0; i < t.GetPlayerCnt(); i++ {
			if t.GetPlayer(i).GetIsHuman() {
				humanTeam = t.GetPlayer(i).GetTeam()
				break
			}
		}
		if winnerTeam == humanTeam {
			return "", "tarneeb.gameEndHumanWin", map[string]string{"team": strconv.Itoa(winnerTeam)}
		}
		return "", "tarneeb.gameEndCpuWin", map[string]string{"team": strconv.Itoa(winnerTeam)}
	}
	switch t.GetPhase() {
	case domain.TarneebPhaseBid:
		return "", "tarneeb.bidPhase", nil
	case domain.TarneebPhaseTrumpDeclaration:
		return "", "tarneeb.trumpPhase", nil
	case domain.TarneebPhasePlay:
		if len(t.GetCurrentTrick()) == 0 {
			return "", "tarneeb.playPhase.lead", nil
		}
		return "", "tarneeb.playPhase.follow", nil
	case domain.TarneebPhaseTrickEnd:
		return "", "tarneeb.trickEnd", nil
	case domain.TarneebPhaseRoundEnd:
		return "", "tarneeb.roundEnd", nil
	}
	return "", "", nil
}

// HintOutput ヒント情報を JSON 出力する
func (p *TarneebWebPresenter) HintOutput(t interfaces.TarneebGame) string {
	hint := t.GetHint()
	resObj := p.buildBase(t)
	if hint != nil {
		resObj.Hint = &controller.TarneebWebOutputHint{
			CardIndex: hint.CardIndex,
			Bid:       hint.Bid,
			TrumpSuit: hint.TrumpSuit,
			Reason:    hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜を JSON 出力
func (p *TarneebWebPresenter) ActionLogOutput(t interfaces.TarneebGame) string {
	return actionLogOutputJSON(t)
}
