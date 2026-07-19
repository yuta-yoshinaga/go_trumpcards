//go:build !js || !wasm || extra

package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// hachihachiFace は花札 (非52枚デッキ) の 1 枚を手続き描画するための自己記述子を返す。
// 色はこいこいと同じ規則 (光→gold, 赤短→red, 青短→blue, タネ→purple, カス→black)。
// deck:"hanafuda" を付与することでフロントエンドは手続き描画パスへ切り替える (ADR-0033)。
func hachihachiFace(card *domain.Card) *CardFace {
	if card == nil {
		return nil
	}
	var color string
	switch domain.HachiHachiCardCategory(card) {
	case domain.HachiHachiBright:
		color = "gold"
	case domain.HachiHachiAnimal:
		color = "purple"
	case domain.HachiHachiRibbon:
		switch domain.HachiHachiCardRibbonColor(card) {
		case domain.HachiHachiRibbonBlue:
			color = "blue"
		default:
			color = "red"
		}
	default:
		color = "black"
	}
	return &CardFace{
		Glyph: domain.HachiHachiCardGlyph(card),
		Label: domain.HachiHachiCardLabel(card),
		Color: color,
		Deck:  "hanafuda",
	}
}

// HachiHachiWebPresenter は八八 (Hachi-Hachi) の Web プレゼンタークラス。
type HachiHachiWebPresenter struct{}

// Output はゲーム状態を JSON 出力する。
func (p *HachiHachiWebPresenter) Output(g interfaces.HachiHachiGame, lastErr error) string {
	resObj := p.buildBase(g)
	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if g.GetGameEndFlag() || g.GetPhase() == domain.HachiHachiPhaseGameEnd {
		resObj.Message = p.buildResultMessage(g)
		resObj.MessageCode = "hachihachi.result.scores"
		resObj.MessageParams = map[string]string{"scores": p.encodeScoresParam(g)}
	}
	// 人間の手番 (プレイフェーズ) ではヒントを埋め、フロントエンドのヒントトグルを
	// 機能させる。GetHint は CPU 手番・ゲーム終了・対象外フェーズでは nil を返すため、
	// その場合 Hint は設定されない。
	p.applyHint(resObj, g)
	return marshalOrError(resObj)
}

// applyHint は人間の手番であればヒント情報を出力オブジェクトへ埋める。
// g.GetHint() は CPU 手番・ゲーム終了・対象外フェーズでは nil を返すため、
// その場合 Hint は設定されず、omitempty により JSON からも省かれる。
func (p *HachiHachiWebPresenter) applyHint(resObj *controller.HachiHachiWebOutput, g interfaces.HachiHachiGame) {
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.HachiHachiWebOutputHint{
			CardIndex:  hint.CardIndex,
			FieldIndex: hint.FieldIndex,
			Reason:     hint.Reason,
		}
	}
}

// buildBase は基本フィールドを埋めた出力オブジェクトを生成する。
func (p *HachiHachiWebPresenter) buildBase(g interfaces.HachiHachiGame) *controller.HachiHachiWebOutput {
	resObj := new(controller.HachiHachiWebOutput)
	resObj.Players = make([]*controller.HachiHachiWebOutputPlayer, 0)
	resObj.PlayableIndices = make([]int, 0)
	resObj.CaptureOptions = make(map[int][]int)

	fieldOut := make([]*controller.WebOutputCard, 0, len(g.GetFieldCards()))
	for _, c := range g.GetFieldCards() {
		fieldOut = append(fieldOut, cardToOutputWithFace(c, hachihachiFace))
	}
	resObj.FieldCards = fieldOut

	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.CurrentTurn = g.GetCurrentTurn()
	resObj.RemainingDeck = g.GetRemainingDeck()
	resObj.Winner = g.GetWinner()
	resObj.Result = int(g.GetResult())
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.IsHumanTurn = g.IsHumanTurn()

	cfg := g.GetConfig()
	resObj.Config = controller.HachiHachiWebConfigOutput{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetRounds:  cfg.TargetRounds,
	}

	if g.GetPhase() == domain.HachiHachiPhasePlay && g.IsHumanTurn() {
		if idx := g.GetPlayableIndices(g.GetCurrentTurn()); idx != nil {
			resObj.PlayableIndices = idx
		}
		if opts := g.GetCaptureOptions(g.GetCurrentTurn()); opts != nil {
			resObj.CaptureOptions = opts
		}
	}

	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		yakus, raw := g.GetYaku(i)
		captured := make([]*controller.WebOutputCard, 0, player.CapturedCount())
		for _, c := range player.GetCapturedCards() {
			captured = append(captured, cardToOutputWithFace(c, hachihachiFace))
		}
		resObj.Players = append(resObj.Players, &controller.HachiHachiWebOutputPlayer{
			ID:            i,
			IsHuman:       player.GetIsHuman(),
			CardCount:     player.GetCardsSize(),
			Cards:         playerCardsToOutputWithFace(player, player.GetIsHuman(), hachihachiFace),
			Captured:      captured,
			CapturedCount: player.CapturedCount(),
			Score:         player.GetScore(),
			RoundDelta:    player.GetRoundDelta(),
			RawScore:      raw,
			Yaku:          hachihachiYakuToWeb(yakus),
		})
	}

	if det := g.GetLastRoundResult(); det != nil {
		scores := make([]*controller.HachiHachiWebOutputPlayerScore, 0, len(det.Scores))
		for _, s := range det.Scores {
			scores = append(scores, &controller.HachiHachiWebOutputPlayerScore{
				PlayerIdx: s.PlayerIdx,
				RawScore:  s.RawScore,
				Yaku:      hachihachiYakuToWeb(s.Yaku),
				Bonus:     s.Bonus,
				Delta:     s.Delta,
			})
		}
		resObj.LastRoundResult = &controller.HachiHachiWebOutputRoundResult{
			Scores: scores,
			Best:   det.Best,
		}
	}
	return resObj
}

// hachihachiYakuToWeb は domain の出来役スライスを Web 出力型へ変換する。
func hachihachiYakuToWeb(yakus []domain.HachiHachiYaku) []*controller.HachiHachiWebOutputYaku {
	out := make([]*controller.HachiHachiWebOutputYaku, 0, len(yakus))
	for _, y := range yakus {
		out = append(out, &controller.HachiHachiWebOutputYaku{Key: y.Key, Points: y.Points})
	}
	return out
}

// encodeScoresParam は累計得点を "0:12,1:-3" 形式の文字列に詰める。
func (p *HachiHachiWebPresenter) encodeScoresParam(g interfaces.HachiHachiGame) string {
	parts := make([]string, 0, g.GetPlayerCnt())
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		parts = append(parts, fmt.Sprintf("%d:%d", i, player.GetScore()))
	}
	return strings.Join(parts, ",")
}

// buildResultMessage はゲーム終了時のフォールバック (英語) メッセージ。
func (p *HachiHachiWebPresenter) buildResultMessage(g interfaces.HachiHachiGame) string {
	msg := "Game over. "
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		name := fmt.Sprintf("CPU%d", i)
		if player.GetIsHuman() {
			name = "You"
		}
		msg += fmt.Sprintf("%s:%d ", name, player.GetScore())
	}
	return msg
}

// HintOutput はヒント情報を JSON 出力する。
func (p *HachiHachiWebPresenter) HintOutput(g interfaces.HachiHachiGame) string {
	resObj := p.buildBase(g)
	p.applyHint(resObj, g)
	return marshalOrError(resObj)
}

// ActionLogOutput は棋譜を JSON 出力する。
func (p *HachiHachiWebPresenter) ActionLogOutput(g interfaces.HachiHachiGame) string {
	return actionLogOutputJSON(g)
}
