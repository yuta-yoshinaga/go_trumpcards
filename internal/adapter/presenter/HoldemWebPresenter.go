package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// Ensure interfaces.HoldemGame satisfies communityCardPresenterGame at compile time.
var _ communityCardPresenterGame = (interfaces.HoldemGame)(nil)

// HoldemWebPresenter テキサスホールデムWebプレゼンタークラス
type HoldemWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (hwp *HoldemWebPresenter) Output(h interfaces.HoldemGame, lastErr error) string {
	resObj := hwp.buildOutput(h, lastErr)
	return marshalOrError(resObj)
}

// buildOutput ゲーム状態をHoldemWebOutputに変換
func (hwp *HoldemWebPresenter) buildOutput(h interfaces.HoldemGame, lastErr error) *controller.HoldemWebOutput {
	resObj := buildCommunityCardBaseOutput(h)
	resObj.Players = buildPokerPlayersOutput(h.GetPhase(), h.GetPlayerCnt(), func(i int) communityCardPresenterPlayer { return h.GetPlayer(i) }, domain.HoldemPhaseShowdown, domain.HoldemPhaseEnd, pokerHandName)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = hwp.buildMessage(h, lastErr)
	return resObj
}

// buildMessage ゲーム結果メッセージを構築
func (hwp *HoldemWebPresenter) buildMessage(h interfaces.HoldemGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if h.IsMuckAvailable() {
		return "Muck or show your hand.", "holdem.muck.prompt", nil
	}
	if h.GetGameEndFlag() {
		msg, code := hwp.buildResultMessage(h)
		return msg, code, nil
	}
	return "", "", nil
}

// buildResultMessage builds the end-of-round message and its i18n code
func (hwp *HoldemWebPresenter) buildResultMessage(h interfaces.HoldemGame) (string, string) {
	results := h.GetRoundResults()
	if len(results) == 0 {
		return "Game over.", "holdem.result.gameOver"
	}

	for _, r := range results {
		if h.GetPlayer(r.PlayerIdx).GetIsHuman() {
			if r.WonAmount > 0 {
				return "You are the winner.", "holdem.result.win"
			}
		}
	}

	// Human not in results (folded)
	for i := 0; i < h.GetPlayerCnt(); i++ {
		if h.GetPlayer(i).GetIsHuman() && h.GetPlayer(i).GetFolded() {
			return "You folded.", "holdem.result.folded"
		}
	}

	// Human mucked
	for _, r := range results {
		if h.GetPlayer(r.PlayerIdx).GetIsHuman() && r.Mucked {
			return "You mucked.", "holdem.result.mucked"
		}
	}

	return "You lose.", "holdem.result.lose"
}

// ActionLogOutput 棋譜をJSON出力
func (hwp *HoldemWebPresenter) ActionLogOutput(h interfaces.HoldemGame) string {
	return actionLogOutputJSON(h)
}
