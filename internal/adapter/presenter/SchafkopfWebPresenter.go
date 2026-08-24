//go:build !js || !wasm || extra4

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// SchafkopfWebPresenter シャーフコップのWebプレゼンタークラス
type SchafkopfWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *SchafkopfWebPresenter) Output(g interfaces.SchafkopfGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **フェーズと手番はここでは見ない。**Schafkopf.GetHint() が自分で
	// 「人間の手番で、かつ行動を選べる状態か」を確かめて nil を返す。
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.SchafkopfWebOutputHint{
			CardIndices: hint.CardIndices,
			Suit:        hint.Suit,
			Pick:        hint.Pick,
			Reason:      hint.Reason,
		}
	}

	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *SchafkopfWebPresenter) buildBase(g interfaces.SchafkopfGame) *controller.SchafkopfWebOutput {
	resObj := new(controller.SchafkopfWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.LeadPlayerIdx = g.GetLeadPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()
	resObj.PickerIdx = g.GetPickerIdx()
	resObj.PassCount = g.GetPassCount()
	resObj.Contract = int(g.GetContract())
	resObj.SoloSuit = g.GetSoloSuit()
	resObj.BeatableContracts = []int{}
	for _, c := range g.GetBeatableContracts() {
		resObj.BeatableContracts = append(resObj.BeatableContracts, int(c))
	}
	resObj.CalledSuit = g.GetCalledSuit()
	resObj.PartnerRevealed = g.IsPartnerRevealed()
	resObj.RoundPickerPoints = g.GetRoundPickerPoints()
	resObj.RoundMultiplier = g.GetRoundMultiplier()
	resObj.RoundPickerWon = g.GetRoundPickerWon()

	phase := g.GetPhase()
	// PartnerIdx: hide (send -1) until the partner has been revealed during play.
	if g.IsPartnerRevealed() || phase == domain.SchafkopfPhaseRoundEnd || phase == domain.SchafkopfPhaseGameEnd {
		resObj.PartnerIdx = g.GetPartnerIdx()
	} else {
		resObj.PartnerIdx = -1
	}

	resObj.CallableSuits = p.callableSuits(g)
	resObj.PlayableIndices = p.playableIndices(g)

	cfg := g.GetConfig()
	resObj.Config = controller.SchafkopfWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		BaseChips:     cfg.BaseChips,
		StartChips:    cfg.StartChips,
		TargetChips:   cfg.TargetChips,
	}

	resObj.CurrentTrick = trickCardsToOutput(g.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// callableSuits 呼び可能スートを返す (Call フェーズ以外は空)
func (p *SchafkopfWebPresenter) callableSuits(g interfaces.SchafkopfGame) []int {
	if g.GetPhase() != domain.SchafkopfPhaseCall {
		return make([]int, 0)
	}
	suits := g.GetCallableSuits()
	if suits == nil {
		return make([]int, 0)
	}
	return suits
}

// playableIndices 人間プレイヤーがプレイできるカードのインデックスを返す
func (p *SchafkopfWebPresenter) playableIndices(g interfaces.SchafkopfGame) []int {
	if g.GetPhase() != domain.SchafkopfPhasePlay || !g.IsHumanTurn() {
		return make([]int, 0)
	}
	idx := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
	if idx == nil {
		return make([]int, 0)
	}
	return idx
}

// buildPlayersOutput プレイヤー情報を構築
func (p *SchafkopfWebPresenter) buildPlayersOutput(g interfaces.SchafkopfGame) []*controller.SchafkopfWebOutputPlayer {
	out := make([]*controller.SchafkopfWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		out = append(out, &controller.SchafkopfWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, player.GetIsHuman()),
			TrickCount: player.GetTrickCount(),
			Chips:      player.GetChips(),
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *SchafkopfWebPresenter) buildMessage(g interfaces.SchafkopfGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		return p.winnerMessage(g)
	}
	switch g.GetPhase() {
	case domain.SchafkopfPhasePick:
		return "", "schafkopf.pickPhase", nil
	case domain.SchafkopfPhaseCall:
		return "", "schafkopf.callPhase", nil
	case domain.SchafkopfPhasePlay:
		if len(g.GetCurrentTrick()) == 0 {
			return "", "schafkopf.playPhase.lead", nil
		}
		return "", "schafkopf.playPhase.follow", nil
	case domain.SchafkopfPhaseTrickEnd:
		return "", "schafkopf.trickEnd", nil
	case domain.SchafkopfPhaseRoundEnd:
		return "", "schafkopf.roundEnd", nil
	}
	return "", "", nil
}

// winnerMessage 勝者メッセージを構築する
func (p *SchafkopfWebPresenter) winnerMessage(g interfaces.SchafkopfGame) (string, string, map[string]string) {
	winnerIdx := g.GetWinnerIdx()
	isHuman := false
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if player := g.GetPlayer(i); player != nil && player.GetIsHuman() && i == winnerIdx {
			isHuman = true
			break
		}
	}
	if isHuman {
		return "ゲーム終了！ あなたの勝ち！", "schafkopf.result.humanWin", nil
	}
	params := map[string]string{"cpuId": fmt.Sprintf("%d", winnerIdx)}
	return fmt.Sprintf("ゲーム終了！ CPU %dの勝ち！", winnerIdx), "schafkopf.result.cpuWin", params
}

// HintOutput ヒント情報をJSON出力する
func (p *SchafkopfWebPresenter) HintOutput(g interfaces.SchafkopfGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = &controller.SchafkopfWebOutputHint{
			CardIndices: hint.CardIndices,
			Suit:        hint.Suit,
			Pick:        hint.Pick,
			Reason:      hint.Reason,
		}
	}
	// **「頼んだヒントか」を CLI が見分けられるようにする。**このゲーム群の
	// `hintAvailable` は画面のラベルとして既に使われているので、別キーを出す (#4483)。
	if hint != nil {
		resObj.MessageCode = "schafkopf.hintRequested"
	} else {
		resObj.MessageCode = "schafkopf.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *SchafkopfWebPresenter) ActionLogOutput(g interfaces.SchafkopfGame) string {
	return actionLogOutputJSON(g)
}
