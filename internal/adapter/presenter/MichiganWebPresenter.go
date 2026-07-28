//go:build !js || !wasm || extra3

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MichiganWebPresenter はミシガン (Michigan) の Web プレゼンタークラス。
type MichiganWebPresenter struct{}

// Output はゲーム状態を JSON 出力する。
func (p *MichiganWebPresenter) Output(g interfaces.MichiganGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	return marshalOrError(resObj)
}

// buildBase は基本フィールドを埋めた出力オブジェクトを生成する。
func (p *MichiganWebPresenter) buildBase(g interfaces.MichiganGame) *controller.MichiganWebOutput {
	resObj := new(controller.MichiganWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.Ante = g.GetAnte()
	resObj.Chips = g.GetChips()
	resObj.BetBudget = g.GetBetBudget()
	resObj.HumanBetPlaced = g.GetHumanBetPlaced()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.LeadPlayerIdx = g.GetLeadPlayerIdx()
	resObj.SeqSuit = g.GetSeqSuit()
	resObj.SeqSuitName = michiganSuitName(g.GetSeqSuit())
	resObj.SeqHighValue = g.GetSeqHighValue()
	resObj.NeedNewSequence = g.GetPhase() == domain.MichiganPhasePlay && g.GetSeqSuit() == 0
	resObj.DeadHandCount = g.GetDeadHandCount()
	resObj.IsHumanTurn = g.IsHumanTurn()
	resObj.PlayableIndices = michiganIntsOrEmpty(g.GetPlayableIndices())
	resObj.WinnerIdx = g.GetWinnerIdx()
	resObj.MatchWinnerIdx = g.GetMatchWinnerIdx()
	resObj.Result = int(g.GetResult())
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.Players = p.buildPlayersOutput(g)
	resObj.Boodles = p.buildBoodlesOutput(g)

	cfg := g.GetConfig()
	resObj.Config = controller.MichiganWebOutputConfig{
		PlayerCount:   cfg.PlayerCount,
		Ante:          cfg.Ante,
		StartingChips: cfg.StartingChips,
		TargetRounds:  cfg.TargetRounds,
	}
	return resObj
}

// michiganIntsOrEmpty は nil を空スライスに変換する。
func michiganIntsOrEmpty(v []int) []int {
	if v == nil {
		return make([]int, 0)
	}
	return v
}

// michiganSuitName はスート値を i18n キー用の文字列に変換する (0 = 空)。
func michiganSuitName(suit int) string {
	switch suit {
	case domain.CardDesignSpade:
		return "spade"
	case domain.CardDesignClover:
		return "clover"
	case domain.CardDesignHeart:
		return "heart"
	case domain.CardDesignDiamond:
		return "diamond"
	default:
		return ""
	}
}

// buildPlayersOutput はプレイヤー情報を構築する。人間は常に手札公開。結果フェーズでは
// 全プレイヤーの手札も公開する。
func (p *MichiganWebPresenter) buildPlayersOutput(g interfaces.MichiganGame) []*controller.MichiganWebOutputPlayer {
	out := make([]*controller.MichiganWebOutputPlayer, 0)
	reveal := g.GetPhase() == domain.MichiganPhaseResult
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		showCards := player.GetIsHuman() || reveal
		out = append(out, &controller.MichiganWebOutputPlayer{
			ID:        i,
			IsHuman:   player.GetIsHuman(),
			Chips:     player.GetChips(),
			RoundBet:  player.GetRoundBet(),
			CardCount: player.GetCardsSize(),
			Cards:     playerCardsToOutput(player, showCards),
			IsCurrent: i == g.GetCurrentPlayerIdx() && g.GetPhase() == domain.MichiganPhasePlay,
			IsWinner:  i == g.GetWinnerIdx(),
		})
	}
	return out
}

// buildBoodlesOutput はブードル情報を構築する。
func (p *MichiganWebPresenter) buildBoodlesOutput(g interfaces.MichiganGame) []*controller.MichiganWebOutputBoodle {
	out := make([]*controller.MichiganWebOutputBoodle, 0, g.GetBoodleCnt())
	for i := 0; i < g.GetBoodleCnt(); i++ {
		b := g.GetBoodle(i)
		if b == nil {
			continue
		}
		out = append(out, &controller.MichiganWebOutputBoodle{
			Card:      cardToOutput(b.GetCard()),
			Chips:     b.GetChips(),
			ClaimedBy: b.GetClaimedBy(),
		})
	}
	return out
}

// buildMessage はゲーム結果メッセージを構築する。
func (p *MichiganWebPresenter) buildMessage(g interfaces.MichiganGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		return p.winnerMessage(g)
	}
	switch g.GetPhase() {
	case domain.MichiganPhaseBet:
		return "", "michigan.betPhase", nil
	case domain.MichiganPhasePlay:
		return "", "michigan.playPhase", nil
	case domain.MichiganPhaseResult:
		return p.roundEndMessage(g)
	}
	return "", "", nil
}

// roundEndMessage はラウンド終了時のメッセージを構築する。
func (p *MichiganWebPresenter) roundEndMessage(g interfaces.MichiganGame) (string, string, map[string]string) {
	switch g.GetResult() {
	case domain.MichiganResultWin:
		return "You had the best round!", "michigan.roundEndHumanWin", nil
	case domain.MichiganResultLose:
		return "A rival did better this round.", "michigan.roundEndHumanLose", nil
	default:
		return "The round is over.", "michigan.roundEnd", nil
	}
}

// winnerMessage は試合終了メッセージを構築する。
func (p *MichiganWebPresenter) winnerMessage(g interfaces.MichiganGame) (string, string, map[string]string) {
	winner := g.GetMatchWinnerIdx()
	pl := g.GetPlayer(winner)
	if pl != nil && pl.GetIsHuman() {
		return "Game over! You win!", "michigan.result.humanWin", nil
	}
	params := map[string]string{"player": fmt.Sprintf("%d", winner)}
	return fmt.Sprintf("Game over! CPU %d wins!", winner), "michigan.result.cpuWin", params
}

// HintOutput はヒント情報を JSON 出力する。
func (p *MichiganWebPresenter) HintOutput(g interfaces.MichiganGame) string {
	resObj := p.buildBase(g)
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.MichiganWebOutputHint{
			CardIndex: hint.CardIndex,
			Reason:    hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput は棋譜を JSON 出力する。
func (p *MichiganWebPresenter) ActionLogOutput(g interfaces.MichiganGame) string {
	return actionLogOutputJSON(g)
}
