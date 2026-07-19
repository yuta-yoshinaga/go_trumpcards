//go:build !js || !wasm || extra

package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// gostopFace は花札 (非52枚デッキ) の 1 枚を手続き描画するための自己記述子を返す。
// 色は札種/띠色に応じて INK_COLORS のトークンへ対応付ける (光→gold, 赤띠→red,
// 青띠→blue, 열끗→purple, 피→black)。deck:"hanafuda" でフロントは手続き描画パスへ
// 切り替える。詳細は ADR-0033。
func gostopFace(card *domain.Card) *CardFace {
	if card == nil {
		return nil
	}
	var color string
	switch domain.GoStopCardCategory(card) {
	case domain.GoStopGwang:
		color = "gold"
	case domain.GoStopYeol:
		color = "purple"
	case domain.GoStopTti:
		switch domain.GoStopCardRibbonColor(card) {
		case domain.GoStopRibbonBlue:
			color = "blue"
		default:
			color = "red"
		}
	default:
		color = "black"
	}
	return &CardFace{
		Glyph: domain.GoStopCardGlyph(card),
		Label: domain.GoStopCardLabel(card),
		Color: color,
		Deck:  "hanafuda",
	}
}

// GoStopWebPresenter はゴーストップ (Go-Stop) の Web プレゼンタークラス。
type GoStopWebPresenter struct{}

// Output はゲーム状態を JSON 出力する。
func (p *GoStopWebPresenter) Output(g interfaces.GoStopGame, lastErr error) string {
	resObj := p.buildBase(g)
	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if g.GetGameEndFlag() || g.GetPhase() == domain.GoStopPhaseGameEnd {
		resObj.Message = p.buildResultMessage(g)
		resObj.MessageCode = "gostop.result.scores"
		resObj.MessageParams = map[string]string{"scores": p.encodeScoresParam(g)}
	} else {
		// 人間手番のプレイ/決断フェーズでは通常出力にもヒントを載せる。
		// GetHint は CPU 手番/ゲーム終了時に nil を返すため、その場合 Hint は付かない。
		p.applyHint(resObj, g)
	}
	return marshalOrError(resObj)
}

// applyHint は人間手番であればドメインのヒントを出力オブジェクトへ写す。CPU 手番や
// 終了時など GetHint が nil を返すケースでは Hint フィールドを設定しない。
func (p *GoStopWebPresenter) applyHint(resObj *controller.GoStopWebOutput, g interfaces.GoStopGame) {
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.GoStopWebOutputHint{
			CardIndex:  hint.CardIndex,
			FieldIndex: hint.FieldIndex,
			Go:         hint.Go,
			Reason:     hint.Reason,
		}
	}
}

// breakdownToWeb は domain の得点内訳を Web 出力型へ変換する。
func breakdownToWeb(bd *domain.GoStopBreakdown) *controller.GoStopWebOutputBreakdown {
	if bd == nil {
		return nil
	}
	return &controller.GoStopWebOutputBreakdown{
		Gwang:       bd.Gwang,
		Godori:      bd.Godori,
		Tti:         bd.Tti,
		Yeol:        bd.Yeol,
		Pi:          bd.Pi,
		Base:        bd.Base,
		GoCount:     bd.GoCount,
		GoMult:      bd.GoMult,
		GoScore:     bd.GoScore,
		BrightCount: bd.BrightCount,
		RibbonCount: bd.RibbonCount,
		AnimalCount: bd.AnimalCount,
		PiCount:     bd.PiCount,
	}
}

// buildBase は基本フィールドを埋めた出力オブジェクトを生成する。
func (p *GoStopWebPresenter) buildBase(g interfaces.GoStopGame) *controller.GoStopWebOutput {
	resObj := new(controller.GoStopWebOutput)
	resObj.Players = make([]*controller.GoStopWebOutputPlayer, 0)
	resObj.PlayableIndices = make([]int, 0)
	resObj.CaptureOptions = make(map[int][]int)

	fieldOut := make([]*controller.WebOutputCard, 0, len(g.GetFieldCards()))
	for _, c := range g.GetFieldCards() {
		fieldOut = append(fieldOut, cardToOutputWithFace(c, gostopFace))
	}
	resObj.FieldCards = fieldOut

	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.CurrentTurn = g.GetCurrentTurn()
	resObj.RemainingDeck = g.GetRemainingDeck()
	resObj.RoundWinner = g.GetRoundWinner()
	resObj.Winner = g.GetWinner()
	resObj.Result = int(g.GetResult())
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.IsHumanTurn = g.IsHumanTurn()
	resObj.PendingPoints = g.GetPendingPoints()
	resObj.PendingBreakdown = breakdownToWeb(g.GetPendingBreakdown())

	cfg := g.GetConfig()
	resObj.Config = controller.GoStopWebConfigOutput{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetScore:   cfg.TargetScore,
	}

	if g.GetPhase() == domain.GoStopPhasePlay && g.IsHumanTurn() {
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
		bd, pts := g.GetScore(i)
		captured := make([]*controller.WebOutputCard, 0, player.CapturedCount())
		for _, c := range player.GetCapturedCards() {
			captured = append(captured, cardToOutputWithFace(c, gostopFace))
		}
		resObj.Players = append(resObj.Players, &controller.GoStopWebOutputPlayer{
			ID:            i,
			IsHuman:       player.GetIsHuman(),
			CardCount:     player.GetCardsSize(),
			Cards:         playerCardsToOutputWithFace(player, player.GetIsHuman(), gostopFace),
			Captured:      captured,
			CapturedCount: player.CapturedCount(),
			Score:         player.GetScore(),
			GoCount:       player.GetGoCount(),
			Breakdown:     breakdownToWeb(bd),
			Points:        pts,
		})
	}

	if det := g.GetLastRoundResult(); det != nil {
		resObj.LastRoundResult = &controller.GoStopWebOutputRoundResult{
			Winner:     det.Winner,
			Breakdown:  breakdownToWeb(det.Breakdown),
			BasePoints: det.BasePoints,
			GoScore:    det.GoScore,
			BakMult:    det.BakMult,
			Total:      det.Total,
			GwangBak:   det.GwangBak,
			PiBak:      det.PiBak,
			GoBak:      det.GoBak,
			GoCount:    det.GoCount,
		}
	}
	return resObj
}

// encodeScoresParam は累計得点を "0:12,1:3" 形式の文字列に詰める。
func (p *GoStopWebPresenter) encodeScoresParam(g interfaces.GoStopGame) string {
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
func (p *GoStopWebPresenter) buildResultMessage(g interfaces.GoStopGame) string {
	msg := "Game over. "
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		name := "CPU"
		if player.GetIsHuman() {
			name = "You"
		}
		msg += fmt.Sprintf("%s:%d ", name, player.GetScore())
	}
	return msg
}

// HintOutput はヒント情報を JSON 出力する。
func (p *GoStopWebPresenter) HintOutput(g interfaces.GoStopGame) string {
	resObj := p.buildBase(g)
	p.applyHint(resObj, g)
	return marshalOrError(resObj)
}

// ActionLogOutput は棋譜を JSON 出力する。
func (p *GoStopWebPresenter) ActionLogOutput(g interfaces.GoStopGame) string {
	return actionLogOutputJSON(g)
}
