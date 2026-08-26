//go:build !js || !wasm || extra2

package presenter

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// JulepeWebPresenter フレペWebプレゼンタークラス
type JulepeWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *JulepeWebPresenter) Output(r interfaces.JulepeGame, lastErr error) string {
	resObj := p.buildBase(r)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(r, lastErr)
	// 受動ヒントは Output() でも埋める (#4483)。
	if hint := r.GetHint(); hint != nil {
		resObj.Hint = &controller.JulepeWebOutputHint{CardIndex: hint.CardIndex, Reason: hint.Reason}
	}
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *JulepeWebPresenter) buildBase(r interfaces.JulepeGame) *controller.JulepeWebOutput {
	resObj := new(controller.JulepeWebOutput)
	resObj.Phase = int(r.GetPhase())
	resObj.RoundNumber = r.GetRoundNumber()
	resObj.TrickNumber = r.GetTrickNumber()
	resObj.Pot = r.GetPot()
	resObj.RequiredTricks = r.GetRequiredTricks()
	resObj.Beast = r.GetBeast()
	if resObj.Beast == nil {
		resObj.Beast = make([]bool, 0)
	}
	resObj.TrumpSuit = r.GetTrumpSuit()
	if up := r.GetUpCard(); up != nil {
		resObj.UpCard = cardToOutput(up)
	}
	resObj.CurrentPlayerIdx = r.GetCurrentPlayerIdx()
	resObj.LeadPlayerIdx = r.GetLeadPlayerIdx()
	resObj.DealerIdx = r.GetDealerIdx()
	resObj.ActiveCount = r.GetActiveCount()
	resObj.ValidPlays = intSliceOrEmpty(r.GetValidPlayIndices(0))
	resObj.GameEndFlag = r.GetGameEndFlag()
	resObj.WinnerIdx = r.GetWinnerIdx()
	resObj.CurrentTrick = trickCardsToOutput(r.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(r)
	resObj.Config = controller.JulepeWebOutputConfig{
		PlayerCnt: r.GetConfig().PlayerCnt,
		Rounds:    r.GetConfig().Rounds,
	}
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築
func (p *JulepeWebPresenter) buildPlayersOutput(r interfaces.JulepeGame) []*controller.JulepeWebOutputPlayer {
	out := make([]*controller.JulepeWebOutputPlayer, 0)
	for i := 0; i < r.GetPlayerCnt(); i++ {
		player := r.GetPlayer(i)
		out = append(out, &controller.JulepeWebOutputPlayer{
			ID:          i,
			IsHuman:     player.GetIsHuman(),
			CardCount:   player.GetCardsSize(),
			Cards:       playerCardsToOutput(player, player.GetIsHuman()),
			Chips:       player.GetChips(),
			InRound:     player.GetInRound(),
			Decided:     player.GetDecided(),
			RoundTricks: player.GetRoundTricks(),
			TrickCount:  player.GetTrickCount(),
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *JulepeWebPresenter) buildMessage(r interfaces.JulepeGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if r.GetGameEndFlag() {
		if r.GetWinnerIdx() < 0 {
			return "", "julepe.result.tie", nil
		}
		return "", "julepe.result.winner", map[string]string{
			"idx":   strconv.Itoa(r.GetWinnerIdx()),
			"chips": strconv.Itoa(r.GetPlayer(r.GetWinnerIdx()).GetChips()),
		}
	}
	switch r.GetPhase() {
	case domain.JulepePhaseDecide:
		return "", "julepe.decidePhase", map[string]string{"pot": strconv.Itoa(r.GetPot())}
	case domain.JulepePhaseRoundEnd:
		return "", "julepe.roundEnd", map[string]string{
			"round": strconv.Itoa(r.GetRoundNumber()),
			"pot":   strconv.Itoa(r.GetPot()),
		}
	}
	// **降りたラウンドは見ているだけ。** 案内を変えないと操作待ちに見える。
	if !r.GetPlayer(0).GetInRound() {
		return "", "julepe.watching", nil
	}
	return "", "julepe.play", map[string]string{"pot": strconv.Itoa(r.GetPot())}
}

// HintOutput ヒント情報をJSON出力する
func (p *JulepeWebPresenter) HintOutput(r interfaces.JulepeGame) string {
	resObj := p.buildBase(r)
	if hint := r.GetHint(); hint != nil {
		resObj.Hint = &controller.JulepeWebOutputHint{CardIndex: hint.CardIndex, Reason: hint.Reason}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *JulepeWebPresenter) ActionLogOutput(r interfaces.JulepeGame) string {
	return actionLogOutputJSON(r)
}
