//go:build !js || !wasm || extra4

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// NapoleonWebPresenter ナポレオンWebプレゼンタークラス
type NapoleonWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *NapoleonWebPresenter) Output(n interfaces.NapoleonGame, lastErr error) string {
	resObj := p.buildBaseOutput(n)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(n, lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **フェーズと手番はここでは見ない。**Napoleon.GetHint() が自分で
	// 「人間の手番で、かつ行動を選べる状態か」を確かめて nil を返す。
	if hint := n.GetHint(); hint != nil {
		resObj.Hint = &controller.NapoleonWebOutputHint{
			CardIndex:     hint.CardIndex,
			Bid:           hint.Bid,
			TrumpSuit:     hint.TrumpSuit,
			AdjutantSuit:  hint.AdjutantSuit,
			AdjutantValue: hint.AdjutantValue,
			DiscardIndex:  hint.DiscardIndex,
			Reason:        hint.Reason,
		}
	}

	return marshalOrError(resObj)
}

// buildBaseOutput 基本出力を構築
func (p *NapoleonWebPresenter) buildBaseOutput(n interfaces.NapoleonGame) *controller.NapoleonWebOutput {
	resObj := new(controller.NapoleonWebOutput)
	resObj.Phase = int(n.GetPhase())
	resObj.RoundNumber = n.GetRoundNumber()
	resObj.TrickNumber = n.GetTrickNumber()
	resObj.CurrentPlayerIdx = n.GetCurrentPlayerIdx()
	resObj.BidPlayerIdx = n.GetBidPlayerIdx()
	resObj.TrumpSuit = n.GetTrumpSuit()
	resObj.NapoleonIdx = n.GetNapoleonIdx()
	resObj.AdjutantIdx = n.GetAdjutantIdx()
	resObj.AdjutantRevealed = n.GetAdjutantRevealed()
	resObj.HighestBid = n.GetHighestBid()
	resObj.HighestBidder = n.GetHighestBidder()
	resObj.GameEndFlag = n.GetGameEndFlag()
	resObj.WinnerTeam = n.GetWinnerTeam()
	resObj.LeadPlayerIdx = n.GetLeadPlayerIdx()

	// 副官カード
	adjCard := n.GetAdjutantCard()
	if adjCard != nil {
		resObj.AdjutantCard = cardToOutput(adjCard)
	}

	// 場札 (ナポレオンの場札交換フェーズでのみ表示)
	if n.GetPhase() == domain.NapoleonPhaseKittyExchange {
		kitty := n.GetKitty()
		if len(kitty) > 0 {
			resObj.Kitty = make([]*controller.WebOutputCard, len(kitty))
			for i, c := range kitty {
				resObj.Kitty[i] = cardToOutput(c)
			}
		}
	}

	// 設定
	cfg := n.GetConfig()
	resObj.Config = controller.NapoleonWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		MinBid:        cfg.MinBid,
		PointLimit:    cfg.PointLimit,
	}

	trick := n.GetCurrentTrick()
	resObj.CurrentTrick = p.buildTrickOutput(trick)
	resObj.Players = p.buildPlayersOutput(n)

	return resObj
}

// buildTrickOutput 現在のトリック情報を構築。
//
// 共有の trickCardsToOutput に寄せられない: Napoleon は KV スナップショットの
// タグ差異（pi/cd）のため domain 側が独自型のままで（#4363 / PR #4431）、この
// ラッパーがその型と共有の WebOutputTrickCard を橋渡ししている。出力型は共有型
// なので REST 形状は他の60ゲームと同一。
func (p *NapoleonWebPresenter) buildTrickOutput(trick []*domain.NapoleonTrickCard) []*controller.WebOutputTrickCard {
	return buildTrickCards(trick, func(tc *domain.NapoleonTrickCard) *controller.WebOutputTrickCard {
		return &controller.WebOutputTrickCard{PlayerIdx: tc.PlayerIdx, Card: cardToOutput(tc.Card)}
	})
}

// buildPlayersOutput プレイヤー情報を構築
func (p *NapoleonWebPresenter) buildPlayersOutput(n interfaces.NapoleonGame) []*controller.NapoleonWebOutputPlayer {
	out := make([]*controller.NapoleonWebOutputPlayer, 0)
	for i := 0; i < n.GetPlayerCnt(); i++ {
		player := n.GetPlayer(i)
		pObj := &controller.NapoleonWebOutputPlayer{
			ID:               i,
			IsHuman:          player.GetIsHuman(),
			CardCount:        player.GetCardsSize(),
			Cards:            playerCardsToOutput(player, player.GetIsHuman()),
			Bid:              player.GetBid(),
			IsNapoleon:       player.GetIsNapoleon(),
			IsAdjutant:       n.GetAdjutantRevealed() && player.GetIsAdjutant(),
			AdjutantRevealed: player.GetAdjutantRevealed(),
			PictureCards:     player.GetPictureCards(),
			RoundScore:       player.GetRoundScore(),
			CumulativeScore:  player.GetCumulativeScore(),
			TrickCount:       player.GetTrickCount(),
		}
		out = append(out, pObj)
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *NapoleonWebPresenter) buildMessage(n interfaces.NapoleonGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if n.GetGameEndFlag() {
		winnerTeam := n.GetWinnerTeam()
		if winnerTeam == domain.NapoleonWinnerNapoleon {
			return "", "napoleon.gameEnd.napoleonWins", nil
		}
		return "", "napoleon.gameEnd.alliedWins", nil
	}
	switch n.GetPhase() {
	case domain.NapoleonPhaseBid:
		return "", "napoleon.bidPhase", nil
	case domain.NapoleonPhaseTrumpDeclaration:
		return "", "napoleon.trumpDeclaration", nil
	case domain.NapoleonPhaseKittyExchange:
		return "", "napoleon.kittyExchange", nil
	case domain.NapoleonPhasePlay:
		trick := n.GetCurrentTrick()
		if len(trick) == 0 {
			return "", "napoleon.playPhase.lead", nil
		}
		return "", "napoleon.playPhase.follow", nil
	case domain.NapoleonPhaseTrickEnd:
		return "", "napoleon.trickEnd", nil
	case domain.NapoleonPhaseRoundEnd:
		return "", "napoleon.roundEnd", nil
	}
	return "", "", nil
}

// HintOutput ヒント情報をJSON出力する
func (p *NapoleonWebPresenter) HintOutput(n interfaces.NapoleonGame) string {
	hint := n.GetHint()
	resObj := p.buildBaseOutput(n)

	if hint != nil {
		resObj.Hint = &controller.NapoleonWebOutputHint{
			CardIndex:     hint.CardIndex,
			Bid:           hint.Bid,
			TrumpSuit:     hint.TrumpSuit,
			AdjutantSuit:  hint.AdjutantSuit,
			AdjutantValue: hint.AdjutantValue,
			DiscardIndex:  hint.DiscardIndex,
			Reason:        hint.Reason,
		}
	}
	// **「頼んだヒントか」を CLI が見分けられるようにする。**このゲーム群の
	// `hintAvailable` は画面のラベルとして既に使われているので、別キーを出す (#4483)。
	if hint != nil {
		resObj.MessageCode = "napoleon.hintRequested"
	} else {
		resObj.MessageCode = "napoleon.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *NapoleonWebPresenter) ActionLogOutput(n interfaces.NapoleonGame) string {
	return actionLogOutputJSON(n)
}
