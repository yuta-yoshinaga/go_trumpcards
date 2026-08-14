//go:build !js || !wasm || extra3

package presenter

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// HasenpfefferWebPresenter ハーゼンプフェファーWebプレゼンタークラス
type HasenpfefferWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *HasenpfefferWebPresenter) Output(h interfaces.HasenpfefferGame, lastErr error) string {
	resObj := p.buildBase(h)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(h, lastErr)
	// 受動ヒントは Output() でも埋める (#4483)。
	if hint := h.GetHint(); hint != nil {
		resObj.Hint = &controller.HasenpfefferWebOutputHint{
			CardIndex: hint.CardIndex, Reason: hint.Reason, Value: hint.Value, Suit: hint.Suit,
		}
	}
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *HasenpfefferWebPresenter) buildBase(h interfaces.HasenpfefferGame) *controller.HasenpfefferWebOutput {
	resObj := new(controller.HasenpfefferWebOutput)
	resObj.Phase = int(h.GetPhase())
	resObj.HandNumber = h.GetHandNumber()
	resObj.TrickNumber = h.GetTrickNumber()
	resObj.TrumpSuit = h.GetTrumpSuit()
	resObj.DeclarerIdx = h.GetDeclarerIdx()
	resObj.Contract = h.GetContract()
	// **押せない宣言額をワイヤに載せる。** 載せないとサーバが必ず拒否する額を
	// 出してしまう (#5304)。0 は「もう宣言できない = 降りるしかない」。
	resObj.MinBid = h.NextBid()
	resObj.MustBid = h.MustBid(0)
	resObj.BlindSize = h.GetBlindSize()
	resObj.LastHandEuchred = h.GetLastHandEuchred()
	resObj.LastHandTricks = h.GetLastHandTricks()
	resObj.CurrentPlayerIdx = h.GetCurrentPlayerIdx()
	resObj.LeadPlayerIdx = h.GetLeadPlayerIdx()
	resObj.DealerIdx = h.GetDealerIdx()
	resObj.ValidPlays = intSliceOrEmpty(h.GetValidPlayIndices(0))
	resObj.GameEndFlag = h.GetGameEndFlag()
	resObj.WinnerTeam = h.GetWinnerTeam()
	resObj.CurrentTrick = trickCardsToOutput(h.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(h)

	scores := make([]int, 0, domain.HasenpfefferTeamCnt)
	tricks := make([]int, 0, domain.HasenpfefferTeamCnt)
	for team := 0; team < domain.HasenpfefferTeamCnt; team++ {
		scores = append(scores, h.GetScore(team))
		tricks = append(tricks, h.TeamTricks(team))
	}
	resObj.Scores = scores
	resObj.TeamTricks = tricks
	resObj.Config = controller.HasenpfefferWebOutputConfig{Target: h.GetConfig().Target}
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築
func (p *HasenpfefferWebPresenter) buildPlayersOutput(h interfaces.HasenpfefferGame) []*controller.HasenpfefferWebOutputPlayer {
	out := make([]*controller.HasenpfefferWebOutputPlayer, 0)
	for i := 0; i < h.GetPlayerCnt(); i++ {
		player := h.GetPlayer(i)
		out = append(out, &controller.HasenpfefferWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			Team:       domain.HasenpfefferTeamOf(i),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, player.GetIsHuman()),
			Bid:        player.GetBid(),
			TrickCount: player.GetTrickCount(),
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *HasenpfefferWebPresenter) buildMessage(h interfaces.HasenpfefferGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if h.GetGameEndFlag() {
		params := map[string]string{
			"t0": strconv.Itoa(h.GetScore(0)),
			"t1": strconv.Itoa(h.GetScore(1)),
		}
		switch h.GetWinnerTeam() {
		case 0:
			return "", "hasenpfeffer.result.team0", params
		case 1:
			return "", "hasenpfeffer.result.team1", params
		default:
			return "", "hasenpfeffer.result.tie", params
		}
	}
	switch h.GetPhase() {
	case domain.HasenpfefferPhaseBid:
		if !h.IsHumanBidTurn() {
			return "", "hasenpfeffer.bid.wait", nil
		}
		// **親は降りられないことがある。** 降りる選択肢が無い場面を別に言う。
		if h.MustBid(0) {
			return "", "hasenpfeffer.bid.must", map[string]string{"min": strconv.Itoa(h.NextBid())}
		}
		if h.NextBid() == 0 {
			return "", "hasenpfeffer.bid.capped", map[string]string{"max": strconv.Itoa(domain.HasenpfefferMaxBid)}
		}
		return "", "hasenpfeffer.bid.choose", map[string]string{"min": strconv.Itoa(h.NextBid())}
	case domain.HasenpfefferPhaseDiscard:
		if h.IsHumanDiscardTurn() {
			return "", "hasenpfeffer.discard.choose", map[string]string{
				"contract": strconv.Itoa(h.GetContract()),
			}
		}
		return "", "hasenpfeffer.discard.wait", nil
	case domain.HasenpfefferPhaseHandEnd:
		params := map[string]string{
			"contract": strconv.Itoa(h.GetContract()),
			"took":     strconv.Itoa(h.GetLastHandTricks()),
		}
		if h.GetLastHandEuchred() {
			return "", "hasenpfeffer.handEnd.euchred", params
		}
		return "", "hasenpfeffer.handEnd.made", params
	default:
		// **ジョーカーが最強の切り札。** これを知らないと打ち方が変わる。
		return "", "hasenpfeffer.play", map[string]string{
			"contract": strconv.Itoa(h.GetContract()),
		}
	}
}

// HintOutput ヒント情報をJSON出力する
func (p *HasenpfefferWebPresenter) HintOutput(h interfaces.HasenpfefferGame) string {
	resObj := p.buildBase(h)
	if hint := h.GetHint(); hint != nil {
		resObj.Hint = &controller.HasenpfefferWebOutputHint{
			CardIndex: hint.CardIndex, Reason: hint.Reason, Value: hint.Value, Suit: hint.Suit,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *HasenpfefferWebPresenter) ActionLogOutput(h interfaces.HasenpfefferGame) string {
	return actionLogOutputJSON(h)
}
