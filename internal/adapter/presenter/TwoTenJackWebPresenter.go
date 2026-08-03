package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// TwoTenJackWebPresenter ツーテンジャックWebプレゼンタークラス
type TwoTenJackWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *TwoTenJackWebPresenter) Output(s interfaces.TwoTenJackGame, lastErr error) string {
	resObj := p.buildBase(s)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(s, lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **フェーズと手番はここでは見ない。**TwoTenJack.GetHint() が自分で
	// 「人間の手番で、かつ行動を選べる状態か」を確かめて nil を返す。
	if hint := s.GetHint(); hint != nil {
		resObj.Hint = &controller.TwoTenJackWebOutputHint{
			CardIndex: hint.CardIndex,
			TrumpSuit: hint.TrumpSuit,
			Reason:    hint.Reason,
		}
	}

	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *TwoTenJackWebPresenter) buildBase(s interfaces.TwoTenJackGame) *controller.TwoTenJackWebOutput {
	resObj := new(controller.TwoTenJackWebOutput)
	resObj.Phase = int(s.GetPhase())
	resObj.RoundNumber = s.GetRoundNumber()
	resObj.TrickNumber = s.GetTrickNumber()
	resObj.CurrentPlayerIdx = s.GetCurrentPlayerIdx()
	resObj.DeclarerIdx = s.GetDeclarerIdx()
	resObj.TrumpSuit = s.GetTrumpSuit()
	resObj.GameEndFlag = s.GetGameEndFlag()
	resObj.WinnerTeam = s.GetWinnerTeam()
	resObj.LeadPlayerIdx = s.GetLeadPlayerIdx()

	cfg := s.GetConfig()
	resObj.Config = controller.TwoTenJackWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		PointLimit:    cfg.PointLimit,
	}

	resObj.CurrentTrick = trickCardsToOutput(s.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(s)
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築
func (p *TwoTenJackWebPresenter) buildPlayersOutput(s interfaces.TwoTenJackGame) []*controller.TwoTenJackWebOutputPlayer {
	out := make([]*controller.TwoTenJackWebOutputPlayer, 0)
	for i := 0; i < s.GetPlayerCnt(); i++ {
		player := s.GetPlayer(i)
		pObj := &controller.TwoTenJackWebOutputPlayer{
			ID:              i,
			IsHuman:         player.GetIsHuman(),
			CardCount:       player.GetCardsSize(),
			Cards:           playerCardsToOutput(player, player.GetIsHuman()),
			RoundScore:      player.GetRoundScore(),
			CumulativeScore: player.GetCumulativeScore(),
			TrickCount:      player.GetTrickCount(),
			CapturedPoints:  player.GetCapturedPointCards(),
		}
		out = append(out, pObj)
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *TwoTenJackWebPresenter) buildMessage(s interfaces.TwoTenJackGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if s.GetGameEndFlag() {
		team := s.GetWinnerTeam()
		// Human is in team 0. Show winner based on team containing seat 0.
		return buildWinnerWebMessage("twotenjack", team, team == 0)
	}
	switch s.GetPhase() {
	case domain.TwoTenJackPhaseDeclare:
		return "", "twotenjack.declarePhase", nil
	case domain.TwoTenJackPhasePlay:
		if len(s.GetCurrentTrick()) == 0 {
			return "", "twotenjack.playPhase.lead", nil
		}
		return "", "twotenjack.playPhase.follow", nil
	case domain.TwoTenJackPhaseTrickEnd:
		return "", "twotenjack.trickEnd", nil
	case domain.TwoTenJackPhaseRoundEnd:
		return "", "twotenjack.roundEnd", nil
	}
	return "", "", nil
}

// HintOutput ヒント情報をJSON出力する
func (p *TwoTenJackWebPresenter) HintOutput(s interfaces.TwoTenJackGame) string {
	hint := s.GetHint()
	resObj := p.buildBase(s)
	if hint != nil {
		resObj.Hint = &controller.TwoTenJackWebOutputHint{
			CardIndex: hint.CardIndex,
			TrumpSuit: hint.TrumpSuit,
			Reason:    hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *TwoTenJackWebPresenter) ActionLogOutput(s interfaces.TwoTenJackGame) string {
	return actionLogOutputJSON(s)
}
