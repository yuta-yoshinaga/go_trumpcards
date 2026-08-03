//go:build !js || !wasm || extra

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// CalabresellaWebPresenter カラブレセッラ (Calabresella) のWebプレゼンタークラス
type CalabresellaWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *CalabresellaWebPresenter) Output(g interfaces.CalabresellaGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **フェーズと手番はここでは見ない。**Calabresella.GetHint() が自分で
	// 「人間の手番で、かつ行動を選べる状態か」を確かめて nil を返す。
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.WebOutputCardHint{
			CardIndices: hint.CardIndices,
			Reason:      hint.Reason,
		}
	}

	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *CalabresellaWebPresenter) buildBase(g interfaces.CalabresellaGame) *controller.CalabresellaWebOutput {
	resObj := new(controller.CalabresellaWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.CurrentBidderIdx = g.GetCurrentBidderIdx()
	resObj.LeadPlayerIdx = g.GetLeadPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.ForehandIdx = g.GetForehandIdx()
	resObj.SoloistIdx = g.GetSoloistIdx()
	resObj.WinningBid = int(g.GetWinningBid())
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerPlayer = g.GetWinnerPlayer()
	resObj.PlayerScores = g.GetPlayerScores()
	resObj.RoundThirds = g.GetRoundThirds()
	resObj.LastTrickWinner = -1
	resObj.IsHumanTurn = g.IsHumanTurn()

	resObj.PlayableIndices = p.playableIndices(g)

	cfg := g.GetConfig()
	resObj.Config = controller.CalabresellaWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetPoints:  cfg.TargetPoints,
	}

	resObj.CurrentTrick = trickCardsToOutput(g.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(g)
	resObj.Monte = p.buildMonteOutput(g)
	return resObj
}

// buildMonteOutput 公開済みモンテ (widow) 4 枚を構築する。
// モンテはソリスト確定後の discard フェーズで取得された時点で全員へ公開されるため、
// bid フェーズ (取得前・次ラウンド開始直後) では表示しない。取得記録は棋譜の
// "monte_take" エントリに残っているので、その最新エントリのカードを返す。
func (p *CalabresellaWebPresenter) buildMonteOutput(g interfaces.CalabresellaGame) []*controller.WebOutputCard {
	if g.GetPhase() == domain.CalabresellaPhaseBid {
		return nil
	}
	log := g.GetActionLog()
	for i := len(log) - 1; i >= 0; i-- {
		entry := log[i]
		if entry == nil || entry.ActionType != "monte_take" {
			continue
		}
		out := make([]*controller.WebOutputCard, 0, len(entry.Cards))
		for _, c := range entry.Cards {
			out = append(out, cardToOutput(c))
		}
		return out
	}
	return nil
}

// playableIndices 人間プレイヤーがプレイできるカードのインデックスを返す
func (p *CalabresellaWebPresenter) playableIndices(g interfaces.CalabresellaGame) []int {
	if g.GetPhase() != domain.CalabresellaPhasePlay || !g.IsHumanTurn() {
		return make([]int, 0)
	}
	idx := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
	if idx == nil {
		return make([]int, 0)
	}
	return idx
}

// buildPlayersOutput プレイヤー情報を構築
func (p *CalabresellaWebPresenter) buildPlayersOutput(g interfaces.CalabresellaGame) []*controller.CalabresellaWebOutputPlayer {
	scores := g.GetPlayerScores()
	thirds := g.GetRoundThirds()
	soloist := g.GetSoloistIdx()
	out := make([]*controller.CalabresellaWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		out = append(out, &controller.CalabresellaWebOutputPlayer{
			ID:          i,
			IsHuman:     player.GetIsHuman(),
			CardCount:   player.GetCardsSize(),
			Cards:       playerCardsToOutput(player, player.GetIsHuman()),
			TrickCount:  player.GetTrickCount(),
			Score:       scores[i],
			IsSoloist:   i == soloist,
			RoundThirds: thirds[i],
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *CalabresellaWebPresenter) buildMessage(g interfaces.CalabresellaGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		return p.winnerMessage(g)
	}
	switch g.GetPhase() {
	case domain.CalabresellaPhaseBid:
		return "", "calabresella.bidPhase", nil
	case domain.CalabresellaPhaseDiscard:
		return "", "calabresella.discardPhase", nil
	case domain.CalabresellaPhasePlay:
		if len(g.GetCurrentTrick()) == 0 {
			return "", "calabresella.playPhase.lead", nil
		}
		return "", "calabresella.playPhase.follow", nil
	case domain.CalabresellaPhaseTrickEnd:
		return "", "calabresella.trickEnd", nil
	case domain.CalabresellaPhaseRoundEnd:
		return "", "calabresella.roundEnd", nil
	}
	return "", "", nil
}

// winnerMessage 勝者プレイヤーメッセージを構築する
func (p *CalabresellaWebPresenter) winnerMessage(g interfaces.CalabresellaGame) (string, string, map[string]string) {
	winner := g.GetWinnerPlayer()
	humanIdx := -1
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if player := g.GetPlayer(i); player != nil && player.GetIsHuman() {
			humanIdx = i
			break
		}
	}
	if humanIdx >= 0 && winner == humanIdx {
		return "", "calabresella.result.humanWin", nil
	}
	params := map[string]string{"player": fmt.Sprintf("%d", winner)}
	return "", "calabresella.result.cpuWin", params
}

// HintOutput ヒント情報をJSON出力する
func (p *CalabresellaWebPresenter) HintOutput(g interfaces.CalabresellaGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = &controller.WebOutputCardHint{
			CardIndices: hint.CardIndices,
			Reason:      hint.Reason,
		}
	}
	// **「頼んだヒントか」を CLI が見分けられるようにする。**このゲーム群の
	// `hintAvailable` は画面のラベルとして既に使われているので、別キーを出す (#4483)。
	if hint != nil {
		resObj.MessageCode = "calabresella.hintRequested"
	} else {
		resObj.MessageCode = "calabresella.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *CalabresellaWebPresenter) ActionLogOutput(g interfaces.CalabresellaGame) string {
	return actionLogOutputJSON(g)
}
