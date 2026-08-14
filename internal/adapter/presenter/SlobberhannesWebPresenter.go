//go:build !js || !wasm || classic

package presenter

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// SlobberhannesWebPresenter スロバーハンネスWebプレゼンタークラス
type SlobberhannesWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *SlobberhannesWebPresenter) Output(s interfaces.SlobberhannesGame, lastErr error) string {
	resObj := p.buildBase(s)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(s, lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は command:"hint" 専用
	// のレスポンスで、ページの state にはマージされない (#4483)。
	if hint := s.GetHint(); hint != nil {
		resObj.Hint = &controller.SlobberhannesWebOutputHint{CardIndex: hint.CardIndex, Reason: hint.Reason}
	}
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *SlobberhannesWebPresenter) buildBase(s interfaces.SlobberhannesGame) *controller.SlobberhannesWebOutput {
	resObj := new(controller.SlobberhannesWebOutput)
	resObj.Phase = int(s.GetPhase())
	resObj.RoundNumber = s.GetRoundNumber()
	resObj.TrickNumber = s.GetTrickNumber()
	resObj.CurrentPlayerIdx = s.GetCurrentPlayerIdx()
	resObj.LeadPlayerIdx = s.GetLeadPlayerIdx()
	resObj.DealerIdx = s.GetDealerIdx()
	resObj.ValidPlays = intSliceOrEmpty(s.GetValidPlayIndices(0))
	resObj.GameEndFlag = s.GetGameEndFlag()
	resObj.WinnerIdx = s.GetWinnerIdx()
	resObj.CurrentTrick = trickCardsToOutput(s.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(s)
	resObj.Config = controller.SlobberhannesWebOutputConfig{Rounds: s.GetConfig().Rounds}
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築
func (p *SlobberhannesWebPresenter) buildPlayersOutput(s interfaces.SlobberhannesGame) []*controller.SlobberhannesWebOutputPlayer {
	out := make([]*controller.SlobberhannesWebOutputPlayer, 0)
	for i := 0; i < s.GetPlayerCnt(); i++ {
		player := s.GetPlayer(i)
		out = append(out, &controller.SlobberhannesWebOutputPlayer{
			ID:             i,
			IsHuman:        player.GetIsHuman(),
			CardCount:      player.GetCardsSize(),
			Cards:          playerCardsToOutput(player, player.GetIsHuman()),
			Score:          player.GetScore(),
			TrickCount:     player.GetTrickCount(),
			TookFirstTrick: player.GetTookFirstTrick(),
			TookLastTrick:  player.GetTookLastTrick(),
			TookQueen:      player.GetTookQueen(),
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *SlobberhannesWebPresenter) buildMessage(s interfaces.SlobberhannesGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if s.GetGameEndFlag() {
		if s.GetWinnerIdx() < 0 {
			return "", "slobberhannes.result.tie", nil
		}
		return "", "slobberhannes.result.winner", map[string]string{
			"idx":   strconv.Itoa(s.GetWinnerIdx()),
			"score": strconv.Itoa(s.GetPlayer(s.GetWinnerIdx()).GetScore()),
		}
	}
	if s.GetPhase() == domain.SlobberhannesPhaseRoundEnd {
		return "", "slobberhannes.roundEnd", map[string]string{
			"round": strconv.Itoa(s.GetRoundNumber()),
		}
	}
	// **最初と最後のトリックは中身に関係なく罰点**なので、そこだけ案内を変える。
	switch s.GetTrickNumber() {
	case 0:
		return "", "slobberhannes.play.firstTrick", nil
	case domain.SlobberhannesTricksPerRound - 1:
		return "", "slobberhannes.play.lastTrick", nil
	}
	return "", "slobberhannes.play.normal", nil
}

// HintOutput ヒント情報をJSON出力する
func (p *SlobberhannesWebPresenter) HintOutput(s interfaces.SlobberhannesGame) string {
	resObj := p.buildBase(s)
	if hint := s.GetHint(); hint != nil {
		resObj.Hint = &controller.SlobberhannesWebOutputHint{CardIndex: hint.CardIndex, Reason: hint.Reason}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *SlobberhannesWebPresenter) ActionLogOutput(s interfaces.SlobberhannesGame) string {
	return actionLogOutputJSON(s)
}
