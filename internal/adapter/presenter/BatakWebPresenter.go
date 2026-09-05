package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// BatakWebPresenter Batak Web プレゼンタークラス
type BatakWebPresenter struct{}

// Output ゲーム状態を JSON 出力
func (p *BatakWebPresenter) Output(cb interfaces.BatakGame, lastErr error) string {
	resObj := p.buildBase(cb)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(cb, cb.GetCurrentTrick(), lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **フェーズと手番はここでは見ない。**Batak.GetHint() が自分で
	// 「人間の手番で、かつ行動を選べる状態か」を確かめて nil を返す。
	if hint := cb.GetHint(); hint != nil {
		resObj.Hint = &controller.BatakWebOutputHint{
			CardIndex: hint.CardIndex,
			Bid:       hint.Bid,
			Reason:    hint.Reason,
		}
	}

	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *BatakWebPresenter) buildBase(cb interfaces.BatakGame) *controller.BatakWebOutput {
	resObj := new(controller.BatakWebOutput)
	resObj.Phase = int(cb.GetPhase())
	resObj.RoundNumber = cb.GetRoundNumber()
	resObj.TrickNumber = cb.GetTrickNumber()
	resObj.CurrentPlayerIdx = cb.GetCurrentPlayerIdx()
	resObj.BidPlayerIdx = cb.GetBidPlayerIdx()
	resObj.SpadesBroken = cb.GetSpadesBroken()
	resObj.GameEndFlag = cb.GetGameEndFlag()
	resObj.WinnerIdx = cb.GetWinnerIdx()
	resObj.LeadPlayerIdx = cb.GetLeadPlayerIdx()
	resObj.DeclarerIdx = cb.GetDeclarerIdx()
	resObj.HighBid = cb.GetHighBid()
	if cb.IsHumanBidTurn() {
		resObj.MinLegalBid = cb.MinLegalBid()
	} else {
		resObj.MinLegalBid = 0
	}

	cfg := cb.GetConfig()
	resObj.Config = controller.BatakWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		MaxRounds:     cfg.MaxRounds,
	}

	resObj.CurrentTrick = trickCardsToOutput(cb.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(cb)

	// Provide the human player's valid play indices so the frontend can grey out
	// cards that the Batak rules (lead-suit / must-trump-spade) forbid.
	for i := 0; i < cb.GetPlayerCnt(); i++ {
		if pl := cb.GetPlayer(i); pl != nil && pl.GetIsHuman() {
			resObj.ValidPlayIndices = cb.GetValidPlayIndices(i)
			break
		}
	}
	if resObj.ValidPlayIndices == nil {
		resObj.ValidPlayIndices = make([]int, 0)
	}
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築
func (p *BatakWebPresenter) buildPlayersOutput(cb interfaces.BatakGame) []*controller.BatakWebOutputPlayer {
	out := make([]*controller.BatakWebOutputPlayer, 0)
	for i := 0; i < cb.GetPlayerCnt(); i++ {
		player := cb.GetPlayer(i)
		pObj := &controller.BatakWebOutputPlayer{
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
func (p *BatakWebPresenter) buildMessage(cb interfaces.BatakGame, trick []*domain.TrickCard, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if cb.GetGameEndFlag() {
		winnerIdx := cb.GetWinnerIdx()
		player := cb.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		return buildWinnerWebMessage("batak", winnerIdx, isHuman)
	}
	switch cb.GetPhase() {
	case domain.BatakPhaseBid:
		return "", "batak.bidPhase", nil
	case domain.BatakPhasePlay:
		if len(trick) == 0 {
			return "", "batak.playPhase.lead", nil
		}
		return "", "batak.playPhase.follow", nil
	case domain.BatakPhaseTrickEnd:
		return "", "batak.trickEnd", nil
	case domain.BatakPhaseRoundEnd:
		return "", "batak.roundEnd", nil
	}
	return "", "", nil
}

// HintOutput ヒント情報を JSON 出力する
func (p *BatakWebPresenter) HintOutput(cb interfaces.BatakGame) string {
	hint := cb.GetHint()
	resObj := p.buildBase(cb)
	if hint != nil {
		resObj.Hint = &controller.BatakWebOutputHint{
			CardIndex: hint.CardIndex,
			Bid:       hint.Bid,
			Reason:    hint.Reason,
		}
	}
	// **「頼んだヒントか」を CLI が見分けられるようにする。**このゲーム群の
	// `hintAvailable` は画面のラベルとして既に使われているので、別キーを出す (#4483)。
	if hint != nil {
		resObj.MessageCode = "batak.hintRequested"
	} else {
		resObj.MessageCode = "batak.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜を JSON 出力
func (p *BatakWebPresenter) ActionLogOutput(cb interfaces.BatakGame) string {
	return actionLogOutputJSON(cb)
}
