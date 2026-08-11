//go:build !js || !wasm || extra3

package presenter

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MinibridgeWebPresenter ミニブリッジWebプレゼンタークラス
type MinibridgeWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *MinibridgeWebPresenter) Output(s interfaces.MinibridgeGame, lastErr error) string {
	resObj := p.buildBase(s)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(s, lastErr)
	// 受動ヒントは Output() でも埋める (#4483)。
	resObj.Hint = minibridgeHintOutput(s)
	return marshalOrError(resObj)
}

// minibridgeHintOutput はヒントを出力形に変換する。
func minibridgeHintOutput(s interfaces.MinibridgeGame) *controller.MinibridgeWebOutputHint {
	hint := s.GetHint()
	if hint == nil {
		return nil
	}
	return &controller.MinibridgeWebOutputHint{
		CardIndex: hint.CardIndex, Reason: hint.Reason, Level: hint.Level, Suit: hint.Suit,
	}
}

// buildBase 共通フィールドを構築
func (p *MinibridgeWebPresenter) buildBase(s interfaces.MinibridgeGame) *controller.MinibridgeWebOutput {
	resObj := new(controller.MinibridgeWebOutput)
	resObj.Phase = int(s.GetPhase())
	resObj.RoundNumber = s.GetRoundNumber()
	resObj.TrickNumber = s.GetTrickNumber()
	resObj.ContractLevel = s.GetContractLevel()
	resObj.ContractSuit = s.GetContractSuit()
	resObj.RequiredTricks = s.RequiredTricks()
	resObj.DeclarerIdx = s.GetDeclarerIdx()
	resObj.DummyIdx = s.GetDummyIdx()
	resObj.DummyHand = cardsToOutputOrEmpty(s.GetDummyHand())
	resObj.LastMade = s.GetLastMade()
	resObj.LastTricks = s.GetLastTricks()
	resObj.TeamScores = minibridgeTeamScores(s)
	resObj.CurrentPlayerIdx = s.GetCurrentPlayerIdx()
	resObj.LeadPlayerIdx = s.GetLeadPlayerIdx()
	resObj.DealerIdx = s.GetDealerIdx()
	// **ダミーの手番も人間が操作する。** 席 0 固定で返すと、ダミーを動かすとき
	// 出せる札が分からない。
	resObj.ValidPlays = intSliceOrEmpty(s.GetValidPlayIndices(minibridgeControlledSeat(s)))
	resObj.GameEndFlag = s.GetGameEndFlag()
	resObj.WinnerTeam = s.GetWinnerTeam()
	resObj.CurrentTrick = trickCardsToOutput(s.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(s)
	resObj.Config = controller.MinibridgeWebOutputConfig{Rounds: s.GetConfig().Rounds}
	return resObj
}

// minibridgeControlledSeat は人間がいま動かしている席を返す。
//
// **人間がデクレアラーならダミーの手番も人間の出番**なので、そのときは
// ダミーの席の合法手を返します。
func minibridgeControlledSeat(s interfaces.MinibridgeGame) int {
	if s.GetPhase() == domain.MinibridgePhasePlay && s.IsHumanTurn() {
		return s.GetCurrentPlayerIdx()
	}
	return 0
}

// minibridgeTeamScores はチーム得点を配列で返す。
func minibridgeTeamScores(s interfaces.MinibridgeGame) []int {
	out := make([]int, 0, domain.MinibridgeTeamCnt)
	for t := range domain.MinibridgeTeamCnt {
		out = append(out, s.GetTeamScore(t))
	}
	return out
}

// buildPlayersOutput プレイヤー情報を構築
func (p *MinibridgeWebPresenter) buildPlayersOutput(s interfaces.MinibridgeGame) []*controller.MinibridgeWebOutputPlayer {
	out := make([]*controller.MinibridgeWebOutputPlayer, 0)
	for i := 0; i < s.GetPlayerCnt(); i++ {
		player := s.GetPlayer(i)
		// **ダミーの手札は契約が決まった時点で公開される。**
		reveal := player.GetIsHuman() || (i == s.GetDummyIdx() && s.GetContractLevel() > 0)
		out = append(out, &controller.MinibridgeWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, reveal),
			Hcp:        player.GetHcp(),
			Team:       player.GetTeam(),
			TrickCount: player.GetTrickCount(),
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *MinibridgeWebPresenter) buildMessage(s interfaces.MinibridgeGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if s.GetGameEndFlag() {
		// **勝敗はペア単位なので、席番号を渡す文言が無い。** 未使用のパラメータは
		// 載せない（レビュー指摘 PR #5313）。
		switch s.GetWinnerTeam() {
		case 0:
			return "", "minibridge.result.you", nil
		case -1:
			return "", "minibridge.result.tie", nil
		default:
			return "", "minibridge.result.cpu", nil
		}
	}
	switch s.GetPhase() {
	case domain.MinibridgePhaseContract:
		params := map[string]string{"hcp": strconv.Itoa(minibridgePairHcp(s))}
		if s.IsHumanContractTurn() {
			return "", "minibridge.contract.choose", params
		}
		return "", "minibridge.contract.wait", params
	case domain.MinibridgePhaseRoundEnd:
		code := "minibridge.roundEnd.down"
		if s.GetLastMade() {
			code = "minibridge.roundEnd.made"
		}
		return "", code, map[string]string{
			"need": strconv.Itoa(s.RequiredTricks()),
			"took": strconv.Itoa(s.GetLastTricks()),
		}
	default:
		return "", "minibridge.play", map[string]string{
			"need": strconv.Itoa(s.RequiredTricks()),
			"took": strconv.Itoa(minibridgeDeclarerTricks(s)),
		}
	}
}

// minibridgePairHcp は落札側ペアの合計 HCP を返す。
func minibridgePairHcp(s interfaces.MinibridgeGame) int {
	decl := s.GetPlayer(s.GetDeclarerIdx())
	if decl == nil {
		return 0
	}
	total := 0
	for i := 0; i < s.GetPlayerCnt(); i++ {
		if p := s.GetPlayer(i); p != nil && p.GetTeam() == decl.GetTeam() {
			total += p.GetHcp()
		}
	}
	return total
}

// minibridgeDeclarerTricks は落札側が取ったトリック数を返す。
//
// **ペアの 2 席ぶんを足す。** 落札者ひとりぶんで判定すると必ず足りない。
func minibridgeDeclarerTricks(s interfaces.MinibridgeGame) int {
	decl := s.GetPlayer(s.GetDeclarerIdx())
	if decl == nil {
		return 0
	}
	total := 0
	for i := 0; i < s.GetPlayerCnt(); i++ {
		if p := s.GetPlayer(i); p != nil && p.GetTeam() == decl.GetTeam() {
			total += p.GetTrickCount()
		}
	}
	return total
}

// HintOutput ヒント情報をJSON出力する
func (p *MinibridgeWebPresenter) HintOutput(s interfaces.MinibridgeGame) string {
	resObj := p.buildBase(s)
	resObj.Hint = minibridgeHintOutput(s)
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *MinibridgeWebPresenter) ActionLogOutput(s interfaces.MinibridgeGame) string {
	return actionLogOutputJSON(s)
}
