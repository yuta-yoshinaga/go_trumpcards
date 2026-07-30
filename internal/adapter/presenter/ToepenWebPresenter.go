//go:build !js || !wasm || extra3

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// ToepenWebPresenter トゥーペンWebプレゼンタークラス
type ToepenWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *ToepenWebPresenter) Output(t interfaces.ToepenGame, lastErr error) string {
	resObj := p.buildBase(t)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(t, lastErr)
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *ToepenWebPresenter) buildBase(t interfaces.ToepenGame) *controller.ToepenWebOutput {
	resObj := new(controller.ToepenWebOutput)
	resObj.Phase = int(t.GetPhase())
	resObj.CurrentPlayerIdx = t.GetCurrentPlayerIdx()
	resObj.LeadPlayerIdx = t.GetLeadPlayerIdx()
	resObj.DealerIdx = t.GetDealerIdx()
	resObj.LeadSuit = t.GetLeadSuit()
	resObj.TrickNumber = t.GetTrickNumber()
	resObj.HandNumber = t.GetHandNumber()
	resObj.Stake = t.GetStake()
	resObj.KnockerIdx = t.GetKnockerIdx()
	resObj.PendingRespondent = t.GetPendingRespondent()
	resObj.LastTrickWinner = t.GetLastTrickWinner()
	resObj.MaxLives = domain.ToepenMaxLives
	resObj.GameEndFlag = t.GetGameEndFlag()
	resObj.WinnerIdx = t.GetWinnerIdx()
	resObj.CurrentTrick = trickCardsToOutput(t.GetCurrentTrick())

	// フォロー義務の判定はここで一度だけ行う。クライアントに再実装させると
	// 規則の実装が 2 つになり、答えも 2 つになる。
	valid := t.GetValidPlayIndices(0)
	resObj.ValidPlayIndices = make([]int, 0, len(valid))
	resObj.ValidPlayIndices = append(resObj.ValidPlayIndices, valid...)

	cfg := t.GetConfig()
	resObj.Config = controller.ToepenWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		PlayerCnt:     cfg.PlayerCnt,
	}
	resObj.Players = p.buildPlayersOutput(t)

	// ヒントは通常のレスポンスにも載せる。他ゲームは HintOutput でしか設定して
	// おらず、フロントは通常の state レスポンスから読むため、どのページも呼んで
	// いない = ヒントトグルが何も表示しない状態になっている。
	if !t.GetGameEndFlag() {
		resObj.Hint = toepenHint(t)
	}
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築する。
//
// CPU の手札は伏せる。Workers はこの JSON をそのままブラウザへ返すので、ここで
// 落とさなかったものは相手の手札がそのまま見えることを意味する。枚数・失点・
// 降参/脱落は公開情報なので常に送る。
func (p *ToepenWebPresenter) buildPlayersOutput(t interfaces.ToepenGame) []*controller.ToepenWebOutputPlayer {
	players := t.GetPlayers()
	out := make([]*controller.ToepenWebOutputPlayer, 0, len(players))
	for i, player := range players {
		if player == nil {
			continue
		}
		reveal := player.GetIsHuman() || t.GetGameEndFlag()
		out = append(out, &controller.ToepenWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, reveal),
			Lives:      t.GetLives(i),
			Folded:     t.IsFolded(i),
			Eliminated: t.IsEliminated(i),
			Hidden:     !reveal,
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *ToepenWebPresenter) buildMessage(t interfaces.ToepenGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if t.GetGameEndFlag() {
		if t.GetWinnerIdx() == 0 {
			return "you win", "toepen.win", nil
		}
		return "you lose", "toepen.lose", nil
	}
	if t.GetPhase() == domain.ToepenPhaseHandEnd {
		return "hand over", "toepen.hand_end", nil
	}
	return "", "", nil
}

// HintOutput ヒント情報を出力する
func (p *ToepenWebPresenter) HintOutput(t interfaces.ToepenGame) string {
	resObj := p.buildBase(t)
	resObj.Hint = toepenHint(t)
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜を出力する
func (p *ToepenWebPresenter) ActionLogOutput(t interfaces.ToepenGame) string {
	return actionLogOutputJSON(t)
}

// toepenHint 人間プレイヤーへの推奨手を返す。CPU と同じ意思決定を通す。
func toepenHint(t interfaces.ToepenGame) *controller.ToepenWebOutputHint {
	if t.GetGameEndFlag() {
		return &controller.ToepenWebOutputHint{Reason: "toepen.hint.game_end"}
	}
	switch t.GetPhase() {
	case domain.ToepenPhaseHandEnd:
		return &controller.ToepenWebOutputHint{Reason: "toepen.hint.hand_end"}
	case domain.ToepenPhaseRespond:
		if t.GetPendingRespondent() != 0 {
			return &controller.ToepenWebOutputHint{Reason: "toepen.hint.not_your_turn"}
		}
		fold := t.ToepenCpuDecide(0).Fold
		reason := "toepen.hint.stay"
		if fold {
			reason = "toepen.hint.fold"
		}
		return &controller.ToepenWebOutputHint{Fold: &fold, Reason: reason}
	default:
		if t.GetCurrentPlayerIdx() != 0 {
			return &controller.ToepenWebOutputHint{Reason: "toepen.hint.not_your_turn"}
		}
		idx := t.ToepenCpuDecide(0).HandIdx
		if idx < 0 {
			return &controller.ToepenWebOutputHint{Reason: "toepen.hint.none"}
		}
		return &controller.ToepenWebOutputHint{CardIndex: &idx, Reason: "toepen.hint.play"}
	}
}
