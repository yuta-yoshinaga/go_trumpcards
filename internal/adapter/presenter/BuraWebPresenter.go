//go:build !js || !wasm || extra3

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// BuraWebPresenter ブラWebプレゼンタークラス
type BuraWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *BuraWebPresenter) Output(b interfaces.BuraGame, lastErr error) string {
	resObj := p.buildBase(b)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(b, lastErr)
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *BuraWebPresenter) buildBase(b interfaces.BuraGame) *controller.BuraWebOutput {
	resObj := new(controller.BuraWebOutput)
	resObj.Phase = int(b.GetPhase())
	resObj.TrickNumber = b.GetTrickNumber()
	resObj.CurrentPlayerIdx = b.GetCurrentPlayerIdx()
	resObj.LeadPlayerIdx = b.GetLeadPlayerIdx()
	resObj.TrumpSuit = b.GetTrumpSuit()
	if tc := b.GetTrumpCard(); tc != nil {
		resObj.TrumpCard = cardToOutput(tc)
	}
	resObj.StockRemaining = len(b.GetStock())
	resObj.WinThreshold = domain.BuraWinThreshold
	resObj.GameEndFlag = b.GetGameEndFlag()
	resObj.WinnerIdx = b.GetWinnerIdx()
	resObj.IsDraw = b.IsDraw()
	// 役の一覧はドメインから。画面側で数え直すと、役を足したとき案内だけが古くなる。
	combos := domain.BuraWinningCombinations()
	resObj.WinningCombinations = make([]string, 0, len(combos))
	for _, c := range combos {
		resObj.WinningCombinations = append(resObj.WinningCombinations, c.Key())
	}

	cfg := b.GetConfig()
	resObj.Config = controller.BuraWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
	}

	resObj.CurrentLead = make([]*controller.WebOutputCard, 0, len(b.GetCurrentLead()))
	for _, c := range b.GetCurrentLead() {
		resObj.CurrentLead = append(resObj.CurrentLead, cardToOutput(c))
	}
	resObj.Players = p.buildPlayersOutput(b)

	// Populate the hint on every response, not only on the `hint` command.
	//
	// The other games set Hint exclusively in HintOutput, which no page calls:
	// the frontend reads it off the ordinary state response, so their hint
	// toggles have nothing to show and silently do nothing. Computing it here
	// is a pure call over a three-card hand. The toggle stays client-side --
	// this is the player's own suggestion, not hidden information.
	if !b.GetGameEndFlag() && b.GetCurrentPlayerIdx() == 0 {
		indices, reason := buraHint(b)
		resObj.Hint = &controller.BuraWebOutputHint{CardIndices: indices, Reason: reason}
	}
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築する。
//
// CPU の手札は伏せる。Workers はこの JSON をそのままブラウザへ返すので、
// ここで落とさなかったものは相手の手札がそのまま見えることを意味する。
// 枚数 (CardCount) だけは常に送る -- 何枚持っているかは公開情報で、
// 伏せてしまうと UI が裏向きの札を何枚描けばよいか分からなくなる。
func (p *BuraWebPresenter) buildPlayersOutput(b interfaces.BuraGame) []*controller.BuraWebOutputPlayer {
	players := b.GetPlayers()
	out := make([]*controller.BuraWebOutputPlayer, 0, len(players))
	for i, player := range players {
		if player == nil {
			continue
		}
		// 終局後は全員の手札を開く。それ以外は人間の手札だけ。
		reveal := player.GetIsHuman() || b.GetGameEndFlag()
		out = append(out, &controller.BuraWebOutputPlayer{
			ID:        i,
			IsHuman:   player.GetIsHuman(),
			CardCount: player.GetCardsSize(),
			Cards:     playerCardsToOutput(player, reveal),
			Points:    b.GetPlayerPoints(i),
			Hidden:    !reveal,
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *BuraWebPresenter) buildMessage(b interfaces.BuraGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if !b.GetGameEndFlag() {
		return "", "", nil
	}
	if b.IsDraw() {
		return "draw", "bura.draw", nil
	}
	switch b.GetWinnerIdx() {
	case 0:
		return "you win", "bura.win", nil
	case -1:
		return "draw", "bura.draw", nil
	default:
		return "you lose", "bura.lose", nil
	}
}

// HintOutput ヒント情報を出力する
func (p *BuraWebPresenter) HintOutput(b interfaces.BuraGame) string {
	resObj := p.buildBase(b)
	indices, reason := buraHint(b)
	resObj.Hint = &controller.BuraWebOutputHint{
		CardIndices: indices,
		Reason:      reason,
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜を出力する
func (p *BuraWebPresenter) ActionLogOutput(b interfaces.BuraGame) string {
	return actionLogOutputJSON(b)
}

// buraHint 人間プレイヤーへの推奨手を返す。
//
// 提案は CPU と同じ意思決定を通す。別の理屈でヒントを組むと、CPU が避ける
// 手を human に勧めることになりかねない。
func buraHint(b interfaces.BuraGame) ([]int, string) {
	if b.GetGameEndFlag() {
		return nil, "bura.hint.game_end"
	}
	if b.GetCurrentPlayerIdx() != 0 {
		return nil, "bura.hint.not_your_turn"
	}
	action := b.BuraCpuDecide(0)
	switch {
	case action.Declare:
		// **役はどれも手札 3 枚すべてで成る。**ブラ(全部切札)・モスクワ(全部 A)・
		// 小モスクワ(全部 6 + 切札の 6)・モロトカ(全部同スート) はいずれも
		// `BuraDetectCombination` が手札全体を見て判定しており、一部の札で
		// 成立することはない。だから「どの 3 枚か」は常に全部で、検出器を
		// 拡張する必要はない。**一番重要な瞬間だけ何も光らない**のを直す (#4909)。
		return buraWholeHandIndices(b), "bura.hint.declare"
	case action.Claim:
		return nil, "bura.hint.claim"
	case len(b.GetCurrentLead()) > 0:
		return action.Indices, "bura.hint.respond"
	default:
		return action.Indices, "bura.hint.lead"
	}
}

// buraWholeHandIndices は人間の手札すべてのインデックスを返す。
func buraWholeHandIndices(b interfaces.BuraGame) []int {
	p := b.GetPlayer(0)
	if p == nil {
		return nil
	}
	out := make([]int, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		out = append(out, i)
	}
	return out
}
