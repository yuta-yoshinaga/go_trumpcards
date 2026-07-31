//go:build !js || !wasm || extra2

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// LobaWebPresenter ロバWebプレゼンタークラス
type LobaWebPresenter struct{}

func lobaCardsOutput(cards []*domain.Card) []*controller.WebOutputCard {
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
func (p *LobaWebPresenter) Output(c interfaces.LobaGame, lastErr error) string {
	resObj := p.buildBase(c)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(c, lastErr)
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *LobaWebPresenter) buildBase(c interfaces.LobaGame) *controller.LobaWebOutput {
	resObj := new(controller.LobaWebOutput)
	resObj.Phase = int(c.GetPhase())
	resObj.CurrentPlayerIdx = c.GetCurrentPlayerIdx()
	resObj.StockCount = c.GetStockCount()
	resObj.RoundNo = c.GetRoundNumber()
	resObj.KnockOut = domain.LobaKnockOut
	resObj.RoundWinner = c.GetRoundWinner()
	resObj.RoundClean = c.IsRoundClean()
	resObj.GameEndFlag = c.GetGameEndFlag()
	resObj.WinnerIdx = c.GetWinnerIdx()

	if top := c.GetDiscardTop(); top != nil {
		resObj.DiscardTop = cardToOutput(top)
	}

	melds := c.GetMelds()
	resObj.Melds = make([]*controller.LobaWebOutputMeld, 0, len(melds))
	for _, m := range melds {
		if m == nil {
			continue
		}
		resObj.Melds = append(resObj.Melds, &controller.LobaWebOutputMeld{
			Owner: m.Owner,
			Kind:  int(m.Kind),
			Cards: lobaCardsOutput(m.Cards),
		})
	}

	cfg := c.GetConfig()
	resObj.Config = controller.LobaWebOutputConfig{CpuDifficulty: int(cfg.CpuDifficulty)}
	resObj.Players = p.buildPlayersOutput(c)

	// ヒントは通常のレスポンスにも載せる。HintOutput にしか設定しないと、
	// フロントは通常の state を読むので何も表示されない。
	if !c.GetGameEndFlag() && c.GetCurrentPlayerIdx() == 0 {
		resObj.Hint = lobaHint(c)
	}
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築する。
//
// CPU の手札は伏せる。**失点と脱落は公開**する -- 101 で脱落するので、誰があと
// 何点なのかが最大の判断材料。
func (p *LobaWebPresenter) buildPlayersOutput(c interfaces.LobaGame) []*controller.LobaWebOutputPlayer {
	players := c.GetPlayers()
	out := make([]*controller.LobaWebOutputPlayer, 0, len(players))
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
		out = append(out, &controller.LobaWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      cards,
			Score:      c.GetScore(i),
			Eliminated: c.IsEliminated(i),
			HasMelded:  c.HasMelded(i),
			Hidden:     !reveal,
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *LobaWebPresenter) buildMessage(c interfaces.LobaGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if !c.GetGameEndFlag() {
		return "", "", nil
	}
	if c.GetWinnerIdx() == 0 {
		return "you are the last one standing", "loba.win", nil
	}
	return "you were knocked out", "loba.lose", nil
}

// HintOutput ヒント情報を出力する
func (p *LobaWebPresenter) HintOutput(c interfaces.LobaGame) string {
	resObj := p.buildBase(c)
	resObj.Hint = lobaHint(c)
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜を出力する
func (p *LobaWebPresenter) ActionLogOutput(c interfaces.LobaGame) string {
	return actionLogOutputJSON(c)
}

// lobaHint 人間プレイヤーへの推奨手を返す。CPU と同じ意思決定を通す。
func lobaHint(c interfaces.LobaGame) *controller.LobaWebOutputHint {
	if c.GetGameEndFlag() {
		return &controller.LobaWebOutputHint{Reason: "loba.hint.game_end"}
	}
	if c.GetCurrentPlayerIdx() != 0 {
		return &controller.LobaWebOutputHint{Reason: "loba.hint.not_your_turn"}
	}
	if c.GetPhase() == domain.LobaPhaseDraw {
		return &controller.LobaWebOutputHint{DrawStock: true, Reason: "loba.hint.draw"}
	}
	action := c.LobaCpuDecide(0)
	if len(action.MeldIdxs) > 0 {
		return &controller.LobaWebOutputHint{CardIndices: action.MeldIdxs, Reason: "loba.hint.meld"}
	}
	if action.DiscardIdx < 0 {
		return &controller.LobaWebOutputHint{Reason: "loba.hint.none"}
	}
	idx := action.DiscardIdx
	return &controller.LobaWebOutputHint{CardIndex: &idx, Reason: "loba.hint.discard"}
}
