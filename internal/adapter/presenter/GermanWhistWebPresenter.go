//go:build !js || !wasm || classic

package presenter

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// GermanWhistWebPresenter ジャーマンホイストWebプレゼンタークラス
type GermanWhistWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *GermanWhistWebPresenter) Output(g interfaces.GermanWhistGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は command:"hint" 専用
	// のレスポンスで、ページの state にはマージされない (#4483)。
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.GermanWhistWebOutputHint{
			CardIndex: hint.CardIndex,
			Reason:    hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *GermanWhistWebPresenter) buildBase(g interfaces.GermanWhistGame) *controller.GermanWhistWebOutput {
	resObj := new(controller.GermanWhistWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.LeadPlayerIdx = g.GetLeadPlayerIdx()
	resObj.TrumpSuit = g.GetTrumpSuit()
	if up := g.GetUpCard(); up != nil {
		resObj.UpCard = cardToOutput(up)
	}
	resObj.StockCount = g.GetStockCount()
	resObj.ValidPlays = intSliceOrEmpty(g.GetValidPlayIndices(0))
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()
	resObj.CurrentTrick = trickCardsToOutput(g.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築
func (p *GermanWhistWebPresenter) buildPlayersOutput(g interfaces.GermanWhistGame) []*controller.GermanWhistWebOutputPlayer {
	out := make([]*controller.GermanWhistWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		out = append(out, &controller.GermanWhistWebOutputPlayer{
			ID:            i,
			IsHuman:       player.GetIsHuman(),
			CardCount:     player.GetCardsSize(),
			Cards:         playerCardsToOutput(player, player.GetIsHuman()),
			TrickCount:    player.GetTrickCount(),
			ScoringTricks: player.GetScoringTricks(),
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *GermanWhistWebPresenter) buildMessage(g interfaces.GermanWhistGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		params := map[string]string{
			"p0": strconv.Itoa(g.GetPlayer(0).GetScoringTricks()),
			"p1": strconv.Itoa(g.GetPlayer(1).GetScoringTricks()),
		}
		switch g.GetWinnerIdx() {
		case 0:
			return "", "germanwhist.result.p0Win", params
		case 1:
			return "", "germanwhist.result.p1Win", params
		default:
			return "", "germanwhist.result.tie", params
		}
	}
	// **前半と後半で案内が変わる。**前半は表向きの札の取り合いで、勝っても
	// 点にならない。後半は取ったトリックがそのまま得点になる。
	if g.GetPhase() == domain.GermanWhistPhaseDraw {
		return "", "germanwhist.phase.draw", nil
	}
	return "", "germanwhist.phase.scoring", nil
}

// HintOutput ヒント情報をJSON出力する
func (p *GermanWhistWebPresenter) HintOutput(g interfaces.GermanWhistGame) string {
	resObj := p.buildBase(g)
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.GermanWhistWebOutputHint{
			CardIndex: hint.CardIndex,
			Reason:    hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *GermanWhistWebPresenter) ActionLogOutput(g interfaces.GermanWhistGame) string {
	return actionLogOutputJSON(g)
}
