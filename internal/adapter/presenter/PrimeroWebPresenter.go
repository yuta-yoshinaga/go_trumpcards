//go:build !js || !wasm || extra3

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// primeroCategoryLabel は役カテゴリ定数を短い役名キーに変換する。フロントエンドは
// この値を `hand.<key>` として i18n 参照する。
func primeroCategoryLabel(category int) string {
	switch category {
	case domain.PrimeroHandFluxus:
		return "fluxus"
	case domain.PrimeroHandSupremus:
		return "supremus"
	case domain.PrimeroHandPrimero:
		return "primero"
	case domain.PrimeroHandNumerus:
		return "numerus"
	default:
		return ""
	}
}

// primeroHandName は手の役名 i18n キーを返す (評価不能時は空文字)。
func primeroHandName(player *domain.PrimeroPlayer) string {
	if player == nil || player.GetCardsSize() != domain.PrimeroHandSize {
		return ""
	}
	cards := make([]*domain.Card, 0, domain.PrimeroHandSize)
	for i := 0; i < player.GetCardsSize(); i++ {
		cards = append(cards, player.GetCard(i))
	}
	category, _ := domain.PrimeroEval(cards)
	return primeroCategoryLabel(category)
}

// PrimeroWebPresenter はプリメロ (Primero) の Web プレゼンタークラス。
type PrimeroWebPresenter struct{}

// Output はゲーム状態を JSON 出力する。
func (p *PrimeroWebPresenter) Output(g interfaces.PrimeroGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	return marshalOrError(resObj)
}

// buildBase は基本フィールドを埋めた出力オブジェクトを生成する。
func (p *PrimeroWebPresenter) buildBase(g interfaces.PrimeroGame) *controller.PrimeroWebOutput {
	resObj := new(controller.PrimeroWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.Pot = g.GetPot()
	resObj.Ante = g.GetAnte()
	resObj.Chips = g.GetChips()
	resObj.CurrentBet = g.GetCurrentBet()
	resObj.RaiseCount = g.GetRaiseCount()
	resObj.MaxRaises = g.GetMaxRaises()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.IsHumanTurn = g.IsHumanTurn()
	resObj.CanRaise = g.CanRaise()
	resObj.WinnerIdx = g.GetWinnerIdx()
	resObj.MatchWinnerIdx = g.GetMatchWinnerIdx()
	resObj.Result = int(g.GetResult())
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.Players = p.buildPlayersOutput(g)

	cfg := g.GetConfig()
	resObj.Config = controller.PrimeroWebOutputConfig{
		PlayerCount:   cfg.PlayerCount,
		Ante:          cfg.Ante,
		StartingChips: cfg.StartingChips,
		TargetRounds:  cfg.TargetRounds,
	}
	return resObj
}

// buildPlayersOutput はプレイヤー情報を構築する。人間は常に手札公開。結果フェーズでは
// フォールドしていない (非脱落) プレイヤーの手も公開する。
func (p *PrimeroWebPresenter) buildPlayersOutput(g interfaces.PrimeroGame) []*controller.PrimeroWebOutputPlayer {
	out := make([]*controller.PrimeroWebOutputPlayer, 0)
	reveal := g.GetPhase() == domain.PrimeroPhaseResult
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		showCards := player.GetIsHuman() || (reveal && !player.GetFolded() && !player.GetOut())
		handName := ""
		if showCards {
			handName = primeroHandName(player)
		}
		out = append(out, &controller.PrimeroWebOutputPlayer{
			ID:        i,
			IsHuman:   player.GetIsHuman(),
			Chips:     player.GetChips(),
			RoundBet:  player.GetRoundBet(),
			Folded:    player.GetFolded(),
			Out:       player.GetOut(),
			CardCount: player.GetCardsSize(),
			Cards:     playerCardsToOutput(player, showCards),
			HandName:  handName,
			IsWinner:  i == g.GetWinnerIdx(),
		})
	}
	return out
}

// buildMessage はゲーム結果メッセージを構築する。
func (p *PrimeroWebPresenter) buildMessage(g interfaces.PrimeroGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		return p.winnerMessage(g)
	}
	switch g.GetPhase() {
	case domain.PrimeroPhaseBetting:
		return "", "primero.bettingPhase", nil
	case domain.PrimeroPhaseResult:
		return p.roundEndMessage(g)
	}
	return "", "", nil
}

// roundEndMessage はラウンド終了時のメッセージを構築する。
func (p *PrimeroWebPresenter) roundEndMessage(g interfaces.PrimeroGame) (string, string, map[string]string) {
	winner := g.GetWinnerIdx()
	if winner < 0 {
		return "The round is over.", "primero.roundEnd", nil
	}
	switch g.GetResult() {
	case domain.PrimeroResultWin:
		return "You win the pot!", "primero.roundEndHumanWin", nil
	case domain.PrimeroResultLose:
		params := map[string]string{"player": fmt.Sprintf("%d", winner)}
		return fmt.Sprintf("CPU %d wins the pot.", winner), "primero.roundEndHumanLose", params
	default:
		params := map[string]string{"player": fmt.Sprintf("%d", winner)}
		return fmt.Sprintf("CPU %d wins the pot.", winner), "primero.roundEndCpuWin", params
	}
}

// winnerMessage は試合終了メッセージを構築する。
func (p *PrimeroWebPresenter) winnerMessage(g interfaces.PrimeroGame) (string, string, map[string]string) {
	winner := g.GetMatchWinnerIdx()
	pl := g.GetPlayer(winner)
	if pl != nil && pl.GetIsHuman() {
		return "Game over! You win!", "primero.result.humanWin", nil
	}
	params := map[string]string{"player": fmt.Sprintf("%d", winner)}
	return fmt.Sprintf("Game over! CPU %d wins!", winner), "primero.result.cpuWin", params
}

// HintOutput はヒント情報を JSON 出力する。
func (p *PrimeroWebPresenter) HintOutput(g interfaces.PrimeroGame) string {
	resObj := p.buildBase(g)
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.PrimeroWebOutputHint{
			Action: hint.Action,
			Reason: hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput は棋譜を JSON 出力する。
func (p *PrimeroWebPresenter) ActionLogOutput(g interfaces.PrimeroGame) string {
	return actionLogOutputJSON(g)
}
