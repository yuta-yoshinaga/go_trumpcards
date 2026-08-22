//go:build !js || !wasm || extra3

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// RamschWebPresenter Ramsch Web presenter.
type RamschWebPresenter struct{}

// Output renders the game state as a JSON string.
func (p *RamschWebPresenter) Output(s interfaces.RamschGame, lastErr error) string {
	resObj := p.buildBaseOutput(s)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(s, lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **フェーズと手番はここでは見ない。**Ramsch.GetHint() が自分で
	// 「人間の手番で、かつ行動を選べる状態か」を確かめて nil を返す。
	if hint := s.GetHint(); hint != nil {
		resObj.Hint = &controller.RamschWebOutputHint{
			CardIndex: hint.CardIndex,
			Reason:    hint.Reason,
		}
	}

	return marshalOrError(resObj)
}

// buildBaseOutput builds the base output struct.
func (p *RamschWebPresenter) buildBaseOutput(s interfaces.RamschGame) *controller.RamschWebOutput {
	resObj := new(controller.RamschWebOutput)
	resObj.Phase = int(s.GetPhase())
	resObj.RoundNumber = s.GetRoundNumber()
	resObj.TrickNumber = s.GetTrickNumber()
	resObj.CurrentPlayerIdx = s.GetCurrentPlayerIdx()
	resObj.ForehandIdx = s.GetForehandIdx()
	resObj.MiddlehandIdx = s.GetMiddlehandIdx()
	resObj.RearhandIdx = s.GetRearhandIdx()
	resObj.DealerIdx = s.GetDealerIdx()
	resObj.LoserIdx = s.GetLoserIdx()
	resObj.Durchmarsch = s.IsDurchmarsch()
	resObj.DurchmarschIdx = s.GetDurchmarschIdx()
	resObj.GameEndFlag = s.GetGameEndFlag()
	resObj.LeadPlayerIdx = s.GetLeadPlayerIdx()

	// **伏せ札はラウンドが終わるまで見せない。** 最終トリックの獲得者が
	// 受け取る 2 枚なので、中身が分かると「最後を取るか避けるか」の判断が
	// 完全情報になり、このゲームの終盤が消える。
	phase := s.GetPhase()
	if phase == domain.RamschPhaseRoundEnd || phase == domain.RamschPhaseGameEnd {
		if skat := s.GetSkat(); len(skat) > 0 {
			resObj.Skat = make([]*controller.WebOutputCard, len(skat))
			for i, c := range skat {
				resObj.Skat[i] = cardToOutput(c)
			}
		}
	}

	cfg := s.GetConfig()
	resObj.Config = controller.RamschWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetScore:   cfg.TargetScore,
	}

	resObj.CurrentTrick = trickCardsToOutput(s.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(s)
	return resObj
}

// buildPlayersOutput builds the per-player output.
func (p *RamschWebPresenter) buildPlayersOutput(s interfaces.RamschGame) []*controller.RamschWebOutputPlayer {
	out := make([]*controller.RamschWebOutputPlayer, 0)
	for i := 0; i < s.GetPlayerCnt(); i++ {
		player := s.GetPlayer(i)
		out = append(out, &controller.RamschWebOutputPlayer{
			ID:              i,
			IsHuman:         player.GetIsHuman(),
			CardCount:       player.GetCardsSize(),
			Cards:           playerCardsToOutput(player, player.GetIsHuman()),
			CardPoints:      player.GetCardPoints(),
			RoundsWon:       player.GetRoundsWon(),
			RoundsLost:      player.GetRoundsLost(),
			RoundScore:      player.GetRoundScore(),
			CumulativeScore: player.GetCumulativeScore(),
			TrickCount:      player.GetTrickCount(),
		})
	}
	return out
}

// buildMessage produces a message and i18n message code for the current state.
func (p *RamschWebPresenter) buildMessage(s interfaces.RamschGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if s.GetGameEndFlag() {
		return "", "ramsch.gameEnd", nil
	}
	switch s.GetPhase() {
	case domain.RamschPhasePlay:
		if len(s.GetCurrentTrick()) == 0 {
			return "", "ramsch.playPhase.lead", nil
		}
		return "", "ramsch.playPhase.follow", nil
	case domain.RamschPhaseTrickEnd:
		return "", "ramsch.trickEnd", nil
	case domain.RamschPhaseRoundEnd:
		return "", "ramsch.roundEnd", nil
	}
	return "", "", nil
}

// HintOutput renders the hint output.
func (p *RamschWebPresenter) HintOutput(s interfaces.RamschGame) string {
	hint := s.GetHint()
	resObj := p.buildBaseOutput(s)
	if hint != nil {
		resObj.Hint = &controller.RamschWebOutputHint{
			CardIndex: hint.CardIndex,
			Reason:    hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput returns the round's action log as JSON.
func (p *RamschWebPresenter) ActionLogOutput(s interfaces.RamschGame) string {
	return actionLogOutputJSON(s)
}
