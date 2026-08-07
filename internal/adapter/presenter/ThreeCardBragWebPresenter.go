//go:build !js || !wasm || casino

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// threeCardBragCategoryLabel は役カテゴリ定数を短い役名キーに変換する。フロントエンドは
// この値を `hand.<key>` として i18n 参照する (locales/{ja,en}/threecardbrag.json の hand.* に対応)。
func threeCardBragCategoryLabel(category int) string {
	switch category {
	case domain.ThreeCardBragPrial:
		return "prial"
	case domain.ThreeCardBragRunningFlush:
		return "runningflush"
	case domain.ThreeCardBragRun:
		return "run"
	case domain.ThreeCardBragFlush:
		return "flush"
	case domain.ThreeCardBragPair:
		return "pair"
	case domain.ThreeCardBragHighCard:
		return "highcard"
	default:
		return ""
	}
}

// threeCardBragHandName は手の役名 i18n キーを返す (評価不能時は空文字)。
func threeCardBragHandName(player *domain.ThreeCardBragPlayer) string {
	if player == nil || player.GetCardsSize() != domain.ThreeCardBragHandSize {
		return ""
	}
	cards := make([]*domain.Card, 0, domain.ThreeCardBragHandSize)
	for i := 0; i < player.GetCardsSize(); i++ {
		cards = append(cards, player.GetCard(i))
	}
	category, _ := domain.ThreeCardBragEval(cards)
	return threeCardBragCategoryLabel(category)
}

// ThreeCardBragWebPresenter スリーカード・ブラグのWebプレゼンタークラス
type ThreeCardBragWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *ThreeCardBragWebPresenter) Output(g interfaces.ThreeCardBragGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **フェーズと手番はここでは見ない。**ThreeCardBrag.GetHint() が自分で
	// 「人間の手番で、かつ行動を選べる状態か」を確かめて nil を返す。
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.ThreeCardBragWebOutputHint{
			Action: hint.Action,
			Reason: hint.Reason,
		}
	}

	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *ThreeCardBragWebPresenter) buildBase(g interfaces.ThreeCardBragGame) *controller.ThreeCardBragWebOutput {
	resObj := new(controller.ThreeCardBragWebOutput)
	resObj.Pot = g.GetPot()
	resObj.Stake = g.GetStake()
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.RoundWinnerIdx = g.GetRoundWinnerIdx()
	resObj.MatchWinnerIdx = g.GetMatchWinnerIdx()
	resObj.IsShowdown = g.IsShowdown()
	resObj.CanShow = g.CanShow()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.IsHumanTurn = g.IsHumanTurn()

	cfg := g.GetConfig()
	resObj.Config = controller.ThreeCardBragWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		Ante:          cfg.Ante,
		StartingChips: cfg.StartingChips,
	}

	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築
func (p *ThreeCardBragWebPresenter) buildPlayersOutput(g interfaces.ThreeCardBragGame) []*controller.ThreeCardBragWebOutputPlayer {
	out := make([]*controller.ThreeCardBragWebOutputPlayer, 0)
	reveal := g.IsShowdown()
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		// 人間は常に公開。ショーダウン時は非フォールドの手も公開する。
		showCards := player.GetIsHuman() || (reveal && !player.GetFolded())
		handName := ""
		if showCards && !player.GetIsHuman() {
			handName = threeCardBragHandName(player)
		}
		out = append(out, &controller.ThreeCardBragWebOutputPlayer{
			ID:        i,
			IsHuman:   player.GetIsHuman(),
			Chips:     player.GetChips(),
			Seen:      player.GetSeen(),
			Folded:    player.GetFolded(),
			Out:       player.GetOut(),
			RoundBet:  player.GetRoundBet(),
			CardCount: player.GetCardsSize(),
			Cards:     playerCardsToOutput(player, showCards),
			HandName:  handName,
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *ThreeCardBragWebPresenter) buildMessage(g interfaces.ThreeCardBragGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		return p.winnerMessage(g)
	}
	switch g.GetPhase() {
	case domain.ThreeCardBragPhaseBetting:
		return "", "threecardbrag.bettingPhase", nil
	case domain.ThreeCardBragPhaseShowdown:
		return "", "threecardbrag.showdownPhase", nil
	case domain.ThreeCardBragPhaseRoundEnd:
		return p.roundEndMessage(g)
	}
	return "", "", nil
}

// roundEndMessage ディール終了時のメッセージを構築する
func (p *ThreeCardBragWebPresenter) roundEndMessage(g interfaces.ThreeCardBragGame) (string, string, map[string]string) {
	winner := g.GetRoundWinnerIdx()
	pl := g.GetPlayer(winner)
	if pl != nil && pl.GetIsHuman() {
		return "", "threecardbrag.roundEndHumanWin", nil
	}
	params := map[string]string{"player": fmt.Sprintf("%d", winner)}
	return "", "threecardbrag.roundEndCpuWin", params
}

// winnerMessage 試合終了メッセージを構築する
func (p *ThreeCardBragWebPresenter) winnerMessage(g interfaces.ThreeCardBragGame) (string, string, map[string]string) {
	winner := g.GetMatchWinnerIdx()
	pl := g.GetPlayer(winner)
	if pl != nil && pl.GetIsHuman() {
		return "ゲーム終了！ あなたの勝ち！", "threecardbrag.result.humanWin", nil
	}
	params := map[string]string{"player": fmt.Sprintf("%d", winner)}
	return fmt.Sprintf("ゲーム終了！ CPU%dの勝ち！", winner), "threecardbrag.result.cpuWin", params
}

// HintOutput ヒント情報をJSON出力する
func (p *ThreeCardBragWebPresenter) HintOutput(g interfaces.ThreeCardBragGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = &controller.ThreeCardBragWebOutputHint{
			Action: hint.Action,
			Reason: hint.Reason,
		}
		// **「頼んだヒントか」を画面が見分けられるようにする。**Output() は
		// 受動ヒントとして常に hint を詰めるので、これが無いとページ側は
		// 要求の有無を区別できず、要求していないヒントを出してしまう。
		// このゲーム群の hintAvailable は画面のラベルとして既に使われている
		// ので、別キーを出す (#4483 と同じ理由)。
		resObj.MessageCode = "threecardbrag.hintRequested"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *ThreeCardBragWebPresenter) ActionLogOutput(g interfaces.ThreeCardBragGame) string {
	return actionLogOutputJSON(g)
}
