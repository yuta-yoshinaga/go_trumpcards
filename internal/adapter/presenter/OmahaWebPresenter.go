package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// OmahaWebPresenter オマハホールデムWebプレゼンタークラス
type OmahaWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (owp *OmahaWebPresenter) Output(o interfaces.OmahaGame, lastErr error) string {
	resObj := owp.buildOutput(o, lastErr)
	return marshalOrError(resObj)
}

// buildOutput ゲーム状態をOmahaWebOutputに変換
func (owp *OmahaWebPresenter) buildOutput(o interfaces.OmahaGame, lastErr error) *controller.HoldemWebOutput {
	resObj := buildCommunityCardBaseOutput(o)
	resObj.Players = buildPokerPlayersOutput(o.GetPhase(), o.GetPlayerCnt(), func(i int) communityCardPresenterPlayer { return o.GetPlayer(i) }, domain.OmahaPhaseShowdown, domain.OmahaPhaseEnd, pokerHandName)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = owp.buildMessage(o, lastErr)
	return resObj
}

func (owp *OmahaWebPresenter) buildMessage(o interfaces.OmahaGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if o.IsMuckAvailable() {
		return "Muck or show your hand.", "omaha.muck.prompt", nil
	}
	if o.GetGameEndFlag() {
		msg, code := owp.buildResultMessage(o)
		return msg, code, nil
	}
	return "", "", nil
}

func (owp *OmahaWebPresenter) buildResultMessage(o interfaces.OmahaGame) (string, string) {
	results := o.GetRoundResults()
	if len(results) == 0 {
		return "Game over.", "omaha.result.gameOver"
	}

	for _, r := range results {
		if o.GetPlayer(r.PlayerIdx).GetIsHuman() {
			if r.WonAmount > 0 {
				return "You are the winner.", "omaha.result.win"
			}
		}
	}

	for i := 0; i < o.GetPlayerCnt(); i++ {
		if o.GetPlayer(i).GetIsHuman() && o.GetPlayer(i).GetFolded() {
			return "You folded.", "omaha.result.folded"
		}
	}

	for _, r := range results {
		if o.GetPlayer(r.PlayerIdx).GetIsHuman() && r.Mucked {
			return "You mucked.", "omaha.result.mucked"
		}
	}

	return "You lose.", "omaha.result.lose"
}

// ActionLogOutput 棋譜をJSON出力
func (owp *OmahaWebPresenter) ActionLogOutput(o interfaces.OmahaGame) string {
	return actionLogOutputJSON(o)
}
