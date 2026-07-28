//go:build !js || !wasm || extra

package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// CinchWebPresenter はチンチ (Cinch) の Web プレゼンタークラス。
type CinchWebPresenter struct{}

// Output はゲーム状態を JSON 出力する。
func (p *CinchWebPresenter) Output(g interfaces.CinchGame, lastErr error) string {
	resObj := p.buildBase(g)
	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if g.GetGameEndFlag() {
		resObj.Message = p.buildResultMessage(g)
		resObj.MessageCode = "cinch.result.scores"
		resObj.MessageParams = map[string]string{"scores": p.encodeScoresParam(g)}
	}
	return marshalOrError(resObj)
}

// buildBase は基本フィールドを埋めた出力オブジェクトを生成する。
func (p *CinchWebPresenter) buildBase(g interfaces.CinchGame) *controller.CinchWebOutput {
	resObj := new(controller.CinchWebOutput)
	resObj.Players = make([]*controller.CinchWebOutputPlayer, 0)
	resObj.CurrentTrick = make([]*controller.WebOutputTrickCard, 0)
	resObj.LastTrick = make([]*controller.WebOutputTrickCard, 0)
	resObj.PlayableIndices = make([]int, 0)
	resObj.RoundWinners = make([]int, 0)

	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.TotalTricks = domain.CinchTotalTricks
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.CurrentTurn = g.GetCurrentTurn()
	resObj.BidPlayerIdx = g.GetBidPlayerIdx()
	resObj.CurrentBid = g.GetCurrentBid()
	resObj.BidWinnerIdx = g.GetBidWinnerIdx()
	resObj.TrumpSuit = g.GetTrumpSuit()
	resObj.LastTrickWinner = g.GetLastTrickWinner()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()
	resObj.IsHumanTurn = g.IsHumanTurn()
	resObj.RoundWinners = append(resObj.RoundWinners, g.GetRoundWinners()...)

	cfg := g.GetConfig()
	resObj.Config = controller.CinchWebConfigOutput{
		CpuDifficulty: int(cfg.CpuDifficulty),
		PointLimit:    cfg.PointLimit,
	}

	resObj.CurrentTrick = cinchTrickToOutput(g.GetCurrentTrick())
	resObj.LastTrick = cinchTrickToOutput(g.GetLastTrick())
	resObj.PlayableIndices = p.playableIndices(g)

	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		resObj.Players = append(resObj.Players, &controller.CinchWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, player.GetIsHuman()),
			TrickCount: player.GetTrickCount(),
			Bid:        player.GetBid(),
			TotalScore: player.GetTotalScore(),
		})
	}

	if det := g.GetLastDealDetail(); det != nil {
		resObj.LastDealDetail = &controller.CinchWebOutputDealDetail{
			TrumpSuit: det.TrumpSuit,
			BidderIdx: det.BidderIdx,
			Bid:       det.Bid,
			SetBack:   det.SetBack,
			Points:    det.Points,
			Gained:    det.Gained,
		}
	}
	return resObj
}

// playableIndices は人間プレイヤーがプレイできるカードのインデックスを返す。
func (p *CinchWebPresenter) playableIndices(g interfaces.CinchGame) []int {
	if g.GetPhase() != domain.CinchPhasePlay || !g.IsHumanTurn() {
		return make([]int, 0)
	}
	idx := g.GetPlayableIndices(g.GetCurrentTurn())
	if idx == nil {
		return make([]int, 0)
	}
	return idx
}

// cinchTrickToOutput はトリックを WebOutput 表現に変換する。
func cinchTrickToOutput(trick []*domain.TrickCard) []*controller.WebOutputTrickCard {
	return buildTrickCards(trick, func(tc *domain.TrickCard) *controller.WebOutputTrickCard {
		if tc == nil {
			return nil
		}
		return &controller.WebOutputTrickCard{PlayerIdx: tc.PlayerIdx, Card: cardToOutput(tc.Card)}
	})
}

// encodeScoresParam は最終スコアを "0:12,1:-3" 形式の文字列に詰める。
func (p *CinchWebPresenter) encodeScoresParam(g interfaces.CinchGame) string {
	parts := make([]string, 0, g.GetPlayerCnt())
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		parts = append(parts, fmt.Sprintf("%d:%d", i, player.GetTotalScore()))
	}
	return strings.Join(parts, ",")
}

// buildResultMessage はゲーム終了時のフォールバック (英語) メッセージ。
func (p *CinchWebPresenter) buildResultMessage(g interfaces.CinchGame) string {
	msg := "Game over. "
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		name := fmt.Sprintf("CPU %d", i)
		if player.GetIsHuman() {
			name = "You"
		}
		msg += fmt.Sprintf("%s:%dpt ", name, player.GetTotalScore())
	}
	return msg
}

// HintOutput はヒント情報を JSON 出力する。
func (p *CinchWebPresenter) HintOutput(g interfaces.CinchGame) string {
	resObj := p.buildBase(g)
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.CinchWebOutputHint{
			CardIndices: hint.CardIndices,
			Bid:         hint.Bid,
			TrumpSuit:   hint.TrumpSuit,
			Reason:      hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput は棋譜を JSON 出力する。
func (p *CinchWebPresenter) ActionLogOutput(g interfaces.CinchGame) string {
	return actionLogOutputJSON(g)
}
