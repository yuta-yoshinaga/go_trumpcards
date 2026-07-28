//go:build !js || !wasm || extra3

package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// koikoiFace は花札 (非52枚デッキ) の 1 枚を手続き描画するための自己記述子を返す。
// すべての花札札が対象 (標準52枚は存在しない)。色は札種/短冊色に応じて INK_COLORS に
// 存在するトークンへ対応付ける (光→gold, 赤短→red, 青短→blue, タネ→purple, カス→black)。
// deck:"hanafuda" を付与することでフロントエンドは手続き描画パスへ切り替える。詳細は
// ADR-0033。
func koikoiFace(card *domain.Card) *CardFace {
	if card == nil {
		return nil
	}
	var color string
	switch domain.KoiKoiCardCategory(card) {
	case domain.KoiKoiBright:
		color = "gold"
	case domain.KoiKoiAnimal:
		color = "purple"
	case domain.KoiKoiRibbon:
		switch domain.KoiKoiCardRibbonColor(card) {
		case domain.KoiKoiRibbonBlue:
			color = "blue"
		default:
			color = "red"
		}
	default:
		color = "black"
	}
	return &CardFace{
		Glyph: domain.KoiKoiCardGlyph(card),
		Label: domain.KoiKoiCardLabel(card),
		Color: color,
		Deck:  "hanafuda",
	}
}

// KoiKoiWebPresenter はこいこい (Koi-Koi) の Web プレゼンタークラス。
type KoiKoiWebPresenter struct{}

// Output はゲーム状態を JSON 出力する。
func (p *KoiKoiWebPresenter) Output(g interfaces.KoiKoiGame, lastErr error) string {
	resObj := p.buildBase(g)
	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if g.GetGameEndFlag() || g.GetPhase() == domain.KoiKoiPhaseGameEnd {
		resObj.Message = p.buildResultMessage(g)
		resObj.MessageCode = "koikoi.result.scores"
		resObj.MessageParams = map[string]string{"scores": p.encodeScoresParam(g)}
	}
	// 人間の手番 (プレイ／こいこい判断) ではヒントを埋め、フロントエンドのヒントトグルを
	// 機能させる。GetHint は CPU 手番・ゲーム終了・対象外フェーズでは nil を返すため、
	// その場合 Hint は設定されない。
	p.applyHint(resObj, g)
	return marshalOrError(resObj)
}

// applyHint は人間の手番であればヒント情報を出力オブジェクトへ埋める。
// g.GetHint() は CPU 手番・ゲーム終了・対象外フェーズでは nil を返すため、
// その場合 Hint は設定されず、omitempty により JSON からも省かれる。
func (p *KoiKoiWebPresenter) applyHint(resObj *controller.KoiKoiWebOutput, g interfaces.KoiKoiGame) {
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.KoiKoiWebOutputHint{
			CardIndex:  hint.CardIndex,
			FieldIndex: hint.FieldIndex,
			KoiKoi:     hint.KoiKoi,
			Reason:     hint.Reason,
		}
	}
}

// buildBase は基本フィールドを埋めた出力オブジェクトを生成する。
func (p *KoiKoiWebPresenter) buildBase(g interfaces.KoiKoiGame) *controller.KoiKoiWebOutput {
	resObj := new(controller.KoiKoiWebOutput)
	resObj.Players = make([]*controller.KoiKoiWebOutputPlayer, 0)
	resObj.PlayableIndices = make([]int, 0)
	resObj.CaptureOptions = make(map[int][]int)
	resObj.PendingYaku = make([]*controller.KoiKoiWebOutputYaku, 0)

	fieldOut := make([]*controller.WebOutputCard, 0, len(g.GetFieldCards()))
	for _, c := range g.GetFieldCards() {
		fieldOut = append(fieldOut, cardToOutputWithFace(c, koikoiFace))
	}
	resObj.FieldCards = fieldOut

	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.CurrentTurn = g.GetCurrentTurn()
	resObj.RemainingDeck = g.GetRemainingDeck()
	resObj.KoikoiCount = g.GetKoikoiCount()
	resObj.RoundWinner = g.GetRoundWinner()
	resObj.Winner = g.GetWinner()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.IsHumanTurn = g.IsHumanTurn()
	resObj.PendingPoints = g.GetPendingPoints()
	resObj.PendingYaku = yakuToWeb(g.GetPendingYaku())

	cfg := g.GetConfig()
	resObj.Config = controller.KoiKoiWebConfigOutput{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetScore:   cfg.TargetScore,
	}

	if g.GetPhase() == domain.KoiKoiPhasePlay && g.IsHumanTurn() {
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
		yakus, pts := g.GetYaku(i)
		captured := make([]*controller.WebOutputCard, 0, player.CapturedCount())
		for _, c := range player.GetCapturedCards() {
			captured = append(captured, cardToOutputWithFace(c, koikoiFace))
		}
		resObj.Players = append(resObj.Players, &controller.KoiKoiWebOutputPlayer{
			ID:            i,
			IsHuman:       player.GetIsHuman(),
			CardCount:     player.GetCardsSize(),
			Cards:         playerCardsToOutputWithFace(player, player.GetIsHuman(), koikoiFace),
			Captured:      captured,
			CapturedCount: player.CapturedCount(),
			Score:         player.GetScore(),
			CalledKoiKoi:  player.GetCalledKoiKoi(),
			Yaku:          yakuToWeb(yakus),
			YakuPoints:    pts,
		})
	}

	if det := g.GetLastRoundResult(); det != nil {
		resObj.LastRoundResult = &controller.KoiKoiWebOutputRoundResult{
			Winner:      det.Winner,
			Yaku:        yakuToWeb(det.Yaku),
			BasePoints:  det.BasePoints,
			Multiplier:  det.Multiplier,
			Total:       det.Total,
			KoikoiCount: det.KoikoiCount,
		}
	}
	return resObj
}

// yakuToWeb は domain の役スライスを Web 出力型へ変換する。
func yakuToWeb(yakus []domain.KoiKoiYaku) []*controller.KoiKoiWebOutputYaku {
	out := make([]*controller.KoiKoiWebOutputYaku, 0, len(yakus))
	for _, y := range yakus {
		out = append(out, &controller.KoiKoiWebOutputYaku{Key: y.Key, Points: y.Points})
	}
	return out
}

// encodeScoresParam は累計得点を "0:12,1:3" 形式の文字列に詰める。
func (p *KoiKoiWebPresenter) encodeScoresParam(g interfaces.KoiKoiGame) string {
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
func (p *KoiKoiWebPresenter) buildResultMessage(g interfaces.KoiKoiGame) string {
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
func (p *KoiKoiWebPresenter) HintOutput(g interfaces.KoiKoiGame) string {
	resObj := p.buildBase(g)
	p.applyHint(resObj, g)
	return marshalOrError(resObj)
}

// ActionLogOutput は棋譜を JSON 出力する。
func (p *KoiKoiWebPresenter) ActionLogOutput(g interfaces.KoiKoiGame) string {
	return actionLogOutputJSON(g)
}
