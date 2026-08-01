//go:build !js || !wasm || casino

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// EcarteWebPresenter エカルテWebプレゼンタークラス
type EcarteWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *EcarteWebPresenter) Output(b interfaces.EcarteGame, lastErr error) string {
	resObj := p.buildBase(b)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(b, lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **フェーズと手番はここでは見ない。**Ecarte.GetHint() が自分で
	// 「人間の手番で、かつ行動を選べる状態か」を確かめて nil を返す。
	if hint := b.GetHint(); hint != nil {
		resObj.Hint = &controller.EcarteWebOutputHint{
			CardIndex: hint.CardIndex,
			Action:    hint.Action,
			Reason:    hint.Reason,
		}
	}

	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *EcarteWebPresenter) buildBase(b interfaces.EcarteGame) *controller.EcarteWebOutput {
	resObj := new(controller.EcarteWebOutput)
	resObj.Phase = int(b.GetPhase())
	resObj.NegStep = int(b.GetNegStep())
	resObj.RoundNumber = b.GetRoundNumber()
	resObj.TrickNumber = b.GetTrickNumber()
	resObj.CurrentPlayerIdx = b.GetCurrentPlayerIdx()
	resObj.DealerIdx = b.GetDealerIdx()
	resObj.ElderIdx = b.GetElderIdx()
	resObj.LeadPlayerIdx = b.GetLeadPlayerIdx()
	resObj.TrumpSuit = b.GetTrumpSuit()
	if tc := b.GetTrumpCard(); tc != nil {
		resObj.TrumpCard = cardToOutput(tc)
	}
	resObj.StockRemaining = b.GetStockRemaining()
	resObj.RefusalByDealer = b.IsRefusalByDealer()
	resObj.GameEndFlag = b.GetGameEndFlag()
	resObj.WinnerIdx = b.GetWinnerIdx()

	cfg := b.GetConfig()
	resObj.Config = controller.EcarteWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetScore:   cfg.TargetScore,
	}

	cnt := b.GetPlayerCnt()
	resObj.DealPoints = make([]int, cnt)
	resObj.MatchScore = make([]int, cnt)
	for i := 0; i < cnt; i++ {
		resObj.DealPoints[i] = b.GetDealPoints(i)
		resObj.MatchScore[i] = b.GetMatchScore(i)
	}

	resObj.CurrentTrick = trickCardsToOutput(b.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(b)
	resObj.ValidPlays = p.buildValidPlays(b)
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築
func (p *EcarteWebPresenter) buildPlayersOutput(b interfaces.EcarteGame) []*controller.EcarteWebOutputPlayer {
	out := make([]*controller.EcarteWebOutputPlayer, 0)
	for i := 0; i < b.GetPlayerCnt(); i++ {
		player := b.GetPlayer(i)
		out = append(out, &controller.EcarteWebOutputPlayer{
			ID:              i,
			IsHuman:         player.GetIsHuman(),
			CardCount:       player.GetCardsSize(),
			Cards:           playerCardsToOutput(player, player.GetIsHuman()),
			RoundScore:      player.GetRoundScore(),
			CumulativeScore: player.GetCumulativeScore(),
			TrickCount:      player.GetTrickCount(),
		})
	}
	return out
}

// buildValidPlays 人間がプレイフェーズで出せるカードのインデックス一覧を構築する (それ以外は空)。
func (p *EcarteWebPresenter) buildValidPlays(b interfaces.EcarteGame) []int {
	if b.GetPhase() != domain.EcartePhasePlay || b.GetCurrentPlayerIdx() != 0 {
		return make([]int, 0)
	}
	if v := b.GetValidPlayIndices(0); v != nil {
		return v
	}
	return make([]int, 0)
}

// buildMessage ゲーム結果メッセージを構築
func (p *EcarteWebPresenter) buildMessage(b interfaces.EcarteGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if b.GetGameEndFlag() {
		m0 := b.GetMatchScore(0)
		m1 := b.GetMatchScore(1)
		params := map[string]string{
			"p0": fmt.Sprintf("%d", m0),
			"p1": fmt.Sprintf("%d", m1),
		}
		switch b.GetWinnerIdx() {
		case 0:
			return fmt.Sprintf("ゲーム終了！ あなたの勝利です (%d-%d)！", m0, m1), "ecarte.result.p0Win", params
		case 1:
			return fmt.Sprintf("ゲーム終了！ CPUの勝利です (%d-%d)。", m0, m1), "ecarte.result.p1Win", params
		default:
			return fmt.Sprintf("ゲーム終了！ 引き分けです (%d-%d)。", m0, m1), "ecarte.result.tie", params
		}
	}
	switch b.GetPhase() {
	case domain.EcartePhaseExchange:
		return "", ecarteNegStepCode(b.GetNegStep()), nil
	case domain.EcartePhasePlay:
		if len(b.GetCurrentTrick()) == 0 {
			return "", "ecarte.playPhase.lead", nil
		}
		return "", "ecarte.playPhase.follow", nil
	case domain.EcartePhaseRoundEnd:
		return "", "ecarte.roundEnd", nil
	}
	return "", "", nil
}

// ecarteNegStepCode 交換ステップに対応するメッセージコードを返す。
func ecarteNegStepCode(step domain.EcarteNegStep) string {
	switch step {
	case domain.EcarteNegElderDecide:
		return "ecarte.exchange.elderDecide"
	case domain.EcarteNegDealerRespond:
		return "ecarte.exchange.dealerRespond"
	case domain.EcarteNegElderDiscard:
		return "ecarte.exchange.elderDiscard"
	default:
		return "ecarte.exchange.dealerDiscard"
	}
}

// HintOutput ヒント情報をJSON出力する
func (p *EcarteWebPresenter) HintOutput(b interfaces.EcarteGame) string {
	hint := b.GetHint()
	resObj := p.buildBase(b)
	if hint != nil {
		resObj.Hint = &controller.EcarteWebOutputHint{
			CardIndex: hint.CardIndex,
			Action:    hint.Action,
			Reason:    hint.Reason,
		}
	}
	// **「頼んだヒントか」を CLI が見分けられるようにする。**このゲーム群の
	// `hintAvailable` は画面のラベルとして既に使われているので、別キーを出す (#4483)。
	if hint != nil {
		resObj.MessageCode = "ecarte.hintRequested"
	} else {
		resObj.MessageCode = "ecarte.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *EcarteWebPresenter) ActionLogOutput(b interfaces.EcarteGame) string {
	return actionLogOutputJSON(b)
}
