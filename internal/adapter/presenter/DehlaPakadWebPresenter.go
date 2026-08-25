//go:build !js || !wasm || extra

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// DehlaPakadWebPresenter はデーラ・パカドの Web プレゼンター。
type DehlaPakadWebPresenter struct{}

// Output はゲーム状態を JSON 出力する。
func (p *DehlaPakadWebPresenter) Output(g interfaces.DehlaPakadGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	// 受動ヒントは Output でも埋める (#4483)。
	if hint := g.GetHint(); hint != nil {
		if len(hint.CardIndices) > 0 {
			resObj.Hint = cardHint(hint.CardIndices, hint.Reason)
		}
		resObj.HintTrumpSuit = hint.TrumpSuit
	}
	return marshalOrError(resObj)
}

// buildBase は共通フィールドを構築する。
func (p *DehlaPakadWebPresenter) buildBase(g interfaces.DehlaPakadGame) *controller.DehlaPakadWebOutput {
	resObj := new(controller.DehlaPakadWebOutput)
	resObj.Phase = g.GetPhase()
	resObj.HandNumber = g.GetHandNumber()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.TrumpChooserIdx = g.GetTrumpChooserIdx()
	resObj.TrumpSuit = g.GetTrumpSuit()
	resObj.TrumpSuitName = domain.DehlaPakadSuitName(g.GetTrumpSuit())
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.TrickCount = domain.DehlaPakadTrickCount
	resObj.CurrentPlayerIdx = g.GetCurrentTurn()
	resObj.LeadPlayerIdx = g.GetLeadPlayerIdx()
	resObj.CurrentTrick = trickCardsToOutput(g.GetCurrentTrick())
	resObj.LastTrick = trickCardsToOutput(g.GetLastTrick())
	resObj.LastTrickWinner = g.GetLastTrickWinner()
	// **山を引き取れるかは「直前も自分が取ったか」で決まる。** 画面に出さないと、
	// なぜいまトリックを取りにいく価値があるのかが読めない。
	resObj.PrevTrickWinner = g.GetPrevTrickWinner()
	resObj.CentrePileCount = len(g.GetCentrePile())
	resObj.CentrePileTens = g.GetCentrePileTens()
	resObj.PlayableIndices = p.playableIndices(g)
	resObj.TeamTens = dehlaPakadIntsOrEmpty(g.GetTeamTens())
	resObj.TeamKots = dehlaPakadIntsOrEmpty(g.GetTeamKots())
	resObj.HumanTeam = p.humanTeam(g)
	resObj.StreakTeam = g.GetStreakTeam()
	resObj.StreakCount = g.GetStreakCount()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerTeam = g.GetWinnerTeam()
	resObj.IsHumanTurn = g.IsHumanTurn()
	resObj.IsTrumpPhase = g.GetPhase() == domain.DehlaPakadPhaseSelectTrump
	resObj.HintTrumpSuit = -1

	resObj.LastHand = dehlaPakadHandToOutput(g.GetLastResult())
	resObj.HandHistory = make([]*controller.DehlaPakadWebOutputHand, 0)
	for _, h := range g.GetHandHistory() {
		if out := dehlaPakadHandToOutput(h); out != nil {
			resObj.HandHistory = append(resObj.HandHistory, out)
		}
	}

	cfg := g.GetConfig()
	resObj.Config = controller.DehlaPakadWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetKots:    cfg.TargetKots,
	}
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// dehlaPakadIntsOrEmpty は nil を空スライスに直す (JSON で null を出さない)。
//
// **同名の共通ヘルパは別タグの中にある。** そちらを呼ぶと extra の TinyGo
// ビルドだけが落ちる ── ホストの `go build ./...` は絶対に落ちない。
func dehlaPakadIntsOrEmpty(v []int) []int {
	if v == nil {
		return make([]int, 0)
	}
	return v
}

// dehlaPakadHandToOutput は 1 ハンドの結果を出力形へ直す。
func dehlaPakadHandToOutput(h *domain.DehlaPakadHandResult) *controller.DehlaPakadWebOutputHand {
	if h == nil {
		return nil
	}
	tens := make([]int, domain.DehlaPakadTeamCnt)
	copy(tens, h.TeamTens[:])
	return &controller.DehlaPakadWebOutputHand{
		WinnerTeam: h.WinnerTeam,
		TeamTens:   tens,
		Kot:        h.Kot,
		KotReason:  h.KotReason,
		DealerIdx:  h.DealerIdx,
		TrumpSuit:  h.TrumpSuit,
	}
}

// humanTeam は人間のチームを返す (居なければ 0)。
func (p *DehlaPakadWebPresenter) humanTeam(g interfaces.DehlaPakadGame) int {
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if pl := g.GetPlayer(i); pl != nil && pl.GetIsHuman() {
			return domain.DehlaPakadTeamOf(i)
		}
	}
	return 0
}

// playableIndices は人間が出せる札のインデックスを返す。
func (p *DehlaPakadWebPresenter) playableIndices(g interfaces.DehlaPakadGame) []int {
	if g.GetPhase() != domain.DehlaPakadPhasePlay || !g.IsHumanTurn() {
		return make([]int, 0)
	}
	return dehlaPakadIntsOrEmpty(g.GetPlayableIndices(g.GetCurrentTurn()))
}

// buildPlayersOutput は席の情報を構築する (人間のみ手札を公開)。
func (p *DehlaPakadWebPresenter) buildPlayersOutput(g interfaces.DehlaPakadGame) []*controller.DehlaPakadWebOutputPlayer {
	dealer := g.GetDealerIdx()
	chooser := g.GetTrumpChooserIdx()
	out := make([]*controller.DehlaPakadWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		gathered := 0
		for _, trick := range player.GetTricksTaken() {
			gathered += len(trick)
		}
		out = append(out, &controller.DehlaPakadWebOutputPlayer{
			ID:             i,
			IsHuman:        player.GetIsHuman(),
			Team:           domain.DehlaPakadTeamOf(i),
			CardCount:      player.GetCardsSize(),
			Cards:          playerCardsToOutput(player, player.GetIsHuman()),
			GatheredCount:  gathered,
			IsDealer:       i == dealer,
			IsTrumpChooser: i == chooser,
		})
	}
	return out
}

// buildMessage はフェーズ / 結果メッセージを構築する。
func (p *DehlaPakadWebPresenter) buildMessage(g interfaces.DehlaPakadGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		code, params := domain.ErrorMessageCode(lastErr)
		return lastErr.Error(), code, params
	}
	if g.GetGameEndFlag() {
		if g.GetWinnerTeam() == p.humanTeam(g) {
			return "", "dehlapakad.result.humanWin", nil
		}
		return "", "dehlapakad.result.cpuWin", nil
	}
	switch g.GetPhase() {
	case domain.DehlaPakadPhaseSelectTrump:
		return "", "dehlapakad.selectTrump", nil
	case domain.DehlaPakadPhasePlay:
		// **山に 10 が乗っているかで、このトリックの重みが変わる。**
		if g.GetCentrePileTens() > 0 {
			return "", "dehlapakad.playPhase.tensAtStake", nil
		}
		return "", "dehlapakad.playPhase", nil
	case domain.DehlaPakadPhaseHandEnd:
		return "", "dehlapakad.handEnd", nil
	}
	return "", "", nil
}

// HintOutput はヒント情報を JSON 出力する。
func (p *DehlaPakadWebPresenter) HintOutput(g interfaces.DehlaPakadGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil && (len(hint.CardIndices) > 0 || hint.TrumpSuit >= 0) {
		if len(hint.CardIndices) > 0 {
			resObj.Hint = cardHint(hint.CardIndices, hint.Reason)
		}
		resObj.HintTrumpSuit = hint.TrumpSuit
		resObj.MessageCode = "dehlapakad.hintRequested"
	} else {
		resObj.MessageCode = "dehlapakad.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput は棋譜を JSON 出力する。
func (p *DehlaPakadWebPresenter) ActionLogOutput(g interfaces.DehlaPakadGame) string {
	return actionLogOutputJSON(g)
}
