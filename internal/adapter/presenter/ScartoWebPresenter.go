//go:build !js || !wasm || extra4

package presenter

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// scartoFace は Scarto の 78 枚デッキ全札を手続き描画するための自己記述子を返す。スート札は
// 隅ラベルにランク (1-10, V/C/D/R)、グリフにスート記号 (♠♥♦♣) を用い、黒スート (スペード/
// クラブ) は black、赤スート (ハート/ダイヤ) は red。切り札は番号 1-21 をラベルに、グリフ
// "✦"、色 purple。エクスキューズはグリフ "★"、ラベル "Excuse"、色 gold。deck:"tarot" を
// 付与しフロントエンドの手続き描画パスへ切り替える (French Tarot と同一マッピング; ADR-0033)。
func scartoFace(card *domain.Card) *CardFace {
	if card == nil {
		return nil
	}
	switch card.GetDesign() {
	case domain.ScartoExcuseDesign:
		return &CardFace{Glyph: "★", Label: "Excuse", Color: "gold", Deck: "tarot"}
	case domain.ScartoTrumpDesign:
		label := strconv.Itoa(card.GetValue())
		return &CardFace{Glyph: "✦", Label: label, Color: "purple", Deck: "tarot"}
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
			Label: scartoRankLabel(card.GetValue()),
			Color: colors[card.GetDesign()],
			Deck:  "tarot",
		}
	}
}

// scartoRankLabel スート札のランクラベルを返す (1-10 は数字、コート札は V/C/D/R)。
func scartoRankLabel(value int) string {
	switch value {
	case 11:
		return "V" // Valet
	case 12:
		return "C" // Cavalier
	case 13:
		return "D" // Dame
	case 14:
		return "R" // Roi
	default:
		return strconv.Itoa(value)
	}
}

// ScartoWebPresenter スカルト (Scarto) のWebプレゼンタークラス
type ScartoWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *ScartoWebPresenter) Output(g interfaces.ScartoGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **Scarto.GetHint() の各フェーズを読んで、席を確かめていることを確認した。**
	// 他ゲームがそうだから、で済ませない —— Pinochle は見ていなかった (#4585)。
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = cardHint(hint.CardIndices, hint.Reason)
	}

	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *ScartoWebPresenter) buildBase(g interfaces.ScartoGame) *controller.ScartoWebOutput {
	resObj := new(controller.ScartoWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.LeadPlayerIdx = g.GetLeadPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.ScartoCount = g.GetScartoCount()
	resObj.Outcome = int(g.GetOutcome())
	resObj.Result = int(g.GetResult())
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerPlayer = g.GetWinnerPlayer()
	resObj.PlayerScores = g.GetPlayerScores()
	resObj.DealScores = g.GetDealScores()
	resObj.LastTrickWinner = -1
	resObj.IsHumanTurn = g.IsHumanTurn()
	resObj.IsHumanScarto = g.IsHumanScartoTurn()

	resObj.PlayableIndices = p.playableIndices(g)

	cfg := g.GetConfig()
	resObj.Config = controller.ScartoWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetDeals:   cfg.TargetDeals,
	}

	resObj.CurrentTrick = trickCardsToOutputWithFace(g.GetCurrentTrick(), scartoFace)
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// playableIndices 人間プレイヤーがプレイできるカードのインデックスを返す
func (p *ScartoWebPresenter) playableIndices(g interfaces.ScartoGame) []int {
	if g.GetPhase() != domain.ScartoPhasePlay || !g.IsHumanTurn() {
		return make([]int, 0)
	}
	idx := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
	if idx == nil {
		return make([]int, 0)
	}
	return idx
}

// buildPlayersOutput プレイヤー情報を構築 (人間のみ手札を公開、伏せたスカルトは非公開)
func (p *ScartoWebPresenter) buildPlayersOutput(g interfaces.ScartoGame) []*controller.ScartoWebOutputPlayer {
	scores := g.GetPlayerScores()
	dealer := g.GetDealerIdx()
	out := make([]*controller.ScartoWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		out = append(out, &controller.ScartoWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutputWithFace(player, player.GetIsHuman(), scartoFace),
			TrickCount: player.GetTrickCount(),
			CardPoints: g.GetCardPoints(i),
			Score:      scores[i],
			IsDealer:   i == dealer,
		})
	}
	return out
}

// buildMessage ゲーム結果/フェーズメッセージを構築
func (p *ScartoWebPresenter) buildMessage(g interfaces.ScartoGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		return p.winnerMessage(g)
	}
	switch g.GetPhase() {
	case domain.ScartoPhaseScarto:
		return "", "scarto.scartoPhase", nil
	case domain.ScartoPhasePlay:
		if len(g.GetCurrentTrick()) == 0 {
			return "", "scarto.playPhase.lead", nil
		}
		return "", "scarto.playPhase.follow", nil
	case domain.ScartoPhaseTrickEnd:
		return "", "scarto.trickEnd", nil
	case domain.ScartoPhaseRoundEnd:
		return "", scartoOutcomeMessageCode(g.GetOutcome()), nil
	}
	return "", "", nil
}

// scartoOutcomeMessageCode ディール結果に対応するメッセージコードを返す。
func scartoOutcomeMessageCode(o domain.ScartoOutcome) string {
	switch o {
	case domain.ScartoOutcomeWin:
		return "scarto.roundEnd.win"
	case domain.ScartoOutcomeLoss:
		return "scarto.roundEnd.loss"
	default:
		return "scarto.roundEnd"
	}
}

// winnerMessage 勝者プレイヤーメッセージを構築する
func (p *ScartoWebPresenter) winnerMessage(g interfaces.ScartoGame) (string, string, map[string]string) {
	winner := g.GetWinnerPlayer()
	if winner < 0 {
		return "ゲーム終了！ 引き分け！", "scarto.result.draw", nil
	}
	humanIdx := -1
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if player := g.GetPlayer(i); player != nil && player.GetIsHuman() {
			humanIdx = i
			break
		}
	}
	if humanIdx >= 0 && winner == humanIdx {
		return "ゲーム終了！ あなたの勝ち！", "scarto.result.humanWin", nil
	}
	params := map[string]string{"player": fmt.Sprintf("%d", winner)}
	return fmt.Sprintf("ゲーム終了！ プレイヤー%dの勝ち！", winner), "scarto.result.cpuWin", params
}

// HintOutput ヒント情報をJSON出力する
func (p *ScartoWebPresenter) HintOutput(g interfaces.ScartoGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = cardHint(hint.CardIndices, hint.Reason)
	}
	// **「頼んだヒントか」を CLI が見分けられるようにする。**このゲーム群の
	// `hintAvailable` は画面のラベルとして既に使われているので、別キーを出す (#4483)。
	if g.GetHint() != nil {
		resObj.MessageCode = "scarto.hintRequested"
	} else {
		resObj.MessageCode = "scarto.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *ScartoWebPresenter) ActionLogOutput(g interfaces.ScartoGame) string {
	return actionLogOutputJSON(g)
}
