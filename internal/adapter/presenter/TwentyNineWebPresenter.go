//go:build !js || !wasm || casino

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// TwentyNineWebPresenter トゥエンティナイン (29) のWebプレゼンタークラス
type TwentyNineWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *TwentyNineWebPresenter) Output(g interfaces.TwentyNineGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **フェーズと手番はここでは見ない。**TwentyNine.GetHint() が自分で
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
func (p *TwentyNineWebPresenter) buildBase(g interfaces.TwentyNineGame) *controller.TwentyNineWebOutput {
	resObj := new(controller.TwentyNineWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.LeadPlayerIdx = g.GetLeadPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.DeclarerIdx = g.GetDeclarerIdx()
	resObj.Contract = int(g.GetContract())
	resObj.TrumpSuit = g.GetTrumpSuit()
	resObj.TrumpRevealed = g.GetTrumpRevealed()
	resObj.Bids = p.bidsOutput(g)
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerTeam = g.GetWinnerTeam()
	resObj.TeamScores = g.GetTeamScores()
	resObj.RoundTeamPoints = g.GetRoundTeamPoints()
	resObj.IsHumanTurn = g.IsHumanTurn()
	resObj.IsHumanBidTurn = g.IsHumanBidTurn()

	resObj.PlayableIndices = p.playableIndices(g)

	cfg := g.GetConfig()
	resObj.Config = controller.TwentyNineWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetPoints:  cfg.TargetPoints,
	}

	resObj.CurrentTrick = trickCardsToOutput(g.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// bidsOutput 各プレイヤーの入札を int 配列に変換する
func (p *TwentyNineWebPresenter) bidsOutput(g interfaces.TwentyNineGame) [domain.TwentyNinePlayerCnt]int {
	bids := g.GetBids()
	var out [domain.TwentyNinePlayerCnt]int
	for i := range bids {
		out[i] = int(bids[i])
	}
	return out
}

// playableIndices 人間プレイヤーがプレイできるカードのインデックスを返す
func (p *TwentyNineWebPresenter) playableIndices(g interfaces.TwentyNineGame) []int {
	if g.GetPhase() != domain.TwentyNinePhasePlay || !g.IsHumanTurn() {
		return make([]int, 0)
	}
	idx := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
	if idx == nil {
		return make([]int, 0)
	}
	return idx
}

// buildPlayersOutput プレイヤー情報を構築
func (p *TwentyNineWebPresenter) buildPlayersOutput(g interfaces.TwentyNineGame) []*controller.TwentyNineWebOutputPlayer {
	teamScores := g.GetTeamScores()
	declarer := g.GetDeclarerIdx()
	out := make([]*controller.TwentyNineWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		out = append(out, &controller.TwentyNineWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, player.GetIsHuman()),
			TrickCount: player.GetTrickCount(),
			TeamScore:  teamScores[domain.TwentyNineTeamOf(i)],
			IsDeclarer: i == declarer,
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *TwentyNineWebPresenter) buildMessage(g interfaces.TwentyNineGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		return p.winnerMessage(g)
	}
	switch g.GetPhase() {
	case domain.TwentyNinePhaseBid:
		return "", "twentynine.bidPhase", nil
	case domain.TwentyNinePhasePlay:
		if len(g.GetCurrentTrick()) == 0 {
			return "", "twentynine.playPhase.lead", nil
		}
		return "", "twentynine.playPhase.follow", nil
	case domain.TwentyNinePhaseTrickEnd:
		return "", "twentynine.trickEnd", nil
	case domain.TwentyNinePhaseRoundEnd:
		return "", "twentynine.roundEnd", nil
	}
	return "", "", nil
}

// winnerMessage 勝者チームメッセージを構築する
func (p *TwentyNineWebPresenter) winnerMessage(g interfaces.TwentyNineGame) (string, string, map[string]string) {
	winnerTeam := g.GetWinnerTeam()
	humanTeam := -1
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if player := g.GetPlayer(i); player != nil && player.GetIsHuman() {
			humanTeam = domain.TwentyNineTeamOf(i)
			break
		}
	}
	if humanTeam >= 0 && winnerTeam == humanTeam {
		return "ゲーム終了！ あなたのチームの勝ち！", "twentynine.result.humanWin", nil
	}
	teamName := twentyNineTeamLabel(winnerTeam)
	params := map[string]string{"team": teamName}
	return fmt.Sprintf("ゲーム終了！ チーム%sの勝ち！", teamName), "twentynine.result.cpuWin", params
}

// HintOutput ヒント情報をJSON出力する
func (p *TwentyNineWebPresenter) HintOutput(g interfaces.TwentyNineGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = &controller.WebOutputCardHint{
			CardIndices: hint.CardIndices,
			Reason:      hint.Reason,
		}
	}
	// **「頼んだヒントか」をフロントが見分けられるようにする。**ページは
	// `isRequestedHint` でこのコードを見てからバナーを出すので (#4605)、
	// 付けないと押しても何も出ない。`hintAvailable` は画面のラベルとして
	// 既に使われているため別キーにする (#4483)。
	if hint != nil {
		resObj.MessageCode = "twentynine.hintRequested"
	} else {
		resObj.MessageCode = "twentynine.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *TwentyNineWebPresenter) ActionLogOutput(g interfaces.TwentyNineGame) string {
	return actionLogOutputJSON(g)
}
