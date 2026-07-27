package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// PitchWebPresenter ピッチWebプレゼンタークラス
type PitchWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *PitchWebPresenter) Output(s interfaces.PitchGame, lastErr error) string {
	resObj := p.buildBase(s)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(s, s.GetCurrentTrick(), lastErr)
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *PitchWebPresenter) buildBase(s interfaces.PitchGame) *controller.PitchWebOutput {
	resObj := new(controller.PitchWebOutput)
	resObj.Phase = int(s.GetPhase())
	resObj.RoundNumber = s.GetRoundNumber()
	resObj.TrickNumber = s.GetTrickNumber()
	resObj.DealerIdx = s.GetDealerIdx()
	resObj.CurrentPlayerIdx = s.GetCurrentPlayerIdx()
	resObj.BidPlayerIdx = s.GetBidPlayerIdx()
	resObj.CurrentBid = s.GetCurrentBid()
	resObj.BidWinnerIdx = s.GetBidWinnerIdx()
	resObj.TrumpSuit = s.GetTrumpSuit()
	resObj.GameEndFlag = s.GetGameEndFlag()
	resObj.WinnerIdx = s.GetWinnerIdx()
	resObj.LeadPlayerIdx = s.GetLeadPlayerIdx()

	cfg := s.GetConfig()
	resObj.Config = controller.PitchWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		PointLimit:    cfg.PointLimit,
	}

	resObj.CurrentTrick = trickCardsToOutput(s.GetCurrentTrick())
	resObj.LastTrick, resObj.LastTrickWinner = p.buildLastTrickOutput(s)
	resObj.Players = p.buildPlayersOutput(s)

	// human の有効プレイインデックスを供給
	for i := 0; i < s.GetPlayerCnt(); i++ {
		if pl := s.GetPlayer(i); pl != nil && pl.GetIsHuman() {
			resObj.ValidPlayIndices = s.GetValidPlayIndices(i)
			break
		}
	}
	if resObj.ValidPlayIndices == nil {
		resObj.ValidPlayIndices = make([]int, 0)
	}
	return resObj
}

// buildLastTrickOutput は直前に解決されたトリック（誰が何を出し誰が取ったか）を
// アクションログから再構築する。ドメインは専用の lastTrick フィールドを持たないが、
// 各トリックの "play" ログ（プレイヤーと札）と "trick_win" ログ（勝者）から
// 現ラウンドの直近トリックを復元できる。ラウンド開始直後（プレイフェーズのトリック 1
// で、この局のトリックがまだ確定していない）は空スライスと -1 を返す。
func (p *PitchWebPresenter) buildLastTrickOutput(s interfaces.PitchGame) ([]*controller.WebOutputTrickCard, int) {
	empty := make([]*controller.WebOutputTrickCard, 0)
	// ラウンド最初のトリックがプレイ中は、この局に確定済みトリックが無いため空を返す。
	if s.GetPhase() == domain.PitchPhasePlay && s.GetTrickNumber() <= 1 {
		return empty, -1
	}

	log := s.GetActionLog()
	winIdx := -1
	for i := len(log) - 1; i >= 0; i-- {
		if log[i] != nil && log[i].ActionType == "trick_win" {
			winIdx = i
			break
		}
	}
	if winIdx < 0 {
		return empty, -1
	}

	// trick_win 直前の "play" ログ（プレイ順）が、そのトリックの各札に対応する。
	var plays []*domain.ActionLogEntry
	for i := 0; i < winIdx; i++ {
		if e := log[i]; e != nil && e.ActionType == "play" && len(e.Cards) > 0 {
			plays = append(plays, e)
		}
	}
	if len(plays) < domain.PitchPlayerCnt {
		return empty, -1
	}
	plays = plays[len(plays)-domain.PitchPlayerCnt:]

	out := make([]*controller.WebOutputTrickCard, 0, len(plays))
	for _, e := range plays {
		out = append(out, &controller.WebOutputTrickCard{
			PlayerIdx: e.PlayerIdx,
			Card:      cardToOutput(e.Cards[0]),
		})
	}
	return out, log[winIdx].PlayerIdx
}

// buildPlayersOutput プレイヤー情報を構築
func (p *PitchWebPresenter) buildPlayersOutput(s interfaces.PitchGame) []*controller.PitchWebOutputPlayer {
	out := make([]*controller.PitchWebOutputPlayer, 0)
	for i := 0; i < s.GetPlayerCnt(); i++ {
		player := s.GetPlayer(i)
		pObj := &controller.PitchWebOutputPlayer{
			ID:              i,
			IsHuman:         player.GetIsHuman(),
			CardCount:       player.GetCardsSize(),
			Cards:           playerCardsToOutput(player, player.GetIsHuman()),
			Bid:             player.GetBid(),
			RoundScore:      player.GetRoundScore(),
			CumulativeScore: player.GetCumulativeScore(),
			TrickCount:      player.GetTrickCount(),
		}
		out = append(out, pObj)
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *PitchWebPresenter) buildMessage(s interfaces.PitchGame, trick []*domain.TrickCard, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if s.GetGameEndFlag() {
		winnerIdx := s.GetWinnerIdx()
		player := s.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		return buildWinnerWebMessage("pitch", winnerIdx, isHuman)
	}
	switch s.GetPhase() {
	case domain.PitchPhaseBid:
		return "", "pitch.bidPhase", nil
	case domain.PitchPhasePlay:
		if len(trick) == 0 {
			return "", "pitch.playPhase.lead", nil
		}
		return "", "pitch.playPhase.follow", nil
	case domain.PitchPhaseTrickEnd:
		return "", "pitch.trickEnd", nil
	case domain.PitchPhaseRoundEnd:
		return "", "pitch.roundEnd", nil
	}
	return "", "", nil
}

// HintOutput ヒント情報をJSON出力する
func (p *PitchWebPresenter) HintOutput(s interfaces.PitchGame) string {
	hint := s.GetHint()
	resObj := p.buildBase(s)
	if hint != nil {
		resObj.Hint = &controller.PitchWebOutputHint{
			CardIndex: hint.CardIndex,
			Bid:       hint.Bid,
			Reason:    hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *PitchWebPresenter) ActionLogOutput(s interfaces.PitchGame) string {
	return actionLogOutputJSON(s)
}
