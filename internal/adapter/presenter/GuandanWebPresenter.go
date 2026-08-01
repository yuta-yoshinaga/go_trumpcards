//go:build !js || !wasm || extra2

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// GuandanWebPresenter 掼蛋 Webプレゼンタークラス
type GuandanWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *GuandanWebPresenter) Output(g interfaces.GuandanGame, lastErr error) string {
	resObj := new(controller.GuandanWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.HandNumber = g.GetHandNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.Level = g.GetLevel()
	resObj.DeclarerTeam = g.GetDeclarerTeam()
	resObj.LastPlayerIdx = g.GetLastPlayerIdx()
	resObj.TributeCancelled = g.IsTributeCancelled()
	resObj.MinLevel = domain.GuandanMinLevel
	resObj.MaxLevel = domain.GuandanMaxLevel
	// **上昇量は 1 / 2 / 4。**上位独占の +4 が読めないと動機が伝わらない。
	resObj.AdvanceFirstSecond = domain.GuandanAdvanceFirstSecond
	resObj.AdvanceFirstThird = domain.GuandanAdvanceFirstThird
	resObj.AdvanceFirstFourth = domain.GuandanAdvanceFirstFourth
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerTeam = g.GetWinnerTeam()

	for team := range domain.GuandanTeamCnt {
		resObj.TeamLevels[team] = g.GetTeamLevel(team)
	}

	if c := g.GetLastCombo(); c != nil {
		resObj.LastCombo = &controller.GuandanWebOutputCombo{
			Kind: int(c.Kind), Rank: c.Rank, Size: c.Size,
		}
	}

	finished := g.GetFinished()
	resObj.Finished = make([]int, 0, len(finished))
	resObj.Finished = append(resObj.Finished, finished...)

	tributes := g.GetTributes()
	resObj.Tributes = make([]*controller.GuandanWebOutputTribute, 0, len(tributes))
	for _, t := range tributes {
		if t == nil {
			continue
		}
		resObj.Tributes = append(resObj.Tributes, &controller.GuandanWebOutputTribute{
			From: t.From, To: t.To,
			Card:     cardToOutput(t.Card),
			Returned: cardToOutput(t.Returned),
		})
	}

	if r := g.GetLastResult(); r != nil {
		resObj.LastResult = &controller.GuandanWebOutputResult{
			Order:       r.Order,
			WinnerTeam:  r.WinnerTeam,
			Advance:     r.Advance,
			FirstSecond: r.FirstSecond,
		}
	}

	cfg := g.GetConfig()
	resObj.Config = controller.GuandanWebOutputConfig{CpuDifficulty: int(cfg.CpuDifficulty)}

	resObj.Players = p.buildPlayersOutput(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)

	return marshalOrError(resObj)
}

// buildPlayersOutput プレイヤー情報を構築
func (p *GuandanWebPresenter) buildPlayersOutput(g interfaces.GuandanGame) []*controller.GuandanWebOutputPlayer {
	players := g.GetPlayers()
	out := make([]*controller.GuandanWebOutputPlayer, 0, len(players))
	reveal := g.GetPhase() == domain.GuandanPhaseHandEnd || g.GetGameEndFlag()
	finished := g.GetFinished()
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
		rank := 0
		for pos, seat := range finished {
			if seat == i {
				rank = pos + 1
			}
		}
		out = append(out, &controller.GuandanWebOutputPlayer{
			ID:            i,
			IsHuman:       player.GetIsHuman(),
			Team:          domain.GuandanTeamOf(i),
			CardCount:     player.GetCardsSize(),
			Cards:         cards,
			FinishedRank:  rank,
			IsCurrentTurn: !g.GetGameEndFlag() && i == g.GetCurrentPlayerIdx(),
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *GuandanWebPresenter) buildMessage(g interfaces.GuandanGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		// **チーム戦なので勝敗は席ではなくチームで見る。**人間は席 0 = チーム 0。
		if g.GetWinnerTeam() == domain.GuandanTeamOf(0) {
			return "your team wins", "guandan.result.humanWin", nil
		}
		return "the other team wins", "guandan.result.cpuWin", nil
	}
	switch g.GetPhase() {
	case domain.GuandanPhaseTribute:
		if g.IsTributeCancelled() {
			return "", "guandan.tributeCancelled", nil
		}
		return "", "guandan.tributePhase", nil
	case domain.GuandanPhasePlay:
		return "", "guandan.playPhase", nil
	case domain.GuandanPhaseHandEnd:
		if r := g.GetLastResult(); r != nil && r.FirstSecond {
			// **上位独占は +4。**そこだけ別のメッセージにする。
			return "", "guandan.handFirstSecond", nil
		}
		return "", "guandan.handEnd", nil
	}
	return "", "", nil
}

// ActionLogOutput 棋譜をJSON出力
func (p *GuandanWebPresenter) ActionLogOutput(g interfaces.GuandanGame) string {
	return actionLogOutputJSON(g)
}
