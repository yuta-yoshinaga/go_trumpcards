//go:build !js || !wasm || extra2

package presenter

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// ShelemWebPresenter シェレムWebプレゼンタークラス
type ShelemWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *ShelemWebPresenter) Output(s interfaces.ShelemGame, lastErr error) string {
	resObj := p.buildBase(s)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(s, lastErr)
	// 受動ヒントは Output() でも埋める (#4483)。
	if hint := s.GetHint(); hint != nil {
		resObj.Hint = &controller.ShelemWebOutputHint{
			CardIndex: hint.CardIndex, Reason: hint.Reason, Value: hint.Value, Suit: hint.Suit,
		}
	}
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *ShelemWebPresenter) buildBase(s interfaces.ShelemGame) *controller.ShelemWebOutput {
	resObj := new(controller.ShelemWebOutput)
	resObj.Phase = int(s.GetPhase())
	resObj.RoundNumber = s.GetRoundNumber()
	resObj.TrickNumber = s.GetTrickNumber()
	resObj.TrumpSuit = s.GetTrumpSuit()
	resObj.DeclarerIdx = s.GetDeclarerIdx()
	resObj.Contract = s.GetContract()
	resObj.ShelemBid = s.GetShelemBid()
	// **次に出せる最小額をワイヤに載せる。** 載せないとクライアントは
	// 上回らない入札を出してしまい、サーバに拒否されるまで分からない。
	resObj.MinBid = shelemNextBid(s.GetContract())
	resObj.WidowSize = s.GetWidowSize()
	resObj.DiscardCount = domain.ShelemWidowSize
	resObj.CurrentPlayerIdx = s.GetCurrentPlayerIdx()
	resObj.BidPlayerIdx = s.GetBidPlayerIdx()
	resObj.LeadPlayerIdx = s.GetLeadPlayerIdx()
	resObj.DealerIdx = s.GetDealerIdx()
	resObj.ValidPlays = intSliceOrEmpty(s.GetValidPlayIndices(0))
	resObj.GameEndFlag = s.GetGameEndFlag()
	resObj.WinnerTeam = s.GetWinnerTeam()
	resObj.CurrentTrick = trickCardsToOutput(s.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(s)

	scores := make([]int, 0, domain.ShelemTeamCnt)
	points := make([]int, 0, domain.ShelemTeamCnt)
	tricks := make([]int, 0, domain.ShelemTeamCnt)
	for team := 0; team < domain.ShelemTeamCnt; team++ {
		scores = append(scores, s.GetScore(team))
		points = append(points, s.GetRoundPoints(team))
		tricks = append(tricks, s.TeamTricks(team))
	}
	resObj.Scores = scores
	resObj.RoundPoints = points
	resObj.TeamTricks = tricks
	resObj.Config = controller.ShelemWebOutputConfig{Target: s.GetConfig().Target}
	return resObj
}

// shelemNextBid 現在の落札額を上回る最小の入札額
func shelemNextBid(contract int) int {
	if contract < domain.ShelemMinBid {
		return domain.ShelemMinBid
	}
	next := contract + domain.ShelemBidStep
	if next > domain.ShelemMaxBid {
		return domain.ShelemMaxBid
	}
	return next
}

// buildPlayersOutput プレイヤー情報を構築
func (p *ShelemWebPresenter) buildPlayersOutput(s interfaces.ShelemGame) []*controller.ShelemWebOutputPlayer {
	out := make([]*controller.ShelemWebOutputPlayer, 0)
	for i := 0; i < s.GetPlayerCnt(); i++ {
		player := s.GetPlayer(i)
		out = append(out, &controller.ShelemWebOutputPlayer{
			ID:             i,
			IsHuman:        player.GetIsHuman(),
			Team:           domain.ShelemTeamOf(i),
			CardCount:      player.GetCardsSize(),
			Cards:          playerCardsToOutput(player, player.GetIsHuman()),
			Bid:            player.GetBid(),
			Passed:         player.GetPassed(),
			DeclaredShelem: player.GetDeclaredShelem(),
			TrickCount:     player.GetTrickCount(),
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *ShelemWebPresenter) buildMessage(s interfaces.ShelemGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if s.GetGameEndFlag() {
		params := map[string]string{
			"t0": strconv.Itoa(s.GetScore(0)),
			"t1": strconv.Itoa(s.GetScore(1)),
		}
		switch s.GetWinnerTeam() {
		case 0:
			return "", "shelem.result.team0", params
		case 1:
			return "", "shelem.result.team1", params
		default:
			return "", "shelem.result.tie", params
		}
	}
	switch s.GetPhase() {
	case domain.ShelemPhaseBid:
		if s.IsHumanBidTurn() {
			return "", "shelem.bid.choose", map[string]string{"min": strconv.Itoa(shelemNextBid(s.GetContract()))}
		}
		return "", "shelem.bid.wait", nil
	case domain.ShelemPhaseDiscard:
		if s.IsHumanDiscardTurn() {
			return "", "shelem.discard.choose", map[string]string{
				"n":        strconv.Itoa(domain.ShelemWidowSize),
				"contract": strconv.Itoa(s.GetContract()),
			}
		}
		return "", "shelem.discard.wait", nil
	case domain.ShelemPhaseRoundEnd:
		// **Shelem と通常契約で結末の意味が違う。** 別の文言にする。
		if s.GetShelemBid() {
			return "", "shelem.roundEnd.shelem", map[string]string{
				"round": strconv.Itoa(s.GetRoundNumber()),
			}
		}
		return "", "shelem.roundEnd", map[string]string{
			"round":    strconv.Itoa(s.GetRoundNumber()),
			"contract": strconv.Itoa(s.GetContract()),
			"got":      strconv.Itoa(s.GetRoundPoints(domain.ShelemTeamOf(s.GetDeclarerIdx()))),
		}
	}
	return "", "shelem.play", nil
}

// HintOutput ヒント情報をJSON出力する
func (p *ShelemWebPresenter) HintOutput(s interfaces.ShelemGame) string {
	resObj := p.buildBase(s)
	if hint := s.GetHint(); hint != nil {
		resObj.Hint = &controller.ShelemWebOutputHint{
			CardIndex: hint.CardIndex, Reason: hint.Reason, Value: hint.Value, Suit: hint.Suit,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *ShelemWebPresenter) ActionLogOutput(s interfaces.ShelemGame) string {
	return actionLogOutputJSON(s)
}
