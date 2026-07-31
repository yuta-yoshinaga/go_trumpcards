//go:build !js || !wasm || extra3

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// DesmocheWebPresenter デスモチェWebプレゼンタークラス
type DesmocheWebPresenter struct{}

func desmocheCardsOutput(cards []*domain.Card) []*controller.WebOutputCard {
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
func (p *DesmocheWebPresenter) Output(c interfaces.DesmocheGame, lastErr error) string {
	resObj := p.buildBase(c)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(c, lastErr)
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *DesmocheWebPresenter) buildBase(c interfaces.DesmocheGame) *controller.DesmocheWebOutput {
	resObj := new(controller.DesmocheWebOutput)
	resObj.Phase = int(c.GetPhase())
	resObj.CurrentPlayerIdx = c.GetCurrentPlayerIdx()
	resObj.StockCount = c.GetStockCount()
	resObj.RoundNo = c.GetRoundNumber()
	resObj.Pot = c.GetPot()
	resObj.GoOutSize = domain.DesmocheGoOutSize
	resObj.RoundWinner = c.GetRoundWinner()
	resObj.RoundExhausted = c.IsRoundExhausted()
	resObj.GameEndFlag = c.GetGameEndFlag()
	resObj.WinnerIdx = c.GetWinnerIdx()

	if top := c.GetDiscardTop(); top != nil {
		resObj.DiscardTop = cardToOutput(top)
	}

	melds := c.GetMelds()
	resObj.Melds = make([]*controller.DesmocheWebOutputMeld, 0, len(melds))
	for _, m := range melds {
		if m == nil {
			continue
		}
		resObj.Melds = append(resObj.Melds, &controller.DesmocheWebOutputMeld{
			Owner: m.Owner,
			Kind:  int(m.Kind),
			Cards: desmocheCardsOutput(m.Cards),
		})
	}

	cfg := c.GetConfig()
	resObj.Config = controller.DesmocheWebOutputConfig{CpuDifficulty: int(cfg.CpuDifficulty)}
	resObj.Players = p.buildPlayersOutput(c)

	// ヒントは通常のレスポンスにも載せる。HintOutput にしか設定しないと、
	// フロントは通常の state を読むので何も表示されない。
	if !c.GetGameEndFlag() && c.GetCurrentPlayerIdx() == 0 {
		resObj.Hint = desmocheHint(c)
	}
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築する。
//
// CPU の手札は伏せるが、**場に出した枚数と収支は公開**する。メルドは表向きに
// 並ぶ規則なので隠しようがなく、10 枚に何枚届いていないかが唯一の進捗表示。
func (p *DesmocheWebPresenter) buildPlayersOutput(c interfaces.DesmocheGame) []*controller.DesmocheWebOutputPlayer {
	players := c.GetPlayers()
	out := make([]*controller.DesmocheWebOutputPlayer, 0, len(players))
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
		out = append(out, &controller.DesmocheWebOutputPlayer{
			ID:          i,
			IsHuman:     player.GetIsHuman(),
			CardCount:   player.GetCardsSize(),
			Cards:       cards,
			Score:       c.GetScore(i),
			MeldedCount: c.MeldedCount(i),
			Hidden:      !reveal,
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *DesmocheWebPresenter) buildMessage(c interfaces.DesmocheGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if !c.GetGameEndFlag() {
		return "", "", nil
	}
	if c.GetWinnerIdx() == 0 {
		return "you finish ahead", "desmoche.win", nil
	}
	return "you finish behind", "desmoche.lose", nil
}

// HintOutput ヒント情報を出力する
func (p *DesmocheWebPresenter) HintOutput(c interfaces.DesmocheGame) string {
	resObj := p.buildBase(c)
	resObj.Hint = desmocheHint(c)
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜を出力する
func (p *DesmocheWebPresenter) ActionLogOutput(c interfaces.DesmocheGame) string {
	return actionLogOutputJSON(c)
}

// desmocheHint 人間プレイヤーへの推奨手を返す。CPU と同じ意思決定を通す。
func desmocheHint(c interfaces.DesmocheGame) *controller.DesmocheWebOutputHint {
	if c.GetGameEndFlag() {
		return &controller.DesmocheWebOutputHint{Reason: "desmoche.hint.game_end"}
	}
	if c.GetCurrentPlayerIdx() != 0 {
		return &controller.DesmocheWebOutputHint{Reason: "desmoche.hint.not_your_turn"}
	}
	if c.GetPhase() == domain.DesmochePhaseDraw {
		return &controller.DesmocheWebOutputHint{DrawStock: true, Reason: "desmoche.hint.draw"}
	}
	action := c.DesmocheCpuDecide(0)
	if len(action.MeldIdxs) > 0 {
		return &controller.DesmocheWebOutputHint{CardIndices: action.MeldIdxs, Reason: "desmoche.hint.meld"}
	}
	if action.DiscardIdx < 0 {
		return &controller.DesmocheWebOutputHint{Reason: "desmoche.hint.none"}
	}
	idx := action.DiscardIdx
	return &controller.DesmocheWebOutputHint{CardIndex: &idx, Reason: "desmoche.hint.discard"}
}
