//go:build !js || !wasm || casino

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// teenPattiCategoryLabel は役カテゴリ定数を短い役名キーに変換する。フロントエンドは
// この値を `hand.<key>` として i18n 参照する (locales/{ja,en}/teenpatti.json の hand.* に対応)。
func teenPattiCategoryLabel(category int) string {
	switch category {
	case domain.ThreeCardBragPrial:
		return "trail"
	case domain.ThreeCardBragRunningFlush:
		return "puresequence"
	case domain.ThreeCardBragRun:
		return "sequence"
	case domain.ThreeCardBragFlush:
		return "color"
	case domain.ThreeCardBragPair:
		return "pair"
	case domain.ThreeCardBragHighCard:
		return "highcard"
	default:
		return ""
	}
}

// teenPattiHandName は手の役名 i18n キーを返す (評価不能時は空文字)。
func teenPattiHandName(player *domain.TeenPattiPlayer) string {
	if player == nil || player.GetCardsSize() != domain.TeenPattiHandSize {
		return ""
	}
	cards := make([]*domain.Card, 0, domain.TeenPattiHandSize)
	for i := 0; i < player.GetCardsSize(); i++ {
		cards = append(cards, player.GetCard(i))
	}
	category, _ := domain.ThreeCardBragEval(cards)
	return teenPattiCategoryLabel(category)
}

// TeenPattiWebPresenter ティーン・パティのWebプレゼンタークラス
type TeenPattiWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *TeenPattiWebPresenter) Output(g interfaces.TeenPattiGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **フェーズと手番はここでは見ない。**TeenPatti.GetHint() が自分で
	// 「人間の手番で、かつ行動を選べる状態か」を確かめて nil を返す。
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.TeenPattiWebOutputHint{
			Action: hint.Action,
			Reason: hint.Reason,
		}
	}

	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *TeenPattiWebPresenter) buildBase(g interfaces.TeenPattiGame) *controller.TeenPattiWebOutput {
	resObj := new(controller.TeenPattiWebOutput)
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
	resObj.CanRequestSideShow = g.CanRequestSideShow()
	resObj.SideShowRequester = g.GetSideShowRequester()
	resObj.SideShowTarget = g.GetSideShowTarget()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.IsHumanTurn = g.IsHumanTurn()

	cfg := g.GetConfig()
	resObj.Config = controller.TeenPattiWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		Ante:          cfg.Ante,
		StartingChips: cfg.StartingChips,
	}

	resObj.Players = p.buildPlayersOutput(g)
	resObj.LastSideShow = p.buildSideShowOutput(g)
	return resObj
}

// buildSideShowOutput 直近で成立したサイドショーの比較結果を構築する。
// 人間が申請者または対象のときのみ両者の手札と役名を公開する (CPU 同士は秘匿)。
func (p *TeenPattiWebPresenter) buildSideShowOutput(g interfaces.TeenPattiGame) *controller.TeenPattiWebOutputSideShow {
	req, tgt, loser, ok := g.GetLastSideShow()
	if !ok {
		return nil
	}
	reqP := g.GetPlayer(req)
	tgtP := g.GetPlayer(tgt)
	if reqP == nil || tgtP == nil {
		return nil
	}
	// 秘匿: 人間が当事者でないサイドショー (CPU 同士) はカードを漏らさない。
	if !reqP.GetIsHuman() && !tgtP.GetIsHuman() {
		return nil
	}
	winner := req
	if loser == req {
		winner = tgt
	}
	return &controller.TeenPattiWebOutputSideShow{
		RequesterIdx: req,
		TargetIdx:    tgt,
		WinnerIdx:    winner,
		LoserIdx:     loser,
		Requester:    teenPattiSideShowHand(req, reqP),
		Target:       teenPattiSideShowHand(tgt, tgtP),
	}
}

// teenPattiSideShowHand はサイドショー参加者 1 人分の公開手札を構築する。
func teenPattiSideShowHand(idx int, player *domain.TeenPattiPlayer) *controller.TeenPattiWebOutputSideShowHand {
	return &controller.TeenPattiWebOutputSideShowHand{
		PlayerIdx: idx,
		HandName:  teenPattiHandName(player),
		Cards:     playerCardsToOutput(player, true),
	}
}

// buildPlayersOutput プレイヤー情報を構築
func (p *TeenPattiWebPresenter) buildPlayersOutput(g interfaces.TeenPattiGame) []*controller.TeenPattiWebOutputPlayer {
	out := make([]*controller.TeenPattiWebOutputPlayer, 0)
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
			handName = teenPattiHandName(player)
		}
		out = append(out, &controller.TeenPattiWebOutputPlayer{
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
func (p *TeenPattiWebPresenter) buildMessage(g interfaces.TeenPattiGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		return p.winnerMessage(g)
	}
	switch g.GetPhase() {
	case domain.TeenPattiPhaseBetting:
		return "", "teenpatti.bettingPhase", nil
	case domain.TeenPattiPhaseSideShow:
		return "", "teenpatti.sideShowPhase", nil
	case domain.TeenPattiPhaseShowdown:
		return "", "teenpatti.showdownPhase", nil
	case domain.TeenPattiPhaseRoundEnd:
		return p.roundEndMessage(g)
	}
	return "", "", nil
}

// roundEndMessage ディール終了時のメッセージを構築する
func (p *TeenPattiWebPresenter) roundEndMessage(g interfaces.TeenPattiGame) (string, string, map[string]string) {
	winner := g.GetRoundWinnerIdx()
	pl := g.GetPlayer(winner)
	if pl != nil && pl.GetIsHuman() {
		return "あなたがポットを獲得しました！", "teenpatti.roundEndHumanWin", nil
	}
	params := map[string]string{"player": fmt.Sprintf("%d", winner)}
	return fmt.Sprintf("CPU%dがポットを獲得しました。", winner), "teenpatti.roundEndCpuWin", params
}

// winnerMessage 試合終了メッセージを構築する
func (p *TeenPattiWebPresenter) winnerMessage(g interfaces.TeenPattiGame) (string, string, map[string]string) {
	winner := g.GetMatchWinnerIdx()
	pl := g.GetPlayer(winner)
	if pl != nil && pl.GetIsHuman() {
		return "ゲーム終了！ あなたの勝ち！", "teenpatti.result.humanWin", nil
	}
	params := map[string]string{"player": fmt.Sprintf("%d", winner)}
	return fmt.Sprintf("ゲーム終了！ CPU%dの勝ち！", winner), "teenpatti.result.cpuWin", params
}

// HintOutput ヒント情報をJSON出力する
func (p *TeenPattiWebPresenter) HintOutput(g interfaces.TeenPattiGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = &controller.TeenPattiWebOutputHint{
			Action: hint.Action,
			Reason: hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *TeenPattiWebPresenter) ActionLogOutput(g interfaces.TeenPattiGame) string {
	return actionLogOutputJSON(g)
}
