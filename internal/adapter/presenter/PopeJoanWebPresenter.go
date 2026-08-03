//go:build !js || !wasm || extra3

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// PopeJoanWebPresenter ポープ・ジョーンWebプレゼンタークラス
type PopeJoanWebPresenter struct{}

func popeJoanCardsOutput(cards []*domain.Card) []*controller.WebOutputCard {
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
func (p *PopeJoanWebPresenter) Output(c interfaces.PopeJoanGame, lastErr error) string {
	resObj := p.buildBase(c)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(c, lastErr)
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *PopeJoanWebPresenter) buildBase(c interfaces.PopeJoanGame) *controller.PopeJoanWebOutput {
	resObj := new(controller.PopeJoanWebOutput)
	resObj.Phase = int(c.GetPhase())
	resObj.CurrentPlayerIdx = c.GetCurrentPlayerIdx()
	resObj.TrumpSuit = c.GetTrumpSuit()
	resObj.RunSuit = c.GetRunSuit()
	resObj.RunRank = c.GetRunRank()
	resObj.DealNo = c.GetDealNumber()
	resObj.DealWinner = c.GetDealWinner()
	resObj.GameEndFlag = c.GetGameEndFlag()
	resObj.WinnerIdx = c.GetWinnerIdx()
	resObj.PlayedPile = popeJoanCardsOutput(c.GetPlayedPile())

	if t := c.GetTurnUp(); t != nil {
		resObj.TurnUp = cardToOutput(t)
	}

	// **8 区画は全部送る。**取られなかった区画は持ち越されるので、いくら
	// 乗っているかがそのまま次のディールの狙いどころになる。
	board := c.GetBoard()
	resObj.Compartments = make([]*controller.PopeJoanWebOutputCompartment, 0, domain.PopeJoanCompartmentCount)
	for i := range domain.PopeJoanCompartmentCount {
		comp := domain.PopeJoanCompartment(i)
		resObj.Compartments = append(resObj.Compartments, &controller.PopeJoanWebOutputCompartment{
			Name: comp.String(), Chips: board.Get(comp),
		})
	}

	awards := c.GetAwards()
	resObj.Awards = make([]*controller.PopeJoanWebOutputAward, 0, len(awards))
	for _, a := range awards {
		if a == nil {
			continue
		}
		resObj.Awards = append(resObj.Awards, &controller.PopeJoanWebOutputAward{
			Compartment: a.Compartment.String(), Player: a.Player, Chips: a.Chips, ByTurnUp: a.ByTurnUp,
		})
	}

	cfg := c.GetConfig()
	resObj.TargetDeals = cfg.TargetDeals
	resObj.Config = controller.PopeJoanWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetDeals:   cfg.TargetDeals,
	}
	resObj.Players = p.buildPlayersOutput(c)

	// ヒントは通常のレスポンスにも載せる。HintOutput にしか設定しないと、
	// フロントは通常の state を読むので何も表示されない。
	if !c.GetGameEndFlag() && c.GetCurrentPlayerIdx() == 0 {
		resObj.Hint = popeJoanHint(c)
	}
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築する。
//
// CPU の手札は伏せるが、**枚数と「Pope を持っているか」は公開**する。出し切られ
// たとき残り札 1 枚につき 1 チップ払い、Pope 保持者だけが免除されるので、両方が
// 見えていないと精算が理解できない。
func (p *PopeJoanWebPresenter) buildPlayersOutput(c interfaces.PopeJoanGame) []*controller.PopeJoanWebOutputPlayer {
	players := c.GetPlayers()
	out := make([]*controller.PopeJoanWebOutputPlayer, 0, len(players))
	for i, player := range players {
		if player == nil {
			continue
		}
		reveal := player.GetIsHuman() || c.GetGameEndFlag()
		cards := make([]*controller.WebOutputCard, 0, player.GetCardsSize())
		holdsPope := false
		for j := range player.GetCardsSize() {
			card := player.GetCard(j)
			if card == nil {
				continue
			}
			if card.GetDesign() == domain.CardDesignDiamond && card.GetValue() == domain.PopeJoanPopeRank {
				holdsPope = true
			}
			if reveal {
				cards = append(cards, cardToOutput(card))
			}
		}
		out = append(out, &controller.PopeJoanWebOutputPlayer{
			ID:        i,
			IsHuman:   player.GetIsHuman(),
			CardCount: player.GetCardsSize(),
			Cards:     cards,
			Chips:     player.GetChips(),
			HoldsPope: holdsPope,
			Hidden:    !reveal,
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *PopeJoanWebPresenter) buildMessage(c interfaces.PopeJoanGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if !c.GetGameEndFlag() {
		return "", "", nil
	}
	if c.GetWinnerIdx() == 0 {
		return "you finish with the most chips", "popejoan.win", nil
	}
	return "you finish behind", "popejoan.lose", nil
}

// HintOutput ヒント情報を出力する
func (p *PopeJoanWebPresenter) HintOutput(c interfaces.PopeJoanGame) string {
	resObj := p.buildBase(c)
	resObj.Hint = popeJoanHint(c)
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜を出力する
func (p *PopeJoanWebPresenter) ActionLogOutput(c interfaces.PopeJoanGame) string {
	return actionLogOutputJSON(c)
}

// popeJoanHint 人間プレイヤーへの推奨手を返す。CPU と同じ意思決定を通す。
func popeJoanHint(c interfaces.PopeJoanGame) *controller.PopeJoanWebOutputHint {
	if c.GetGameEndFlag() {
		return &controller.PopeJoanWebOutputHint{Reason: "popejoan.hint.game_end"}
	}
	if c.GetPhase() != domain.PopeJoanPhasePlay {
		return &controller.PopeJoanWebOutputHint{Reason: "popejoan.hint.deal_end"}
	}
	if c.GetCurrentPlayerIdx() != 0 {
		return &controller.PopeJoanWebOutputHint{Reason: "popejoan.hint.not_your_turn"}
	}
	idx := c.PopeJoanCpuDecide(0)
	if idx < 0 {
		return &controller.PopeJoanWebOutputHint{Reason: "popejoan.hint.none"}
	}
	// 並びの途中なら続きは一意。止まっていれば「最も低い札」しか出せない。
	reason := "popejoan.hint.follow"
	if c.GetRunSuit() < 0 {
		reason = "popejoan.hint.lead"
	}
	return &controller.PopeJoanWebOutputHint{CardIndex: &idx, Reason: reason}
}
