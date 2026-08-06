package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// WhistWebPresenter ホイストWebプレゼンタークラス
type WhistWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *WhistWebPresenter) Output(w interfaces.WhistGame, lastErr error) string {
	resObj := p.buildBase(w)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(w, w.GetCurrentTrick(), lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **フェーズと手番はここでは見ない。**Whist.GetHint() が自分で
	// 「人間の手番で、かつ行動を選べる状態か」を確かめて nil を返す。
	if hint := w.GetHint(); hint != nil {
		resObj.Hint = &controller.WhistWebOutputHint{
			CardIndex: hint.CardIndex,
			Reason:    hint.Reason,
		}
	}

	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
// validPlayIndices は人間がいま出せる手札の位置を返す。
//
// **判定はドメインの GetValidPlayIndices をそのまま呼ぶ。**フォロースートの
// 規則をフロントに複製すると、ドメインだけ直したときに黙って食い違う。
// プレイフェーズで人間の手番でなければ空 -- 呼び出し側は空を「制限なし」とは
// 解釈せず、手番かどうかで先に分岐する (#4742)。
func (p *WhistWebPresenter) validPlayIndices(w interfaces.WhistGame) []int {
	if w.GetPhase() != domain.WhistPhasePlay || !w.IsHumanTurn() {
		return make([]int, 0)
	}
	idx := w.GetValidPlayIndices(w.GetCurrentPlayerIdx())
	if idx == nil {
		return make([]int, 0)
	}
	return idx
}

func (p *WhistWebPresenter) buildBase(w interfaces.WhistGame) *controller.WhistWebOutput {
	resObj := new(controller.WhistWebOutput)
	resObj.Phase = int(w.GetPhase())
	resObj.RoundNumber = w.GetRoundNumber()
	resObj.TrickNumber = w.GetTrickNumber()
	resObj.CurrentPlayerIdx = w.GetCurrentPlayerIdx()
	resObj.TrumpSuit = w.GetTrumpSuit()
	resObj.DealerIdx = w.GetDealerIdx()
	resObj.TeamScores = [2]int{w.GetTeamScore(0), w.GetTeamScore(1)}
	resObj.GameEndFlag = w.GetGameEndFlag()
	resObj.WinnerTeam = w.GetWinnerTeam()
	resObj.LeadPlayerIdx = w.GetLeadPlayerIdx()
	resObj.ValidPlayIndices = p.validPlayIndices(w)

	// 設定
	cfg := w.GetConfig()
	resObj.Config = controller.WhistWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		PointLimit:    cfg.PointLimit,
	}

	resObj.CurrentTrick = trickCardsToOutput(w.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(w)
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築
func (p *WhistWebPresenter) buildPlayersOutput(w interfaces.WhistGame) []*controller.WhistWebOutputPlayer {
	out := make([]*controller.WhistWebOutputPlayer, 0)
	for i := 0; i < w.GetPlayerCnt(); i++ {
		player := w.GetPlayer(i)
		pObj := &controller.WhistWebOutputPlayer{
			ID:              i,
			IsHuman:         player.GetIsHuman(),
			CardCount:       player.GetCardsSize(),
			Cards:           playerCardsToOutput(player, player.GetIsHuman()),
			RoundScore:      player.GetRoundScore(),
			CumulativeScore: player.GetCumulativeScore(),
			TrickCount:      player.GetTrickCount(),
			Team:            player.GetTeam(),
		}
		out = append(out, pObj)
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *WhistWebPresenter) buildMessage(w interfaces.WhistGame, trick []*domain.TrickCard, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if w.GetGameEndFlag() {
		winnerTeam := w.GetWinnerTeam()
		msg := fmt.Sprintf("ゲーム終了！ チーム%dの勝ち！", winnerTeam)
		code := fmt.Sprintf("whist.result.team%dWin", winnerTeam)
		params := map[string]string{"team": fmt.Sprintf("%d", winnerTeam)}
		return msg, code, params
	}
	switch w.GetPhase() {
	case domain.WhistPhasePlay:
		if len(trick) == 0 {
			return "", "whist.playPhase.lead", nil
		}
		return "", "whist.playPhase.follow", nil
	case domain.WhistPhaseTrickEnd:
		return "", "whist.trickEnd", nil
	case domain.WhistPhaseRoundEnd:
		return "", "whist.roundEnd", nil
	}
	return "", "", nil
}

// HintOutput ヒント情報をJSON出力する
func (p *WhistWebPresenter) HintOutput(w interfaces.WhistGame) string {
	hint := w.GetHint()
	resObj := p.buildBase(w)
	if hint != nil {
		resObj.Hint = &controller.WhistWebOutputHint{
			CardIndex: hint.CardIndex,
			Reason:    hint.Reason,
		}
	}
	// **「頼んだヒントか」を CLI が見分けられるようにする。**このゲーム群の
	// `hintAvailable` は画面のラベルとして既に使われているので、別キーを出す (#4483)。
	if hint != nil {
		resObj.MessageCode = "whist.hintRequested"
	} else {
		resObj.MessageCode = "whist.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *WhistWebPresenter) ActionLogOutput(w interfaces.WhistGame) string {
	return actionLogOutputJSON(w)
}
