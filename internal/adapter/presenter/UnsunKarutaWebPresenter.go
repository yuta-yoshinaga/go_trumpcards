//go:build !js || !wasm || classic

package presenter

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// unsunKarutaFace は 75 枚のカルタを手続き描画するための自己記述子を返す。
//
// **既存の 4 スートに収まらない。** 第 5 スート「クル」と、数札 1〜9 + 6 枚の
// 絵札という並びは 52 枚デッキの記号では書けないので、スートごとの記号と位の
// ラベルを自前で持つ (ADR-0033 の手続き描画パス)。
func unsunKarutaFace(card *domain.Card) *CardFace {
	if card == nil {
		return nil
	}
	glyphs := map[int]string{
		domain.UnsunKarutaSuitPao:   "棒",
		domain.UnsunKarutaSuitIsu:   "剣",
		domain.UnsunKarutaSuitKotsu: "杯",
		domain.UnsunKarutaSuitOru:   "金",
		domain.UnsunKarutaSuitKuru:  "巴",
	}
	labels := map[int]string{
		domain.UnsunKarutaSota:  "ソウタ",
		domain.UnsunKarutaUma:   "ウマ",
		domain.UnsunKarutaKiri:  "キリ",
		domain.UnsunKarutaUn:    "ウン",
		domain.UnsunKarutaSun:   "スン",
		domain.UnsunKarutaRobai: "ロバイ",
	}
	label, ok := labels[card.GetValue()]
	if !ok {
		label = strconv.Itoa(card.GetValue())
	}
	glyph, ok := glyphs[card.GetDesign()]
	if !ok {
		glyph = "?"
	}
	// **丸物と長物を色で分ける。** 数札の強さが逆になる 2 系統なので、
	// 画面で見分けられないと 1 と 9 のどちらが強いのか毎回考えることになる。
	colour := "black"
	if domain.UnsunKarutaIsRoundSuit(card.GetDesign()) {
		colour = "red"
	}
	return &CardFace{Glyph: glyph, Label: label, Color: colour, Deck: "unsun"}
}

// UnsunKarutaWebPresenter はうんすんカルタの Web プレゼンター。
type UnsunKarutaWebPresenter struct{}

// Output はゲーム状態を JSON 出力する。
func (p *UnsunKarutaWebPresenter) Output(g interfaces.UnsunKarutaGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	// 受動ヒントは Output でも埋める (#4483)。
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = cardHint(hint.CardIndices, hint.Reason)
	}
	return marshalOrError(resObj)
}

// buildBase は共通フィールドを構築する。
func (p *UnsunKarutaWebPresenter) buildBase(g interfaces.UnsunKarutaGame) *controller.UnsunKarutaWebOutput {
	resObj := new(controller.UnsunKarutaWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.TrickCount = domain.UnsunKarutaTrickCount
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.LeadPlayerIdx = g.GetLeadPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.TrumpSuit = g.GetTrumpSuit()
	resObj.TrumpSuitName = domain.UnsunKarutaSuitName(g.GetTrumpSuit())
	if card := g.TrumpCard(); card != nil {
		resObj.TrumpCard = cardToOutputWithFace(card, unsunKarutaFace)
	}
	resObj.MustFollow = g.IsMustFollow()
	resObj.Declared = g.IsDeclaredThisTrick()
	resObj.CanDeclare = g.CanDeclare()
	resObj.TeamTricks = intsOrEmptySlice(g.GetTeamTricks())
	resObj.TeamScores = intsOrEmptySlice(g.GetTeamScores())
	resObj.LastTrickWinner = g.GetLastTrickWinner()
	resObj.Result = int(g.GetResult())
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerTeam = g.GetWinnerTeam()
	resObj.IsHumanTurn = g.IsHumanTurn()
	resObj.PlayableIndices = p.playableIndices(g)
	resObj.HumanTeam = p.humanTeam(g)

	cfg := g.GetConfig()
	resObj.Config = controller.UnsunKarutaWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetDeals:   cfg.TargetDeals,
	}
	resObj.CurrentTrick = trickCardsToOutputWithFace(g.GetCurrentTrick(), unsunKarutaFace)
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// intsOrEmptySlice は nil を空スライスに直す (JSON で null を出さない)。
func intsOrEmptySlice(v []int) []int {
	if v == nil {
		return make([]int, 0)
	}
	return v
}

// humanTeam は人間のチームを返す (居なければ 0)。
func (p *UnsunKarutaWebPresenter) humanTeam(g interfaces.UnsunKarutaGame) int {
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if pl := g.GetPlayer(i); pl != nil && pl.GetIsHuman() {
			return domain.UnsunKarutaTeamOf(i)
		}
	}
	return 0
}

// playableIndices は人間が出せる札のインデックスを返す。
func (p *UnsunKarutaWebPresenter) playableIndices(g interfaces.UnsunKarutaGame) []int {
	if g.GetPhase() != domain.UnsunKarutaPhasePlay || !g.IsHumanTurn() {
		return make([]int, 0)
	}
	idx := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
	if idx == nil {
		return make([]int, 0)
	}
	return idx
}

// buildPlayersOutput は席の情報を構築する (人間のみ手札を公開)。
func (p *UnsunKarutaWebPresenter) buildPlayersOutput(g interfaces.UnsunKarutaGame) []*controller.UnsunKarutaWebOutputPlayer {
	dealer := g.GetDealerIdx()
	out := make([]*controller.UnsunKarutaWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		out = append(out, &controller.UnsunKarutaWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			Team:       domain.UnsunKarutaTeamOf(i),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutputWithFace(player, player.GetIsHuman(), unsunKarutaFace),
			TrickCount: player.GetTrickCount(),
			IsDealer:   i == dealer,
		})
	}
	return out
}

// buildMessage はフェーズ / 結果メッセージを構築する。
func (p *UnsunKarutaWebPresenter) buildMessage(g interfaces.UnsunKarutaGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		code, params := domain.ErrorMessageCode(lastErr)
		return lastErr.Error(), code, params
	}
	if g.GetGameEndFlag() {
		if g.GetWinnerTeam() < 0 {
			return "", "unsunkaruta.result.draw", nil
		}
		if g.GetWinnerTeam() == p.humanTeam(g) {
			return "", "unsunkaruta.result.humanWin", nil
		}
		return "", "unsunkaruta.result.cpuWin", nil
	}
	switch g.GetPhase() {
	case domain.UnsunKarutaPhasePlay:
		if len(g.GetCurrentTrick()) == 0 {
			return "", "unsunkaruta.playPhase.lead", nil
		}
		if g.IsMustFollow() {
			return "", "unsunkaruta.playPhase.mustFollow", nil
		}
		return "", "unsunkaruta.playPhase.free", nil
	case domain.UnsunKarutaPhaseTrickEnd:
		return "", "unsunkaruta.trickEnd", nil
	case domain.UnsunKarutaPhaseRoundEnd:
		return "", "unsunkaruta.roundEnd", nil
	}
	return "", "", nil
}

// HintOutput はヒント情報を JSON 出力する。
func (p *UnsunKarutaWebPresenter) HintOutput(g interfaces.UnsunKarutaGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = cardHint(hint.CardIndices, hint.Reason)
		resObj.MessageCode = "unsunkaruta.hintRequested"
	} else {
		resObj.MessageCode = "unsunkaruta.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput は棋譜を JSON 出力する。
func (p *UnsunKarutaWebPresenter) ActionLogOutput(g interfaces.UnsunKarutaGame) string {
	return actionLogOutputJSON(g)
}
