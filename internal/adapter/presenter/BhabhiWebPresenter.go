//go:build !js || !wasm || extra4

package presenter

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// BhabhiWebPresenter バービーWebプレゼンタークラス
type BhabhiWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *BhabhiWebPresenter) Output(b interfaces.BhabhiGame, lastErr error) string {
	resObj := p.buildBase(b)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(b, lastErr)
	// 受動ヒントは Output() でも埋める (#4483)。
	if hint := b.GetHint(); hint != nil {
		resObj.Hint = &controller.BhabhiWebOutputHint{CardIndex: hint.CardIndex, Reason: hint.Reason}
	}
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *BhabhiWebPresenter) buildBase(b interfaces.BhabhiGame) *controller.BhabhiWebOutput {
	resObj := new(controller.BhabhiWebOutput)
	resObj.Phase = int(b.GetPhase())
	resObj.TrickNumber = b.GetTrickNumber()
	resObj.LeadSuit = b.GetLeadSuit()
	resObj.LastPickupIdx = b.GetLastPickupIdx()
	resObj.LastPickupSize = b.GetLastPickupSize()
	resObj.CurrentPlayerIdx = b.GetCurrentPlayerIdx()
	resObj.LeadPlayerIdx = b.GetLeadPlayerIdx()
	resObj.ValidPlays = intSliceOrEmpty(b.GetValidPlayIndices(0))
	resObj.AliveCount = b.GetAliveCount()
	resObj.GameEndFlag = b.GetGameEndFlag()
	resObj.BhabhiIdx = b.GetBhabhiIdx()
	resObj.Stalemate = b.IsStalemate()
	resObj.StalemateTricks = domain.BhabhiStalemateTricks
	resObj.Pile = trickCardsToOutput(b.GetPile())
	resObj.Players = p.buildPlayersOutput(b)
	resObj.Config = controller.BhabhiWebOutputConfig{PlayerCnt: b.GetConfig().PlayerCnt}
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築
func (p *BhabhiWebPresenter) buildPlayersOutput(b interfaces.BhabhiGame) []*controller.BhabhiWebOutputPlayer {
	out := make([]*controller.BhabhiWebOutputPlayer, 0)
	for i := 0; i < b.GetPlayerCnt(); i++ {
		player := b.GetPlayer(i)
		out = append(out, &controller.BhabhiWebOutputPlayer{
			ID:        i,
			IsHuman:   player.GetIsHuman(),
			CardCount: player.GetCardsSize(),
			Cards:     playerCardsToOutput(player, player.GetIsHuman()),
			Rank:      player.GetRank(),
			Pickups:   player.GetPickups(),
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *BhabhiWebPresenter) buildMessage(b interfaces.BhabhiGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if b.GetGameEndFlag() {
		params := map[string]string{"idx": strconv.Itoa(b.GetBhabhiIdx())}
		// **膠着で終わったことは盤面から読めない。** 別のコードで言う。
		if b.IsStalemate() {
			params["tricks"] = strconv.Itoa(b.GetTrickNumber())
			if b.GetBhabhiIdx() == 0 {
				return "", "bhabhi.result.stalemateYou", params
			}
			return "", "bhabhi.result.stalemateCpu", params
		}
		if b.GetBhabhiIdx() == 0 {
			return "", "bhabhi.result.you", params
		}
		return "", "bhabhi.result.cpu", params
	}
	// **フォローできないと場札を全部引き取る。** 何枚積まれているかが要点。
	if len(b.GetPile()) > 0 {
		return "", "bhabhi.play.pile", map[string]string{"n": strconv.Itoa(len(b.GetPile()))}
	}
	return "", "bhabhi.play.lead", nil
}

// HintOutput ヒント情報をJSON出力する
func (p *BhabhiWebPresenter) HintOutput(b interfaces.BhabhiGame) string {
	resObj := p.buildBase(b)
	if hint := b.GetHint(); hint != nil {
		resObj.Hint = &controller.BhabhiWebOutputHint{CardIndex: hint.CardIndex, Reason: hint.Reason}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *BhabhiWebPresenter) ActionLogOutput(b interfaces.BhabhiGame) string {
	return actionLogOutputJSON(b)
}
