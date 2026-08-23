//go:build !js || !wasm || classic

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// BrusquembilleWebPresenter ブリュスカンビーユWebプレゼンタークラス
type BrusquembilleWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *BrusquembilleWebPresenter) Output(b interfaces.BrusquembilleGame, lastErr error) string {
	resObj := p.buildBase(b)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(b, lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **フェーズと手番はここでは見ない。**Brusquembille.GetHint() が自分で
	// 「プレイ中かつ人間の手番」を確かめて nil を返す。ソリティア側のゲートを
	// 持ち込むと、ドメインが既に持つ判定を二重に書くことになる。
	if hint := b.GetHint(); hint != nil {
		resObj.Hint = &controller.BrusquembilleWebOutputHint{
			CardIndex: hint.CardIndex,
			Reason:    hint.Reason,
		}
	}

	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *BrusquembilleWebPresenter) buildBase(b interfaces.BrusquembilleGame) *controller.BrusquembilleWebOutput {
	resObj := new(controller.BrusquembilleWebOutput)
	resObj.Phase = int(b.GetPhase())
	resObj.TrickNumber = b.GetTrickNumber()
	resObj.CurrentPlayerIdx = b.GetCurrentPlayerIdx()
	resObj.TrumpSuit = b.GetTrumpSuit()
	if tc := b.GetTrumpCard(); tc != nil {
		resObj.TrumpCard = cardToOutput(tc)
	}
	resObj.DealerIdx = b.GetDealerIdx()
	resObj.LeadPlayerIdx = b.GetLeadPlayerIdx()
	resObj.StockRemaining = b.GetStockRemaining()
	resObj.GameEndFlag = b.GetGameEndFlag()
	resObj.WinnerIdx = b.GetWinnerIdx()
	// 人間 (席 0) の合法手と、いま追従義務があるか。
	resObj.ValidIndices = b.GetValidPlayIndices(0)
	if resObj.ValidIndices == nil {
		resObj.ValidIndices = []int{}
	}
	resObj.FollowRequired = b.IsFollowRequired()

	cfg := b.GetConfig()
	resObj.Config = controller.BrusquembilleWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		PlayerCnt:     cfg.PlayerCnt,
	}

	resObj.CurrentTrick = trickCardsToOutput(b.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(b)
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築
func (p *BrusquembilleWebPresenter) buildPlayersOutput(b interfaces.BrusquembilleGame) []*controller.BrusquembilleWebOutputPlayer {
	out := make([]*controller.BrusquembilleWebOutputPlayer, 0)
	for i := 0; i < b.GetPlayerCnt(); i++ {
		player := b.GetPlayer(i)
		out = append(out, &controller.BrusquembilleWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, player.GetIsHuman()),
			Points:     b.GetPlayerPoints(i),
			TrickCount: player.GetTrickCount(),
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *BrusquembilleWebPresenter) buildMessage(b interfaces.BrusquembilleGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if b.GetGameEndFlag() {
		p0 := b.GetPlayerPoints(0)
		p1 := b.GetPlayerPoints(1)
		params := map[string]string{
			"p0": fmt.Sprintf("%d", p0),
			"p1": fmt.Sprintf("%d", p1),
		}
		switch b.GetWinnerIdx() {
		case 0:
			return "", "brusquembille.result.p0Win", params
		case 1:
			return "", "brusquembille.result.p1Win", params
		default:
			return "", "brusquembille.result.tie", params
		}
	}
	switch b.GetPhase() {
	case domain.BrusquembillePhasePlay:
		if len(b.GetCurrentTrick()) == 0 {
			return "", "brusquembille.playPhase.lead", nil
		}
		return "", "brusquembille.playPhase.follow", nil
	case domain.BrusquembillePhaseTrickEnd:
		return "", "brusquembille.trickEnd", nil
	}
	return "", "", nil
}

// HintOutput ヒント情報をJSON出力する
func (p *BrusquembilleWebPresenter) HintOutput(b interfaces.BrusquembilleGame) string {
	hint := b.GetHint()
	resObj := p.buildBase(b)
	if hint != nil {
		resObj.Hint = &controller.BrusquembilleWebOutputHint{
			CardIndex: hint.CardIndex,
			Reason:    hint.Reason,
		}
	}
	// **「頼んだヒントか」をフロントが見分けられるようにする。**ページは
	// `isRequestedHint` でこのコードを見てからバナーを出すので (#4605)、
	// 付けないと押しても何も出ない。`hintAvailable` は画面のラベルとして
	// 既に使われているため別キーにする (#4483)。
	if hint != nil {
		resObj.MessageCode = "brusquembille.hintRequested"
	} else {
		resObj.MessageCode = "brusquembille.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *BrusquembilleWebPresenter) ActionLogOutput(b interfaces.BrusquembilleGame) string {
	return actionLogOutputJSON(b)
}
