//go:build !js || !wasm || extra3

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// NainJauneWebPresenter ル・ナン・ジョーヌWebプレゼンタークラス
type NainJauneWebPresenter struct{}

func nainJauneCardsOutput(cards []*domain.Card) []*controller.WebOutputCard {
	out := make([]*controller.WebOutputCard, 0, len(cards))
	for _, c := range cards {
		if c == nil {
			continue
		}
		out = append(out, cardToOutput(c))
	}
	return out
}

// nainJauneBoxCard は区画を取る札を返す。**スートまで一致していなければ
// 取れない**ので、クライアントに書き写させずここから送る。
func nainJauneBoxCard(box domain.NainJauneBox) *domain.Card {
	for v := 1; v <= 13; v++ {
		for _, d := range []int{
			domain.CardDesignSpade, domain.CardDesignClover,
			domain.CardDesignHeart, domain.CardDesignDiamond,
		} {
			c := domain.NewCard(d, v, true)
			if got, ok := domain.NainJauneBoxForCard(c); ok && got == box {
				return c
			}
		}
	}
	return nil
}

// Output ゲーム状態をJSON出力
func (p *NainJauneWebPresenter) Output(c interfaces.NainJauneGame, lastErr error) string {
	resObj := p.buildBase(c)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(c, lastErr)
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *NainJauneWebPresenter) buildBase(c interfaces.NainJauneGame) *controller.NainJauneWebOutput {
	resObj := new(controller.NainJauneWebOutput)
	resObj.Phase = int(c.GetPhase())
	resObj.CurrentPlayerIdx = c.GetCurrentPlayerIdx()
	resObj.TalonCount = c.GetTalonCount()
	resObj.RunRank = c.GetRunRank()
	resObj.DealNo = c.GetDealNumber()
	resObj.DealWinner = c.GetDealWinner()
	resObj.GameEndFlag = c.GetGameEndFlag()
	resObj.WinnerIdx = c.GetWinnerIdx()
	resObj.PlayedPile = nainJauneCardsOutput(c.GetPlayedPile())

	// **5 区画は全部送る。**取られなかった区画は持ち越されるので、いくら
	// 乗っているかがそのまま狙いどころになる。
	board := c.GetBoard()
	resObj.Boxes = make([]*controller.NainJauneWebOutputBox, 0, domain.NainJauneBoxCount)
	for i := range domain.NainJauneBoxCount {
		box := domain.NainJauneBox(i)
		out := &controller.NainJauneWebOutputBox{Name: box.String(), Chips: board.Get(box)}
		if card := nainJauneBoxCard(box); card != nil {
			out.Card = cardToOutput(card)
		}
		resObj.Boxes = append(resObj.Boxes, out)
	}

	awards := c.GetAwards()
	resObj.Awards = make([]*controller.NainJauneWebOutputAward, 0, len(awards))
	for _, a := range awards {
		if a == nil {
			continue
		}
		resObj.Awards = append(resObj.Awards, &controller.NainJauneWebOutputAward{
			Box: a.Box.String(), Player: a.Player, Chips: a.Chips,
		})
	}

	cfg := c.GetConfig()
	resObj.TargetDeals = cfg.TargetDeals
	resObj.Config = controller.NainJauneWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetDeals:   cfg.TargetDeals,
	}
	resObj.Players = p.buildPlayersOutput(c)

	// ヒントは通常のレスポンスにも載せる。HintOutput にしか設定しないと、
	// フロントは通常の state を読むので何も表示されない。
	if !c.GetGameEndFlag() && c.GetCurrentPlayerIdx() == 0 {
		resObj.Hint = nainJauneHint(c)
	}
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築する。
//
// CPU の手札は伏せるが、**枚数と失点は公開**する。支払いは枚数ではなく点数な
// ので、枚数だけでは相手がいくら抱えているのか読めない。
func (p *NainJauneWebPresenter) buildPlayersOutput(c interfaces.NainJauneGame) []*controller.NainJauneWebOutputPlayer {
	players := c.GetPlayers()
	out := make([]*controller.NainJauneWebOutputPlayer, 0, len(players))
	for i, player := range players {
		if player == nil {
			continue
		}
		reveal := player.GetIsHuman() || c.GetGameEndFlag()
		cards := make([]*controller.WebOutputCard, 0, player.GetCardsSize())
		points := 0
		for j := range player.GetCardsSize() {
			card := player.GetCard(j)
			if card == nil {
				continue
			}
			points += domain.NainJaunePoints(card)
			if reveal {
				cards = append(cards, cardToOutput(card))
			}
		}
		out = append(out, &controller.NainJauneWebOutputPlayer{
			ID:        i,
			IsHuman:   player.GetIsHuman(),
			CardCount: player.GetCardsSize(),
			Cards:     cards,
			Chips:     player.GetChips(),
			Points:    points,
			Hidden:    !reveal,
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *NainJauneWebPresenter) buildMessage(c interfaces.NainJauneGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if !c.GetGameEndFlag() {
		return "", "", nil
	}
	if c.GetWinnerIdx() == 0 {
		return "you finish with the most chips", "nainjaune.win", nil
	}
	return "you finish behind", "nainjaune.lose", nil
}

// HintOutput ヒント情報を出力する
func (p *NainJauneWebPresenter) HintOutput(c interfaces.NainJauneGame) string {
	resObj := p.buildBase(c)
	resObj.Hint = nainJauneHint(c)
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜を出力する
func (p *NainJauneWebPresenter) ActionLogOutput(c interfaces.NainJauneGame) string {
	return actionLogOutputJSON(c)
}

// nainJauneHint 人間プレイヤーへの推奨手を返す。CPU と同じ意思決定を通す。
func nainJauneHint(c interfaces.NainJauneGame) *controller.NainJauneWebOutputHint {
	if c.GetGameEndFlag() {
		return &controller.NainJauneWebOutputHint{Reason: "nainjaune.hint.game_end"}
	}
	if c.GetPhase() != domain.NainJaunePhasePlay {
		return &controller.NainJauneWebOutputHint{Reason: "nainjaune.hint.deal_end"}
	}
	if c.GetCurrentPlayerIdx() != 0 {
		return &controller.NainJauneWebOutputHint{Reason: "nainjaune.hint.not_your_turn"}
	}
	idx := c.NainJauneCpuDecide(0)
	if idx < 0 {
		return &controller.NainJauneWebOutputHint{Reason: "nainjaune.hint.none"}
	}
	// 区画を取れる札なら、それが理由。取れないなら並びの都合。
	reason := "nainjaune.hint.lead"
	if c.GetRunRank() > 0 {
		reason = "nainjaune.hint.follow"
	}
	if p := c.GetPlayer(0); p != nil && idx < p.GetCardsSize() {
		if box, ok := domain.NainJauneBoxForCard(p.GetCard(idx)); ok && c.GetBoard().Get(box) > 0 {
			reason = "nainjaune.hint.box"
		}
	}
	return &controller.NainJauneWebOutputHint{CardIndex: &idx, Reason: reason}
}
