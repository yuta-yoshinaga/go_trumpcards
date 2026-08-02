//go:build !js || !wasm || extra3

package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// LooWebPresenter はルー (Loo) の Web プレゼンタークラス。
type LooWebPresenter struct{}

// Output はゲーム状態を JSON 出力する。
func (p *LooWebPresenter) Output(g interfaces.LooGame, lastErr error) string {
	resObj := p.buildBase(g)
	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if g.GetPhase() == domain.LooPhaseRoundEnd {
		resObj.Message = p.buildResultMessage(g)
		resObj.MessageCode = "loo.result.chips"
		resObj.MessageParams = map[string]string{"chips": p.encodeChipsParam(g)}
	}
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **Loo.GetHint() の各フェーズを読んで、席を確かめていることを確認した。**
	// 他ゲームがそうだから、で済ませない —— Pinochle は見ていなかった (#4585)。
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.LooWebOutputHint{
			CardIndices: hint.CardIndices,
			Decision:    hint.Decision,
			Reason:      hint.Reason,
		}
	}

	return marshalOrError(resObj)
}

// buildBase は基本フィールドを埋めた出力オブジェクトを生成する。
func (p *LooWebPresenter) buildBase(g interfaces.LooGame) *controller.LooWebOutput {
	resObj := new(controller.LooWebOutput)
	resObj.Players = make([]*controller.LooWebOutputPlayer, 0)
	resObj.CurrentTrick = make([]*controller.WebOutputTrickCard, 0)
	resObj.LastTrick = make([]*controller.WebOutputTrickCard, 0)
	resObj.PlayableIndices = make([]int, 0)

	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.TotalTricks = domain.LooTrickCount
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.CurrentTurn = g.GetCurrentTurn()
	resObj.DecidePlayerIdx = g.GetDecidePlayerIdx()
	resObj.TrumpSuit = g.GetTrumpSuit()
	resObj.Pot = g.GetPot()
	resObj.PotStart = g.GetPotStart()
	resObj.LastTrickWinner = g.GetLastTrickWinner()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.IsHumanTurn = g.IsHumanTurn()

	if tu := g.GetTurnUp(); tu != nil {
		resObj.TurnUp = cardToOutput(tu)
	}

	cfg := g.GetConfig()
	resObj.Config = controller.LooWebConfigOutput{
		CpuDifficulty: int(cfg.CpuDifficulty),
		Ante:          cfg.Ante,
	}

	resObj.CurrentTrick = looTrickToOutput(g.GetCurrentTrick())
	resObj.LastTrick = looTrickToOutput(g.GetLastTrick())
	resObj.PlayableIndices = p.playableIndices(g)

	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		resObj.Players = append(resObj.Players, &controller.LooWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, player.GetIsHuman()),
			TrickCount: player.GetTrickCount(),
			Playing:    player.GetPlaying(),
			Chips:      player.GetChips(),
		})
	}

	if det := g.GetLastDealDetail(); det != nil {
		resObj.LastDealDetail = &controller.LooWebOutputDealDetail{
			PotStart:  det.PotStart,
			TrumpSuit: det.TrumpSuit,
			Playing:   det.Playing,
			Tricks:    det.Tricks,
			Gained:    det.Gained,
			Looed:     det.Looed,
			PotCarry:  det.PotCarry,
		}
	}
	return resObj
}

// playableIndices は人間プレイヤーがプレイできるカードのインデックスを返す。
func (p *LooWebPresenter) playableIndices(g interfaces.LooGame) []int {
	if g.GetPhase() != domain.LooPhasePlay || !g.IsHumanTurn() {
		return make([]int, 0)
	}
	idx := g.GetPlayableIndices(g.GetCurrentTurn())
	if idx == nil {
		return make([]int, 0)
	}
	return idx
}

// looTrickToOutput はトリックを WebOutput 表現に変換する。
func looTrickToOutput(trick []*domain.TrickCard) []*controller.WebOutputTrickCard {
	return buildTrickCards(trick, func(tc *domain.TrickCard) *controller.WebOutputTrickCard {
		if tc == nil {
			return nil
		}
		return &controller.WebOutputTrickCard{PlayerIdx: tc.PlayerIdx, Card: cardToOutput(tc.Card)}
	})
}

// encodeChipsParam は最終チップを "0:12,1:-3" 形式の文字列に詰める。
func (p *LooWebPresenter) encodeChipsParam(g interfaces.LooGame) string {
	parts := make([]string, 0, g.GetPlayerCnt())
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		parts = append(parts, fmt.Sprintf("%d:%d", i, player.GetChips()))
	}
	return strings.Join(parts, ",")
}

// buildResultMessage はディール終了時のフォールバック (英語) メッセージ。
func (p *LooWebPresenter) buildResultMessage(g interfaces.LooGame) string {
	msg := "Deal over. "
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		name := fmt.Sprintf("CPU %d", i)
		if player.GetIsHuman() {
			name = "You"
		}
		msg += fmt.Sprintf("%s:%d ", name, player.GetChips())
	}
	return msg
}

// HintOutput はヒント情報を JSON 出力する。
func (p *LooWebPresenter) HintOutput(g interfaces.LooGame) string {
	resObj := p.buildBase(g)
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.LooWebOutputHint{
			CardIndices: hint.CardIndices,
			Decision:    hint.Decision,
			Reason:      hint.Reason,
		}
	}
	// **「頼んだヒントか」を CLI が見分けられるようにする。**このゲーム群の
	// `hintAvailable` は画面のラベルとして既に使われているので、別キーを出す (#4483)。
	if g.GetHint() != nil {
		resObj.MessageCode = "loo.hintRequested"
	} else {
		resObj.MessageCode = "loo.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput は棋譜を JSON 出力する。
func (p *LooWebPresenter) ActionLogOutput(g interfaces.LooGame) string {
	return actionLogOutputJSON(g)
}
