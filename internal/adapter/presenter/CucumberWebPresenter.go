//go:build !js || !wasm || classic

package presenter

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// CucumberWebPresenter キューカンバーWebプレゼンタークラス
type CucumberWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *CucumberWebPresenter) Output(s interfaces.CucumberGame, lastErr error) string {
	resObj := p.buildBase(s)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(s, lastErr)
	// 受動ヒントは Output() でも埋める (#4483)。
	resObj.Hint = cucumberHintOutput(s)
	return marshalOrError(resObj)
}

// cucumberHintOutput はヒントを出力形に変換する。
func cucumberHintOutput(s interfaces.CucumberGame) *controller.CucumberWebOutputHint {
	hint := s.GetHint()
	if hint == nil {
		return nil
	}
	return &controller.CucumberWebOutputHint{CardIndex: hint.CardIndex, Reason: hint.Reason}
}

// buildBase 共通フィールドを構築
func (p *CucumberWebPresenter) buildBase(s interfaces.CucumberGame) *controller.CucumberWebOutput {
	resObj := new(controller.CucumberWebOutput)
	resObj.Phase = int(s.GetPhase())
	resObj.ValidPlays = intSliceOrEmpty(s.GetValidPlayIndices(0))
	resObj.HighestInTrick = s.HighestInTrick()
	resObj.Forced = s.IsForcedLowest(0)
	resObj.CurrentTrick = trickCardsToOutput(s.GetCurrentTrick())
	resObj.CurrentPlayerIdx = s.GetCurrentPlayerIdx()
	resObj.LeadPlayerIdx = s.GetLeadPlayerIdx()
	resObj.TrickNumber = s.GetTrickNumber()
	resObj.TotalTricks = domain.CucumberHandSize
	resObj.RoundNumber = s.GetRoundNumber()
	resObj.LastTrickWinnerIdx = s.GetLastTrickWinnerIdx()
	resObj.LastPenalty = s.GetLastPenalty()
	resObj.GameEndFlag = s.GetGameEndFlag()
	resObj.WinnerIdx = s.GetWinnerIdx()
	resObj.Players = p.buildPlayersOutput(s)
	cfg := s.GetConfig()
	resObj.Config = controller.CucumberWebOutputConfig{
		PlayerCnt:   cfg.PlayerCnt,
		TargetScore: cfg.TargetScore,
	}
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築
func (p *CucumberWebPresenter) buildPlayersOutput(s interfaces.CucumberGame) []*controller.CucumberWebOutputPlayer {
	out := make([]*controller.CucumberWebOutputPlayer, 0)
	for i := 0; i < s.GetPlayerCnt(); i++ {
		player := s.GetPlayer(i)
		out = append(out, &controller.CucumberWebOutputPlayer{
			ID:        i,
			IsHuman:   player.GetIsHuman(),
			CardCount: player.GetCardsSize(),
			Cards:     playerCardsToOutput(player, player.GetIsHuman()),
			Penalty:   player.GetPenalty(),
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *CucumberWebPresenter) buildMessage(s interfaces.CucumberGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if s.GetGameEndFlag() {
		params := map[string]string{"n": strconv.Itoa(s.GetPlayer(s.GetWinnerIdx()).GetPenalty())}
		if s.GetWinnerIdx() == 0 {
			return "", "cucumber.result.you", params
		}
		params["idx"] = strconv.Itoa(s.GetWinnerIdx())
		return "", "cucumber.result.cpu", params
	}

	if s.GetPhase() == domain.CucumberPhaseRoundEnd {
		// **失点はラウンドに 1 回だけの出来事。** 配り直す前に読ませます。
		loser := s.GetLastTrickWinnerIdx()
		params := map[string]string{"n": strconv.Itoa(s.GetLastPenalty())}
		if loser == 0 {
			return "", "cucumber.round.you", params
		}
		params["idx"] = strconv.Itoa(loser)
		return "", "cucumber.round.cpu", params
	}
	if !s.IsHumanTurn() {
		return "", "cucumber.waiting", nil
	}
	// **出す札が決まっている場面は、選べる場面と言い分けます。**
	if s.IsForcedLowest(0) {
		return "", "cucumber.forced", nil
	}
	if s.HighestInTrick() == 0 {
		return "", "cucumber.lead", nil
	}
	return "", "cucumber.beat", map[string]string{"n": strconv.Itoa(s.HighestInTrick())}
}

// HintOutput ヒント情報をJSON出力する
func (p *CucumberWebPresenter) HintOutput(s interfaces.CucumberGame) string {
	resObj := p.buildBase(s)
	resObj.Hint = cucumberHintOutput(s)
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *CucumberWebPresenter) ActionLogOutput(s interfaces.CucumberGame) string {
	return actionLogOutputJSON(s)
}
