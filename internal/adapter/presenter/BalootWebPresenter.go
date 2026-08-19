//go:build !js || !wasm || extra2

package presenter

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// BalootWebPresenter バルートWebプレゼンタークラス
type BalootWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *BalootWebPresenter) Output(b interfaces.BalootGame, lastErr error) string {
	resObj := p.buildBase(b)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(b, lastErr)
	// 受動ヒントは Output() でも埋める (#4483)。
	if hint := b.GetHint(); hint != nil {
		resObj.Hint = &controller.BalootWebOutputHint{CardIndex: hint.CardIndex, Reason: hint.Reason, Suit: hint.Suit}
	}
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *BalootWebPresenter) buildBase(b interfaces.BalootGame) *controller.BalootWebOutput {
	resObj := new(controller.BalootWebOutput)
	resObj.Phase = int(b.GetPhase())
	resObj.Mode = int(b.GetMode())
	resObj.RoundNumber = b.GetRoundNumber()
	resObj.TrickNumber = b.GetTrickNumber()
	resObj.TrumpSuit = b.GetTrumpSuit()
	resObj.DeclarerIdx = b.GetDeclarerIdx()
	resObj.CurrentPlayerIdx = b.GetCurrentPlayerIdx()
	resObj.LeadPlayerIdx = b.GetLeadPlayerIdx()
	resObj.DealerIdx = b.GetDealerIdx()
	resObj.ValidPlays = intSliceOrEmpty(b.GetValidPlayIndices(0))
	resObj.GameEndFlag = b.GetGameEndFlag()
	resObj.WinnerTeam = b.GetWinnerTeam()
	resObj.CurrentTrick = trickCardsToOutput(b.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(b)

	scores := make([]int, 0, domain.BalootTeamCnt)
	roundPoints := make([]int, 0, domain.BalootTeamCnt)
	for team := 0; team < domain.BalootTeamCnt; team++ {
		scores = append(scores, b.GetScore(team))
		roundPoints = append(roundPoints, b.GetRoundPoints(team))
	}
	resObj.Scores = scores
	resObj.RoundPoints = roundPoints
	resObj.Config = controller.BalootWebOutputConfig{Target: b.GetConfig().Target}
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築
func (p *BalootWebPresenter) buildPlayersOutput(b interfaces.BalootGame) []*controller.BalootWebOutputPlayer {
	out := make([]*controller.BalootWebOutputPlayer, 0)
	for i := 0; i < b.GetPlayerCnt(); i++ {
		player := b.GetPlayer(i)
		out = append(out, &controller.BalootWebOutputPlayer{
			ID:        i,
			IsHuman:   player.GetIsHuman(),
			Team:      domain.BalootTeamOf(i),
			CardCount: player.GetCardsSize(),
			Cards:     playerCardsToOutput(player, player.GetIsHuman()),
			// **伏せている席は保有そのものを送らない。**フロントで隠すだけだと
			// レスポンスを見れば分かってしまう (#5750)。
			HasBaloot:      player.GetBalootRevealed() && player.GetHasBaloot(),
			BalootRevealed: player.GetBalootRevealed(),
			Declared:       player.GetDeclared(),
			TrickCount:     player.GetTrickCount(),
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *BalootWebPresenter) buildMessage(b interfaces.BalootGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if b.GetGameEndFlag() {
		params := map[string]string{
			"t0": strconv.Itoa(b.GetScore(0)),
			"t1": strconv.Itoa(b.GetScore(1)),
		}
		switch b.GetWinnerTeam() {
		case 0:
			return "", "baloot.result.team0", params
		case 1:
			return "", "baloot.result.team1", params
		default:
			return "", "baloot.result.tie", params
		}
	}
	switch b.GetPhase() {
	case domain.BalootPhaseDeclare:
		// **親は見送れない。** 案内を変えないと選べない選択肢を出すことになる。
		if b.GetDealerIdx() == 0 && b.IsHumanDeclareTurn() {
			return "", "baloot.declare.dealerStuck", nil
		}
		return "", "baloot.declare.choose", nil
	case domain.BalootPhaseRoundEnd:
		return "", "baloot.roundEnd", map[string]string{
			"round": strconv.Itoa(b.GetRoundNumber()),
			"t0":    strconv.Itoa(b.GetRoundPoints(0)),
			"t1":    strconv.Itoa(b.GetRoundPoints(1)),
		}
	}
	return "", "baloot.play", nil
}

// HintOutput ヒント情報をJSON出力する
func (p *BalootWebPresenter) HintOutput(b interfaces.BalootGame) string {
	resObj := p.buildBase(b)
	if hint := b.GetHint(); hint != nil {
		resObj.Hint = &controller.BalootWebOutputHint{CardIndex: hint.CardIndex, Reason: hint.Reason, Suit: hint.Suit}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *BalootWebPresenter) ActionLogOutput(b interfaces.BalootGame) string {
	return actionLogOutputJSON(b)
}
