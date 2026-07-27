package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// BriscolaWebPresenter ブリスコラWebプレゼンタークラス
type BriscolaWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *BriscolaWebPresenter) Output(b interfaces.BriscolaGame, lastErr error) string {
	resObj := p.buildBase(b)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(b, lastErr)
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *BriscolaWebPresenter) buildBase(b interfaces.BriscolaGame) *controller.BriscolaWebOutput {
	resObj := new(controller.BriscolaWebOutput)
	resObj.Phase = int(b.GetPhase())
	resObj.TrickNumber = b.GetTrickNumber()
	resObj.CurrentPlayerIdx = b.GetCurrentPlayerIdx()
	resObj.TrumpSuit = b.GetTrumpSuit()
	if tc := b.GetTrumpCard(); tc != nil {
		resObj.TrumpCard = cardToOutput(tc)
	}
	resObj.DealerIdx = b.GetDealerIdx()
	resObj.LeadPlayerIdx = b.GetLeadPlayerIdx()
	resObj.StockRemaining = b.GetStockRemaining()
	resObj.GameEndFlag = b.GetGameEndFlag()
	resObj.WinnerIdx = b.GetWinnerIdx()

	cfg := b.GetConfig()
	resObj.Config = controller.BriscolaWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
	}

	resObj.CurrentTrick = p.buildTrickOutput(b.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(b)
	return resObj
}

// buildTrickOutput 現在のトリック情報を構築
func (p *BriscolaWebPresenter) buildTrickOutput(trick []*domain.TrickCard) []*controller.BriscolaWebOutputTrickCard {
	return buildTrickCards(trick, func(tc *domain.TrickCard) *controller.BriscolaWebOutputTrickCard {
		return &controller.BriscolaWebOutputTrickCard{PlayerIdx: tc.PlayerIdx, Card: cardToOutput(tc.Card)}
	})
}

// buildPlayersOutput プレイヤー情報を構築
func (p *BriscolaWebPresenter) buildPlayersOutput(b interfaces.BriscolaGame) []*controller.BriscolaWebOutputPlayer {
	out := make([]*controller.BriscolaWebOutputPlayer, 0)
	for i := 0; i < b.GetPlayerCnt(); i++ {
		player := b.GetPlayer(i)
		out = append(out, &controller.BriscolaWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, player.GetIsHuman()),
			Points:     b.GetPlayerPoints(i),
			TrickCount: player.GetTrickCount(),
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *BriscolaWebPresenter) buildMessage(b interfaces.BriscolaGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if b.GetGameEndFlag() {
		p0 := b.GetPlayerPoints(0)
		p1 := b.GetPlayerPoints(1)
		params := map[string]string{
			"p0": fmt.Sprintf("%d", p0),
			"p1": fmt.Sprintf("%d", p1),
		}
		switch b.GetWinnerIdx() {
		case 0:
			return "", "briscola.result.p0Win", params
		case 1:
			return "", "briscola.result.p1Win", params
		default:
			return "", "briscola.result.tie", params
		}
	}
	switch b.GetPhase() {
	case domain.BriscolaPhasePlay:
		if len(b.GetCurrentTrick()) == 0 {
			return "", "briscola.playPhase.lead", nil
		}
		return "", "briscola.playPhase.follow", nil
	case domain.BriscolaPhaseTrickEnd:
		return "", "briscola.trickEnd", nil
	}
	return "", "", nil
}

// HintOutput ヒント情報をJSON出力する
func (p *BriscolaWebPresenter) HintOutput(b interfaces.BriscolaGame) string {
	hint := b.GetHint()
	resObj := p.buildBase(b)
	if hint != nil {
		resObj.Hint = &controller.BriscolaWebOutputHint{
			CardIndex: hint.CardIndex,
			Reason:    hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *BriscolaWebPresenter) ActionLogOutput(b interfaces.BriscolaGame) string {
	return actionLogOutputJSON(b)
}
