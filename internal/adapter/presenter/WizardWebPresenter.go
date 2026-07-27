package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// wizardFace は非52枚デッキのウィザード/ジェスター札を手続き的に描画するための
// 自己記述子を返す。標準52枚デッキの札には nil を返し (従来どおりPNGパスで描画)、
// ウィザード/ジェスター札にのみ Glyph/Label/Color/Deck を付与する。ADR-0033 参照。
func wizardFace(card *domain.Card) *CardFace {
	if card == nil {
		return nil
	}
	switch card.GetDesign() {
	case domain.WizardDesignWizard:
		return &CardFace{Glyph: "✦", Label: "Wizard", Color: "purple", Deck: "wizard"}
	case domain.WizardDesignJester:
		return &CardFace{Glyph: "☺", Label: "Jester", Color: "green", Deck: "wizard"}
	}
	return nil // 標準52枚デッキ ⇒ nil ⇒ PNGパス (design/value のみ)
}

// WizardWebPresenter ウィザードWebプレゼンタークラス
type WizardWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *WizardWebPresenter) Output(o interfaces.WizardGame, lastErr error) string {
	resObj := p.buildBase(o)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(o, lastErr)
	return marshalOrError(resObj)
}

// HintOutput ヒント情報をJSON出力する
func (p *WizardWebPresenter) HintOutput(o interfaces.WizardGame) string {
	hint := o.GetHint()
	resObj := p.buildBase(o)

	if hint != nil {
		resObj.Hint = &controller.WizardWebOutputHint{
			CardIndex: hint.CardIndex,
			Bid:       hint.Bid,
			Reason:    hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *WizardWebPresenter) ActionLogOutput(o interfaces.WizardGame) string {
	return actionLogOutputJSON(o)
}

// buildBase 共通フィールドを構築
func (p *WizardWebPresenter) buildBase(o interfaces.WizardGame) *controller.WizardWebOutput {
	resObj := new(controller.WizardWebOutput)
	resObj.Phase = int(o.GetPhase())
	resObj.RoundNumber = o.GetRoundNumber()
	resObj.TotalRounds = o.GetTotalRounds()
	resObj.HandSize = o.GetHandSize()
	resObj.TrickNumber = o.GetTrickNumber()
	resObj.CurrentPlayerIdx = o.GetCurrentPlayerIdx()
	resObj.BidPlayerIdx = o.GetBidPlayerIdx()
	resObj.DealerIdx = o.GetDealerIdx()
	resObj.TrumpCard = cardToOutputWithFace(o.GetTrumpCard(), wizardFace)
	resObj.TrumpSuit = o.GetTrumpSuit()
	resObj.RestrictedBid = o.GetRestrictedBid()
	resObj.GameEndFlag = o.GetGameEndFlag()
	resObj.WinnerIdx = o.GetWinnerIdx()
	resObj.LeadPlayerIdx = o.GetLeadPlayerIdx()

	cfg := o.GetConfig()
	resObj.Config = controller.WizardWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
	}

	resObj.CurrentTrick = p.buildTrickOutput(o.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(o)
	return resObj
}

// buildTrickOutput 現在のトリック情報を構築
func (p *WizardWebPresenter) buildTrickOutput(trick []*domain.TrickCard) []*controller.WizardWebOutputTrickCard {
	return buildTrickCards(trick, func(tc *domain.TrickCard) *controller.WizardWebOutputTrickCard {
		return &controller.WizardWebOutputTrickCard{PlayerIdx: tc.PlayerIdx, Card: cardToOutputWithFace(tc.Card, wizardFace)}
	})
}

// buildPlayersOutput プレイヤー情報を構築
func (p *WizardWebPresenter) buildPlayersOutput(o interfaces.WizardGame) []*controller.WizardWebOutputPlayer {
	out := make([]*controller.WizardWebOutputPlayer, 0)
	for i := 0; i < o.GetPlayerCnt(); i++ {
		player := o.GetPlayer(i)
		pObj := &controller.WizardWebOutputPlayer{
			ID:              i,
			IsHuman:         player.GetIsHuman(),
			CardCount:       player.GetCardsSize(),
			Cards:           playerCardsToOutputWithFace(player, player.GetIsHuman(), wizardFace),
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
func (p *WizardWebPresenter) buildMessage(o interfaces.WizardGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if o.GetGameEndFlag() {
		winnerIdx := o.GetWinnerIdx()
		player := o.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		return buildWinnerWebMessage("wizard", winnerIdx, isHuman)
	}
	switch o.GetPhase() {
	case domain.WizardPhaseBid:
		return "", "wizard.bidPhase", nil
	case domain.WizardPhasePlay:
		if len(o.GetCurrentTrick()) == 0 {
			return "", "wizard.playPhase.lead", nil
		}
		return "", "wizard.playPhase.follow", nil
	case domain.WizardPhaseTrickEnd:
		return "", "wizard.trickEnd", nil
	case domain.WizardPhaseRoundEnd:
		return "", "wizard.roundEnd", nil
	}
	return "", "", nil
}
