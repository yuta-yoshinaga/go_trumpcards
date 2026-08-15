//go:build !js || !wasm || extra3

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// OmbreWebPresenter オンブル (Ombre) のWebプレゼンタークラス
type OmbreWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *OmbreWebPresenter) Output(g interfaces.OmbreGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **フェーズと手番はここでは見ない。**Ombre.GetHint() が自分で
	// 「人間の手番で、かつ行動を選べる状態か」を確かめて nil を返す。
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = cardHint(hint.CardIndices, hint.Reason)
	}

	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *OmbreWebPresenter) buildBase(g interfaces.OmbreGame) *controller.OmbreWebOutput {
	resObj := new(controller.OmbreWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.CurrentBidderIdx = g.GetCurrentBidderIdx()
	resObj.LeadPlayerIdx = g.GetLeadPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.ForehandIdx = g.GetForehandIdx()
	resObj.OmbreIdx = g.GetOmbreIdx()
	resObj.WinningBid = int(g.GetWinningBid())
	resObj.TrumpSuit = g.GetTrumpSuit()
	resObj.Outcome = int(g.GetOutcome())
	resObj.Result = int(g.GetResult())
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerPlayer = g.GetWinnerPlayer()
	resObj.PlayerScores = g.GetPlayerScores()
	resObj.LastTrickWinner = -1
	resObj.IsHumanTurn = g.IsHumanTurn()
	resObj.IsHumanBidTurn = g.IsHumanBidTurn()

	resObj.PlayableIndices = p.playableIndices(g)

	cfg := g.GetConfig()
	resObj.Config = controller.OmbreWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetRounds:  cfg.TargetRounds,
	}

	resObj.CurrentTrick = trickCardsToOutput(g.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// playableIndices 人間プレイヤーがプレイできるカードのインデックスを返す
func (p *OmbreWebPresenter) playableIndices(g interfaces.OmbreGame) []int {
	if g.GetPhase() != domain.OmbrePhasePlay || !g.IsHumanTurn() {
		return make([]int, 0)
	}
	idx := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
	if idx == nil {
		return make([]int, 0)
	}
	return idx
}

// buildPlayersOutput プレイヤー情報を構築
func (p *OmbreWebPresenter) buildPlayersOutput(g interfaces.OmbreGame) []*controller.OmbreWebOutputPlayer {
	scores := g.GetPlayerScores()
	ombre := g.GetOmbreIdx()
	out := make([]*controller.OmbreWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		out = append(out, &controller.OmbreWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, player.GetIsHuman()),
			TrickCount: player.GetTrickCount(),
			Score:      scores[i],
			IsOmbre:    i == ombre,
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *OmbreWebPresenter) buildMessage(g interfaces.OmbreGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		return p.winnerMessage(g)
	}
	switch g.GetPhase() {
	case domain.OmbrePhaseBid:
		return "", "ombre.bidPhase", nil
	case domain.OmbrePhasePlay:
		if len(g.GetCurrentTrick()) == 0 {
			return "", "ombre.playPhase.lead", nil
		}
		return "", "ombre.playPhase.follow", nil
	case domain.OmbrePhaseTrickEnd:
		return "", "ombre.trickEnd", nil
	case domain.OmbrePhaseRoundEnd:
		return "", ombreOutcomeMessageCode(g.GetOutcome()), nil
	}
	return "", "", nil
}

// ombreOutcomeMessageCode ディール結果に対応するメッセージコードを返す。
func ombreOutcomeMessageCode(o domain.OmbreOutcome) string {
	switch o {
	case domain.OmbreOutcomeSacar:
		return "ombre.roundEnd.sacar"
	case domain.OmbreOutcomePuesta:
		return "ombre.roundEnd.puesta"
	case domain.OmbreOutcomeCodille:
		return "ombre.roundEnd.codille"
	default:
		return "ombre.roundEnd"
	}
}

// winnerMessage 勝者プレイヤーメッセージを構築する
func (p *OmbreWebPresenter) winnerMessage(g interfaces.OmbreGame) (string, string, map[string]string) {
	winner := g.GetWinnerPlayer()
	humanIdx := -1
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if player := g.GetPlayer(i); player != nil && player.GetIsHuman() {
			humanIdx = i
			break
		}
	}
	if humanIdx >= 0 && winner == humanIdx {
		return "", "ombre.result.humanWin", nil
	}
	params := map[string]string{"player": fmt.Sprintf("%d", winner)}
	return "", "ombre.result.cpuWin", params
}

// HintOutput ヒント情報をJSON出力する
func (p *OmbreWebPresenter) HintOutput(g interfaces.OmbreGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = cardHint(hint.CardIndices, hint.Reason)
	}
	// **「頼んだヒントか」を CLI が見分けられるようにする。**このゲーム群の
	// `hintAvailable` は画面のラベルとして既に使われているので、別キーを出す (#4483)。
	if hint != nil {
		resObj.MessageCode = "ombre.hintRequested"
	} else {
		resObj.MessageCode = "ombre.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *OmbreWebPresenter) ActionLogOutput(g interfaces.OmbreGame) string {
	return actionLogOutputJSON(g)
}
