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

// TappTarockWebPresenter タップ・タロックの Web プレゼンター。
type TappTarockWebPresenter struct{}

// Output ゲーム状態を JSON 出力する。
func (p *TappTarockWebPresenter) Output(g interfaces.TappTarockGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	// **受動ヒントは Output() でも埋める。** HintOutput() は command:"hint" 専用の
	// レスポンスでページの state にはマージされないので、ここで埋めないと
	// state.hint が常に undefined になる (#4483)。
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.TappTarockWebOutputHint{
			Bid:            hint.Bid,
			CardIndex:      hint.CardIndex,
			DiscardIndices: hint.DiscardIndices,
			Reason:         hint.Reason,
		}
		if resObj.Hint.DiscardIndices == nil {
			resObj.Hint.DiscardIndices = make([]int, 0)
		}
	}
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築する。
func (p *TappTarockWebPresenter) buildBase(g interfaces.TappTarockGame) *controller.TappTarockWebOutput {
	resObj := new(controller.TappTarockWebOutput)
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
	resObj.ContractName = domain.TappTarockBidName(g.GetContract())
	resObj.TalonCount = g.GetTalonSize()
	resObj.LastTrickWinner = g.GetLastTrickWinner()
	resObj.Outcome = int(g.GetOutcome())
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerPlayer = g.GetWinnerPlayer()
	resObj.IsHumanTurn = g.IsHumanTurn()
	resObj.Config = controller.TappTarockWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetDeals:   cfg.TargetDeals,
	}

	resObj.PlayableIndices = p.playableIndices(g)
	resObj.DiscardableIndices = p.discardableIndices(g)
	resObj.CurrentTrick = trickCardsToOutputWithFace(g.GetCurrentTrick(), koenigrufenFace)
	resObj.LastTrickCards = make([]*controller.WebOutputCard, 0, len(g.GetLastTrickCards()))
	for _, c := range g.GetLastTrickCards() {
		resObj.LastTrickCards = append(resObj.LastTrickCards, cardToOutputWithFace(c, koenigrufenFace))
	}
	resObj.Players = p.buildPlayersOutput(g)
	resObj.Breakdown = p.buildBreakdown(g)
	return resObj
}

// discardableIndices 人間のデクレアラーが伏せられる手札のインデックスを返す。
func (p *TappTarockWebPresenter) discardableIndices(g interfaces.TappTarockGame) []int {
	if g.GetDeclarerIdx() != 0 || g.GetPhase() != domain.TappTarockPhaseTalon {
		return make([]int, 0)
	}
	if idx := g.GetDiscardableIndices(); idx != nil {
		return idx
	}
	return make([]int, 0)
}

// playableIndices 人間が出せる手札のインデックスを返す。
func (p *TappTarockWebPresenter) playableIndices(g interfaces.TappTarockGame) []int {
	if g.GetPhase() != domain.TappTarockPhasePlay || !g.IsHumanTurn() {
		return make([]int, 0)
	}
	if idx := g.GetValidPlayIndices(g.GetCurrentPlayerIdx()); idx != nil {
		return idx
	}
	return make([]int, 0)
}

// buildPlayersOutput 席の情報を構築する (人間のみ手札を公開)。
func (p *TappTarockWebPresenter) buildPlayersOutput(g interfaces.TappTarockGame) []*controller.TappTarockWebOutputPlayer {
	declarer := g.GetDeclarerIdx()
	out := make([]*controller.TappTarockWebOutputPlayer, 0, g.GetPlayerCnt())
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		out = append(out, &controller.TappTarockWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutputWithFace(player, player.GetIsHuman(), koenigrufenFace),
			TrickCount: player.GetTrickCount(),
			CardPoints: g.GetCardPoints(i),
			Score:      g.GetPlayerScore(i),
			IsDeclarer: declarer >= 0 && i == declarer,
		})
	}
	return out
}

// buildBreakdown 直近ディールの精算内訳を構築する。
func (p *TappTarockWebPresenter) buildBreakdown(g interfaces.TappTarockGame) *controller.TappTarockWebOutputBreakdown {
	bd := g.GetBreakdown()
	if bd == nil {
		return nil
	}
	seats := make([]int, len(bd.Seats))
	copy(seats, bd.Seats)
	return &controller.TappTarockWebOutputBreakdown{
		Contract:   int(bd.Contract),
		TeamPoints: bd.TeamPoints,
		Threshold:  bd.Threshold,
		Won:        bd.Won,
		Solo:       bd.Solo,
		Base:       bd.Base,
		Seats:      seats,
		Loser:      bd.Loser,
		Name:       domain.TappTarockBidName(bd.Contract),
	}
}

// buildMessage 結果メッセージを構築する。
func (p *TappTarockWebPresenter) buildMessage(g interfaces.TappTarockGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		return p.buildResultMessage(g), "tapptarock.result.scores",
			map[string]string{"scores": p.encodeScoresParam(g)}
	}
	return "", "", nil
}

// encodeScoresParam 通算得点を "0:12,1:-4" 形式に詰める。
func (p *TappTarockWebPresenter) encodeScoresParam(g interfaces.TappTarockGame) string {
	parts := make([]string, 0, g.GetPlayerCnt())
	for i := 0; i < g.GetPlayerCnt(); i++ {
		parts = append(parts, strconv.Itoa(i)+":"+strconv.Itoa(g.GetPlayerScore(i)))
	}
	return strings.Join(parts, ",")
}

// buildResultMessage 終局時のフォールバック (英語) メッセージ。
func (p *TappTarockWebPresenter) buildResultMessage(g interfaces.TappTarockGame) string {
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
func (p *TappTarockWebPresenter) HintOutput(g interfaces.TappTarockGame) string {
	return p.Output(g, nil)
}

// ActionLogOutput 棋譜を JSON 出力する。
func (p *TappTarockWebPresenter) ActionLogOutput(g interfaces.TappTarockGame) string {
	return actionLogOutputJSON(g)
}
