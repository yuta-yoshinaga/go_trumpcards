//go:build !js || !wasm || casino

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// DramahaWebPresenter ドラマハホールデムWebプレゼンタークラス
type DramahaWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (owp *DramahaWebPresenter) Output(o interfaces.DramahaGame, lastErr error) string {
	resObj := owp.buildOutput(o, lastErr)
	return marshalOrError(resObj)
}

// buildOutput ゲーム状態をDramahaWebOutputに変換
func (owp *DramahaWebPresenter) buildOutput(o interfaces.DramahaGame, lastErr error) *controller.HoldemWebOutput {
	resObj := buildCommunityCardBaseOutput(o)
	resObj.Players = buildPokerPlayersOutput(o.GetPhase(), o.GetPlayerCnt(), func(i int) communityCardPresenterPlayer { return o.GetPlayer(i) }, domain.DramahaPhaseShowdown, domain.DramahaPhaseEnd, pokerHandName)
	resObj.IsHiLo = true /* ドラマハは常に二分する */
	resObj.Message, resObj.MessageCode, resObj.MessageParams = owp.buildMessage(o, lastErr)
	return resObj
}

func (owp *DramahaWebPresenter) buildMessage(o interfaces.DramahaGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if o.IsMuckAvailable() {
		return "", "dramaha.muck.prompt", nil
	}
	if o.GetGameEndFlag() {
		msg, code := owp.buildResultMessage(o)
		return msg, code, nil
	}
	return "", "", nil
}

func (owp *DramahaWebPresenter) buildResultMessage(o interfaces.DramahaGame) (string, string) {
	results := o.GetRoundResults()
	if len(results) == 0 {
		return "", "dramaha.result.gameOver"
	}

	hiLo := true /* ドラマハは常に二分する */
	for _, r := range results {
		if o.GetPlayer(r.PlayerIdx).GetIsHuman() {
			if r.WonAmount > 0 {
				if hiLo {
					switch {
					case r.HiWonAmount > 0 && r.LowWonAmount > 0:
						return "", "dramahahilo.result.scoop"
					case r.LowWonAmount > 0:
						return "", "dramahahilo.result.lowWin"
					case r.HiWonAmount > 0:
						return "", "dramahahilo.result.hiWin"
					}
				}
				return "", "dramaha.result.win"
			}
		}
	}

	for i := 0; i < o.GetPlayerCnt(); i++ {
		if o.GetPlayer(i).GetIsHuman() && o.GetPlayer(i).GetFolded() {
			return "", "dramaha.result.folded"
		}
	}

	for _, r := range results {
		if o.GetPlayer(r.PlayerIdx).GetIsHuman() && r.Mucked {
			return "", "dramaha.result.mucked"
		}
	}

	return "", "dramaha.result.lose"
}

// ActionLogOutput 棋譜をJSON出力
func (owp *DramahaWebPresenter) ActionLogOutput(o interfaces.DramahaGame) string {
	return actionLogOutputJSON(o)
}
