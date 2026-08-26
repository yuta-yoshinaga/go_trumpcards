//go:build !js || !wasm || classic

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// DilotiWebPresenter はディロティの Web プレゼンター。
type DilotiWebPresenter struct{}

// Output はゲーム状態を JSON 出力する。
func (p *DilotiWebPresenter) Output(g interfaces.DilotiGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	p.applyHint(g, resObj)
	return marshalOrError(resObj)
}

// dilotiIntsOrEmpty は nil を空スライスに直す (JSON で null を出さない)。
//
// **同名の共通ヘルパは別タグの中にある。** そちらを呼ぶと classic の TinyGo
// ビルドだけが落ちる ── ホストの `go build ./...` は絶対に落ちない。
func dilotiIntsOrEmpty(v []int) []int {
	if v == nil {
		return make([]int, 0)
	}
	return v
}

// applyHint はヒントを出力へ写す。
func (p *DilotiWebPresenter) applyHint(g interfaces.DilotiGame, resObj *controller.DilotiWebOutput) {
	hint := g.GetHint()
	if hint == nil {
		return
	}
	resObj.HintHandIdx = hint.Move.HandIdx
	resObj.HintAction = hint.Move.Action
	resObj.HintTableIdxs = dilotiIntsOrEmpty(hint.Move.TableIdxs)
	resObj.HintDeclIdxs = dilotiIntsOrEmpty(hint.Move.DeclIdxs)
	resObj.HintDeclValue = hint.Move.Value
	resObj.HintReason = hint.Move.Reason
}

// buildBase は共通フィールドを構築する。
func (p *DilotiWebPresenter) buildBase(g interfaces.DilotiGame) *controller.DilotiWebOutput {
	human := p.humanIdx(g)
	resObj := new(controller.DilotiWebOutput)
	resObj.Phase = g.GetPhase()
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.Table = cardsToOutput(g.GetTable())
	resObj.Declarations = p.declarations(g)
	resObj.DeckRemaining = g.GetDeckRemaining()
	resObj.LastCapturer = g.GetLastCapturer()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()
	resObj.IsHumanTurn = g.IsHumanTurn()
	resObj.HintHandIdx = -1
	resObj.HintTableIdxs = make([]int, 0)
	resObj.HintDeclIdxs = make([]int, 0)

	// **どの札で何ができるかは画面側で解かせない。** 同ランク・合計一致・宣言が
	// 絡むので、規則をフロントに二重に持たせると必ずずれる。
	resObj.TakeOptions, resObj.DeclareOptions, resObj.CanTrail = p.moveOptions(g, human)
	resObj.LastResult = p.lastResult(g)

	cfg := g.GetConfig()
	resObj.Config = controller.DilotiWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetScore:   cfg.TargetScore,
	}
	resObj.Players = p.buildPlayersOutput(g, human)
	return resObj
}

// humanIdx は人間の席を返す (居なければ 0)。
func (p *DilotiWebPresenter) humanIdx(g interfaces.DilotiGame) int {
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if pl := g.GetPlayer(i); pl != nil && pl.GetIsHuman() {
			return i
		}
	}
	return 0
}

// declarations は場に積まれた宣言を出力形へ直す。
func (p *DilotiWebPresenter) declarations(g interfaces.DilotiGame) []*controller.DilotiWebOutputDeclaration {
	out := make([]*controller.DilotiWebOutputDeclaration, 0)
	for _, d := range g.GetDeclarations() {
		if d == nil {
			continue
		}
		groups := make([][]*controller.WebOutputCard, 0, len(d.Groups))
		for _, grp := range d.Groups {
			groups = append(groups, cardsToOutput(grp))
		}
		out = append(out, &controller.DilotiWebOutputDeclaration{
			OwnerIdx: d.OwnerIdx, Value: d.Value, Groups: groups, IsGroup: d.IsGroup,
		})
	}
	return out
}

// moveOptions は人間の手札ごとの取り手・宣言候補・場に置けるかを返す。
func (p *DilotiWebPresenter) moveOptions(g interfaces.DilotiGame, human int) (
	[][]*controller.DilotiWebOutputTake, [][]*controller.DilotiWebOutputDeclCandidate, []bool) {
	takes := make([][]*controller.DilotiWebOutputTake, 0)
	decls := make([][]*controller.DilotiWebOutputDeclCandidate, 0)
	trail := make([]bool, 0)
	player := g.GetPlayer(human)
	if player == nil || g.GetPhase() != domain.DilotiPhasePlay || !g.IsHumanTurn() {
		return takes, decls, trail
	}
	for i := 0; i < player.GetCardsSize(); i++ {
		t := make([]*controller.DilotiWebOutputTake, 0)
		for _, opt := range g.GetTakeOptions(human, i) {
			t = append(t, &controller.DilotiWebOutputTake{
				TableIdxs: dilotiIntsOrEmpty(opt.TableIdxs),
				DeclIdxs:  dilotiIntsOrEmpty(opt.DeclIdxs),
			})
		}
		takes = append(takes, t)

		d := make([]*controller.DilotiWebOutputDeclCandidate, 0)
		for _, cand := range g.GetDeclareOptions(human, i) {
			d = append(d, &controller.DilotiWebOutputDeclCandidate{
				Value: cand.Value, TableIdxs: dilotiIntsOrEmpty(cand.TableIdxs),
			})
		}
		decls = append(decls, d)
		trail = append(trail, g.CanTrail(human, i))
	}
	return takes, decls, trail
}

// lastResult は直前の局の集計を出力形へ直す。
func (p *DilotiWebPresenter) lastResult(g interfaces.DilotiGame) *controller.DilotiWebOutputResult {
	res := g.GetLastResult()
	if res == nil {
		return nil
	}
	lines := make([]*controller.DilotiWebOutputScoreLine, 0, len(res.Lines))
	for _, l := range res.Lines {
		points := make([]int, len(l.Points))
		copy(points, l.Points)
		lines = append(lines, &controller.DilotiWebOutputScoreLine{Key: l.Key, Points: points})
	}
	totals := make([]int, domain.DilotiPlayerCnt)
	copy(totals, res.Totals[:])
	counts := make([]int, domain.DilotiPlayerCnt)
	copy(counts, res.CardCounts[:])
	xeris := make([]int, domain.DilotiPlayerCnt)
	copy(xeris, res.Xeris[:])
	return &controller.DilotiWebOutputResult{
		Lines: lines, Totals: totals, CardCounts: counts, Xeris: xeris,
	}
}

// buildPlayersOutput は席の情報を構築する (人間のみ手札を公開)。
func (p *DilotiWebPresenter) buildPlayersOutput(g interfaces.DilotiGame, human int) []*controller.DilotiWebOutputPlayer {
	dealer := g.GetDealerIdx()
	out := make([]*controller.DilotiWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		out = append(out, &controller.DilotiWebOutputPlayer{
			ID:            i,
			IsHuman:       player.GetIsHuman(),
			Cards:         playerCardsToOutput(player, i == human),
			CardCount:     player.GetCardsSize(),
			CapturedCount: len(player.GetCaptured()),
			Xeri:          player.GetXeri(),
			Score:         player.GetScore(),
			IsDealer:      i == dealer,
		})
	}
	return out
}

// buildMessage はフェーズ / 結果メッセージを構築する。
func (p *DilotiWebPresenter) buildMessage(g interfaces.DilotiGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		code, params := domain.ErrorMessageCode(lastErr)
		return lastErr.Error(), code, params
	}
	human := p.humanIdx(g)
	if g.GetGameEndFlag() {
		if g.GetWinnerIdx() == human {
			return "", "diloti.result.humanWin", nil
		}
		return "", "diloti.result.cpuWin", nil
	}
	switch g.GetPhase() {
	case domain.DilotiPhasePlay:
		// **場を 1 枚で払えるならそれが最優先。** クセリ 1 回で 10 点あり、
		// 固定点の合計 11 点に匹敵するので、見落とすと局が決まる。
		if p.canXeri(g, human) {
			return "", "diloti.playPhase.canXeri", nil
		}
		return "", "diloti.playPhase", nil
	case domain.DilotiPhaseRoundEnd:
		return "", "diloti.roundEnd", nil
	}
	return "", "", nil
}

// canXeri は人間が 1 枚で場を払える手を持つかを返す。
func (p *DilotiWebPresenter) canXeri(g interfaces.DilotiGame, human int) bool {
	table, decls := len(g.GetTable()), len(g.GetDeclarations())
	if table+decls == 0 {
		return false
	}
	player := g.GetPlayer(human)
	if player == nil {
		return false
	}
	for i := 0; i < player.GetCardsSize(); i++ {
		for _, opt := range g.GetTakeOptions(human, i) {
			if len(opt.TableIdxs) == table && len(opt.DeclIdxs) == decls {
				return true
			}
		}
	}
	return false
}

// HintOutput はヒント情報を JSON 出力する。
func (p *DilotiWebPresenter) HintOutput(g interfaces.DilotiGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil && hint.Move.HandIdx >= 0 {
		p.applyHint(g, resObj)
		resObj.MessageCode = "diloti.hintRequested"
	} else {
		resObj.MessageCode = "diloti.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput は棋譜を JSON 出力する。
func (p *DilotiWebPresenter) ActionLogOutput(g interfaces.DilotiGame) string {
	return actionLogOutputJSON(g)
}
