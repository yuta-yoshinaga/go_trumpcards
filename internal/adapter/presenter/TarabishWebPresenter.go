//go:build !js || !wasm || extra3

package presenter

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// TarabishWebPresenter タラビッシュWebプレゼンタークラス
type TarabishWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *TarabishWebPresenter) Output(t interfaces.TarabishGame, lastErr error) string {
	resObj := p.buildBase(t)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(t, lastErr)
	// 受動ヒントは Output() でも埋める (#4483)。
	if hint := t.GetHint(); hint != nil {
		resObj.Hint = &controller.TarabishWebOutputHint{CardIndex: hint.CardIndex, Reason: hint.Reason}
	}
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *TarabishWebPresenter) buildBase(t interfaces.TarabishGame) *controller.TarabishWebOutput {
	resObj := new(controller.TarabishWebOutput)
	resObj.Phase = int(t.GetPhase())
	resObj.RoundNumber = t.GetRoundNumber()
	resObj.TrickNumber = t.GetTrickNumber()
	resObj.TrumpSuit = t.GetTrumpSuit()
	if up := t.GetUpCard(); up != nil {
		resObj.UpCard = cardToOutput(up)
	}
	resObj.TrumpTakerIdx = t.GetTrumpTakerIdx()
	resObj.CurrentPlayerIdx = t.GetCurrentPlayerIdx()
	resObj.LeadPlayerIdx = t.GetLeadPlayerIdx()
	resObj.DealerIdx = t.GetDealerIdx()
	resObj.ValidPlays = intSliceOrEmpty(t.GetValidPlayIndices(0))
	resObj.GameEndFlag = t.GetGameEndFlag()
	resObj.WinnerTeam = t.GetWinnerTeam()
	resObj.CurrentTrick = trickCardsToOutput(t.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(t)

	scores := make([]int, 0, domain.TarabishTeamCnt)
	roundPoints := make([]int, 0, domain.TarabishTeamCnt)
	for team := 0; team < domain.TarabishTeamCnt; team++ {
		scores = append(scores, t.GetScore(team))
		roundPoints = append(roundPoints, t.GetRoundPoints(team))
	}
	resObj.Scores = scores
	resObj.RoundPoints = roundPoints
	resObj.Config = controller.TarabishWebOutputConfig{Target: t.GetConfig().Target}
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築
func (p *TarabishWebPresenter) buildPlayersOutput(t interfaces.TarabishGame) []*controller.TarabishWebOutputPlayer {
	out := make([]*controller.TarabishWebOutputPlayer, 0)
	for i := 0; i < t.GetPlayerCnt(); i++ {
		player := t.GetPlayer(i)
		out = append(out, &controller.TarabishWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			Team:       domain.TarabishTeamOf(i),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, player.GetIsHuman()),
			MeldPoints: player.GetMeldPoints(),
			RunLen:     player.GetRunLen(),
			HasBella:   player.GetHasBella(),
			TrickCount: player.GetTrickCount(),
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *TarabishWebPresenter) buildMessage(t interfaces.TarabishGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if t.GetGameEndFlag() {
		params := map[string]string{
			"t0": strconv.Itoa(t.GetScore(0)),
			"t1": strconv.Itoa(t.GetScore(1)),
		}
		switch t.GetWinnerTeam() {
		case 0:
			return "", "tarabish.result.team0", params
		case 1:
			return "", "tarabish.result.team1", params
		default:
			return "", "tarabish.result.tie", params
		}
	}
	switch t.GetPhase() {
	case domain.TarabishPhaseBid:
		// **親は見送れない。** 案内を変えないと選べない選択肢を出すことになる。
		if t.GetDealerIdx() == 0 && t.IsHumanBidTurn() {
			return "", "tarabish.bid.dealerStuck", nil
		}
		return "", "tarabish.bid.choose", nil
	case domain.TarabishPhaseRoundEnd:
		return "", "tarabish.roundEnd", map[string]string{
			"round": strconv.Itoa(t.GetRoundNumber()),
			"t0":    strconv.Itoa(t.GetRoundPoints(0)),
			"t1":    strconv.Itoa(t.GetRoundPoints(1)),
		}
	}
	return "", "tarabish.play", nil
}

// HintOutput ヒント情報をJSON出力する
func (p *TarabishWebPresenter) HintOutput(t interfaces.TarabishGame) string {
	resObj := p.buildBase(t)
	if hint := t.GetHint(); hint != nil {
		resObj.Hint = &controller.TarabishWebOutputHint{CardIndex: hint.CardIndex, Reason: hint.Reason}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *TarabishWebPresenter) ActionLogOutput(t interfaces.TarabishGame) string {
	return actionLogOutputJSON(t)
}
