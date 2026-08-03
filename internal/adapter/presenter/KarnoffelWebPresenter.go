//go:build !js || !wasm || classic

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// KarnoffelWebPresenter カルニッフェル Webプレゼンタークラス
type KarnoffelWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *KarnoffelWebPresenter) Output(g interfaces.KarnoffelGame, lastErr error) string {
	resObj := new(controller.KarnoffelWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.HandNumber = g.GetHandNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.ChosenSuit = g.GetChosenSuit()
	resObj.TrickLeaderIdx = g.GetTrickLeaderIdx()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.TricksToWin = domain.KarnoffelTricksToWin
	resObj.HandSize = domain.KarnoffelHandSize
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerTeam = g.GetWinnerTeam()

	for team := range domain.KarnoffelTeamCnt {
		resObj.TeamTricks[team] = g.KarnoffelTeamTricks(team)
		resObj.HandsWon[team] = g.GetHandsWon(team)
	}

	trick := g.GetTrick()
	resObj.Trick = make([]*controller.WebOutputCard, 0, len(trick))
	for _, c := range trick {
		if out := cardToOutput(c); out != nil {
			resObj.Trick = append(resObj.Trick, out)
		}
	}

	if r := g.GetLastResult(); r != nil {
		resObj.LastResult = &controller.KarnoffelWebOutputResult{
			WinnerTeam: r.WinnerTeam,
			Tricks:     r.Tricks,
			ChosenSuit: r.ChosenSuit,
		}
	}

	// **出せる札はサーバーが決める。**追随の義務は無いが、第 1 トリックの
	// リードに悪魔は使えないので、フロントで再現するとずれる。
	resObj.ValidPlays = make([]int, 0)
	if g.GetPhase() == domain.KarnoffelPhasePlay && g.IsHumanTurn() {
		resObj.ValidPlays = append(resObj.ValidPlays, g.KarnoffelValidPlays(g.GetCurrentPlayerIdx())...)
	}

	cfg := g.GetConfig()
	resObj.TargetHands = cfg.TargetHands
	resObj.Config = controller.KarnoffelWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetHands:   cfg.TargetHands,
	}

	resObj.Players = p.buildPlayersOutput(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)

	return marshalOrError(resObj)
}

// karnoffelWebReveal は全員の手札を公開する局面かを返す。
func karnoffelWebReveal(g interfaces.KarnoffelGame) bool {
	phase := g.GetPhase()
	return phase == domain.KarnoffelPhaseHandEnd || phase == domain.KarnoffelPhaseGameEnd
}

// buildPlayersOutput プレイヤー情報を構築
func (p *KarnoffelWebPresenter) buildPlayersOutput(g interfaces.KarnoffelGame) []*controller.KarnoffelWebOutputPlayer {
	players := g.GetPlayers()
	out := make([]*controller.KarnoffelWebOutputPlayer, 0, len(players))
	reveal := karnoffelWebReveal(g)
	for i := range players {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		cards := make([]*controller.WebOutputCard, 0, player.GetCardsSize())
		if player.GetIsHuman() || reveal {
			for j := range player.GetCardsSize() {
				if c := cardToOutput(player.GetCard(j)); c != nil {
					cards = append(cards, c)
				}
			}
		}
		out = append(out, &controller.KarnoffelWebOutputPlayer{
			ID:      i,
			IsHuman: player.GetIsHuman(),
			Team:    domain.KarnoffelTeamOf(i),
			// **表向きの札は全員ぶん公開される。**切札の根拠が見えないと
			// 盤面が読めない。
			CardCount:     player.GetCardsSize(),
			Cards:         cards,
			UpCard:        cardToOutput(g.GetUpCard(i)),
			TricksWon:     g.GetTricksWon(i),
			IsDealer:      i == g.GetDealerIdx(),
			IsCurrentTurn: g.GetPhase() == domain.KarnoffelPhasePlay && i == g.GetCurrentPlayerIdx(),
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *KarnoffelWebPresenter) buildMessage(g interfaces.KarnoffelGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		// **チーム戦なので勝敗は席ではなくチームで見る。**人間は席 0 = チーム 0。
		if g.GetWinnerTeam() == domain.KarnoffelTeamOf(0) {
			return "your team wins", "karnoffel.result.humanWin", nil
		}
		return "the other team wins", "karnoffel.result.cpuWin", nil
	}
	switch g.GetPhase() {
	case domain.KarnoffelPhasePlay:
		return "", "karnoffel.playPhase", nil
	case domain.KarnoffelPhaseHandEnd:
		if r := g.GetLastResult(); r != nil && r.WinnerTeam < 0 {
			return "", "karnoffel.handDrawn", nil
		}
		return "", "karnoffel.handEnd", nil
	}
	return "", "", nil
}

// ActionLogOutput 棋譜をJSON出力
func (p *KarnoffelWebPresenter) ActionLogOutput(g interfaces.KarnoffelGame) string {
	return actionLogOutputJSON(g)
}
