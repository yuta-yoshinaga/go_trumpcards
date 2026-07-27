//go:build !js || !wasm || extra

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// OmbreWebPresenter オンブル (Ombre) のWebプレゼンタークラス
type OmbreWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *OmbreWebPresenter) Output(g interfaces.OmbreGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *OmbreWebPresenter) buildBase(g interfaces.OmbreGame) *controller.OmbreWebOutput {
	resObj := new(controller.OmbreWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.CurrentBidderIdx = g.GetCurrentBidderIdx()
	resObj.LeadPlayerIdx = g.GetLeadPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.ForehandIdx = g.GetForehandIdx()
	resObj.OmbreIdx = g.GetOmbreIdx()
	resObj.WinningBid = int(g.GetWinningBid())
	resObj.TrumpSuit = g.GetTrumpSuit()
	resObj.Outcome = int(g.GetOutcome())
	resObj.Result = int(g.GetResult())
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerPlayer = g.GetWinnerPlayer()
	resObj.PlayerScores = g.GetPlayerScores()
	resObj.LastTrickWinner = -1
	resObj.IsHumanTurn = g.IsHumanTurn()
	resObj.IsHumanBidTurn = g.IsHumanBidTurn()

	resObj.PlayableIndices = p.playableIndices(g)

	cfg := g.GetConfig()
	resObj.Config = controller.OmbreWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetRounds:  cfg.TargetRounds,
	}

	resObj.CurrentTrick = p.buildTrickOutput(g.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// playableIndices 人間プレイヤーがプレイできるカードのインデックスを返す
func (p *OmbreWebPresenter) playableIndices(g interfaces.OmbreGame) []int {
	if g.GetPhase() != domain.OmbrePhasePlay || !g.IsHumanTurn() {
		return make([]int, 0)
	}
	idx := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
	if idx == nil {
		return make([]int, 0)
	}
	return idx
}

// buildTrickOutput 現在のトリック情報を構築
func (p *OmbreWebPresenter) buildTrickOutput(trick []*domain.OmbreTrickCard) []*controller.OmbreWebOutputTrickCard {
	return buildTrickCards(trick, func(tc *domain.OmbreTrickCard) *controller.OmbreWebOutputTrickCard {
		return &controller.OmbreWebOutputTrickCard{PlayerIdx: tc.PlayerIdx, Card: cardToOutput(tc.Card)}
	})
}

// buildPlayersOutput プレイヤー情報を構築
func (p *OmbreWebPresenter) buildPlayersOutput(g interfaces.OmbreGame) []*controller.OmbreWebOutputPlayer {
	scores := g.GetPlayerScores()
	ombre := g.GetOmbreIdx()
	out := make([]*controller.OmbreWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		out = append(out, &controller.OmbreWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, player.GetIsHuman()),
			TrickCount: player.GetTrickCount(),
			Score:      scores[i],
			IsOmbre:    i == ombre,
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *OmbreWebPresenter) buildMessage(g interfaces.OmbreGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		return p.winnerMessage(g)
	}
	switch g.GetPhase() {
	case domain.OmbrePhaseBid:
		return "", "ombre.bidPhase", nil
	case domain.OmbrePhasePlay:
		if len(g.GetCurrentTrick()) == 0 {
			return "", "ombre.playPhase.lead", nil
		}
		return "", "ombre.playPhase.follow", nil
	case domain.OmbrePhaseTrickEnd:
		return "", "ombre.trickEnd", nil
	case domain.OmbrePhaseRoundEnd:
		return "", ombreOutcomeMessageCode(g.GetOutcome()), nil
	}
	return "", "", nil
}

// ombreOutcomeMessageCode ディール結果に対応するメッセージコードを返す。
func ombreOutcomeMessageCode(o domain.OmbreOutcome) string {
	switch o {
	case domain.OmbreOutcomeSacar:
		return "ombre.roundEnd.sacar"
	case domain.OmbreOutcomePuesta:
		return "ombre.roundEnd.puesta"
	case domain.OmbreOutcomeCodille:
		return "ombre.roundEnd.codille"
	default:
		return "ombre.roundEnd"
	}
}

// winnerMessage 勝者プレイヤーメッセージを構築する
func (p *OmbreWebPresenter) winnerMessage(g interfaces.OmbreGame) (string, string, map[string]string) {
	winner := g.GetWinnerPlayer()
	humanIdx := -1
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if player := g.GetPlayer(i); player != nil && player.GetIsHuman() {
			humanIdx = i
			break
		}
	}
	if humanIdx >= 0 && winner == humanIdx {
		return "", "ombre.result.humanWin", nil
	}
	params := map[string]string{"player": fmt.Sprintf("%d", winner)}
	return "", "ombre.result.cpuWin", params
}

// HintOutput ヒント情報をJSON出力する
func (p *OmbreWebPresenter) HintOutput(g interfaces.OmbreGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = &controller.OmbreWebOutputHint{
			CardIndices: hint.CardIndices,
			Reason:      hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *OmbreWebPresenter) ActionLogOutput(g interfaces.OmbreGame) string {
	return actionLogOutputJSON(g)
}
