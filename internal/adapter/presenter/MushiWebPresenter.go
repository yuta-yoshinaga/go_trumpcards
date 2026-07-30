//go:build !js || !wasm || extra2

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MushiWebPresenter 虫Webプレゼンタークラス
type MushiWebPresenter struct{}

// mushiFace は花札 1 枚を手続き描画するための自己記述子を返す (ADR-0033)。
//
// 花札には専用 PNG が無く、design は**スートではなく月**なので、標準の
// cardToOutput をそのまま使うと design が "CLOVER"/"JOKER" に化けて描画パスが
// 壊れる。deck:"hanafuda" を付けてフロントを手続き描画へ切り替える。
func mushiFace(card *domain.Card) *CardFace {
	if card == nil {
		return nil
	}
	var color string
	switch domain.MushiCardCategory(card) {
	case domain.MushiBright:
		color = "gold"
	case domain.MushiAnimal:
		color = "purple"
	case domain.MushiRibbon:
		color = "red"
	default:
		color = "black"
	}
	if domain.MushiIsWild(card) {
		color = "gold" // 雷札はカスだが役割が特別なので目立たせる
	}
	return &CardFace{
		Glyph: domain.MushiCardGlyph(card),
		Label: domain.MushiCardName(card),
		Color: color,
		Deck:  "hanafuda",
	}
}

// mushiCardOutput は 1 枚を、種別・点・ワイルドかどうかまで含めて出力する。
// 点はサーバーが計算して渡す -- 40 枚の札種表をクライアントにもう一部持たせない。
func mushiCardOutput(c *domain.Card) *controller.MushiWebOutputCard {
	if c == nil {
		return nil
	}
	return &controller.MushiWebOutputCard{
		WebOutputCard: cardToOutputWithFace(c, mushiFace),
		Month:         c.GetDesign(),
		Index:         c.GetValue(),
		Category:      int(domain.MushiCardCategory(c)),
		Points:        domain.MushiCardPoints(c),
		IsWild:        domain.MushiIsWild(c),
	}
}

func mushiCardsOutput(cards []*domain.Card) []*controller.MushiWebOutputCard {
	out := make([]*controller.MushiWebOutputCard, 0, len(cards))
	for _, c := range cards {
		if oc := mushiCardOutput(c); oc != nil {
			out = append(out, oc)
		}
	}
	return out
}

// Output ゲーム状態をJSON出力
func (p *MushiWebPresenter) Output(m interfaces.MushiGame, lastErr error) string {
	resObj := p.buildBase(m)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(m, lastErr)
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *MushiWebPresenter) buildBase(m interfaces.MushiGame) *controller.MushiWebOutput {
	resObj := new(controller.MushiWebOutput)
	resObj.Phase = int(m.GetPhase())
	resObj.CurrentPlayerIdx = m.GetCurrentPlayerIdx()
	resObj.DealerIdx = m.GetDealerIdx()
	resObj.RoundNumber = m.GetRoundNumber()
	resObj.StockCount = m.GetStockCount()
	resObj.GameEndFlag = m.GetGameEndFlag()
	resObj.WinnerIdx = m.GetWinnerIdx()
	resObj.Field = mushiCardsOutput(m.GetField())
	resObj.PendingCard = mushiCardOutput(m.GetPendingCard())

	sel := m.GetSelectableIndices()
	resObj.SelectableIndices = make([]int, 0, len(sel))
	resObj.SelectableIndices = append(resObj.SelectableIndices, sel...)

	cfg := m.GetConfig()
	resObj.Config = controller.MushiWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetRounds:  cfg.TargetRounds,
	}
	resObj.TargetRounds = cfg.TargetRounds
	resObj.Players = p.buildPlayersOutput(m)

	// ヒントは通常のレスポンスにも載せる。他ゲームは HintOutput でしか設定して
	// おらず、フロントは通常の state レスポンスから読むため、どのページも呼んで
	// いない = ヒントトグルが何も表示しない状態になっている。40 枚に対する純粋な
	// 計算なので毎回計算してよい。
	if !m.GetGameEndFlag() && m.GetCurrentPlayerIdx() == 0 {
		resObj.Hint = mushiHint(m)
	}
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築する。
//
// CPU の手札は伏せる。Workers はこの JSON をそのままブラウザへ返すので、ここで
// 落とさなかったものは相手の手札がそのまま見えることを意味する。取り札は
// **公開情報**なので常に送る -- 花札は互いの取り札を見ながら打つゲームで、
// 隠すと役の読み合いが成立しない。
func (p *MushiWebPresenter) buildPlayersOutput(m interfaces.MushiGame) []*controller.MushiWebOutputPlayer {
	players := m.GetPlayers()
	out := make([]*controller.MushiWebOutputPlayer, 0, len(players))
	for i, player := range players {
		if player == nil {
			continue
		}
		reveal := player.GetIsHuman() || m.GetGameEndFlag()
		captured := m.GetCaptured(i)
		pts := 0
		for _, c := range captured {
			pts += domain.MushiCardPoints(c)
		}
		cards := make([]*controller.MushiWebOutputCard, 0, player.GetCardsSize())
		if reveal {
			for j := range player.GetCardsSize() {
				if oc := mushiCardOutput(player.GetCard(j)); oc != nil {
					cards = append(cards, oc)
				}
			}
		}
		out = append(out, &controller.MushiWebOutputPlayer{
			ID:             i,
			IsHuman:        player.GetIsHuman(),
			CardCount:      player.GetCardsSize(),
			Cards:          cards,
			Captured:       mushiCardsOutput(captured),
			CapturedPoints: pts,
			Score:          m.GetScore(i),
			RoundResult:    m.GetRoundResult(i),
			Hidden:         !reveal,
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *MushiWebPresenter) buildMessage(m interfaces.MushiGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if m.GetGameEndFlag() {
		switch m.GetWinnerIdx() {
		case 0:
			return "you win", "mushi.win", nil
		case -1:
			return "draw", "mushi.draw", nil
		default:
			return "you lose", "mushi.lose", nil
		}
	}
	if m.GetPhase() == domain.MushiPhaseRoundEnd {
		return "round over", "mushi.round_end", nil
	}
	return "", "", nil
}

// HintOutput ヒント情報を出力する
func (p *MushiWebPresenter) HintOutput(m interfaces.MushiGame) string {
	resObj := p.buildBase(m)
	resObj.Hint = mushiHint(m)
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜を出力する
func (p *MushiWebPresenter) ActionLogOutput(m interfaces.MushiGame) string {
	return actionLogOutputJSON(m)
}

// mushiHint 人間プレイヤーへの推奨手を返す。
//
// CPU と同じ意思決定を通す。別の理屈で組むと、CPU が避ける手を human に勧める。
func mushiHint(m interfaces.MushiGame) *controller.MushiWebOutputHint {
	if m.GetGameEndFlag() {
		return &controller.MushiWebOutputHint{Reason: "mushi.hint.game_end"}
	}
	switch m.GetPhase() {
	case domain.MushiPhaseRoundEnd:
		return &controller.MushiWebOutputHint{Reason: "mushi.hint.round_end"}
	case domain.MushiPhaseSelect, domain.MushiPhaseWildSelect:
		if m.GetCurrentPlayerIdx() != 0 {
			return &controller.MushiWebOutputHint{Reason: "mushi.hint.not_your_turn"}
		}
		idx := m.MushiCpuDecide(0).FieldIdx
		if idx < 0 {
			return &controller.MushiWebOutputHint{Reason: "mushi.hint.none"}
		}
		return &controller.MushiWebOutputHint{FieldIndex: &idx, Reason: "mushi.hint.select"}
	default:
		if m.GetCurrentPlayerIdx() != 0 {
			return &controller.MushiWebOutputHint{Reason: "mushi.hint.not_your_turn"}
		}
		idx := m.MushiCpuDecide(0).HandIdx
		if idx < 0 {
			return &controller.MushiWebOutputHint{Reason: "mushi.hint.none"}
		}
		return &controller.MushiWebOutputHint{CardIndex: &idx, Reason: "mushi.hint.play"}
	}
}
