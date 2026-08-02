//go:build !js || !wasm || extra3

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// bouillotteCategoryLabel は役カテゴリ定数を短い役名キーに変換する。フロントエンドは
// この値を `hand.<key>` として i18n 参照する。
func bouillotteCategoryLabel(category int) string {
	switch category {
	case domain.BouillotteHandBrelan:
		return "brelan"
	case domain.BouillotteHandHighCard:
		return "highcard"
	default:
		return ""
	}
}

// bouillotteHandName は手の役名 i18n キーを返す (評価不能時は空文字)。retourne を共有カードとして含む。
func bouillotteHandName(player *domain.BouillottePlayer, retourne *domain.Card) string {
	if player == nil || player.GetCardsSize() != domain.BouillotteHandSize {
		return ""
	}
	cards := make([]*domain.Card, 0, domain.BouillotteHandSize)
	for i := 0; i < player.GetCardsSize(); i++ {
		cards = append(cards, player.GetCard(i))
	}
	category, _ := domain.BouillotteEval(cards, retourne)
	return bouillotteCategoryLabel(category)
}

// BouillotteWebPresenter はブイヨット (Bouillotte) の Web プレゼンタークラス。
type BouillotteWebPresenter struct{}

// Output はゲーム状態を JSON 出力する。
func (p *BouillotteWebPresenter) Output(g interfaces.BouillotteGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **Bouillotte.GetHint() は賭けフェーズかつ currentPlayer == 0 に限る。席 0 は常に人間。**
	// 他ゲームがそうだから、で済ませない —— Pinochle は見ていなかった (#4585)。
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.BouillotteWebOutputHint{
			Action: hint.Action,
			Reason: hint.Reason,
		}
	}

	return marshalOrError(resObj)
}

// buildBase は基本フィールドを埋めた出力オブジェクトを生成する。
func (p *BouillotteWebPresenter) buildBase(g interfaces.BouillotteGame) *controller.BouillotteWebOutput {
	resObj := new(controller.BouillotteWebOutput)
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
	resObj.Retourne = cardToOutput(g.GetRetourne())
	resObj.IsHumanTurn = g.IsHumanTurn()
	resObj.CanRaise = g.CanRaise()
	resObj.WinnerIdx = g.GetWinnerIdx()
	resObj.MatchWinnerIdx = g.GetMatchWinnerIdx()
	resObj.Result = int(g.GetResult())
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.Players = p.buildPlayersOutput(g)

	cfg := g.GetConfig()
	resObj.Config = controller.BouillotteWebOutputConfig{
		PlayerCount:   cfg.PlayerCount,
		Ante:          cfg.Ante,
		StartingChips: cfg.StartingChips,
		TargetRounds:  cfg.TargetRounds,
	}
	return resObj
}

// buildPlayersOutput はプレイヤー情報を構築する。人間は常に手札公開。結果フェーズでは
// フォールドしていない (非脱落) プレイヤーの手も公開する。
func (p *BouillotteWebPresenter) buildPlayersOutput(g interfaces.BouillotteGame) []*controller.BouillotteWebOutputPlayer {
	out := make([]*controller.BouillotteWebOutputPlayer, 0)
	reveal := g.GetPhase() == domain.BouillottePhaseResult
	retourne := g.GetRetourne()
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		showCards := player.GetIsHuman() || (reveal && !player.GetFolded() && !player.GetOut())
		handName := ""
		if showCards {
			handName = bouillotteHandName(player, retourne)
		}
		out = append(out, &controller.BouillotteWebOutputPlayer{
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
func (p *BouillotteWebPresenter) buildMessage(g interfaces.BouillotteGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		return p.winnerMessage(g)
	}
	switch g.GetPhase() {
	case domain.BouillottePhaseBetting:
		return "", "bouillotte.bettingPhase", nil
	case domain.BouillottePhaseResult:
		return p.roundEndMessage(g)
	}
	return "", "", nil
}

// roundEndMessage はラウンド終了時のメッセージを構築する。
func (p *BouillotteWebPresenter) roundEndMessage(g interfaces.BouillotteGame) (string, string, map[string]string) {
	winner := g.GetWinnerIdx()
	if winner < 0 {
		return "The round is over.", "bouillotte.roundEnd", nil
	}
	switch g.GetResult() {
	case domain.BouillotteResultWin:
		return "You win the pot!", "bouillotte.roundEndHumanWin", nil
	case domain.BouillotteResultLose:
		params := map[string]string{"player": fmt.Sprintf("%d", winner)}
		return fmt.Sprintf("CPU %d wins the pot.", winner), "bouillotte.roundEndHumanLose", params
	default:
		params := map[string]string{"player": fmt.Sprintf("%d", winner)}
		return fmt.Sprintf("CPU %d wins the pot.", winner), "bouillotte.roundEndCpuWin", params
	}
}

// winnerMessage は試合終了メッセージを構築する。
func (p *BouillotteWebPresenter) winnerMessage(g interfaces.BouillotteGame) (string, string, map[string]string) {
	winner := g.GetMatchWinnerIdx()
	pl := g.GetPlayer(winner)
	if pl != nil && pl.GetIsHuman() {
		return "Game over! You win!", "bouillotte.result.humanWin", nil
	}
	params := map[string]string{"player": fmt.Sprintf("%d", winner)}
	return fmt.Sprintf("Game over! CPU %d wins!", winner), "bouillotte.result.cpuWin", params
}

// HintOutput はヒント情報を JSON 出力する。
func (p *BouillotteWebPresenter) HintOutput(g interfaces.BouillotteGame) string {
	resObj := p.buildBase(g)
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.BouillotteWebOutputHint{
			Action: hint.Action,
			Reason: hint.Reason,
		}
	}
	// **「頼んだヒントか」を CLI が見分けられるようにする。**このゲーム群の
	// `hintAvailable` は画面のラベルとして既に使われているので、別キーを出す (#4483)。
	if g.GetHint() != nil {
		resObj.MessageCode = "bouillotte.hintRequested"
	} else {
		resObj.MessageCode = "bouillotte.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput は棋譜を JSON 出力する。
func (p *BouillotteWebPresenter) ActionLogOutput(g interfaces.BouillotteGame) string {
	return actionLogOutputJSON(g)
}
