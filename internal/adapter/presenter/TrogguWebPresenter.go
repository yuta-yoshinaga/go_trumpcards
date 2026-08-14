//go:build !js || !wasm || extra

package presenter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// TrogguWebPresenter トロッグの Web プレゼンター。
type TrogguWebPresenter struct{}

// Output ゲーム状態を JSON 出力する。
func (p *TrogguWebPresenter) Output(g interfaces.TrogguGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	// **受動ヒントは Output() でも埋める。** HintOutput() は command:"hint" 専用の
	// レスポンスでページの state にはマージされない (#4483)。
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.TrogguWebOutputHint{
			Bid:       hint.Bid,
			CardIndex: hint.CardIndex,
			Reason:    hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築する。
func (p *TrogguWebPresenter) buildBase(g interfaces.TrogguGame) *controller.TrogguWebOutput {
	resObj := new(controller.TrogguWebOutput)
	cfg := g.GetConfig()
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TotalRounds = cfg.TargetDeals
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.BidPlayerIdx = g.GetBidPlayerIdx()
	resObj.HighestBid = int(g.GetHighestBid())
	resObj.DeclarerIdx = g.GetDeclarerIdx()
	resObj.Contract = int(g.GetContract())
	resObj.ContractName = domain.TrogguBidName(g.GetContract())
	resObj.TalonCount = g.GetTalonSize()
	resObj.LastTrickWinner = g.GetLastTrickWinner()
	resObj.Outcome = int(g.GetOutcome())
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerPlayer = g.GetWinnerPlayer()
	resObj.IsHumanTurn = g.IsHumanTurn()
	resObj.Config = controller.TrogguWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetDeals:   cfg.TargetDeals,
	}

	resObj.PlayableIndices = p.playableIndices(g)
	resObj.CurrentTrick = trickCardsToOutputWithFace(g.GetCurrentTrick(), frenchTarotFace)
	resObj.LastTrickCards = make([]*controller.WebOutputCard, 0, len(g.GetLastTrickCards()))
	for _, c := range g.GetLastTrickCards() {
		resObj.LastTrickCards = append(resObj.LastTrickCards, cardToOutputWithFace(c, frenchTarotFace))
	}
	resObj.Players = p.buildPlayersOutput(g)
	resObj.Breakdown = p.buildBreakdown(g)
	return resObj
}

// playableIndices 人間が出せる手札のインデックスを返す。
func (p *TrogguWebPresenter) playableIndices(g interfaces.TrogguGame) []int {
	if g.GetPhase() != domain.TrogguPhasePlay || !g.IsHumanTurn() {
		return make([]int, 0)
	}
	if idx := g.GetValidPlayIndices(g.GetCurrentPlayerIdx()); idx != nil {
		return idx
	}
	return make([]int, 0)
}

// buildPlayersOutput 席の情報を構築する (人間のみ手札を公開)。
func (p *TrogguWebPresenter) buildPlayersOutput(g interfaces.TrogguGame) []*controller.TrogguWebOutputPlayer {
	declarer := g.GetDeclarerIdx()
	out := make([]*controller.TrogguWebOutputPlayer, 0, g.GetPlayerCnt())
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		out = append(out, &controller.TrogguWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutputWithFace(player, player.GetIsHuman(), frenchTarotFace),
			TrickCount: player.GetTrickCount(),
			CardPoints: g.GetCardPoints(i),
			Score:      g.GetPlayerScore(i),
			IsDeclarer: declarer >= 0 && i == declarer,
		})
	}
	return out
}

// buildBreakdown 直近ディールの精算内訳を構築する。
func (p *TrogguWebPresenter) buildBreakdown(g interfaces.TrogguGame) *controller.TrogguWebOutputBreakdown {
	bd := g.GetBreakdown()
	if bd == nil {
		return nil
	}
	seats := make([]int, len(bd.Seats))
	copy(seats, bd.Seats)
	return &controller.TrogguWebOutputBreakdown{
		Contract:       int(bd.Contract),
		ContractName:   bd.ContractName,
		DeclarerPoints: bd.DeclarerPoints,
		DeclarerTricks: bd.DeclarerTricks,
		Target:         bd.Target,
		TargetIsTricks: bd.TargetIsTricks,
		Won:            bd.Won,
		Base:           bd.Base,
		Seats:          seats,
	}
}

// buildMessage 結果メッセージを構築する。
func (p *TrogguWebPresenter) buildMessage(g interfaces.TrogguGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		return p.buildResultMessage(g), "troggu.result.scores",
			map[string]string{"scores": p.encodeScoresParam(g)}
	}
	return "", "", nil
}

// encodeScoresParam 通算得点を "0:30,1:-10" 形式に詰める。
func (p *TrogguWebPresenter) encodeScoresParam(g interfaces.TrogguGame) string {
	parts := make([]string, 0, g.GetPlayerCnt())
	for i := 0; i < g.GetPlayerCnt(); i++ {
		parts = append(parts, strconv.Itoa(i)+":"+strconv.Itoa(g.GetPlayerScore(i)))
	}
	return strings.Join(parts, ",")
}

// buildResultMessage 終局時のフォールバック (英語) メッセージ。
func (p *TrogguWebPresenter) buildResultMessage(g interfaces.TrogguGame) string {
	msg := "Game over. "
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		name := "CPU"
		if player.GetIsHuman() {
			name = "You"
		}
		msg += fmt.Sprintf("%s:%d ", name, g.GetPlayerScore(i))
	}
	return msg
}

// HintOutput ヒント情報を JSON 出力する。
func (p *TrogguWebPresenter) HintOutput(g interfaces.TrogguGame) string {
	return p.Output(g, nil)
}

// ActionLogOutput 棋譜を JSON 出力する。
func (p *TrogguWebPresenter) ActionLogOutput(g interfaces.TrogguGame) string {
	return actionLogOutputJSON(g)
}
