//go:build !js || !wasm || extra2

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// LaughAndLieDownWebPresenter ラフ・アンド・ライダウンWebプレゼンタークラス
type LaughAndLieDownWebPresenter struct{}

func laughAndLieDownCardsOutput(cards []*domain.Card) []*controller.WebOutputCard {
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
func (p *LaughAndLieDownWebPresenter) Output(c interfaces.LaughAndLieDownGame, lastErr error) string {
	resObj := p.buildBase(c)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(c, lastErr)
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *LaughAndLieDownWebPresenter) buildBase(c interfaces.LaughAndLieDownGame) *controller.LaughAndLieDownWebOutput {
	resObj := new(controller.LaughAndLieDownWebOutput)
	resObj.Phase = int(c.GetPhase())
	resObj.CurrentPlayerIdx = c.GetCurrentPlayerIdx()
	resObj.Layout = laughAndLieDownCardsOutput(c.GetLayout())
	resObj.DealerIdx = c.GetDealerIdx()
	resObj.LastInIdx = c.GetLastInIdx()
	resObj.LastInBonus = domain.LaughAndLieDownLastInBonus
	resObj.Pot = domain.LaughAndLieDownPot
	resObj.GameEndFlag = c.GetGameEndFlag()

	valid := c.GetValidPlayIndices(0)
	resObj.ValidIndices = make([]int, 0, len(valid))
	// **3 枚取りができる添字は別に送る。**「1 枚か 3 枚」という原典の選択肢を
	// クライアントが場札を数え直して再現する必要をなくす。
	resObj.ThreeTakeIndices = make([]int, 0, len(valid))
	if !c.GetGameEndFlag() && c.GetCurrentPlayerIdx() == 0 {
		resObj.ValidIndices = append(resObj.ValidIndices, valid...)
		for _, i := range valid {
			if c.CanTakeThree(0, i) {
				resObj.ThreeTakeIndices = append(resObj.ThreeTakeIndices, i)
			}
		}
	}

	cfg := c.GetConfig()
	resObj.Config = controller.LaughAndLieDownWebOutputConfig{CpuDifficulty: int(cfg.CpuDifficulty)}
	resObj.Players = p.buildPlayersOutput(c)

	// ヒントは通常のレスポンスにも載せる。HintOutput にしか設定しないと、
	// フロントは通常の state を読むので何も表示されない。
	if !c.GetGameEndFlag() && c.GetCurrentPlayerIdx() == 0 {
		resObj.Hint = laughAndLieDownHint(c)
	}
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築する。
//
// CPU の手札は伏せる。**取得枚数は公開**する -- 8 枚との差がそのまま精算になる
// 数字で、卓上でも数えられる。
func (p *LaughAndLieDownWebPresenter) buildPlayersOutput(c interfaces.LaughAndLieDownGame) []*controller.LaughAndLieDownWebOutputPlayer {
	players := c.GetPlayers()
	out := make([]*controller.LaughAndLieDownWebOutputPlayer, 0, len(players))
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
		out = append(out, &controller.LaughAndLieDownWebOutputPlayer{
			ID:        i,
			IsHuman:   player.GetIsHuman(),
			CardCount: player.GetCardsSize(),
			Cards:     cards,
			WonCount:  c.GetWonCount(i),
			LaidDown:  c.IsLaidDown(i),
			Score:     c.GetScore(i),
			Hidden:    !reveal,
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *LaughAndLieDownWebPresenter) buildMessage(c interfaces.LaughAndLieDownGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if !c.GetGameEndFlag() {
		return "", "", nil
	}
	switch {
	case c.GetScore(0) > 0:
		return "you finish ahead", "laughandliedown.win", nil
	case c.GetScore(0) == 0:
		return "you break even", "laughandliedown.even", nil
	default:
		return "you finish behind", "laughandliedown.lose", nil
	}
}

// HintOutput ヒント情報を出力する
func (p *LaughAndLieDownWebPresenter) HintOutput(c interfaces.LaughAndLieDownGame) string {
	resObj := p.buildBase(c)
	resObj.Hint = laughAndLieDownHint(c)
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜を出力する
func (p *LaughAndLieDownWebPresenter) ActionLogOutput(c interfaces.LaughAndLieDownGame) string {
	return actionLogOutputJSON(c)
}

// laughAndLieDownHint 人間プレイヤーへの推奨手を返す。CPU と同じ意思決定を通す。
func laughAndLieDownHint(c interfaces.LaughAndLieDownGame) *controller.LaughAndLieDownWebOutputHint {
	if c.GetGameEndFlag() {
		return &controller.LaughAndLieDownWebOutputHint{TakeCount: 1, Reason: "laughandliedown.hint.game_end"}
	}
	if c.GetCurrentPlayerIdx() != 0 {
		return &controller.LaughAndLieDownWebOutputHint{TakeCount: 1, Reason: "laughandliedown.hint.not_your_turn"}
	}
	action := c.LaughAndLieDownCpuDecide(0)
	if action.HandIdx < 0 {
		// 取れないなら降りるしかない。選択肢ではないので、その旨だけ伝える。
		return &controller.LaughAndLieDownWebOutputHint{TakeCount: 1, Reason: "laughandliedown.hint.must_lie_down"}
	}
	idx := action.HandIdx
	reason := "laughandliedown.hint.take_one"
	if action.TakeCount == 3 {
		reason = "laughandliedown.hint.take_three"
	}
	return &controller.LaughAndLieDownWebOutputHint{CardIndex: &idx, TakeCount: action.TakeCount, Reason: reason}
}
