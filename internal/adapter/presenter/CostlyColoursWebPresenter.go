//go:build !js || !wasm || extra

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// CostlyColoursWebPresenter はコストリー・カラーズの Web プレゼンター。
type CostlyColoursWebPresenter struct{}

// Output はゲーム状態を JSON 出力する。
func (p *CostlyColoursWebPresenter) Output(g interfaces.CostlyColoursGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	if hint := g.GetHint(); hint != nil {
		resObj.HintHandIdx = hint.HandIdx
		resObj.HintAcceptMog = hint.AcceptMog
		resObj.HintReason = hint.Reason
	}
	return marshalOrError(resObj)
}

// costlyColoursIntsOrEmpty は nil を空スライスに直す (JSON で null を出さない)。
//
// **同名の共通ヘルパは別タグの中にある。** そちらを呼ぶと extra の TinyGo
// ビルドだけが落ちる ── ホストの `go build ./...` は絶対に落ちない。
func costlyColoursIntsOrEmpty(v []int) []int {
	if v == nil {
		return make([]int, 0)
	}
	return v
}

// buildBase は共通フィールドを構築する。
func (p *CostlyColoursWebPresenter) buildBase(g interfaces.CostlyColoursGame) *controller.CostlyColoursWebOutput {
	human := p.humanIdx(g)
	resObj := new(controller.CostlyColoursWebOutput)
	resObj.Phase = g.GetPhase()
	resObj.DealNumber = g.GetDealNumber()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	// **表の 1 枚は見せる。** ショーの色役もトランプの J / 2 もこれ次第。
	if t := g.GetTurnUp(); t != nil {
		if out := cardsToOutput([]*domain.Card{t}); len(out) > 0 {
			resObj.TurnUp = out[0]
		}
	}
	resObj.Pile = cardsToOutput(g.GetPile())
	resObj.Total = g.GetTotal()
	resObj.WentOut = g.GetWentOut()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()
	resObj.IsHumanTurn = g.IsHumanTurn()
	resObj.HintHandIdx = -1

	// **出せる札はサーバが数える。** 31 を超える札を並べると押しても弾かれる。
	resObj.PlayableIdxs = make([]int, 0)
	if g.GetPhase() == domain.CostlyColoursPhasePlay && g.IsHumanTurn() {
		resObj.PlayableIdxs = costlyColoursIntsOrEmpty(g.PlayableIdxs(human))
	}
	resObj.LastResult = p.lastResult(g)

	cfg := g.GetConfig()
	resObj.Config = controller.CostlyColoursWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetScore:   cfg.TargetScore,
	}
	resObj.Players = p.buildPlayersOutput(g, human)
	return resObj
}

// humanIdx は人間の席を返す (居なければ 0)。
func (p *CostlyColoursWebPresenter) humanIdx(g interfaces.CostlyColoursGame) int {
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if pl := g.GetPlayer(i); pl != nil && pl.GetIsHuman() {
			return i
		}
	}
	return 0
}

// lastResult は直前のディールの集計を出力形へ直す。
func (p *CostlyColoursWebPresenter) lastResult(g interfaces.CostlyColoursGame) *controller.CostlyColoursWebOutputResult {
	res := g.GetLastResult()
	if res == nil {
		return nil
	}
	lines := make([]*controller.CostlyColoursWebOutputScoreLine, 0, len(res.Lines))
	for _, l := range res.Lines {
		points := make([]int, len(l.Points))
		copy(points, l.Points)
		lines = append(lines, &controller.CostlyColoursWebOutputScoreLine{Key: l.Key, Points: points})
	}
	totals := make([]int, domain.CostlyColoursPlayerCnt)
	copy(totals, res.Totals[:])
	combos := make([]string, domain.CostlyColoursPlayerCnt)
	copy(combos, res.Combos[:])
	return &controller.CostlyColoursWebOutputResult{Lines: lines, Totals: totals, Combos: combos}
}

// buildPlayersOutput は席の情報を構築する (人間のみ手札を公開)。
func (p *CostlyColoursWebPresenter) buildPlayersOutput(g interfaces.CostlyColoursGame, human int) []*controller.CostlyColoursWebOutputPlayer {
	dealer := g.GetDealerIdx()
	out := make([]*controller.CostlyColoursWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		// **出した札は伏せない。** 相手のぶんも数え上げで場に出ているので、
		// 隠すとショーの内訳が追えなくなるだけ ── 伏せるのは手札だけ。
		played := cardsToOutput(player.GetPlayed())
		out = append(out, &controller.CostlyColoursWebOutputPlayer{
			ID:        i,
			IsHuman:   player.GetIsHuman(),
			Cards:     playerCardsToOutput(player, i == human),
			CardCount: player.GetCardsSize(),
			Played:    played,
			Score:     player.GetScore(),
			IsDealer:  i == dealer,
			MoggedIn:  player.IsMoggedIn(),
		})
	}
	return out
}

// buildMessage はフェーズ / 結果メッセージを構築する。
func (p *CostlyColoursWebPresenter) buildMessage(g interfaces.CostlyColoursGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		code, params := domain.ErrorMessageCode(lastErr)
		return lastErr.Error(), code, params
	}
	human := p.humanIdx(g)
	if g.GetGameEndFlag() {
		if g.GetWinnerIdx() == human {
			return "", "costlycolours.result.humanWin", nil
		}
		return "", "costlycolours.result.cpuWin", nil
	}
	switch g.GetPhase() {
	case domain.CostlyColoursPhaseMog:
		return "", "costlycolours.mogPhase", nil
	case domain.CostlyColoursPhasePlay:
		if !g.IsHumanTurn() {
			return "", "costlycolours.playPhase", nil
		}
		// **出せる札が無いなら「ゴー」。** その区別を先に伝える。
		if len(g.PlayableIdxs(human)) == 0 {
			return "", "costlycolours.playPhase.go", nil
		}
		return "", "costlycolours.playPhase", nil
	case domain.CostlyColoursPhaseShow:
		return "", "costlycolours.showPhase", nil
	}
	return "", "", nil
}

// HintOutput はヒント情報を JSON 出力する。
func (p *CostlyColoursWebPresenter) HintOutput(g interfaces.CostlyColoursGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil && hint.Reason != "none" {
		resObj.HintHandIdx = hint.HandIdx
		resObj.HintAcceptMog = hint.AcceptMog
		resObj.HintReason = hint.Reason
		resObj.MessageCode = "costlycolours.hintRequested"
	} else {
		resObj.MessageCode = "costlycolours.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput は棋譜を JSON 出力する。
func (p *CostlyColoursWebPresenter) ActionLogOutput(g interfaces.CostlyColoursGame) string {
	return actionLogOutputJSON(g)
}
