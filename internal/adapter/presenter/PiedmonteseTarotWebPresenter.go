//go:build !js || !wasm || extra4

package presenter

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// piedmonteseTarotFace は 78 枚デッキの札を手続き描画するための自己記述子を返す。
// スート札は隅ラベルにランク (1-10, V/C/D/R)、グリフにスート記号。切り札は番号と
// "✦"、Matto は "★"。`deck:"tarot"` でフロントの手続き描画へ切り替える (ADR-0033)。
func piedmonteseTarotFace(card *domain.Card) *CardFace {
	if card == nil {
		return nil
	}
	switch card.GetDesign() {
	case domain.Tarot78ExcuseDesign:
		return &CardFace{Glyph: "★", Label: "Matto", Color: "gold", Deck: "tarot"}
	case domain.Tarot78TrumpDesign:
		return &CardFace{Glyph: "✦", Label: strconv.Itoa(card.GetValue()), Color: "purple", Deck: "tarot"}
	default:
		glyphs := map[int]string{
			domain.CardDesignSpade:   "♠",
			domain.CardDesignClover:  "♣",
			domain.CardDesignHeart:   "♥",
			domain.CardDesignDiamond: "♦",
		}
		colors := map[int]string{
			domain.CardDesignSpade:   "black",
			domain.CardDesignClover:  "black",
			domain.CardDesignHeart:   "red",
			domain.CardDesignDiamond: "red",
		}
		return &CardFace{
			Glyph: glyphs[card.GetDesign()],
			Label: piedmonteseTarotRankLabel(card.GetValue()),
			Color: colors[card.GetDesign()],
			Deck:  "tarot",
		}
	}
}

// PiedmonteseTarotWebPresenter はピエモンテ・タロッコの Web プレゼンター。
type PiedmonteseTarotWebPresenter struct{}

// Output はゲーム状態を JSON 出力する。
func (p *PiedmonteseTarotWebPresenter) Output(g interfaces.PiedmonteseTarotGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	// **受動ヒントは Output() でも埋める。** HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない (#4483)。
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = cardHint(hint.CardIndices, hint.Reason)
	}
	return marshalOrError(resObj)
}

// buildBase は共通フィールドを構築する。
func (p *PiedmonteseTarotWebPresenter) buildBase(g interfaces.PiedmonteseTarotGame) *controller.PiedmonteseTarotWebOutput {
	resObj := new(controller.PiedmonteseTarotWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.TrickCount = g.HandSize()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.LeadPlayerIdx = g.GetLeadPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.ScartoCount = g.GetScartoCount()
	resObj.TalonSize = g.TalonSize()
	resObj.Outcome = int(g.GetOutcome())
	resObj.Result = int(g.GetResult())
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerPlayer = g.GetWinnerPlayer()
	resObj.PlayerScores = intsOrEmpty(g.GetPlayerScores())
	resObj.DealScores = intsOrEmpty(g.GetDealScores())
	resObj.LastTrickWinner = g.GetLastTrickWinner()
	resObj.IsHumanTurn = g.IsHumanTurn()
	resObj.IsHumanScarto = g.IsHumanScartoTurn()
	resObj.PlayableIndices = p.playableIndices(g)

	cfg := g.GetConfig()
	resObj.Config = controller.PiedmonteseTarotWebOutputConfig{
		Seats:         cfg.Seats,
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetDeals:   cfg.TargetDeals,
	}

	resObj.CurrentTrick = trickCardsToOutputWithFace(g.GetCurrentTrick(), piedmonteseTarotFace)
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// intsOrEmpty は nil を空スライスに直す (JSON で null を出さない)。
func intsOrEmpty(v []int) []int {
	if v == nil {
		return make([]int, 0)
	}
	return v
}

// playableIndices は人間が出せる札のインデックスを返す。
func (p *PiedmonteseTarotWebPresenter) playableIndices(g interfaces.PiedmonteseTarotGame) []int {
	if g.GetPhase() != domain.PiedmonteseTarotPhasePlay || !g.IsHumanTurn() {
		return make([]int, 0)
	}
	idx := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
	if idx == nil {
		return make([]int, 0)
	}
	return idx
}

// buildPlayersOutput は席の情報を構築する (人間のみ手札を公開)。
func (p *PiedmonteseTarotWebPresenter) buildPlayersOutput(g interfaces.PiedmonteseTarotGame) []*controller.PiedmonteseTarotWebOutputPlayer {
	scores := g.GetPlayerScores()
	dealer := g.GetDealerIdx()
	out := make([]*controller.PiedmonteseTarotWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		score := 0
		if i < len(scores) {
			score = scores[i]
		}
		thirds := g.GetCardThirds(i)
		out = append(out, &controller.PiedmonteseTarotWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutputWithFace(player, player.GetIsHuman(), piedmonteseTarotFace),
			TrickCount: player.GetTrickCount(),
			CardThirds: thirds,
			// **画面に出すのは読める形。** 1/3 単位の生の数をそのまま出すと、
			// 78 点のゲームで 234 という数字が並ぶ。
			CardPoints: domain.PiedmonteseTarotFormatThirds(thirds),
			Score:      score,
			IsDealer:   i == dealer,
		})
	}
	return out
}

// buildMessage はフェーズ/結果メッセージを構築する。
func (p *PiedmonteseTarotWebPresenter) buildMessage(g interfaces.PiedmonteseTarotGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		// **翻訳鍵を持つエラーはそのまま渡す。** 生の英文を message に入れると、
		// 画面には言語に関係なくその 1 行が出る。
		code, params := domain.ErrorMessageCode(lastErr)
		return lastErr.Error(), code, params
	}
	if g.GetGameEndFlag() {
		return p.winnerMessage(g)
	}
	switch g.GetPhase() {
	case domain.PiedmonteseTarotPhaseScarto:
		return "", "piedmontesetarot.scartoPhase", nil
	case domain.PiedmonteseTarotPhasePlay:
		if len(g.GetCurrentTrick()) == 0 {
			return "", "piedmontesetarot.playPhase.lead", nil
		}
		return "", "piedmontesetarot.playPhase.follow", nil
	case domain.PiedmonteseTarotPhaseTrickEnd:
		return "", "piedmontesetarot.trickEnd", nil
	case domain.PiedmonteseTarotPhaseRoundEnd:
		return "", piedmonteseTarotOutcomeMessageCode(g.GetOutcome()), nil
	}
	return "", "", nil
}

// piedmonteseTarotOutcomeMessageCode はディール結果のメッセージコードを返す。
func piedmonteseTarotOutcomeMessageCode(o domain.PiedmonteseTarotOutcome) string {
	switch o {
	case domain.PiedmonteseTarotOutcomeWin:
		return "piedmontesetarot.roundEnd.win"
	case domain.PiedmonteseTarotOutcomeLoss:
		return "piedmontesetarot.roundEnd.loss"
	default:
		return "piedmontesetarot.roundEnd"
	}
}

// winnerMessage は勝者メッセージを構築する。
func (p *PiedmonteseTarotWebPresenter) winnerMessage(g interfaces.PiedmonteseTarotGame) (string, string, map[string]string) {
	winner := g.GetWinnerPlayer()
	if winner < 0 {
		return "ゲーム終了！ 引き分け！", "piedmontesetarot.result.draw", nil
	}
	humanIdx := -1
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if player := g.GetPlayer(i); player != nil && player.GetIsHuman() {
			humanIdx = i
			break
		}
	}
	if humanIdx >= 0 && winner == humanIdx {
		return "ゲーム終了！ あなたの勝ち！", "piedmontesetarot.result.humanWin", nil
	}
	params := map[string]string{"player": fmt.Sprintf("%d", winner)}
	return fmt.Sprintf("ゲーム終了！ プレイヤー%dの勝ち！", winner), "piedmontesetarot.result.cpuWin", params
}

// HintOutput はヒント情報を JSON 出力する。
func (p *PiedmonteseTarotWebPresenter) HintOutput(g interfaces.PiedmonteseTarotGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = cardHint(hint.CardIndices, hint.Reason)
		resObj.MessageCode = "piedmontesetarot.hintRequested"
	} else {
		resObj.MessageCode = "piedmontesetarot.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput は棋譜を JSON 出力する。
func (p *PiedmonteseTarotWebPresenter) ActionLogOutput(g interfaces.PiedmonteseTarotGame) string {
	return actionLogOutputJSON(g)
}
