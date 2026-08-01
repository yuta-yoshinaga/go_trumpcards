//go:build !js || !wasm || classic

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// BeziqueWebPresenter ベジークWebプレゼンタークラス
type BeziqueWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *BeziqueWebPresenter) Output(b interfaces.BeziqueGame, lastErr error) string {
	resObj := p.buildBase(b)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(b, lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **フェーズと手番はここでは見ない。**Bezique.GetHint() が自分で
	// 「人間の手番で、かつ行動を選べる状態か」を確かめて nil を返す。
	if hint := b.GetHint(); hint != nil {
		resObj.Hint = &controller.BeziqueWebOutputHint{
			CardIndex: hint.CardIndex,
			MeldIndex: hint.MeldIndex,
			Reason:    hint.Reason,
		}
	}

	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *BeziqueWebPresenter) buildBase(b interfaces.BeziqueGame) *controller.BeziqueWebOutput {
	resObj := new(controller.BeziqueWebOutput)
	resObj.Phase = int(b.GetPhase())
	resObj.RoundNumber = b.GetRoundNumber()
	resObj.TrickNumber = b.GetTrickNumber()
	resObj.CurrentPlayerIdx = b.GetCurrentPlayerIdx()
	resObj.LeadPlayerIdx = b.GetLeadPlayerIdx()
	resObj.DealerIdx = b.GetDealerIdx()
	resObj.TrumpSuit = b.GetTrumpSuit()
	if tc := b.GetTrumpCard(); tc != nil {
		resObj.TrumpCard = cardToOutput(tc)
	}
	resObj.StockRemaining = b.GetStockRemaining()
	resObj.IsEndgame = b.IsEndgame()
	resObj.GameEndFlag = b.GetGameEndFlag()
	resObj.WinnerIdx = b.GetWinnerIdx()

	cfg := b.GetConfig()
	resObj.Config = controller.BeziqueWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetScore:   cfg.TargetScore,
	}

	cnt := b.GetPlayerCnt()
	resObj.DealPoints = make([]int, cnt)
	resObj.DealMeldPoints = make([]int, cnt)
	resObj.MatchScore = make([]int, cnt)
	for i := 0; i < cnt; i++ {
		resObj.DealPoints[i] = b.GetDealPoints(i)
		resObj.DealMeldPoints[i] = b.GetDealMeldPoints(i)
		resObj.MatchScore[i] = b.GetMatchScore(i)
	}

	resObj.CurrentTrick = trickCardsToOutput(b.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(b)
	resObj.AvailableMelds = p.buildMeldsOutput(b)
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築
func (p *BeziqueWebPresenter) buildPlayersOutput(b interfaces.BeziqueGame) []*controller.BeziqueWebOutputPlayer {
	out := make([]*controller.BeziqueWebOutputPlayer, 0)
	for i := 0; i < b.GetPlayerCnt(); i++ {
		player := b.GetPlayer(i)
		out = append(out, &controller.BeziqueWebOutputPlayer{
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

// buildMeldsOutput 人間が役宣言フェーズで宣言できる役一覧を構築する (それ以外は空)。
func (p *BeziqueWebPresenter) buildMeldsOutput(b interfaces.BeziqueGame) []*controller.BeziqueWebOutputMeld {
	out := make([]*controller.BeziqueWebOutputMeld, 0)
	if b.GetPhase() != domain.BeziquePhaseMeld || b.GetCurrentPlayerIdx() != 0 {
		return out
	}
	for _, m := range b.GetAvailableMelds(0) {
		out = append(out, &controller.BeziqueWebOutputMeld{
			Type:   int(m.Type),
			Suit:   m.Suit,
			Points: m.Points,
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *BeziqueWebPresenter) buildMessage(b interfaces.BeziqueGame, lastErr error) (string, string, map[string]string) {
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
			return fmt.Sprintf("ゲーム終了！ あなたの勝利です (%d-%d)！", m0, m1), "bezique.result.p0Win", params
		case 1:
			return fmt.Sprintf("ゲーム終了！ CPUの勝利です (%d-%d)。", m0, m1), "bezique.result.p1Win", params
		default:
			return fmt.Sprintf("ゲーム終了！ 引き分けです (%d-%d)。", m0, m1), "bezique.result.tie", params
		}
	}
	switch b.GetPhase() {
	case domain.BeziquePhasePlay:
		if len(b.GetCurrentTrick()) == 0 {
			return "", "bezique.playPhase.lead", nil
		}
		return "", "bezique.playPhase.follow", nil
	case domain.BeziquePhaseMeld:
		return "", "bezique.meldPhase", nil
	case domain.BeziquePhaseRoundEnd:
		return "", "bezique.roundEnd", nil
	}
	return "", "", nil
}

// HintOutput ヒント情報をJSON出力する
func (p *BeziqueWebPresenter) HintOutput(b interfaces.BeziqueGame) string {
	hint := b.GetHint()
	resObj := p.buildBase(b)
	if hint != nil {
		resObj.Hint = &controller.BeziqueWebOutputHint{
			CardIndex: hint.CardIndex,
			MeldIndex: hint.MeldIndex,
			Reason:    hint.Reason,
		}
	}
	// **「頼んだヒントか」を CLI が見分けられるようにする。**このゲーム群の
	// `hintAvailable` は画面のラベルとして既に使われているので、別キーを出す (#4483)。
	if hint != nil {
		resObj.MessageCode = "bezique.hintRequested"
	} else {
		resObj.MessageCode = "bezique.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *BeziqueWebPresenter) ActionLogOutput(b interfaces.BeziqueGame) string {
	return actionLogOutputJSON(b)
}
