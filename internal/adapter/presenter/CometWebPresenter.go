//go:build !js || !wasm || solo

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// CometWebPresenter はコメットの Web プレゼンター。
type CometWebPresenter struct{}

// Output はゲーム状態を JSON 出力する。
func (p *CometWebPresenter) Output(g interfaces.CometGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	if hint := g.GetHint(); hint != nil {
		resObj.HintHandIdx = hint.HandIdx
		resObj.HintReason = hint.Reason
	}
	return marshalOrError(resObj)
}

// cometIntsOrEmpty は nil を空スライスに直す (JSON で null を出さない)。
//
// **同名の共通ヘルパは別タグの中にある。** そちらを呼ぶと solo の TinyGo
// ビルドだけが落ちる ── ホストの `go build ./...` は絶対に落ちない。
func cometIntsOrEmpty(v []int) []int {
	if v == nil {
		return make([]int, 0)
	}
	return v
}

// buildBase は共通フィールドを構築する。
func (p *CometWebPresenter) buildBase(g interfaces.CometGame) *controller.CometWebOutput {
	human := p.humanIdx(g)
	resObj := new(controller.CometWebOutput)
	resObj.Phase = g.GetPhase()
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.Pile = cardsToOutput(g.GetPile())
	resObj.Need = g.GetNeed()
	resObj.DeadCount = g.GetDeadCount()
	resObj.LastPlayerIdx = g.GetLastPlayerIdx()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()
	resObj.IsHumanTurn = g.IsHumanTurn()
	resObj.HintHandIdx = -1

	// **出せる札はサーバが数える。** コメットがどのランクの代わりにもなるので、
	// 画面側で組み直すと必ずずれる。手番でなければ空。
	resObj.PlayableIdxs = make([]int, 0)
	if g.IsHumanTurn() {
		resObj.PlayableIdxs = cometIntsOrEmpty(g.PlayableIdxs(human))
	}
	resObj.LastResult = p.lastResult(g)

	cfg := g.GetConfig()
	resObj.Config = controller.CometWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		Players:       cfg.Players,
		TargetScore:   cfg.TargetScore,
	}
	resObj.Players = p.buildPlayersOutput(g, human)
	return resObj
}

// humanIdx は人間の席を返す (居なければ 0)。
func (p *CometWebPresenter) humanIdx(g interfaces.CometGame) int {
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if pl := g.GetPlayer(i); pl != nil && pl.GetIsHuman() {
			return i
		}
	}
	return 0
}

// lastResult は直前の局の集計を出力形へ直す。
func (p *CometWebPresenter) lastResult(g interfaces.CometGame) *controller.CometWebOutputResult {
	res := g.GetLastResult()
	if res == nil {
		return nil
	}
	return &controller.CometWebOutputResult{
		WinnerIdx:     res.WinnerIdx,
		CardsLeft:     cometIntsOrEmpty(res.CardsLeft),
		Gained:        cometIntsOrEmpty(res.Gained),
		UnplayedKings: res.UnplayedKings,
		HeldWildIdx:   res.HeldWildIdx,
	}
}

// buildPlayersOutput は席の情報を構築する (人間のみ手札を公開)。
func (p *CometWebPresenter) buildPlayersOutput(g interfaces.CometGame, human int) []*controller.CometWebOutputPlayer {
	dealer := g.GetDealerIdx()
	out := make([]*controller.CometWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		out = append(out, &controller.CometWebOutputPlayer{
			ID:        i,
			IsHuman:   player.GetIsHuman(),
			Cards:     playerCardsToOutput(player, i == human),
			CardCount: player.GetCardsSize(),
			Score:     player.GetScore(),
			IsDealer:  i == dealer,
		})
	}
	return out
}

// buildMessage はフェーズ / 結果メッセージを構築する。
func (p *CometWebPresenter) buildMessage(g interfaces.CometGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		code, params := domain.ErrorMessageCode(lastErr)
		return lastErr.Error(), code, params
	}
	human := p.humanIdx(g)
	if g.GetGameEndFlag() {
		if g.GetWinnerIdx() == human {
			return "", "comet.result.humanWin", nil
		}
		return "", "comet.result.cpuWin", nil
	}
	switch g.GetPhase() {
	case domain.CometPhasePlay:
		if !g.IsHumanTurn() {
			return "", "comet.playPhase", nil
		}
		// **出せる札が無いならパスしかない。** その区別を先に伝えないと、
		// 押せるボタンを探して盤面を睨むことになる。
		if len(g.PlayableIdxs(human)) == 0 {
			return "", "comet.playPhase.mustPass", nil
		}
		if g.GetNeed() <= 0 {
			return "", "comet.playPhase.lead", nil
		}
		return "", "comet.playPhase", nil
	case domain.CometPhaseRoundEnd:
		return "", "comet.roundEnd", nil
	}
	return "", "", nil
}

// HintOutput はヒント情報を JSON 出力する。
func (p *CometWebPresenter) HintOutput(g interfaces.CometGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil && hint.HandIdx >= 0 {
		resObj.HintHandIdx = hint.HandIdx
		resObj.HintReason = hint.Reason
		resObj.MessageCode = "comet.hintRequested"
	} else {
		resObj.MessageCode = "comet.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput は棋譜を JSON 出力する。
func (p *CometWebPresenter) ActionLogOutput(g interfaces.CometGame) string {
	return actionLogOutputJSON(g)
}
