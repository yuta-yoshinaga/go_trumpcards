package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// ShortDeckWebPresenter ショートデックホールデムWebプレゼンタークラス
type ShortDeckWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (owp *ShortDeckWebPresenter) Output(o interfaces.ShortDeckGame, lastErr error) string {
	resObj := owp.buildOutput(o, lastErr)
	return marshalOrError(resObj)
}

// buildOutput ゲーム状態をShortDeckWebOutputに変換
func (owp *ShortDeckWebPresenter) buildOutput(o interfaces.ShortDeckGame, lastErr error) *controller.HoldemWebOutput {
	resObj := buildCommunityCardBaseOutput(o)
	resObj.Players = buildPokerPlayersOutput(o.GetPhase(), o.GetPlayerCnt(), func(i int) communityCardPresenterPlayer { return o.GetPlayer(i) }, domain.ShortDeckPhaseShowdown, domain.ShortDeckPhaseEnd, shortDeckHandName)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = owp.buildMessage(o, lastErr)
	return resObj
}

func (owp *ShortDeckWebPresenter) buildMessage(o interfaces.ShortDeckGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if o.IsMuckAvailable() {
		return "Muck or show your hand.", "shortdeck.muck.prompt", nil
	}
	if o.GetGameEndFlag() {
		msg, code := owp.buildResultMessage(o)
		return msg, code, nil
	}
	return "", "", nil
}

func (owp *ShortDeckWebPresenter) buildResultMessage(o interfaces.ShortDeckGame) (string, string) {
	results := o.GetRoundResults()
	if len(results) == 0 {
		return "Game over.", "shortdeck.result.gameOver"
	}

	for _, r := range results {
		if o.GetPlayer(r.PlayerIdx).GetIsHuman() {
			if r.WonAmount > 0 {
				return "You are the winner.", "shortdeck.result.win"
			}
		}
	}

	for i := 0; i < o.GetPlayerCnt(); i++ {
		if o.GetPlayer(i).GetIsHuman() && o.GetPlayer(i).GetFolded() {
			return "You folded.", "shortdeck.result.folded"
		}
	}

	for _, r := range results {
		if o.GetPlayer(r.PlayerIdx).GetIsHuman() && r.Mucked {
			return "You mucked.", "shortdeck.result.mucked"
		}
	}

	return "You lose.", "shortdeck.result.lose"
}

// ActionLogOutput 棋譜をJSON出力
func (owp *ShortDeckWebPresenter) ActionLogOutput(o interfaces.ShortDeckGame) string {
	return actionLogOutputJSON(o)
}
