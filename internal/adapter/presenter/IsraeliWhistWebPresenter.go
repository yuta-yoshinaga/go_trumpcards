//go:build !js || !wasm || extra2

package presenter

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// IsraeliWhistWebPresenter イスラエリホイストWebプレゼンタークラス
type IsraeliWhistWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *IsraeliWhistWebPresenter) Output(w interfaces.IsraeliWhistGame, lastErr error) string {
	resObj := p.buildBase(w)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(w, lastErr)
	// 受動ヒントは Output() でも埋める (#4483)。
	if hint := w.GetHint(); hint != nil {
		resObj.Hint = &controller.IsraeliWhistWebOutputHint{
			CardIndex: hint.CardIndex, Reason: hint.Reason, Value: hint.Value, Suit: hint.Suit,
		}
	}
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *IsraeliWhistWebPresenter) buildBase(w interfaces.IsraeliWhistGame) *controller.IsraeliWhistWebOutput {
	resObj := new(controller.IsraeliWhistWebOutput)
	resObj.Phase = int(w.GetPhase())
	resObj.RoundNumber = w.GetRoundNumber()
	resObj.Doubled = w.GetRoundDoubled()
	resObj.DoubledAllExact = w.GetRoundAllExact()
	resObj.TrickNumber = w.GetTrickNumber()
	resObj.TrumpSuit = w.GetTrumpSuit()
	resObj.DeclarerIdx = w.GetDeclarerIdx()
	resObj.HighBid = w.GetHighBid()
	resObj.HighSuit = w.GetHighSuit()
	// **落札者のノルマと禁止値をワイヤに載せる。** 載せないとクライアントは
	// 押せない宣言を出してしまい、サーバに拒否されるまで分からない。
	resObj.MinimumBid = w.MinimumBidFor(0)
	resObj.RestrictedBid = w.GetRestrictedBid()
	resObj.CurrentPlayerIdx = w.GetCurrentPlayerIdx()
	resObj.AuctionPlayerIdx = w.GetAuctionPlayerIdx()
	resObj.BidPlayerIdx = w.GetBidPlayerIdx()
	resObj.LeadPlayerIdx = w.GetLeadPlayerIdx()
	resObj.DealerIdx = w.GetDealerIdx()
	resObj.ValidPlays = intSliceOrEmpty(w.GetValidPlayIndices(0))
	resObj.GameEndFlag = w.GetGameEndFlag()
	resObj.WinnerIdx = w.GetWinnerIdx()
	resObj.CurrentTrick = trickCardsToOutput(w.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(w)
	resObj.Config = controller.IsraeliWhistWebOutputConfig{Rounds: w.GetConfig().Rounds}
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築
func (p *IsraeliWhistWebPresenter) buildPlayersOutput(w interfaces.IsraeliWhistGame) []*controller.IsraeliWhistWebOutputPlayer {
	out := make([]*controller.IsraeliWhistWebOutputPlayer, 0)
	for i := 0; i < w.GetPlayerCnt(); i++ {
		player := w.GetPlayer(i)
		out = append(out, &controller.IsraeliWhistWebOutputPlayer{
			ID:          i,
			IsHuman:     player.GetIsHuman(),
			CardCount:   player.GetCardsSize(),
			Cards:       playerCardsToOutput(player, player.GetIsHuman()),
			AuctionBid:  player.GetAuctionBid(),
			AuctionSuit: player.GetAuctionSuit(),
			Passed:      player.GetPassed(),
			Bid:         player.GetBid(),
			TrickCount:  player.GetTrickCount(),
			RoundScore:  player.GetRoundScore(),
			TotalScore:  player.GetTotalScore(),
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *IsraeliWhistWebPresenter) buildMessage(w interfaces.IsraeliWhistGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if w.GetGameEndFlag() {
		if w.GetWinnerIdx() == 0 {
			return "", "israeliwhist.result.win", map[string]string{"score": strconv.Itoa(w.GetPlayer(0).GetTotalScore())}
		}
		if w.GetWinnerIdx() < 0 {
			return "", "israeliwhist.result.tie", nil
		}
		return "", "israeliwhist.result.lose", map[string]string{"idx": strconv.Itoa(w.GetWinnerIdx())}
	}
	switch w.GetPhase() {
	case domain.IsraeliWhistPhaseAuction:
		if w.IsHumanAuctionTurn() {
			return "", "israeliwhist.auction.choose", map[string]string{"high": strconv.Itoa(w.GetHighBid())}
		}
		return "", "israeliwhist.auction.wait", nil
	case domain.IsraeliWhistPhaseBid:
		// **落札者のノルマが先。** ノルマ未満は押せないので、それを先に伝える。
		if m := w.MinimumBidFor(0); m > 0 && w.IsHumanBidTurn() {
			return "", "israeliwhist.bid.quota", map[string]string{"n": strconv.Itoa(m)}
		}
		if r := w.GetRestrictedBid(); r >= 0 && w.IsHumanBidTurn() {
			return "", "israeliwhist.bid.restricted", map[string]string{"n": strconv.Itoa(r)}
		}
		return "", "israeliwhist.bid.choose", nil
	case domain.IsraeliWhistPhaseRoundEnd:
		return "", "israeliwhist.roundEnd", map[string]string{
			"round": strconv.Itoa(w.GetRoundNumber()),
			"score": strconv.Itoa(w.GetPlayer(0).GetRoundScore()),
		}
	}
	return "", "israeliwhist.play", nil
}

// HintOutput ヒント情報をJSON出力する
func (p *IsraeliWhistWebPresenter) HintOutput(w interfaces.IsraeliWhistGame) string {
	resObj := p.buildBase(w)
	if hint := w.GetHint(); hint != nil {
		resObj.Hint = &controller.IsraeliWhistWebOutputHint{
			CardIndex: hint.CardIndex, Reason: hint.Reason, Value: hint.Value, Suit: hint.Suit,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *IsraeliWhistWebPresenter) ActionLogOutput(w interfaces.IsraeliWhistGame) string {
	return actionLogOutputJSON(w)
}
