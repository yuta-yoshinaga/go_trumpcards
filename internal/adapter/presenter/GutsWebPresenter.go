//go:build !js || !wasm || extra4

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// gutsCategoryLabel は役カテゴリ定数を短い役名キーに変換する。フロントエンドは
// この値を `hand.<key>` として i18n 参照する。
func gutsCategoryLabel(category int) string {
	switch category {
	case domain.GutsHandPair:
		return "pair"
	case domain.GutsHandHighCard:
		return "highcard"
	default:
		return ""
	}
}

// gutsHandName は手の役名 i18n キーを返す (評価不能時は空文字)。
func gutsHandName(player *domain.GutsPlayer) string {
	if player == nil || player.GetCardsSize() != domain.GutsHandSize {
		return ""
	}
	cards := make([]*domain.Card, 0, domain.GutsHandSize)
	for i := 0; i < player.GetCardsSize(); i++ {
		cards = append(cards, player.GetCard(i))
	}
	category, _ := domain.GutsEval(cards)
	return gutsCategoryLabel(category)
}

// GutsWebPresenter はガッツ (Guts) の Web プレゼンタークラス。
type GutsWebPresenter struct{}

// Output はゲーム状態を JSON 出力する。
func (p *GutsWebPresenter) Output(g interfaces.GutsGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **Guts.GetHint() は宣言フェーズかつ席 0（NewGutsPlayer(i == 0) で常に人間）に限る。**
	// 他ゲームがそうだから、で済ませない —— Pinochle は見ていなかった (#4585)。
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.GutsWebOutputHint{
			Declaration: int(hint.Declaration),
			Reason:      hint.Reason,
		}
	}

	return marshalOrError(resObj)
}

// buildBase は基本フィールドを埋めた出力オブジェクトを生成する。
func (p *GutsWebPresenter) buildBase(g interfaces.GutsGame) *controller.GutsWebOutput {
	resObj := new(controller.GutsWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.Pot = g.GetPot()
	resObj.CarryPot = g.GetCarryPot()
	resObj.CarryCount = g.GetCarryCount()
	resObj.Ante = g.GetAnte()
	resObj.Chips = g.GetChips()
	resObj.WinnerIdx = g.GetWinnerIdx()
	resObj.MatchWinnerIdx = g.GetMatchWinnerIdx()
	resObj.Result = int(g.GetResult())
	resObj.Matchers = gutsMatchersOrEmpty(g.GetMatchers())
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.Players = p.buildPlayersOutput(g)

	cfg := g.GetConfig()
	resObj.Config = controller.GutsWebOutputConfig{
		PlayerCount:   cfg.PlayerCount,
		Ante:          cfg.Ante,
		StartingChips: cfg.StartingChips,
		TargetRounds:  cfg.TargetRounds,
	}
	return resObj
}

// gutsMatchersOrEmpty は nil を空スライスに正規化する。
func gutsMatchersOrEmpty(m []int) []int {
	if m == nil {
		return make([]int, 0)
	}
	return m
}

// buildPlayersOutput はプレイヤー情報を構築する。人間は常に手札公開。結果フェーズでは
// 「イン」で残った (非脱落) プレイヤーの手も公開する。
func (p *GutsWebPresenter) buildPlayersOutput(g interfaces.GutsGame) []*controller.GutsWebOutputPlayer {
	out := make([]*controller.GutsWebOutputPlayer, 0)
	reveal := g.GetPhase() == domain.GutsPhaseResult
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		showCards := player.GetIsHuman() || (reveal && player.GetIn() && !player.GetOut())
		handName := ""
		if showCards {
			handName = gutsHandName(player)
		}
		out = append(out, &controller.GutsWebOutputPlayer{
			ID:        i,
			IsHuman:   player.GetIsHuman(),
			Chips:     player.GetChips(),
			In:        player.GetIn(),
			Out:       player.GetOut(),
			RoundBet:  player.GetRoundBet(),
			CardCount: player.GetCardsSize(),
			Cards:     playerCardsToOutput(player, showCards),
			HandName:  handName,
			IsWinner:  i == g.GetWinnerIdx(),
			IsMatcher: g.IsMatcher(i),
		})
	}
	return out
}

// buildMessage はゲーム結果メッセージを構築する。
func (p *GutsWebPresenter) buildMessage(g interfaces.GutsGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		return p.winnerMessage(g)
	}
	switch g.GetPhase() {
	case domain.GutsPhaseDeclare:
		return "", "guts.declarePhase", nil
	case domain.GutsPhaseResult:
		return p.roundEndMessage(g)
	}
	return "", "", nil
}

// roundEndMessage はラウンド終了時のメッセージを構築する。
func (p *GutsWebPresenter) roundEndMessage(g interfaces.GutsGame) (string, string, map[string]string) {
	winner := g.GetWinnerIdx()
	if winner < 0 {
		return "Nobody stayed; the pot carries over.", "guts.roundEndCarry", nil
	}
	switch g.GetResult() {
	case domain.GutsResultWin:
		return "You win the pot!", "guts.roundEndHumanWin", nil
	case domain.GutsResultLose:
		return "You lost and must match the pot.", "guts.roundEndHumanLose", nil
	default:
		params := map[string]string{"player": fmt.Sprintf("%d", winner)}
		return fmt.Sprintf("CPU %d wins the pot.", winner), "guts.roundEndCpuWin", params
	}
}

// winnerMessage は試合終了メッセージを構築する。
func (p *GutsWebPresenter) winnerMessage(g interfaces.GutsGame) (string, string, map[string]string) {
	winner := g.GetMatchWinnerIdx()
	pl := g.GetPlayer(winner)
	if pl != nil && pl.GetIsHuman() {
		return "Game over! You win!", "guts.result.humanWin", nil
	}
	params := map[string]string{"player": fmt.Sprintf("%d", winner)}
	return fmt.Sprintf("Game over! CPU %d wins!", winner), "guts.result.cpuWin", params
}

// HintOutput はヒント情報を JSON 出力する。
func (p *GutsWebPresenter) HintOutput(g interfaces.GutsGame) string {
	resObj := p.buildBase(g)
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.GutsWebOutputHint{
			Declaration: int(hint.Declaration),
			Reason:      hint.Reason,
		}
	}
	// **「頼んだヒントか」を CLI が見分けられるようにする。**このゲーム群の
	// `hintAvailable` は画面のラベルとして既に使われているので、別キーを出す (#4483)。
	if g.GetHint() != nil {
		resObj.MessageCode = "guts.hintRequested"
	} else {
		resObj.MessageCode = "guts.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput は棋譜を JSON 出力する。
func (p *GutsWebPresenter) ActionLogOutput(g interfaces.GutsGame) string {
	return actionLogOutputJSON(g)
}
