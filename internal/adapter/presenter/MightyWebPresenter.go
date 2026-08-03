//go:build !js || !wasm || extra2

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MightyWebPresenter マイティWebプレゼンタークラス
type MightyWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *MightyWebPresenter) Output(m interfaces.MightyGame, lastErr error) string {
	resObj := p.buildBaseOutput(m)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(m, lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **フェーズと手番はここでは見ない。**Mighty.GetHint() が自分で
	// 「人間の手番で、かつ行動を選べる状態か」を確かめて nil を返す。
	if hint := m.GetHint(); hint != nil {
		resObj.Hint = &controller.MightyWebOutputHint{
			CardIndex:      hint.CardIndex,
			Bid:            hint.Bid,
			BidNoTrump:     hint.BidNoTrump,
			TrumpSuit:      hint.TrumpSuit,
			PartnerSuit:    hint.PartnerSuit,
			PartnerValue:   hint.PartnerValue,
			DiscardIndices: hint.DiscardIndices,
			JokerLeadSuit:  hint.JokerLeadSuit,
			Reason:         hint.Reason,
		}
	}

	return marshalOrError(resObj)
}

// buildBaseOutput 基本出力を構築
func (p *MightyWebPresenter) buildBaseOutput(m interfaces.MightyGame) *controller.MightyWebOutput {
	resObj := new(controller.MightyWebOutput)
	resObj.Phase = int(m.GetPhase())
	resObj.RoundNumber = m.GetRoundNumber()
	resObj.TrickNumber = m.GetTrickNumber()
	resObj.CurrentPlayerIdx = m.GetCurrentPlayerIdx()
	resObj.BidPlayerIdx = m.GetBidPlayerIdx()
	resObj.TrumpSuit = m.GetTrumpSuit()
	resObj.DeclarerIdx = m.GetDeclarerIdx()
	resObj.PartnerIdx = m.GetPartnerIdx()
	resObj.PartnerRevealed = m.GetPartnerRevealed()
	resObj.HighestBid = m.GetHighestBid()
	resObj.HighestBidder = m.GetHighestBidder()
	resObj.WinningBidNoTrump = m.GetWinningBidNoTrump()
	resObj.GameEndFlag = m.GetGameEndFlag()
	resObj.WinnerTeam = m.GetWinnerTeam()
	resObj.LeadPlayerIdx = m.GetLeadPlayerIdx()

	if partnerCard := m.GetPartnerCard(); partnerCard != nil {
		resObj.PartnerCard = cardToOutput(partnerCard)
	}

	if m.GetPhase() == domain.MightyPhaseKittyExchange {
		kitty := m.GetKitty()
		if len(kitty) > 0 {
			resObj.Kitty = make([]*controller.WebOutputCard, len(kitty))
			for i, c := range kitty {
				resObj.Kitty[i] = cardToOutput(c)
			}
		}
	}

	cfg := m.GetConfig()
	resObj.Config = controller.MightyWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		MinBid:        cfg.MinBid,
		NoTrumpExtra:  cfg.NoTrumpExtra,
		PointLimit:    cfg.PointLimit,
	}

	trick := m.GetCurrentTrick()
	resObj.CurrentTrick = p.buildTrickOutput(trick)
	resObj.Players = p.buildPlayersOutput(m)

	return resObj
}

// buildTrickOutput 現在のトリック情報を構築。
//
// 共有の trickCardsToOutput に寄せられない: Mighty は domain / controller とも
// IsJokerLead と LeadDemandSuit を追加で持つ正当な別形状（#4363）。
func (p *MightyWebPresenter) buildTrickOutput(trick []*domain.MightyTrickCard) []*controller.MightyWebOutputTrickCard {
	return buildTrickCards(trick, func(tc *domain.MightyTrickCard) *controller.MightyWebOutputTrickCard {
		out := &controller.MightyWebOutputTrickCard{PlayerIdx: tc.PlayerIdx, Card: cardToOutput(tc.Card)}
		if tc.IsJokerLead {
			out.IsJokerLead = true
			out.LeadDemandSuit = tc.LeadDemandSuit
		}
		return out
	})
}

// buildPlayersOutput プレイヤー情報を構築
func (p *MightyWebPresenter) buildPlayersOutput(m interfaces.MightyGame) []*controller.MightyWebOutputPlayer {
	out := make([]*controller.MightyWebOutputPlayer, 0)
	for i := 0; i < m.GetPlayerCnt(); i++ {
		player := m.GetPlayer(i)
		pObj := &controller.MightyWebOutputPlayer{
			ID:              i,
			IsHuman:         player.GetIsHuman(),
			CardCount:       player.GetCardsSize(),
			Cards:           playerCardsToOutput(player, player.GetIsHuman()),
			Bid:             player.GetBid(),
			BidNoTrump:      player.GetBidNoTrump(),
			IsDeclarer:      player.GetIsDeclarer(),
			IsPartner:       m.GetPartnerRevealed() && player.GetIsPartner(),
			PartnerRevealed: player.GetPartnerRevealed(),
			PointCards:      player.GetPointCards(),
			RoundScore:      player.GetRoundScore(),
			CumulativeScore: player.GetCumulativeScore(),
			TrickCount:      player.GetTrickCount(),
		}
		out = append(out, pObj)
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *MightyWebPresenter) buildMessage(m interfaces.MightyGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if m.GetGameEndFlag() {
		if m.GetWinnerTeam() == domain.MightyWinnerDeclarer {
			return "", "mighty.gameEnd.declarerWins", nil
		}
		return "", "mighty.gameEnd.oppositionWins", nil
	}
	switch m.GetPhase() {
	case domain.MightyPhaseBid:
		return "", "mighty.bidPhase", nil
	case domain.MightyPhaseTrumpAndFriend:
		return "", "mighty.trumpAndFriendPhase", nil
	case domain.MightyPhaseKittyExchange:
		return "", "mighty.kittyExchange", nil
	case domain.MightyPhasePlay:
		trick := m.GetCurrentTrick()
		if len(trick) == 0 {
			return "", "mighty.playPhase.lead", nil
		}
		return "", "mighty.playPhase.follow", nil
	case domain.MightyPhaseTrickEnd:
		return "", "mighty.trickEnd", nil
	case domain.MightyPhaseRoundEnd:
		return "", "mighty.roundEnd", nil
	}
	return "", "", nil
}

// HintOutput ヒント情報をJSON出力する
func (p *MightyWebPresenter) HintOutput(m interfaces.MightyGame) string {
	hint := m.GetHint()
	resObj := p.buildBaseOutput(m)

	if hint != nil {
		resObj.Hint = &controller.MightyWebOutputHint{
			CardIndex:      hint.CardIndex,
			Bid:            hint.Bid,
			BidNoTrump:     hint.BidNoTrump,
			TrumpSuit:      hint.TrumpSuit,
			PartnerSuit:    hint.PartnerSuit,
			PartnerValue:   hint.PartnerValue,
			DiscardIndices: hint.DiscardIndices,
			JokerLeadSuit:  hint.JokerLeadSuit,
			Reason:         hint.Reason,
		}
	}
	// **「頼んだヒントか」を CLI が見分けられるようにする。**このゲーム群の
	// `hintAvailable` は画面のラベルとして既に使われているので、別キーを出す (#4483)。
	if hint != nil {
		resObj.MessageCode = "mighty.hintRequested"
	} else {
		resObj.MessageCode = "mighty.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *MightyWebPresenter) ActionLogOutput(m interfaces.MightyGame) string {
	return actionLogOutputJSON(m)
}
