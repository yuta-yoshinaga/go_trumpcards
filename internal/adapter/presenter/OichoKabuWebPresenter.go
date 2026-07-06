//go:build !js || !wasm || extra

package presenter

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// oichoKabuFace は非52枚デッキ（カブ札）の札を手続き的に描画するための
// 自己記述子を返す。カブ札には専用PNGアートが無いため、全ての札を procedural に
// 描画する。Label/Glyph は札の数字（1〜10）。Deck="kabu"。ADR-0033 参照。
func oichoKabuFace(card *domain.Card) *CardFace {
	if card == nil {
		return nil
	}
	label := strconv.Itoa(card.GetValue())
	return &CardFace{Glyph: label, Label: label, Color: "black", Deck: "kabu"}
}

// oichoKabuHandOutput は手札スライスを oichoKabuFace 付きで WebOutputCard に変換する。
func oichoKabuHandOutput(hand []*domain.Card) []*controller.WebOutputCard {
	out := make([]*controller.WebOutputCard, 0, len(hand))
	for _, c := range hand {
		out = append(out, cardToOutputWithFace(c, oichoKabuFace))
	}
	return out
}

// OichoKabuWebPresenter おいちょかぶWebプレゼンタークラス
type OichoKabuWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *OichoKabuWebPresenter) Output(o interfaces.OichoKabuGame, lastErr error) string {
	resObj := new(controller.OichoKabuWebOutput)

	resObj.PlayerHand = oichoKabuHandOutput(o.GetPlayerHand())
	resObj.PlayerRank = o.GetPlayerRank()

	// 親の手は結果（終了）まで伏せる（ディーラーのホールカードに相当）。
	if o.GetGameEndFlag() {
		resObj.BankerHand = oichoKabuHandOutput(o.GetBankerHand())
		resObj.BankerRank = o.GetBankerRank()
	} else {
		resObj.BankerHand = make([]*controller.WebOutputCard, 0)
	}

	resObj.Phase = o.GetPhase()
	resObj.Chips = o.GetChips()
	resObj.Bet = o.GetBet()
	resObj.Result = int(o.GetResult())
	resObj.TotalPayout = o.GetTotalPayout()

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if o.GetGameEndFlag() {
		resObj.Message, resObj.MessageCode = oichoKabuEndMessage(o)
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *OichoKabuWebPresenter) ActionLogOutput(o interfaces.OichoKabuGame) string {
	return actionLogOutputJSON(o)
}

// oichoKabuEndMessage は終了時の表示メッセージと i18n キーを返す。
func oichoKabuEndMessage(o interfaces.OichoKabuGame) (string, string) {
	switch o.GetResult() {
	case domain.GameResultWin:
		return "Player wins!", "oichokabu.result.playerWins"
	case domain.GameResultLose:
		return "Banker wins.", "oichokabu.result.bankerWins"
	default:
		return "Push.", "oichokabu.result.push"
	}
}
