//go:build !js || !wasm || extra3

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// CoincheWebPresenter コワンシュWebプレゼンタークラス
type CoincheWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *CoincheWebPresenter) Output(b interfaces.CoincheGame, lastErr error) string {
	resObj := p.buildBase(b)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(b, b.GetCurrentTrick(), lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **フェーズと手番はここでは見ない。**Coinche.GetHint() が自分で
	// 「人間の手番で、かつ行動を選べる状態か」を確かめて nil を返す。
	if hint := b.GetHint(); hint != nil {
		resObj.Hint = &controller.CoincheWebOutputHint{
			CardIndex: hint.CardIndex,
			Bid:       hint.Bid,
			Suit:      hint.Suit,
			Reason:    hint.Reason,
		}
	}

	return marshalOrError(resObj)
}

func (p *CoincheWebPresenter) buildBase(b interfaces.CoincheGame) *controller.CoincheWebOutput {
	resObj := new(controller.CoincheWebOutput)
	resObj.Phase = int(b.GetPhase())
	resObj.RoundNumber = b.GetRoundNumber()
	resObj.TrickNumber = b.GetTrickNumber()
	resObj.CurrentPlayerIdx = b.GetCurrentPlayerIdx()
	resObj.BidPlayerIdx = b.GetBidPlayerIdx()
	resObj.DealerIdx = b.GetDealerIdx()
	resObj.TrumpSuit = b.GetTrumpSuit()
	resObj.ContractPoints = b.GetContractPoints()
	resObj.Multiplier = b.GetMultiplier()
	resObj.Double = int(b.GetDouble())
	resObj.BiddablePoints = append([]int{}, b.GetBiddablePoints()...)
	resObj.MakerTeam = b.GetMakerTeam()
	resObj.MakerPlayerIdx = b.GetMakerPlayerIdx()
	resObj.TeamScores = [2]int{b.GetTeamScore(0), b.GetTeamScore(1)}
	resObj.RoundPoints = [2]int{b.GetRoundPoints(0), b.GetRoundPoints(1)}
	resObj.RoundBeloteBonus = [2]int{b.GetRoundBeloteBonus(0), b.GetRoundBeloteBonus(1)}
	resObj.GameEndFlag = b.GetGameEndFlag()
	resObj.WinnerTeam = b.GetWinnerTeam()
	resObj.LeadPlayerIdx = b.GetLeadPlayerIdx()

	cfg := b.GetConfig()
	resObj.Config = controller.CoincheWebOutputConfig{
		CpuDifficulty:        int(cfg.CpuDifficulty),
		TargetScore:          cfg.TargetScore,
		DixDeDer:             cfg.DixDeDer,
		EnableBeloteRebelote: cfg.EnableBeloteRebelote,
	}

	resObj.CurrentTrick = trickCardsToOutput(b.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(b)
	return resObj
}

func (p *CoincheWebPresenter) buildPlayersOutput(b interfaces.CoincheGame) []*controller.CoincheWebOutputPlayer {
	out := make([]*controller.CoincheWebOutputPlayer, 0)
	for i := 0; i < b.GetPlayerCnt(); i++ {
		player := b.GetPlayer(i)
		pObj := &controller.CoincheWebOutputPlayer{
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

func (p *CoincheWebPresenter) buildMessage(b interfaces.CoincheGame, trick []*domain.TrickCard, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if b.GetGameEndFlag() {
		winnerTeam := b.GetWinnerTeam()
		msg := fmt.Sprintf("ゲーム終了！ チーム%dの勝ち！", winnerTeam)
		code := fmt.Sprintf("coinche.result.team%dWin", winnerTeam)
		params := map[string]string{"team": fmt.Sprintf("%d", winnerTeam)}
		return msg, code, params
	}
	switch b.GetPhase() {
	case domain.CoinchePhaseBid:
		return "", "coinche.bidPhase", nil
	case domain.CoinchePhaseDouble:
		return "", "coinche.doublePhase", nil
	case domain.CoinchePhasePlay:
		if len(trick) == 0 {
			return "", "coinche.playPhase.lead", nil
		}
		return "", "coinche.playPhase.follow", nil
	case domain.CoinchePhaseTrickEnd:
		return "", "coinche.trickEnd", nil
	case domain.CoinchePhaseRoundEnd:
		return "", "coinche.roundEnd", nil
	}
	return "", "", nil
}

// HintOutput ヒント情報をJSON出力する
func (p *CoincheWebPresenter) HintOutput(b interfaces.CoincheGame) string {
	hint := b.GetHint()
	resObj := p.buildBase(b)
	if hint != nil {
		resObj.Hint = &controller.CoincheWebOutputHint{
			CardIndex: hint.CardIndex,
			Bid:       hint.Bid,
			Suit:      hint.Suit,
			Reason:    hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *CoincheWebPresenter) ActionLogOutput(b interfaces.CoincheGame) string {
	return actionLogOutputJSON(b)
}
