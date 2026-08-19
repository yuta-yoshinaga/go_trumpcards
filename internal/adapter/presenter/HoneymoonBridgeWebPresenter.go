//go:build !js || !wasm || solo

package presenter

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// HoneymoonBridgeWebPresenter ハネムーンブリッジWebプレゼンタークラス
type HoneymoonBridgeWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *HoneymoonBridgeWebPresenter) Output(s interfaces.HoneymoonBridgeGame, lastErr error) string {
	resObj := p.buildBase(s)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(s, lastErr)
	// 受動ヒントは Output() でも埋める (#4483)。
	resObj.Hint = honeymoonBridgeHintOutput(s)
	return marshalOrError(resObj)
}

// honeymoonBridgeHintOutput はヒントを出力形に変換する。
func honeymoonBridgeHintOutput(s interfaces.HoneymoonBridgeGame) *controller.HoneymoonBridgeWebOutputHint {
	hint := s.GetHint()
	if hint == nil {
		return nil
	}
	return &controller.HoneymoonBridgeWebOutputHint{
		CardIndex: hint.CardIndex, Reason: hint.Reason, Level: hint.Level, Suit: hint.Suit,
	}
}

// buildBase 共通フィールドを構築
func (p *HoneymoonBridgeWebPresenter) buildBase(s interfaces.HoneymoonBridgeGame) *controller.HoneymoonBridgeWebOutput {
	resObj := new(controller.HoneymoonBridgeWebOutput)
	resObj.Phase = int(s.GetPhase())
	resObj.RoundNumber = s.GetRoundNumber()
	resObj.TrickNumber = s.GetTrickNumber()
	resObj.StockSize = s.GetStockSize()
	resObj.TrumpSuit = s.GetTrumpSuit()
	resObj.DeclarerIdx = s.GetDeclarerIdx()
	resObj.ContractLevel = s.GetContractLevel()
	resObj.RequiredTricks = s.RequiredTricks()
	// **競り中だけ「次に通る最小の宣言」を載せる。** これが無いとページが
	// 序列を自前で作り直すことになり、サーバが必ず拒否する値を送る。
	if s.GetPhase() == domain.HoneymoonBridgePhaseBid {
		resObj.MinBidLevel, resObj.MinBidSuit = s.NextBid()
	}
	resObj.LastMade = s.GetLastMade()
	resObj.LastTricks = s.GetLastTricks()
	resObj.LastPoints = s.GetLastPoints()
	resObj.CurrentPlayerIdx = s.GetCurrentPlayerIdx()
	resObj.LeadPlayerIdx = s.GetLeadPlayerIdx()
	resObj.DealerIdx = s.GetDealerIdx()
	resObj.ValidPlays = intSliceOrEmpty(s.GetValidPlayIndices(0))
	resObj.GameEndFlag = s.GetGameEndFlag()
	resObj.WinnerIdx = s.GetWinnerIdx()
	resObj.CurrentTrick = trickCardsToOutput(s.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(s)
	resObj.Config = controller.HoneymoonBridgeWebOutputConfig{Target: s.GetConfig().Target}
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築
func (p *HoneymoonBridgeWebPresenter) buildPlayersOutput(s interfaces.HoneymoonBridgeGame) []*controller.HoneymoonBridgeWebOutputPlayer {
	out := make([]*controller.HoneymoonBridgeWebOutputPlayer, 0)
	for i := 0; i < s.GetPlayerCnt(); i++ {
		player := s.GetPlayer(i)
		out = append(out, &controller.HoneymoonBridgeWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, player.GetIsHuman()),
			BidLevel:   player.GetBidLevel(),
			BidSuit:    player.GetBidSuit(),
			TrickCount: player.GetTrickCount(),
			Score:      player.GetScore(),
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *HoneymoonBridgeWebPresenter) buildMessage(s interfaces.HoneymoonBridgeGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if s.GetGameEndFlag() {
		params := map[string]string{"idx": strconv.Itoa(s.GetWinnerIdx())}
		switch s.GetWinnerIdx() {
		case 0:
			return "", "honeymoonbridge.result.you", params
		case -1:
			return "", "honeymoonbridge.result.tie", params
		default:
			return "", "honeymoonbridge.result.cpu", params
		}
	}
	switch s.GetPhase() {
	case domain.HoneymoonBridgePhaseDraw:
		// **前半は得点にならない。** 山札の残りだけが意味を持つ。
		return "", "honeymoonbridge.draw", map[string]string{
			"stock": strconv.Itoa(s.GetStockSize()),
		}
	case domain.HoneymoonBridgePhaseBid:
		if s.IsHumanBidTurn() {
			return "", "honeymoonbridge.bid.choose", nil
		}
		return "", "honeymoonbridge.bid.wait", nil
	case domain.HoneymoonBridgePhaseRoundEnd:
		if s.GetContractLevel() == 0 {
			// **両者パスならディールは流れる。**
			return "", "honeymoonbridge.roundEnd.passedOut", nil
		}
		code := "honeymoonbridge.roundEnd.down"
		if s.GetLastMade() {
			code = "honeymoonbridge.roundEnd.made"
		}
		return "", code, map[string]string{
			"need": strconv.Itoa(s.RequiredTricks()),
			"took": strconv.Itoa(s.GetLastTricks()),
		}
	default:
		// **落札者が取った数を出す。** 引き合いの直後は落札者がいないので 0。
		took := 0
		if decl := s.GetPlayer(s.GetDeclarerIdx()); decl != nil {
			took = decl.GetTrickCount()
		}
		return "", "honeymoonbridge.play", map[string]string{
			"need": strconv.Itoa(s.RequiredTricks()),
			"took": strconv.Itoa(took),
		}
	}
}

// HintOutput ヒント情報をJSON出力する
func (p *HoneymoonBridgeWebPresenter) HintOutput(s interfaces.HoneymoonBridgeGame) string {
	resObj := p.buildBase(s)
	resObj.Hint = honeymoonBridgeHintOutput(s)
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *HoneymoonBridgeWebPresenter) ActionLogOutput(s interfaces.HoneymoonBridgeGame) string {
	return actionLogOutputJSON(s)
}
