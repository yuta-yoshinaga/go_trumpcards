//go:build !js || !wasm || extra3

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// CirullaWebPresenter はチルッラの Web プレゼンター。
type CirullaWebPresenter struct{}

// Output はゲーム状態を JSON 出力する。
func (p *CirullaWebPresenter) Output(g interfaces.CirullaGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	if hint := g.GetHint(); hint != nil {
		resObj.HintHandIdx = hint.HandIdx
		resObj.HintCaptureIdxs = cirullaIntsOrEmpty(hint.CaptureIdxs)
		resObj.HintReason = hint.Reason
	}
	return marshalOrError(resObj)
}

// cirullaIntsOrEmpty は nil を空スライスに直す (JSON で null を出さない)。
//
// **同名の共通ヘルパは別タグの中にある。** そちらを呼ぶと extra3 の TinyGo
// ビルドだけが落ちる ── ホストの `go build ./...` は絶対に落ちない。
func cirullaIntsOrEmpty(v []int) []int {
	if v == nil {
		return make([]int, 0)
	}
	return v
}

// buildBase は共通フィールドを構築する。
func (p *CirullaWebPresenter) buildBase(g interfaces.CirullaGame) *controller.CirullaWebOutput {
	human := p.humanIdx(g)
	resObj := new(controller.CirullaWebOutput)
	resObj.Phase = g.GetPhase()
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.Table = cardsToOutput(g.GetTable())
	resObj.DeckRemaining = g.GetDeckRemaining()
	resObj.LastCapturer = g.GetLastCapturer()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()
	resObj.IsHumanTurn = g.IsHumanTurn()
	resObj.HintHandIdx = -1
	resObj.HintCaptureIdxs = make([]int, 0)

	// **どの札で何を取れるかは画面側で解かせない。** 合計 15 と合計一致と
	// アッソの総取りが混ざるので、規則をフロントに二重に持たせると必ずずれる。
	resObj.CaptureOptions = p.captureOptions(g, human)
	resObj.LastResult = p.lastResult(g)

	cfg := g.GetConfig()
	resObj.Config = controller.CirullaWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetScore:   cfg.TargetScore,
	}
	resObj.Players = p.buildPlayersOutput(g, human)
	return resObj
}

// humanIdx は人間の席を返す (居なければ 0)。
func (p *CirullaWebPresenter) humanIdx(g interfaces.CirullaGame) int {
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if pl := g.GetPlayer(i); pl != nil && pl.GetIsHuman() {
			return i
		}
	}
	return 0
}

// captureOptions は人間の手札ごとの捕獲候補を返す。
func (p *CirullaWebPresenter) captureOptions(g interfaces.CirullaGame, human int) [][][]int {
	out := make([][][]int, 0)
	player := g.GetPlayer(human)
	if player == nil || g.GetPhase() != domain.CirullaPhasePlay || !g.IsHumanTurn() {
		return out
	}
	for i := 0; i < player.GetCardsSize(); i++ {
		groups := g.GetCaptureOptions(human, i)
		if groups == nil {
			groups = make([][]int, 0)
		}
		out = append(out, groups)
	}
	return out
}

// lastResult は直前ラウンドの集計を出力形へ直す。
func (p *CirullaWebPresenter) lastResult(g interfaces.CirullaGame) *controller.CirullaWebOutputResult {
	res := g.GetLastResult()
	if res == nil {
		return nil
	}
	lines := make([]*controller.CirullaWebOutputScoreLine, 0, len(res.Lines))
	for _, l := range res.Lines {
		points := make([]int, domain.CirullaPlayerCnt)
		copy(points, l.Points[:])
		lines = append(lines, &controller.CirullaWebOutputScoreLine{Key: l.Key, Points: points})
	}
	totals := make([]int, domain.CirullaPlayerCnt)
	copy(totals, res.Totals[:])
	return &controller.CirullaWebOutputResult{Lines: lines, Totals: totals, SweptDenari: res.SweptDenari}
}

// buildPlayersOutput は席の情報を構築する (人間のみ手札を公開)。
func (p *CirullaWebPresenter) buildPlayersOutput(g interfaces.CirullaGame, human int) []*controller.CirullaWebOutputPlayer {
	dealer := g.GetDealerIdx()
	bonuses := g.GetLastBonus()
	out := make([]*controller.CirullaWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		denari, sette := 0, false
		for _, card := range player.GetCaptured() {
			if domain.CirullaIsDenari(card) {
				denari++
			}
			if domain.CirullaIsSetteBello(card) {
				sette = true
			}
		}
		bonus := ""
		if i < len(bonuses) {
			bonus = bonuses[i]
		}
		out = append(out, &controller.CirullaWebOutputPlayer{
			ID:            i,
			IsHuman:       player.GetIsHuman(),
			Cards:         playerCardsToOutput(player, i == human),
			CardCount:     player.GetCardsSize(),
			CapturedCount: len(player.GetCaptured()),
			DenariCount:   denari,
			HasSetteBello: sette,
			Scope:         player.GetScope(),
			BonusPoints:   player.GetBonusPoints(),
			Score:         player.GetScore(),
			IsDealer:      i == dealer,
			LastBonus:     bonus,
		})
	}
	return out
}

// buildMessage はフェーズ / 結果メッセージを構築する。
func (p *CirullaWebPresenter) buildMessage(g interfaces.CirullaGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		code, params := domain.ErrorMessageCode(lastErr)
		return lastErr.Error(), code, params
	}
	human := p.humanIdx(g)
	if g.GetGameEndFlag() {
		if g.GetWinnerIdx() == human {
			return "", "cirulla.result.humanWin", nil
		}
		return "", "cirulla.result.cpuWin", nil
	}
	switch g.GetPhase() {
	case domain.CirullaPhasePlay:
		// **取れる札があるなら置けない。** その区別を先に伝える。
		for _, groups := range p.captureOptions(g, human) {
			if len(groups) > 0 {
				return "", "cirulla.playPhase.canCapture", nil
			}
		}
		return "", "cirulla.playPhase", nil
	case domain.CirullaPhaseRoundEnd:
		return "", "cirulla.roundEnd", nil
	}
	return "", "", nil
}

// HintOutput はヒント情報を JSON 出力する。
func (p *CirullaWebPresenter) HintOutput(g interfaces.CirullaGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil && hint.HandIdx >= 0 {
		resObj.HintHandIdx = hint.HandIdx
		resObj.HintCaptureIdxs = cirullaIntsOrEmpty(hint.CaptureIdxs)
		resObj.HintReason = hint.Reason
		resObj.MessageCode = "cirulla.hintRequested"
	} else {
		resObj.MessageCode = "cirulla.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput は棋譜を JSON 出力する。
func (p *CirullaWebPresenter) ActionLogOutput(g interfaces.CirullaGame) string {
	return actionLogOutputJSON(g)
}
