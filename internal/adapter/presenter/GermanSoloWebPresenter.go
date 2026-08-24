//go:build !js || !wasm || classic

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// GermanSoloWebPresenter ジャーマン・ソロ (GermanSolo) のWebプレゼンタークラス
type GermanSoloWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *GermanSoloWebPresenter) Output(g interfaces.GermanSoloGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **フェーズと手番はここでは見ない。**GermanSolo.GetHint() が自分で
	// 「人間の手番で、かつ行動を選べる状態か」を確かめて nil を返す。
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = cardHint(hint.CardIndices, hint.Reason)
	}

	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *GermanSoloWebPresenter) buildBase(g interfaces.GermanSoloGame) *controller.GermanSoloWebOutput {
	resObj := new(controller.GermanSoloWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.CurrentBidderIdx = g.GetCurrentBidderIdx()
	resObj.LeadPlayerIdx = g.GetLeadPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.ForehandIdx = g.GetForehandIdx()
	resObj.DeclarerIdx = g.GetDeclarerIdx()
	resObj.WinningBid = int(g.GetWinningBid())
	resObj.HighestBid = int(g.GetHighestBid())
	resObj.BiddableBids = g.GetBiddableBids()
	if resObj.BiddableBids == nil {
		resObj.BiddableBids = make([]int, 0)
	}
	resObj.RequiredTricks = g.RequiredTricks()
	resObj.DeclarerTricks, resObj.DefenderTricks = g.GetSideTrickCounts()
	resObj.TrumpSuit = g.GetTrumpSuit()
	resObj.Outcome = int(g.GetOutcome())
	resObj.Result = int(g.GetResult())
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerPlayer = g.GetWinnerPlayer()
	resObj.PlayerScores = g.GetPlayerScores()
	resObj.LastTrickWinner = -1
	resObj.IsHumanTurn = g.IsHumanTurn()
	resObj.IsHumanBidTurn = g.IsHumanBidTurn()
	resObj.IsHumanAceCallTurn = g.IsHumanAceCallTurn()
	resObj.CalledAceSuit = g.GetCalledAceSuit()
	resObj.CallableAceSuits = g.GetCallableAceSuits()
	if resObj.CallableAceSuits == nil {
		resObj.CallableAceSuits = make([]int, 0)
	}
	// **伏せたままの相方は -1 で出る。** GetPartnerIdx がそれを担保する。
	resObj.PartnerIdx = g.GetPartnerIdx()
	resObj.PlaysAlone = g.IsPlayingAlone()

	resObj.PlayableIndices = p.playableIndices(g)

	cfg := g.GetConfig()
	resObj.Config = controller.GermanSoloWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetRounds:  cfg.TargetRounds,
	}

	resObj.CurrentTrick = trickCardsToOutput(g.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// playableIndices 人間プレイヤーがプレイできるカードのインデックスを返す
func (p *GermanSoloWebPresenter) playableIndices(g interfaces.GermanSoloGame) []int {
	if g.GetPhase() != domain.GermanSoloPhasePlay || !g.IsHumanTurn() {
		return make([]int, 0)
	}
	idx := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
	if idx == nil {
		return make([]int, 0)
	}
	return idx
}

// buildPlayersOutput プレイヤー情報を構築
func (p *GermanSoloWebPresenter) buildPlayersOutput(g interfaces.GermanSoloGame) []*controller.GermanSoloWebOutputPlayer {
	scores := g.GetPlayerScores()
	declarer := g.GetDeclarerIdx()
	out := make([]*controller.GermanSoloWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		out = append(out, &controller.GermanSoloWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, player.GetIsHuman()),
			TrickCount: player.GetTrickCount(),
			Score:      scores[i],
			IsDeclarer: i == declarer,
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *GermanSoloWebPresenter) buildMessage(g interfaces.GermanSoloGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		return p.winnerMessage(g)
	}
	switch g.GetPhase() {
	case domain.GermanSoloPhaseBid:
		return "", "germansolo.bidPhase", nil
	case domain.GermanSoloPhasePlay:
		if len(g.GetCurrentTrick()) == 0 {
			return "", "germansolo.playPhase.lead", nil
		}
		return "", "germansolo.playPhase.follow", nil
	case domain.GermanSoloPhaseTrickEnd:
		return "", "germansolo.trickEnd", nil
	case domain.GermanSoloPhaseRoundEnd:
		return "", germanSoloOutcomeMessageCode(g.GetOutcome()), nil
	}
	return "", "", nil
}

// germanSoloOutcomeMessageCode ディール結果に対応するメッセージコードを返す。
func germanSoloOutcomeMessageCode(o domain.GermanSoloOutcome) string {
	switch o {
	case domain.GermanSoloOutcomeMade:
		return "germansolo.roundEnd.made"
	case domain.GermanSoloOutcomeFailed:
		return "germansolo.roundEnd.failed"
	default:
		return "germansolo.roundEnd"
	}
}

// winnerMessage 勝者プレイヤーメッセージを構築する
func (p *GermanSoloWebPresenter) winnerMessage(g interfaces.GermanSoloGame) (string, string, map[string]string) {
	winner := g.GetWinnerPlayer()
	humanIdx := -1
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if player := g.GetPlayer(i); player != nil && player.GetIsHuman() {
			humanIdx = i
			break
		}
	}
	if humanIdx >= 0 && winner == humanIdx {
		return "", "germansolo.result.humanWin", nil
	}
	params := map[string]string{"player": fmt.Sprintf("%d", winner)}
	return "", "germansolo.result.cpuWin", params
}

// HintOutput ヒント情報をJSON出力する
func (p *GermanSoloWebPresenter) HintOutput(g interfaces.GermanSoloGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = cardHint(hint.CardIndices, hint.Reason)
	}
	// **「頼んだヒントか」を CLI が見分けられるようにする。**このゲーム群の
	// `hintAvailable` は画面のラベルとして既に使われているので、別キーを出す (#4483)。
	if hint != nil {
		resObj.MessageCode = "germansolo.hintRequested"
	} else {
		resObj.MessageCode = "germansolo.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *GermanSoloWebPresenter) ActionLogOutput(g interfaces.GermanSoloGame) string {
	return actionLogOutputJSON(g)
}
