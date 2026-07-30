//go:build !js || !wasm || extra3

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// SkitgubbeWebPresenter シートグッベWebプレゼンタークラス
type SkitgubbeWebPresenter struct{}

func skitgubbeCardsOutput(cards []*domain.Card) []*controller.WebOutputCard {
	out := make([]*controller.WebOutputCard, 0, len(cards))
	for _, c := range cards {
		if c == nil {
			continue
		}
		out = append(out, cardToOutput(c))
	}
	return out
}

// Output ゲーム状態をJSON出力
func (p *SkitgubbeWebPresenter) Output(c interfaces.SkitgubbeGame, lastErr error) string {
	resObj := p.buildBase(c)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(c, lastErr)
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *SkitgubbeWebPresenter) buildBase(c interfaces.SkitgubbeGame) *controller.SkitgubbeWebOutput {
	resObj := new(controller.SkitgubbeWebOutput)
	resObj.Phase = int(c.GetPhase())
	resObj.CurrentPlayerIdx = c.GetCurrentPlayerIdx()
	resObj.StockCount = c.GetStockCount()
	resObj.TrumpSuit = c.GetTrumpSuit()
	resObj.Duel = skitgubbeCardsOutput(c.GetDuel())
	resObj.DuelLeader = c.GetDuelLeader()
	resObj.Pile = skitgubbeCardsOutput(c.GetPile())
	resObj.GameEndFlag = c.GetGameEndFlag()
	resObj.LoserIdx = c.GetLoserIdx()

	valid := c.GetValidPlayIndices(0)
	resObj.ValidIndices = make([]int, 0, len(valid))
	if !c.GetGameEndFlag() && c.GetCurrentPlayerIdx() == 0 {
		resObj.ValidIndices = append(resObj.ValidIndices, valid...)
	}
	// 引き取れるのは「第2フェーズで、場に札があり、出せる札が 1 枚もない」とき
	// だけ。弱い札を出して逃げることはできない規則の裏返しなので、判定はここで
	// 一度だけ行い、クライアントに再実装させない。
	resObj.CanPickUp = !c.GetGameEndFlag() &&
		c.GetCurrentPlayerIdx() == 0 &&
		c.GetPhase() == domain.SkitgubbePhaseShed &&
		len(c.GetPile()) > 0 &&
		len(valid) == 0

	cfg := c.GetConfig()
	resObj.Config = controller.SkitgubbeWebOutputConfig{CpuDifficulty: int(cfg.CpuDifficulty)}
	resObj.Players = p.buildPlayersOutput(c)

	// ヒントは通常のレスポンスにも載せる。HintOutput にしか設定しないと、
	// フロントは通常の state レスポンスから読むので何も表示されない。
	if !c.GetGameEndFlag() && c.GetCurrentPlayerIdx() == 0 {
		resObj.Hint = skitgubbeHint(c)
	}
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築する。
//
// CPU の手札は伏せる。Workers はこの JSON をそのままブラウザへ返すので、ここで
// 落とさなかったものは相手の手札がそのまま見えることを意味する。枚数と
// 「集めた枚数」は公開する -- どちらも卓上で数えられる情報。
func (p *SkitgubbeWebPresenter) buildPlayersOutput(c interfaces.SkitgubbeGame) []*controller.SkitgubbeWebOutputPlayer {
	players := c.GetPlayers()
	out := make([]*controller.SkitgubbeWebOutputPlayer, 0, len(players))
	for i, player := range players {
		if player == nil {
			continue
		}
		reveal := player.GetIsHuman() || c.GetGameEndFlag()
		cards := make([]*controller.WebOutputCard, 0, player.GetCardsSize())
		if reveal {
			for j := range player.GetCardsSize() {
				if card := player.GetCard(j); card != nil {
					cards = append(cards, cardToOutput(card))
				}
			}
		}
		out = append(out, &controller.SkitgubbeWebOutputPlayer{
			ID:             i,
			IsHuman:        player.GetIsHuman(),
			CardCount:      player.GetCardsSize(),
			Cards:          cards,
			CollectedCount: c.GetCollectedCount(i),
			Finished:       c.IsFinished(i),
			Hidden:         !reveal,
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *SkitgubbeWebPresenter) buildMessage(c interfaces.SkitgubbeGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if !c.GetGameEndFlag() {
		return "", "", nil
	}
	if c.GetLoserIdx() == 0 {
		return "you are the skitgubbe", "skitgubbe.lose", nil
	}
	return "you got rid of your cards", "skitgubbe.win", nil
}

// HintOutput ヒント情報を出力する
func (p *SkitgubbeWebPresenter) HintOutput(c interfaces.SkitgubbeGame) string {
	resObj := p.buildBase(c)
	resObj.Hint = skitgubbeHint(c)
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜を出力する
func (p *SkitgubbeWebPresenter) ActionLogOutput(c interfaces.SkitgubbeGame) string {
	return actionLogOutputJSON(c)
}

// skitgubbeHint 人間プレイヤーへの推奨手を返す。CPU と同じ意思決定を通す。
func skitgubbeHint(c interfaces.SkitgubbeGame) *controller.SkitgubbeWebOutputHint {
	if c.GetGameEndFlag() {
		return &controller.SkitgubbeWebOutputHint{Reason: "skitgubbe.hint.game_end"}
	}
	if c.GetCurrentPlayerIdx() != 0 {
		return &controller.SkitgubbeWebOutputHint{Reason: "skitgubbe.hint.not_your_turn"}
	}
	action := c.SkitgubbeCpuDecide(0)
	if action.PickUp {
		return &controller.SkitgubbeWebOutputHint{PickUp: true, Reason: "skitgubbe.hint.pickup"}
	}
	if action.HandIdx < 0 {
		return &controller.SkitgubbeWebOutputHint{Reason: "skitgubbe.hint.none"}
	}
	idx := action.HandIdx
	reason := "skitgubbe.hint.beat"
	if c.GetPhase() == domain.SkitgubbePhaseCollect {
		reason = "skitgubbe.hint.duel"
	}
	return &controller.SkitgubbeWebOutputHint{CardIndex: &idx, Reason: reason}
}
