//go:build !js || !wasm || extra

package presenter

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// ganjifaFace は 96 枚 8 スートデッキの札を手続き描画するための自己記述子を返す。
//
// **標準の cardToOutput には掛けられない。**あれは design 5..8 を既定の JOKER に
// 落とすので、弱いスート群の 48 枚が全部ジョーカーとして描かれる。
//
// 色はスートではなく**群**を表す。強い群 (1..4) は black、弱い群 (5..8) は blue。
// スートの区別はグリフが担うので、色にはこのゲームで唯一読み違えると詰む情報
// ——札の強さがどちら向きに読めるか——を割り当てる。
func ganjifaFace(card *domain.Card) *CardFace {
	if card == nil {
		return nil
	}
	design := card.GetDesign()
	inkColor := "blue"
	if domain.GanjifaIsStrongSuit(design) {
		inkColor = "black"
	}
	return &CardFace{
		Glyph: domain.GanjifaSuitGlyph(design),
		Label: strconv.Itoa(card.GetValue()),
		Color: inkColor,
		Deck:  "ganjifa",
	}
}

// GanjifaWebPresenter ガンジファのWebプレゼンタークラス
type GanjifaWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *GanjifaWebPresenter) Output(g interfaces.GanjifaGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **フェーズと手番はここでは見ない。**Ganjifa.GetHint() が自分で
	// 「人間の手番で、かつ行動を選べる状態か」を確かめて nil を返す。
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.WebOutputCardHint{
			CardIndices: hint.CardIndices,
			Reason:      hint.Reason,
		}
	}

	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *GanjifaWebPresenter) buildBase(g interfaces.GanjifaGame) *controller.GanjifaWebOutput {
	resObj := new(controller.GanjifaWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.LeadPlayerIdx = g.GetLeadPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.TrumpSuit = g.GetTrumpSuit()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerPlayer = g.GetWinnerPlayer()
	resObj.PlayerScores = g.GetPlayerScores()
	resObj.RoundTricks = g.GetRoundTricks()
	resObj.IsHumanTurn = g.IsHumanTurn()

	resObj.PlayableIndices = p.playableIndices(g)

	cfg := g.GetConfig()
	resObj.Config = controller.GanjifaWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetRounds:  cfg.TargetRounds,
	}

	resObj.CurrentTrick = trickCardsToOutputWithFace(g.GetCurrentTrick(), ganjifaFace)
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// playableIndices 人間プレイヤーがプレイできるカードのインデックスを返す
func (p *GanjifaWebPresenter) playableIndices(g interfaces.GanjifaGame) []int {
	if g.GetPhase() != domain.GanjifaPhasePlay || !g.IsHumanTurn() {
		return make([]int, 0)
	}
	idx := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
	if idx == nil {
		return make([]int, 0)
	}
	return idx
}

// buildPlayersOutput プレイヤー情報を構築
func (p *GanjifaWebPresenter) buildPlayersOutput(g interfaces.GanjifaGame) []*controller.GanjifaWebOutputPlayer {
	scores := g.GetPlayerScores()
	out := make([]*controller.GanjifaWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		out = append(out, &controller.GanjifaWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutputWithFace(player, player.GetIsHuman(), ganjifaFace),
			TrickCount: player.GetTrickCount(),
			Score:      scores[i],
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *GanjifaWebPresenter) buildMessage(g interfaces.GanjifaGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		return p.winnerMessage(g)
	}
	switch g.GetPhase() {
	case domain.GanjifaPhasePlay:
		if len(g.GetCurrentTrick()) == 0 {
			return "", "ganjifa.playPhase.lead", nil
		}
		return "", "ganjifa.playPhase.follow", nil
	case domain.GanjifaPhaseTrickEnd:
		return "", "ganjifa.trickEnd", nil
	case domain.GanjifaPhaseRoundEnd:
		return "", "ganjifa.roundEnd", nil
	}
	return "", "", nil
}

// winnerMessage 勝者プレイヤーメッセージを構築する
func (p *GanjifaWebPresenter) winnerMessage(g interfaces.GanjifaGame) (string, string, map[string]string) {
	winner := g.GetWinnerPlayer()
	humanIdx := -1
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if player := g.GetPlayer(i); player != nil && player.GetIsHuman() {
			humanIdx = i
			break
		}
	}
	if humanIdx >= 0 && winner == humanIdx {
		return "ゲーム終了！ あなたの勝ち！", "ganjifa.result.humanWin", nil
	}
	params := map[string]string{"player": fmt.Sprintf("%d", winner)}
	return fmt.Sprintf("ゲーム終了！ プレイヤー%dの勝ち！", winner), "ganjifa.result.cpuWin", params
}

// HintOutput ヒント情報をJSON出力する
func (p *GanjifaWebPresenter) HintOutput(g interfaces.GanjifaGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = &controller.WebOutputCardHint{
			CardIndices: hint.CardIndices,
			Reason:      hint.Reason,
		}
	}
	// **「頼んだヒントか」を CLI が見分けられるようにする。**このゲーム群の
	// `hintAvailable` は画面のラベルとして既に使われているので、別キーを出す (#4483)。
	if hint != nil {
		resObj.MessageCode = "ganjifa.hintRequested"
	} else {
		resObj.MessageCode = "ganjifa.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *GanjifaWebPresenter) ActionLogOutput(g interfaces.GanjifaGame) string {
	return actionLogOutputJSON(g)
}
