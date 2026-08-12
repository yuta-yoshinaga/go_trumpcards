//go:build !js || !wasm || extra3

package presenter

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// RollingStoneWebPresenter ローリングストーンWebプレゼンタークラス
type RollingStoneWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *RollingStoneWebPresenter) Output(s interfaces.RollingStoneGame, lastErr error) string {
	resObj := p.buildBase(s)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(s, lastErr)
	// 受動ヒントは Output() でも埋める (#4483)。
	resObj.Hint = rollingStoneHintOutput(s)
	return marshalOrError(resObj)
}

// rollingStoneHintOutput はヒントを出力形に変換する。
func rollingStoneHintOutput(s interfaces.RollingStoneGame) *controller.RollingStoneWebOutputHint {
	hint := s.GetHint()
	if hint == nil {
		return nil
	}
	return &controller.RollingStoneWebOutputHint{CardIndex: hint.CardIndex, Reason: hint.Reason}
}

// buildBase 共通フィールドを構築
func (p *RollingStoneWebPresenter) buildBase(s interfaces.RollingStoneGame) *controller.RollingStoneWebOutput {
	resObj := new(controller.RollingStoneWebOutput)
	resObj.Phase = int(s.GetPhase())
	// **出せる札が無いことを別のフラグで出す。** 空の validPlays は「まだ手番でない」
	// とも読めてしまうので、引き取りが要る局面はそう名乗ります。
	resObj.MustPickUp = s.MustPickUp(0)
	resObj.ValidPlays = intSliceOrEmpty(s.GetValidPlayIndices(0))
	resObj.CurrentTrick = trickCardsToOutput(s.GetCurrentTrick())
	resObj.CurrentPlayerIdx = s.GetCurrentPlayerIdx()
	resObj.LeadPlayerIdx = s.GetLeadPlayerIdx()
	resObj.TrickNumber = s.GetTrickNumber()
	resObj.LastPickupIdx = s.GetLastPickupIdx()
	resObj.FinishedCnt = s.GetFinishedCnt()
	resObj.DeckSize = s.GetDeckSize()
	resObj.Discarded = s.GetDiscarded()
	resObj.GameEndFlag = s.GetGameEndFlag()
	resObj.WinnerIdx = s.GetWinnerIdx()
	resObj.Players = p.buildPlayersOutput(s)
	resObj.Config = controller.RollingStoneWebOutputConfig{PlayerCnt: s.GetConfig().PlayerCnt}
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築
func (p *RollingStoneWebPresenter) buildPlayersOutput(s interfaces.RollingStoneGame) []*controller.RollingStoneWebOutputPlayer {
	out := make([]*controller.RollingStoneWebOutputPlayer, 0)
	for i := 0; i < s.GetPlayerCnt(); i++ {
		player := s.GetPlayer(i)
		out = append(out, &controller.RollingStoneWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, player.GetIsHuman()),
			Pickups:    player.GetPickups(),
			FinishedAt: player.GetFinishedAt(),
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *RollingStoneWebPresenter) buildMessage(s interfaces.RollingStoneGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if s.GetGameEndFlag() {
		// **上限で切った局は「上がった」わけではない。** 言い分ける。
		if s.GetWinnerIdx() >= 0 && s.GetPlayer(s.GetWinnerIdx()).GetCardsSize() > 0 {
			return "", "rollingstone.result.stalemate", map[string]string{
				"idx": strconv.Itoa(s.GetWinnerIdx()),
			}
		}
		if s.GetWinnerIdx() == 0 {
			return "", "rollingstone.result.you", nil
		}
		return "", "rollingstone.result.cpu", map[string]string{"idx": strconv.Itoa(s.GetWinnerIdx())}
	}
	if s.MustPickUp(0) {
		return "", "rollingstone.pickup", map[string]string{
			"n": strconv.Itoa(len(s.GetCurrentTrick())),
		}
	}
	return "", "rollingstone.play", map[string]string{
		"n": strconv.Itoa(s.GetPlayer(0).GetCardsSize()),
	}
}

// HintOutput ヒント情報をJSON出力する
func (p *RollingStoneWebPresenter) HintOutput(s interfaces.RollingStoneGame) string {
	resObj := p.buildBase(s)
	resObj.Hint = rollingStoneHintOutput(s)
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *RollingStoneWebPresenter) ActionLogOutput(s interfaces.RollingStoneGame) string {
	return actionLogOutputJSON(s)
}
