//go:build !js || !wasm || extra

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// anacondaCategoryLabel はポーカー役カテゴリ定数を短い役名キーに変換する。フロントエンドは
// この値を `hand.<key>` として i18n 参照する。
func anacondaCategoryLabel(category int) string {
	switch category {
	case domain.AnacondaStraightFlush:
		return "straightflush"
	case domain.AnacondaFourKind:
		return "fourkind"
	case domain.AnacondaFullHouse:
		return "fullhouse"
	case domain.AnacondaFlush:
		return "flush"
	case domain.AnacondaStraight:
		return "straight"
	case domain.AnacondaThreeKind:
		return "threekind"
	case domain.AnacondaTwoPair:
		return "twopair"
	case domain.AnacondaOnePair:
		return "onepair"
	case domain.AnacondaHighCard:
		return "highcard"
	default:
		return ""
	}
}

// anacondaHandName は完全公開された手の役名 i18n キーを返す (未公開時は空文字)。
func anacondaHandName(cards []*domain.Card) string {
	if len(cards) != domain.AnacondaKeepSize {
		return ""
	}
	category, _ := domain.AnacondaEval(cards)
	return anacondaCategoryLabel(category)
}

// AnacondaWebPresenter はアナコンダ (Anaconda) の Web プレゼンタークラス。
type AnacondaWebPresenter struct{}

// Output はゲーム状態を JSON 出力する。
func (p *AnacondaWebPresenter) Output(g interfaces.AnacondaGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **Anaconda.GetHint() は席 0（NewAnacondaPlayer(i == 0) で常に人間）に限る。パス／セットは全員同時なので手番の概念が無く、ロールのみ currentPlayer を見る。**
	// 他ゲームがそうだから、で済ませない —— Pinochle は見ていなかった (#4585)。
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.AnacondaWebOutputHint{
			Action:      hint.Action,
			CardIndices: hint.CardIndices,
			Reason:      hint.Reason,
		}
	}

	return marshalOrError(resObj)
}

// buildBase は基本フィールドを埋めた出力オブジェクトを生成する。
func (p *AnacondaWebPresenter) buildBase(g interfaces.AnacondaGame) *controller.AnacondaWebOutput {
	resObj := new(controller.AnacondaWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.CurrentPlayer = g.GetCurrentPlayerIdx()
	resObj.PassCount = g.GetPassCount()
	resObj.RollIndex = g.GetRollIndex()
	resObj.Pot = g.GetPot()
	resObj.CurrentBet = g.GetCurrentBet()
	resObj.RaiseCount = g.GetRaiseCount()
	resObj.MaxRaises = g.GetMaxRaises()
	resObj.Ante = g.GetAnte()
	resObj.Chips = g.GetChips()
	resObj.WinnerIdx = g.GetWinnerIdx()
	resObj.MatchWinnerIdx = g.GetMatchWinnerIdx()
	resObj.Result = int(g.GetResult())
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.IsHumanTurn = g.IsHumanTurn()
	resObj.CanRaise = g.CanRaise()
	resObj.Players = p.buildPlayersOutput(g)

	cfg := g.GetConfig()
	resObj.Config = controller.AnacondaWebOutputConfig{
		PlayerCount:   cfg.PlayerCount,
		Ante:          cfg.Ante,
		StartingChips: cfg.StartingChips,
		TargetRounds:  cfg.TargetRounds,
	}
	return resObj
}

// buildPlayersOutput はプレイヤー情報を構築する。公開カードはドメインの GetRevealedCards が
// フェーズに応じて決める (人間は全手札、CPU はロールで公開された分、結果で非フォールドの全 5 枚)。
func (p *AnacondaWebPresenter) buildPlayersOutput(g interfaces.AnacondaGame) []*controller.AnacondaWebOutputPlayer {
	out := make([]*controller.AnacondaWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		revealed := g.GetRevealedCards(i)
		out = append(out, &controller.AnacondaWebOutputPlayer{
			ID:        i,
			IsHuman:   player.GetIsHuman(),
			Chips:     player.GetChips(),
			Folded:    player.GetFolded(),
			Out:       player.GetOut(),
			RoundBet:  player.GetRoundBet(),
			StreetBet: player.GetStreetBet(),
			CardCount: player.GetCardsSize(),
			Cards:     cardsToOutputOrEmpty(revealed),
			HandName:  anacondaHandName(revealed),
			IsWinner:  i == g.GetWinnerIdx(),
		})
	}
	return out
}

// buildMessage はゲーム状態メッセージを構築する。
func (p *AnacondaWebPresenter) buildMessage(g interfaces.AnacondaGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		return p.winnerMessage(g)
	}
	switch g.GetPhase() {
	case domain.AnacondaPhasePass:
		return "", "anaconda.passPhase", map[string]string{"count": fmt.Sprintf("%d", g.GetPassCount())}
	case domain.AnacondaPhaseSet:
		return "", "anaconda.setPhase", nil
	case domain.AnacondaPhaseRoll:
		return "", "anaconda.rollPhase", map[string]string{"revealed": fmt.Sprintf("%d", g.GetRollIndex())}
	case domain.AnacondaPhaseResult:
		return p.roundEndMessage(g)
	}
	return "", "", nil
}

// roundEndMessage はラウンド終了時のメッセージを構築する。
func (p *AnacondaWebPresenter) roundEndMessage(g interfaces.AnacondaGame) (string, string, map[string]string) {
	winner := g.GetWinnerIdx()
	if winner < 0 {
		return "The round ended with no winner.", "anaconda.roundEndNone", nil
	}
	switch g.GetResult() {
	case domain.AnacondaResultWin:
		return "You win the pot!", "anaconda.roundEndHumanWin", nil
	case domain.AnacondaResultLose:
		return "You lost this round.", "anaconda.roundEndHumanLose", nil
	default:
		params := map[string]string{"player": fmt.Sprintf("%d", winner)}
		return fmt.Sprintf("CPU %d wins the pot.", winner), "anaconda.roundEndCpuWin", params
	}
}

// winnerMessage は試合終了メッセージを構築する。
func (p *AnacondaWebPresenter) winnerMessage(g interfaces.AnacondaGame) (string, string, map[string]string) {
	winner := g.GetMatchWinnerIdx()
	pl := g.GetPlayer(winner)
	if pl != nil && pl.GetIsHuman() {
		return "Game over! You win!", "anaconda.result.humanWin", nil
	}
	params := map[string]string{"player": fmt.Sprintf("%d", winner)}
	return fmt.Sprintf("Game over! CPU %d wins!", winner), "anaconda.result.cpuWin", params
}

// HintOutput はヒント情報を JSON 出力する。
func (p *AnacondaWebPresenter) HintOutput(g interfaces.AnacondaGame) string {
	resObj := p.buildBase(g)
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.AnacondaWebOutputHint{
			Action:      hint.Action,
			CardIndices: hint.CardIndices,
			Reason:      hint.Reason,
		}
	}
	// **「頼んだヒントか」を CLI が見分けられるようにする。**このゲーム群の
	// `hintAvailable` は画面のラベルとして既に使われているので、別キーを出す (#4483)。
	if g.GetHint() != nil {
		resObj.MessageCode = "anaconda.hintRequested"
	} else {
		resObj.MessageCode = "anaconda.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput は棋譜を JSON 出力する。
func (p *AnacondaWebPresenter) ActionLogOutput(g interfaces.AnacondaGame) string {
	return actionLogOutputJSON(g)
}
