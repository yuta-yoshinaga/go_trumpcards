package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// OhHellWebPresenter オー・ヘルWebプレゼンタークラス
type OhHellWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *OhHellWebPresenter) Output(o interfaces.OhHellGame, lastErr error) string {
	resObj := p.buildBase(o)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(o, lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **フェーズと手番はここでは見ない。**OhHell.GetHint() が自分で
	// 「人間の手番で、かつ行動を選べる状態か」を確かめて nil を返す。
	if hint := o.GetHint(); hint != nil {
		resObj.Hint = &controller.OhHellWebOutputHint{
			CardIndex: hint.CardIndex,
			Bid:       hint.Bid,
			Reason:    hint.Reason,
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒント情報をJSON出力する
func (p *OhHellWebPresenter) HintOutput(o interfaces.OhHellGame) string {
	hint := o.GetHint()
	resObj := p.buildBase(o)

	if hint != nil {
		resObj.Hint = &controller.OhHellWebOutputHint{
			CardIndex: hint.CardIndex,
			Bid:       hint.Bid,
			Reason:    hint.Reason,
		}
	}
	// **「頼んだヒントか」を CLI が見分けられるようにする。**このゲーム群の
	// `hintAvailable` は画面のラベルとして既に使われているので、別キーを出す (#4483)。
	if hint != nil {
		resObj.MessageCode = "ohhell.hintRequested"
	} else {
		resObj.MessageCode = "ohhell.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *OhHellWebPresenter) ActionLogOutput(o interfaces.OhHellGame) string {
	return actionLogOutputJSON(o)
}

// buildBase 共通フィールドを構築
func (p *OhHellWebPresenter) buildBase(o interfaces.OhHellGame) *controller.OhHellWebOutput {
	resObj := new(controller.OhHellWebOutput)
	resObj.Phase = int(o.GetPhase())
	resObj.RoundNumber = o.GetRoundNumber()
	resObj.TotalRounds = o.GetTotalRounds()
	resObj.HandSize = o.GetHandSize()
	resObj.TrickNumber = o.GetTrickNumber()
	resObj.CurrentPlayerIdx = o.GetCurrentPlayerIdx()
	resObj.BidPlayerIdx = o.GetBidPlayerIdx()
	resObj.DealerIdx = o.GetDealerIdx()
	resObj.TrumpCard = cardToOutput(o.GetTrumpCard())
	resObj.TrumpSuit = o.GetTrumpSuit()
	resObj.RestrictedBid = o.GetRestrictedBid()
	resObj.GameEndFlag = o.GetGameEndFlag()
	resObj.WinnerIdx = o.GetWinnerIdx()
	resObj.LeadPlayerIdx = o.GetLeadPlayerIdx()

	cfg := o.GetConfig()
	resObj.Config = controller.OhHellWebOutputConfig{
		CpuDifficulty:  int(cfg.CpuDifficulty),
		MaxHandSize:    cfg.MaxHandSize,
		ScoringVariant: int(cfg.ScoringVariant),
		RoundDirection: int(cfg.RoundDirection),
	}

	resObj.CurrentTrick = trickCardsToOutput(o.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(o)
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築
func (p *OhHellWebPresenter) buildPlayersOutput(o interfaces.OhHellGame) []*controller.OhHellWebOutputPlayer {
	out := make([]*controller.OhHellWebOutputPlayer, 0)
	for i := 0; i < o.GetPlayerCnt(); i++ {
		player := o.GetPlayer(i)
		pObj := &controller.OhHellWebOutputPlayer{
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
func (p *OhHellWebPresenter) buildMessage(o interfaces.OhHellGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if o.GetGameEndFlag() {
		winnerIdx := o.GetWinnerIdx()
		player := o.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		return buildWinnerWebMessage("ohhell", winnerIdx, isHuman)
	}
	switch o.GetPhase() {
	case domain.OhHellPhaseBid:
		return "", "ohhell.bidPhase", nil
	case domain.OhHellPhasePlay:
		if len(o.GetCurrentTrick()) == 0 {
			return "", "ohhell.playPhase.lead", nil
		}
		return "", "ohhell.playPhase.follow", nil
	case domain.OhHellPhaseTrickEnd:
		return "", "ohhell.trickEnd", nil
	case domain.OhHellPhaseRoundEnd:
		return "", "ohhell.roundEnd", nil
	}
	return "", "", nil
}
