//go:build !js || !wasm || casino

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MusWebPresenter ムスのWebプレゼンタークラス
type MusWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *MusWebPresenter) Output(g interfaces.MusGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **フェーズと手番はここでは見ない。**Mus.GetHint() が自分で
	// 「人間の手番で、かつ行動を選べる状態か」を確かめて nil を返す。
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.MusWebOutputHint{
			Mus:     hint.Mus,
			Action:  hint.Action,
			Amount:  hint.Amount,
			Indices: hint.Indices,
			Reason:  hint.Reason,
		}
	}

	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *MusWebPresenter) buildBase(g interfaces.MusGame) *controller.MusWebOutput {
	resObj := new(controller.MusWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.ManoIdx = g.GetManoIdx()
	resObj.BetTeam = g.GetBetTeam()
	resObj.PendingStake = g.GetPendingStake()
	resObj.LastBettorTeam = g.GetLastBettorTeam()
	resObj.MusTurn = g.GetMusTurn()
	resObj.DiscardTurn = g.GetDiscardTurn()
	resObj.MusCycle = g.GetMusCycle()
	resObj.Amarrakos = g.GetAmarrakos()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerTeam = g.GetWinnerTeam()
	resObj.IsHumanTurn = g.IsHumanTurn()

	// humanTeam: team of the first human player (-1 if none)
	resObj.HumanTeam = -1
	for i := 0; i < g.GetPlayerCnt(); i++ {
		pl := g.GetPlayer(i)
		if pl != nil && pl.GetIsHuman() {
			resObj.HumanTeam = domain.MusTeamOf(i)
			break
		}
	}

	// Results for all 4 betting rounds
	for ri := 0; ri < domain.MusRoundCnt; ri++ {
		r := g.GetResult(ri)
		resObj.Results[ri] = controller.MusWebOutputResult{
			Kind:  r.Kind,
			Stake: r.Stake,
			Team:  r.Team,
		}
	}

	// Available bet actions depend on pendingStake
	phase := g.GetPhase()
	inBetting := phase >= domain.MusPhaseGrande && phase <= domain.MusPhaseJuego
	if inBetting && g.IsHumanTurn() {
		pending := g.GetPendingStake()
		resObj.CanPaso = pending == 0
		resObj.CanEnvido = true
		resObj.CanOrdago = true
		resObj.CanQuiero = pending != 0
		resObj.CanNoQuiero = pending != 0
	}

	// Config
	cfg := g.GetConfig()
	resObj.Config = controller.MusWebOutputConfig{
		CpuDifficulty:   int(cfg.CpuDifficulty),
		TargetAmarrakos: cfg.TargetAmarrakos,
	}

	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築
func (p *MusWebPresenter) buildPlayersOutput(g interfaces.MusGame) []*controller.MusWebOutputPlayer {
	out := make([]*controller.MusWebOutputPlayer, 0)
	amarrakos := g.GetAmarrakos()
	phase := g.GetPhase()
	reveal := phase == domain.MusPhaseShowdown || phase == domain.MusPhaseRoundEnd || phase == domain.MusPhaseGameEnd

	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		team := domain.MusTeamOf(i)
		showCards := player.GetIsHuman() || reveal
		out = append(out, &controller.MusWebOutputPlayer{
			ID:        i,
			IsHuman:   player.GetIsHuman(),
			CardCount: player.GetCardsSize(),
			Cards:     playerCardsToOutput(player, showCards),
			TeamScore: amarrakos[team],
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *MusWebPresenter) buildMessage(g interfaces.MusGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		return p.winnerMessage(g)
	}
	switch g.GetPhase() {
	case domain.MusPhaseMus:
		return "", "mus.musPhase", nil
	case domain.MusPhaseDiscard:
		return "", "mus.discardPhase", nil
	case domain.MusPhaseGrande:
		return "", "mus.grandePhase", nil
	case domain.MusPhaseChica:
		return "", "mus.chicaPhase", nil
	case domain.MusPhasePares:
		return "", "mus.paresPhase", nil
	case domain.MusPhaseJuego:
		return "", "mus.juegoPhase", nil
	case domain.MusPhaseShowdown:
		return "", "mus.showdownPhase", nil
	case domain.MusPhaseRoundEnd:
		return "", "mus.roundEnd", nil
	}
	return "", "", nil
}

// winnerMessage 勝者メッセージを構築する
func (p *MusWebPresenter) winnerMessage(g interfaces.MusGame) (string, string, map[string]string) {
	winnerTeam := g.GetWinnerTeam()
	// Check whether the human is on the winning team.
	for i := 0; i < g.GetPlayerCnt(); i++ {
		pl := g.GetPlayer(i)
		if pl != nil && pl.GetIsHuman() && domain.MusTeamOf(i) == winnerTeam {
			return "ゲーム終了！ あなたのチームの勝ち！", "mus.result.humanWin", nil
		}
	}
	params := map[string]string{"team": fmt.Sprintf("%d", winnerTeam)}
	return fmt.Sprintf("ゲーム終了！ チーム%dの勝ち！", winnerTeam), "mus.result.cpuWin", params
}

// HintOutput ヒント情報をJSON出力する
func (p *MusWebPresenter) HintOutput(g interfaces.MusGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = &controller.MusWebOutputHint{
			Mus:     hint.Mus,
			Action:  hint.Action,
			Amount:  hint.Amount,
			Indices: hint.Indices,
			Reason:  hint.Reason,
		}
	}
	// **「頼んだヒントか」を CLI が見分けられるようにする。**このゲーム群の
	// `hintAvailable` は画面のラベルとして既に使われているので、別キーを出す (#4483)。
	if hint != nil {
		resObj.MessageCode = "mus.hintRequested"
	} else {
		resObj.MessageCode = "mus.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *MusWebPresenter) ActionLogOutput(g interfaces.MusGame) string {
	return actionLogOutputJSON(g)
}
