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
	for i := 0; i < o.GetPlayerCnt(); i++ {
		player := o.GetPlayer(i)
		if player == nil || !player.GetIsHuman() || player.GetFolded() {
			continue
		}
		best := player.GetBestHand()
		if o.GetPhase() != domain.OmahaPhaseShowdown && o.GetPhase() != domain.OmahaPhaseEnd {
			_, best = player.PeekBestHand(o.GetCommunityCards())
		}
		hole := make([]*domain.Card, player.GetCardsSize())
		for idx := range hole {
			hole[idx] = player.GetCard(idx)
		}
		resObj.Players[i].LiveBestHandHoleIndices, resObj.Players[i].LiveBestHandBoardIndices = liveBestHandIndices(hole, o.GetCommunityCards(), best)
		break
	}
	resObj.IsHiLo = o.GetIsHiLo()
	resObj.PotAwards = make([]*controller.HoldemWebOutputPotAward, 0, len(o.GetPotAwards()))
	for _, a := range o.GetPotAwards() {
		resObj.PotAwards = append(resObj.PotAwards, &controller.HoldemWebOutputPotAward{Amount: a.Amount, Eligible: a.Eligible, HiWinners: a.HiWinners, HiPayouts: a.HiPayouts, LoWinners: a.LoWinners, LoPayouts: a.LoPayouts})
	}
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
	return actionLogToJSON(o.GetActionLog())
}
