//go:build !js || !wasm || solo

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// EuchreWebPresenter ユーカーWebプレゼンタークラス
type EuchreWebPresenter struct{}

// Output ゲーム状態をJSON出力
// euchreThresholdIf returns v only when a score is present, so the thresholds
// appear exactly where they can be read against something.
func euchreThresholdIf(score *int, v int) *int {
	if score == nil {
		return nil
	}
	return &v
}

func (p *EuchreWebPresenter) Output(e interfaces.EuchreGame, lastErr error) string {
	resObj := p.buildBase(e)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(e, e.GetCurrentTrick(), lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **フェーズと手番はここでは見ない。**Euchre.GetHint() が自分で
	// 「人間の手番で、かつ行動を選べる状態か」を確かめて nil を返す。
	if hint := e.GetHint(); hint != nil {
		resObj.Hint = &controller.EuchreWebOutputHint{
			CardIndex: hint.CardIndex,
			OrderUp:   hint.OrderUp,
			Suit:      hint.Suit,
			GoAlone:   hint.GoAlone,
			Reason:    hint.Reason,
			Score:     hint.Score,
			// Send the thresholds alongside the score so the client never has to
			// hardcode them; a copy on the other side drifts the first time one
			// of these constants changes.
			OrderUpScore: euchreThresholdIf(hint.Score, domain.EuchreOrderUpScore),
			GoAloneScore: euchreThresholdIf(hint.Score, domain.EuchreGoAloneScore),
		}
	}

	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *EuchreWebPresenter) buildBase(e interfaces.EuchreGame) *controller.EuchreWebOutput {
	resObj := new(controller.EuchreWebOutput)
	resObj.Phase = int(e.GetPhase())
	resObj.RoundNumber = e.GetRoundNumber()
	resObj.TrickNumber = e.GetTrickNumber()
	resObj.CurrentPlayerIdx = e.GetCurrentPlayerIdx()
	resObj.BidPlayerIdx = e.GetBidPlayerIdx()
	resObj.DealerIdx = e.GetDealerIdx()
	resObj.TrumpSuit = e.GetTrumpSuit()
	resObj.FaceUpCard = cardToOutput(e.GetFaceUpCard())
	resObj.MakerTeam = e.GetMakerTeam()
	resObj.GoingAlone = e.GetGoingAlone()
	resObj.GoingAlonePlayerIdx = e.GetGoingAlonePlayerIdx()
	resObj.TeamScores = [2]int{e.GetTeamScore(0), e.GetTeamScore(1)}
	resObj.GameEndFlag = e.GetGameEndFlag()
	resObj.WinnerTeam = e.GetWinnerTeam()
	resObj.LeadPlayerIdx = e.GetLeadPlayerIdx()

	// 設定
	cfg := e.GetConfig()
	resObj.Config = controller.EuchreWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		PointLimit:    cfg.PointLimit,
	}

	resObj.CurrentTrick = trickCardsToOutput(e.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(e)
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築
func (p *EuchreWebPresenter) buildPlayersOutput(e interfaces.EuchreGame) []*controller.EuchreWebOutputPlayer {
	out := make([]*controller.EuchreWebOutputPlayer, 0)
	for i := 0; i < e.GetPlayerCnt(); i++ {
		player := e.GetPlayer(i)
		pObj := &controller.EuchreWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, player.GetIsHuman()),
			Team:       player.GetTeam(),
			TrickCount: player.GetTrickCount(),
		}
		out = append(out, pObj)
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *EuchreWebPresenter) buildMessage(e interfaces.EuchreGame, trick []*domain.TrickCard, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if e.GetGameEndFlag() {
		winnerTeam := e.GetWinnerTeam()
		msg := fmt.Sprintf("ゲーム終了！ チーム%dの勝ち！", winnerTeam)
		code := fmt.Sprintf("euchre.result.team%dWin", winnerTeam)
		params := map[string]string{"team": fmt.Sprintf("%d", winnerTeam)}
		return msg, code, params
	}
	switch e.GetPhase() {
	case domain.EuchrePhasePickUp:
		return "", "euchre.pickUpPhase", nil
	case domain.EuchrePhaseCallTrump:
		return "", "euchre.callTrumpPhase", nil
	case domain.EuchrePhaseDiscard:
		return "", "euchre.discardPhase", nil
	case domain.EuchrePhasePlay:
		if len(trick) == 0 {
			return "", "euchre.playPhase.lead", nil
		}
		return "", "euchre.playPhase.follow", nil
	case domain.EuchrePhaseTrickEnd:
		return "", "euchre.trickEnd", nil
	case domain.EuchrePhaseRoundEnd:
		return "", "euchre.roundEnd", nil
	}
	return "", "", nil
}

// HintOutput ヒント情報をJSON出力する
func (p *EuchreWebPresenter) HintOutput(e interfaces.EuchreGame) string {
	hint := e.GetHint()
	resObj := p.buildBase(e)
	if hint != nil {
		resObj.Hint = &controller.EuchreWebOutputHint{
			CardIndex: hint.CardIndex,
			OrderUp:   hint.OrderUp,
			Suit:      hint.Suit,
			GoAlone:   hint.GoAlone,
			Reason:    hint.Reason,
			Score:     hint.Score,
			// Send the thresholds alongside the score so the client never has to
			// hardcode them; a copy on the other side drifts the first time one
			// of these constants changes.
			OrderUpScore: euchreThresholdIf(hint.Score, domain.EuchreOrderUpScore),
			GoAloneScore: euchreThresholdIf(hint.Score, domain.EuchreGoAloneScore),
		}
	}
	// **「頼んだヒントか」を CLI が見分けられるようにする。**このゲーム群の
	// `hintAvailable` は画面のラベルとして既に使われているので、別キーを出す (#4483)。
	if hint != nil {
		resObj.MessageCode = "euchre.hintRequested"
	} else {
		resObj.MessageCode = "euchre.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *EuchreWebPresenter) ActionLogOutput(e interfaces.EuchreGame) string {
	return actionLogOutputJSON(e)
}
