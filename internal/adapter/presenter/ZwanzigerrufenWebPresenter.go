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

// ZwanzigerrufenWebPresenter ツヴァンツィガールーフェンの Web プレゼンター。
type ZwanzigerrufenWebPresenter struct{}

// Output ゲーム状態を JSON 出力する。
func (p *ZwanzigerrufenWebPresenter) Output(g interfaces.ZwanzigerrufenGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	// **受動ヒントは Output() でも埋める。** HintOutput() は command:"hint" 専用の
	// レスポンスでページの state にはマージされないので、ここで埋めないと
	// state.hint が常に undefined になる (#4483)。
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.ZwanzigerrufenWebOutputHint{
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

// zwanzigerrufenVisiblePartner 公開してよいパートナーの席を返す。
//
// **判明するまでは -1。** 呼び札が場に出るまで正体を隠すのがこのゲームの骨格なので、
// 画面が出さないのではなく、サーバが送らないことで守る。
func zwanzigerrufenVisiblePartner(g interfaces.ZwanzigerrufenGame) int {
	if !g.GetPartnerRevealed() {
		return -1
	}
	return g.GetPartnerIdx()
}

// buildBase 共通フィールドを構築する。
func (p *ZwanzigerrufenWebPresenter) buildBase(g interfaces.ZwanzigerrufenGame) *controller.ZwanzigerrufenWebOutput {
	resObj := new(controller.ZwanzigerrufenWebOutput)
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
	resObj.ContractName = domain.ZwanzigerrufenBidName(g.GetContract())
	resObj.CalledTrump = g.GetCalledTrump()
	resObj.PartnerRevealed = g.GetPartnerRevealed()
	resObj.PartnerIdx = zwanzigerrufenVisiblePartner(g)
	resObj.TalonCount = g.GetTalonSize()
	resObj.LastTrickWinner = g.GetLastTrickWinner()
	resObj.Outcome = int(g.GetOutcome())
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerPlayer = g.GetWinnerPlayer()
	resObj.IsHumanTurn = g.IsHumanTurn()
	resObj.Config = controller.ZwanzigerrufenWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetDeals:   cfg.TargetDeals,
	}

	resObj.PlayableIndices = p.playableIndices(g)
	resObj.CurrentTrick = trickCardsToOutputWithFace(g.GetCurrentTrick(), koenigrufenFace)
	resObj.LastTrickCards = make([]*controller.WebOutputCard, 0, len(g.GetLastTrickCards()))
	for _, c := range g.GetLastTrickCards() {
		resObj.LastTrickCards = append(resObj.LastTrickCards, cardToOutputWithFace(c, koenigrufenFace))
	}
	resObj.Players = p.buildPlayersOutput(g)
	resObj.Breakdown = p.buildBreakdown(g)
	return resObj
}

// playableIndices 人間が出せる手札のインデックスを返す。
func (p *ZwanzigerrufenWebPresenter) playableIndices(g interfaces.ZwanzigerrufenGame) []int {
	if g.GetPhase() != domain.ZwanzigerrufenPhasePlay || !g.IsHumanTurn() {
		return make([]int, 0)
	}
	if idx := g.GetValidPlayIndices(g.GetCurrentPlayerIdx()); idx != nil {
		return idx
	}
	return make([]int, 0)
}

// buildPlayersOutput 席の情報を構築する (人間のみ手札を公開)。
func (p *ZwanzigerrufenWebPresenter) buildPlayersOutput(g interfaces.ZwanzigerrufenGame) []*controller.ZwanzigerrufenWebOutputPlayer {
	declarer := g.GetDeclarerIdx()
	partner := zwanzigerrufenVisiblePartner(g)
	out := make([]*controller.ZwanzigerrufenWebOutputPlayer, 0, g.GetPlayerCnt())
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		out = append(out, &controller.ZwanzigerrufenWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutputWithFace(player, player.GetIsHuman(), koenigrufenFace),
			TrickCount: player.GetTrickCount(),
			CardPoints: g.GetCardPoints(i),
			Score:      g.GetPlayerScore(i),
			IsDeclarer: declarer >= 0 && i == declarer,
			IsPartner:  partner >= 0 && i == partner,
		})
	}
	return out
}

// buildBreakdown 直近ディールの精算内訳を構築する。
func (p *ZwanzigerrufenWebPresenter) buildBreakdown(g interfaces.ZwanzigerrufenGame) *controller.ZwanzigerrufenWebOutputBreakdown {
	bd := g.GetBreakdown()
	if bd == nil {
		return nil
	}
	seats := make([]int, len(bd.Seats))
	copy(seats, bd.Seats)
	return &controller.ZwanzigerrufenWebOutputBreakdown{
		Contract:   int(bd.Contract),
		TeamPoints: bd.TeamPoints,
		Threshold:  bd.Threshold,
		Won:        bd.Won,
		Solo:       bd.Solo,
		Base:       bd.Base,
		Seats:      seats,
		Loser:      bd.Loser,
		Name:       domain.ZwanzigerrufenBidName(bd.Contract),
	}
}

// buildMessage 結果メッセージを構築する。
func (p *ZwanzigerrufenWebPresenter) buildMessage(g interfaces.ZwanzigerrufenGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		return p.buildResultMessage(g), "zwanzigerrufen.result.scores",
			map[string]string{"scores": p.encodeScoresParam(g)}
	}
	return "", "", nil
}

// encodeScoresParam 通算得点を "0:12,1:-4" 形式に詰める。
func (p *ZwanzigerrufenWebPresenter) encodeScoresParam(g interfaces.ZwanzigerrufenGame) string {
	parts := make([]string, 0, g.GetPlayerCnt())
	for i := 0; i < g.GetPlayerCnt(); i++ {
		parts = append(parts, strconv.Itoa(i)+":"+strconv.Itoa(g.GetPlayerScore(i)))
	}
	return strings.Join(parts, ",")
}

// buildResultMessage 終局時のフォールバック (英語) メッセージ。
func (p *ZwanzigerrufenWebPresenter) buildResultMessage(g interfaces.ZwanzigerrufenGame) string {
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
func (p *ZwanzigerrufenWebPresenter) HintOutput(g interfaces.ZwanzigerrufenGame) string {
	return p.Output(g, nil)
}

// ActionLogOutput 棋譜を JSON 出力する。
func (p *ZwanzigerrufenWebPresenter) ActionLogOutput(g interfaces.ZwanzigerrufenGame) string {
	return actionLogOutputJSON(g)
}
