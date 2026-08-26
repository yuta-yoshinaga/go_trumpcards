//go:build !js || !wasm || extra4

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// QuadrilleWebPresenter カドリール (Quadrille) のWebプレゼンタークラス
type QuadrilleWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *QuadrilleWebPresenter) Output(g interfaces.QuadrilleGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **フェーズと手番はここでは見ない。**Quadrille.GetHint() が自分で
	// 「人間の手番で、かつ行動を選べる状態か」を確かめて nil を返す。
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = cardHint(hint.CardIndices, hint.Reason)
	}

	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *QuadrilleWebPresenter) buildBase(g interfaces.QuadrilleGame) *controller.QuadrilleWebOutput {
	resObj := new(controller.QuadrilleWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.CurrentBidderIdx = g.GetCurrentBidderIdx()
	resObj.LeadPlayerIdx = g.GetLeadPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.ForehandIdx = g.GetForehandIdx()
	resObj.QuadrilleIdx = g.GetQuadrilleIdx()
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
	resObj.IsHumanKingCallTurn = g.IsHumanKingCallTurn()
	resObj.CalledKingSuit = g.GetCalledKingSuit()
	resObj.CallableKingSuits = g.GetCallableKingSuits()
	if resObj.CallableKingSuits == nil {
		resObj.CallableKingSuits = make([]int, 0)
	}
	// **伏せたままの相方は -1 で出る。** GetPartnerIdx がそれを担保する。
	resObj.PartnerIdx = g.GetPartnerIdx()
	resObj.RoiSeul = g.IsRoiSeul()

	resObj.PlayableIndices = p.playableIndices(g)

	cfg := g.GetConfig()
	resObj.Config = controller.QuadrilleWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetRounds:  cfg.TargetRounds,
	}

	resObj.CurrentTrick = trickCardsToOutput(g.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// playableIndices 人間プレイヤーがプレイできるカードのインデックスを返す
func (p *QuadrilleWebPresenter) playableIndices(g interfaces.QuadrilleGame) []int {
	if g.GetPhase() != domain.QuadrillePhasePlay || !g.IsHumanTurn() {
		return make([]int, 0)
	}
	idx := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
	if idx == nil {
		return make([]int, 0)
	}
	return idx
}

// buildPlayersOutput プレイヤー情報を構築
func (p *QuadrilleWebPresenter) buildPlayersOutput(g interfaces.QuadrilleGame) []*controller.QuadrilleWebOutputPlayer {
	scores := g.GetPlayerScores()
	quadrille := g.GetQuadrilleIdx()
	out := make([]*controller.QuadrilleWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		out = append(out, &controller.QuadrilleWebOutputPlayer{
			ID:          i,
			IsHuman:     player.GetIsHuman(),
			CardCount:   player.GetCardsSize(),
			Cards:       playerCardsToOutput(player, player.GetIsHuman()),
			TrickCount:  player.GetTrickCount(),
			Score:       scores[i],
			IsQuadrille: i == quadrille,
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *QuadrilleWebPresenter) buildMessage(g interfaces.QuadrilleGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		return p.winnerMessage(g)
	}
	switch g.GetPhase() {
	case domain.QuadrillePhaseBid:
		return "", "quadrille.bidPhase", nil
	case domain.QuadrillePhasePlay:
		if len(g.GetCurrentTrick()) == 0 {
			return "", "quadrille.playPhase.lead", nil
		}
		return "", "quadrille.playPhase.follow", nil
	case domain.QuadrillePhaseTrickEnd:
		return "", "quadrille.trickEnd", nil
	case domain.QuadrillePhaseRoundEnd:
		return "", quadrilleOutcomeMessageCode(g.GetOutcome()), nil
	}
	return "", "", nil
}

// quadrilleOutcomeMessageCode ディール結果に対応するメッセージコードを返す。
func quadrilleOutcomeMessageCode(o domain.QuadrilleOutcome) string {
	switch o {
	case domain.QuadrilleOutcomeSacar:
		return "quadrille.roundEnd.sacar"
	case domain.QuadrilleOutcomePuesta:
		return "quadrille.roundEnd.puesta"
	case domain.QuadrilleOutcomeCodille:
		return "quadrille.roundEnd.codille"
	default:
		return "quadrille.roundEnd"
	}
}

// winnerMessage 勝者プレイヤーメッセージを構築する
func (p *QuadrilleWebPresenter) winnerMessage(g interfaces.QuadrilleGame) (string, string, map[string]string) {
	winner := g.GetWinnerPlayer()
	humanIdx := -1
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if player := g.GetPlayer(i); player != nil && player.GetIsHuman() {
			humanIdx = i
			break
		}
	}
	if humanIdx >= 0 && winner == humanIdx {
		return "", "quadrille.result.humanWin", nil
	}
	params := map[string]string{"player": fmt.Sprintf("%d", winner)}
	return "", "quadrille.result.cpuWin", params
}

// HintOutput ヒント情報をJSON出力する
func (p *QuadrilleWebPresenter) HintOutput(g interfaces.QuadrilleGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = cardHint(hint.CardIndices, hint.Reason)
	}
	// **「頼んだヒントか」を CLI が見分けられるようにする。**このゲーム群の
	// `hintAvailable` は画面のラベルとして既に使われているので、別キーを出す (#4483)。
	if hint != nil {
		resObj.MessageCode = "quadrille.hintRequested"
	} else {
		resObj.MessageCode = "quadrille.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *QuadrilleWebPresenter) ActionLogOutput(g interfaces.QuadrilleGame) string {
	return actionLogOutputJSON(g)
}
