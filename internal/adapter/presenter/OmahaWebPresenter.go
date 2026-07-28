//go:build !js || !wasm || casino

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
	resObj.IsHiLo = o.GetIsHiLo()
	resObj.Message, resObj.MessageCode, resObj.MessageParams = owp.buildMessage(o, lastErr)
	return resObj
}

func (owp *OmahaWebPresenter) buildMessage(o interfaces.OmahaGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if o.IsMuckAvailable() {
		return "", "omaha.muck.prompt", nil
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
		return "", "omaha.result.gameOver"
	}

	hiLo := o.GetIsHiLo()
	for _, r := range results {
		if o.GetPlayer(r.PlayerIdx).GetIsHuman() {
			if r.WonAmount > 0 {
				if hiLo {
					switch {
					case r.HiWonAmount > 0 && r.LowWonAmount > 0:
						return "", "omahahilo.result.scoop"
					case r.LowWonAmount > 0:
						return "", "omahahilo.result.lowWin"
					case r.HiWonAmount > 0:
						return "", "omahahilo.result.hiWin"
					}
				}
				return "", "omaha.result.win"
			}
		}
	}

	for i := 0; i < o.GetPlayerCnt(); i++ {
		if o.GetPlayer(i).GetIsHuman() && o.GetPlayer(i).GetFolded() {
			return "", "omaha.result.folded"
		}
	}

	for _, r := range results {
		if o.GetPlayer(r.PlayerIdx).GetIsHuman() && r.Mucked {
			return "", "omaha.result.mucked"
		}
	}

	return "", "omaha.result.lose"
}

// ActionLogOutput 棋譜をJSON出力
func (owp *OmahaWebPresenter) ActionLogOutput(o interfaces.OmahaGame) string {
	return actionLogOutputJSON(o)
}
