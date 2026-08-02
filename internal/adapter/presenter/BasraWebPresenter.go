//go:build !js || !wasm || extra3

package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// BasraWebPresenter はバスラ (Basra) の Web プレゼンタークラス。
type BasraWebPresenter struct{}

// Output はゲーム状態を JSON 出力する。
func (p *BasraWebPresenter) Output(g interfaces.BasraGame, lastErr error) string {
	resObj := p.buildBase(g)
	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if g.GetPhase() == domain.BasraPhaseGameEnd {
		resObj.Message = p.buildResultMessage(g)
		resObj.MessageCode = "basra.result.scores"
		resObj.MessageParams = map[string]string{"scores": p.encodeScoresParam(g)}
	}
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **Basra.GetHint() はプレイ中かつ currentTurn == 人間席に限る。**
	// 他ゲームがそうだから、で済ませない —— Pinochle は見ていなかった (#4585)。
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.BasraWebOutputHint{
			CardIndices:  hint.CardIndices,
			TableIndices: hint.TableIndices,
			Reason:       hint.Reason,
		}
	}

	return marshalOrError(resObj)
}

// buildBase は基本フィールドを埋めた出力オブジェクトを生成する。
func (p *BasraWebPresenter) buildBase(g interfaces.BasraGame) *controller.BasraWebOutput {
	resObj := new(controller.BasraWebOutput)
	resObj.Players = make([]*controller.BasraWebOutputPlayer, 0)
	resObj.TableCards = cardsToOutputOrEmpty(g.GetTableCards())
	resObj.PlayableIndices = make([]int, 0)
	resObj.CaptureOptions = make(map[int][]int)
	resObj.Winners = make([]int, 0)

	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.CurrentTurn = g.GetCurrentTurn()
	resObj.LastCaptureIdx = g.GetLastCaptureIdx()
	resObj.RemainingDeck = g.GetRemainingDeck()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.IsHumanTurn = g.IsHumanTurn()

	if w := g.GetWinners(); w != nil {
		resObj.Winners = w
	}

	cfg := g.GetConfig()
	resObj.Config = controller.BasraWebConfigOutput{
		CpuDifficulty: int(cfg.CpuDifficulty),
	}

	if g.GetPhase() == domain.BasraPhasePlay && g.IsHumanTurn() {
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
		resObj.Players = append(resObj.Players, &controller.BasraWebOutputPlayer{
			ID:            i,
			IsHuman:       player.GetIsHuman(),
			CardCount:     player.GetCardsSize(),
			Cards:         playerCardsToOutput(player, player.GetIsHuman()),
			CapturedCount: player.CapturedCount(),
			BasraCount:    player.GetBasraCount(),
			Score:         player.GetScore(),
		})
	}

	if det := g.GetLastDealDetail(); det != nil {
		resObj.LastDealDetail = &controller.BasraWebOutputScoreDetail{
			Cards:            det.Cards,
			Aces:             det.Aces,
			Basras:           det.Basras,
			HasSevenDiamonds: det.HasSevenDiamonds,
			HasTenDiamonds:   det.HasTenDiamonds,
			MostCards:        det.MostCards,
			Gained:           det.Gained,
		}
	}
	return resObj
}

// encodeScoresParam は最終得点を "0:12,1:3" 形式の文字列に詰める。
func (p *BasraWebPresenter) encodeScoresParam(g interfaces.BasraGame) string {
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
func (p *BasraWebPresenter) buildResultMessage(g interfaces.BasraGame) string {
	msg := "Game over. "
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		name := fmt.Sprintf("CPU %d", i)
		if player.GetIsHuman() {
			name = "You"
		}
		msg += fmt.Sprintf("%s:%d ", name, player.GetScore())
	}
	return msg
}

// HintOutput はヒント情報を JSON 出力する。
func (p *BasraWebPresenter) HintOutput(g interfaces.BasraGame) string {
	resObj := p.buildBase(g)
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.BasraWebOutputHint{
			CardIndices:  hint.CardIndices,
			TableIndices: hint.TableIndices,
			Reason:       hint.Reason,
		}
	}
	// **「頼んだヒントか」を CLI が見分けられるようにする。**このゲーム群の
	// `hintAvailable` は画面のラベルとして既に使われているので、別キーを出す (#4483)。
	if g.GetHint() != nil {
		resObj.MessageCode = "basra.hintRequested"
	} else {
		resObj.MessageCode = "basra.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput は棋譜を JSON 出力する。
func (p *BasraWebPresenter) ActionLogOutput(g interfaces.BasraGame) string {
	return actionLogOutputJSON(g)
}
