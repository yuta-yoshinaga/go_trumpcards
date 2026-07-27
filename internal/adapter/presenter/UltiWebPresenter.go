//go:build !js || !wasm || solo

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// UltiWebPresenter ウルティ (Ulti) のWebプレゼンタークラス
type UltiWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *UltiWebPresenter) Output(g interfaces.UltiGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *UltiWebPresenter) buildBase(g interfaces.UltiGame) *controller.UltiWebOutput {
	resObj := new(controller.UltiWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.LeadPlayerIdx = g.GetLeadPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.DeclarerIdx = g.GetDeclarerIdx()
	resObj.Contract = int(g.GetContract())
	resObj.TrumpSuit = g.GetTrumpSuit()
	resObj.TalonCount = g.GetTalonCount()
	resObj.TalonTaken = g.GetTalonTaken()
	resObj.DiscardCount = g.GetDiscardCount()
	resObj.Outcome = int(g.GetOutcome())
	resObj.Result = int(g.GetResult())
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerPlayer = g.GetWinnerPlayer()
	resObj.PlayerCoins = g.GetPlayerCoins()
	resObj.LastTrickWinner = -1
	resObj.IsHumanTurn = g.IsHumanTurn()
	resObj.IsHumanBidTurn = g.IsHumanBidTurn()

	resObj.PlayableIndices = p.playableIndices(g)

	cfg := g.GetConfig()
	resObj.Config = controller.UltiWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetRounds:  cfg.TargetRounds,
	}

	resObj.CurrentTrick = p.buildTrickOutput(g.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// playableIndices 人間プレイヤーがプレイできるカードのインデックスを返す
func (p *UltiWebPresenter) playableIndices(g interfaces.UltiGame) []int {
	if g.GetPhase() != domain.UltiPhasePlay || !g.IsHumanTurn() {
		return make([]int, 0)
	}
	idx := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
	if idx == nil {
		return make([]int, 0)
	}
	return idx
}

// buildTrickOutput 現在のトリック情報を構築
func (p *UltiWebPresenter) buildTrickOutput(trick []*domain.UltiTrickCard) []*controller.UltiWebOutputTrickCard {
	return buildTrickCards(trick, func(tc *domain.UltiTrickCard) *controller.UltiWebOutputTrickCard {
		return &controller.UltiWebOutputTrickCard{PlayerIdx: tc.PlayerIdx, Card: cardToOutput(tc.Card)}
	})
}

// buildPlayersOutput プレイヤー情報を構築
func (p *UltiWebPresenter) buildPlayersOutput(g interfaces.UltiGame) []*controller.UltiWebOutputPlayer {
	coins := g.GetPlayerCoins()
	declarer := g.GetDeclarerIdx()
	out := make([]*controller.UltiWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		out = append(out, &controller.UltiWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, player.GetIsHuman()),
			TrickCount: player.GetTrickCount(),
			CardPoints: g.GetCardPoints(i),
			Coins:      coins[i],
			IsDeclarer: i == declarer,
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *UltiWebPresenter) buildMessage(g interfaces.UltiGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		return p.winnerMessage(g)
	}
	switch g.GetPhase() {
	case domain.UltiPhaseBid:
		return "", "ulti.bidPhase", nil
	case domain.UltiPhaseDiscard:
		return "", "ulti.discardPhase", nil
	case domain.UltiPhasePlay:
		if len(g.GetCurrentTrick()) == 0 {
			return "", "ulti.playPhase.lead", nil
		}
		return "", "ulti.playPhase.follow", nil
	case domain.UltiPhaseTrickEnd:
		return "", "ulti.trickEnd", nil
	case domain.UltiPhaseRoundEnd:
		return "", ultiOutcomeMessageCode(g.GetOutcome()), nil
	}
	return "", "", nil
}

// ultiOutcomeMessageCode ディール結果に対応するメッセージコードを返す。
func ultiOutcomeMessageCode(o domain.UltiOutcome) string {
	switch o {
	case domain.UltiOutcomeWin:
		return "ulti.roundEnd.win"
	case domain.UltiOutcomeLoss:
		return "ulti.roundEnd.loss"
	default:
		return "ulti.roundEnd"
	}
}

// winnerMessage 勝者プレイヤーメッセージを構築する
func (p *UltiWebPresenter) winnerMessage(g interfaces.UltiGame) (string, string, map[string]string) {
	winner := g.GetWinnerPlayer()
	humanIdx := -1
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if player := g.GetPlayer(i); player != nil && player.GetIsHuman() {
			humanIdx = i
			break
		}
	}
	if humanIdx >= 0 && winner == humanIdx {
		return "", "ulti.result.humanWin", nil
	}
	params := map[string]string{"player": fmt.Sprintf("%d", winner)}
	return "", "ulti.result.cpuWin", params
}

// HintOutput ヒント情報をJSON出力する
func (p *UltiWebPresenter) HintOutput(g interfaces.UltiGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = &controller.UltiWebOutputHint{
			CardIndices: hint.CardIndices,
			Reason:      hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *UltiWebPresenter) ActionLogOutput(g interfaces.UltiGame) string {
	return actionLogOutputJSON(g)
}
