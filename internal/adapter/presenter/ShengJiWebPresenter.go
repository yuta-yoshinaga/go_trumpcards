//go:build !js || !wasm || extra4

package presenter

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// ShengJiWebPresenter 升级 Webプレゼンタークラス
type ShengJiWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *ShengJiWebPresenter) Output(g interfaces.ShengJiGame, lastErr error) string {
	resObj := new(controller.ShengJiWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.HandNumber = g.GetHandNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.Level = g.GetLevel()
	resObj.DeclarerTeam = g.GetDeclarerTeam()
	resObj.TrumpSuit = g.GetTrumpSuit()
	resObj.KittySize = g.GetKittySize()
	resObj.TrickLeader = g.GetTrickLeader()
	resObj.TrickCount = g.GetTrickCount()
	resObj.LastTrickWinner = g.GetLastTrickWinner()
	resObj.MinLevel = domain.ShengJiMinLevel
	resObj.MaxLevel = domain.ShengJiMaxLevel
	resObj.KittySizeMax = domain.ShengJiKittySize
	// **80 点は 200 点の 4 割。**この 2 つが読めないと守備側の目標が伝わらない。
	resObj.TotalPoints = domain.ShengJiTotalPoints
	resObj.DefenderTarget = domain.ShengJiDefenderTarget
	resObj.AdvanceStep = domain.ShengJiAdvanceStep
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerTeam = g.GetWinnerTeam()

	for team := range domain.ShengJiTeamCnt {
		resObj.TeamLevels[team] = g.GetTeamLevel(team)
		resObj.TeamPoints[team] = g.GetTeamPoints(team)
	}

	if d := g.GetDeclaration(); d != nil {
		resObj.Declaration = &controller.ShengJiWebOutputDeclaration{
			Seat: d.Seat, Suit: d.Suit, Strength: d.Strength,
		}
	}
	if c := g.GetLeadCombo(); c != nil {
		resObj.LeadCombo = &controller.ShengJiWebOutputCombo{
			Kind: int(c.Kind), Rank: c.Rank, Size: c.Size, Trump: c.Trump, Suit: c.Suit,
		}
	}

	// **底牌は終局まで送らない。**送ると宣言側の埋め方が筒抜けになる。
	resObj.Kitty = make([]*controller.WebOutputCard, 0, domain.ShengJiKittySize)
	for _, c := range g.GetKitty() {
		if out := cardToOutput(c); out != nil {
			resObj.Kitty = append(resObj.Kitty, out)
		}
	}

	resObj.Trick = p.buildTrickOutput(g)
	resObj.DeclarableSuits = p.buildDeclarableSuits(g)

	if r := g.GetLastResult(); r != nil {
		resObj.LastResult = &controller.ShengJiWebOutputResult{
			DeclarerTeam: r.DeclarerTeam, DefenderPoints: r.DefenderPoints,
			KittyPoints: r.KittyPoints, KittyMultiplier: r.KittyMultiplier,
			DeclarerHeld: r.DeclarerHeld, Advance: r.Advance, AdvancingTeam: r.AdvancingTeam,
		}
	}

	cfg := g.GetConfig()
	resObj.Config = controller.ShengJiWebOutputConfig{CpuDifficulty: int(cfg.CpuDifficulty)}

	resObj.Players = p.buildPlayersOutput(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)

	return marshalOrError(resObj)
}

// buildPlayersOutput プレイヤー情報を構築
func (p *ShengJiWebPresenter) buildPlayersOutput(g interfaces.ShengJiGame) []*controller.ShengJiWebOutputPlayer {
	players := g.GetPlayers()
	out := make([]*controller.ShengJiWebOutputPlayer, 0, len(players))
	reveal := g.GetPhase() == domain.ShengJiPhaseHandEnd || g.GetGameEndFlag()
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
		out = append(out, &controller.ShengJiWebOutputPlayer{
			ID:            i,
			IsHuman:       player.GetIsHuman(),
			Team:          domain.ShengJiTeamOf(i),
			CardCount:     player.GetCardsSize(),
			Cards:         cards,
			IsDeclarer:    domain.ShengJiTeamOf(i) == g.GetDeclarerTeam(),
			IsCurrentTurn: !g.GetGameEndFlag() && i == g.GetCurrentPlayerIdx(),
		})
	}
	return out
}

// buildTrickOutput いまのトリックを構築
func (p *ShengJiWebPresenter) buildTrickOutput(g interfaces.ShengJiGame) []*controller.ShengJiWebOutputPlay {
	trick := g.GetTrick()
	out := make([]*controller.ShengJiWebOutputPlay, 0, len(trick))
	for i, play := range trick {
		cards := make([]*controller.WebOutputCard, 0, len(play))
		for _, c := range play {
			if out := cardToOutput(c); out != nil {
				cards = append(cards, out)
			}
		}
		out = append(out, &controller.ShengJiWebOutputPlay{
			// トリックはリード順なので、席は先頭からの距離で決まる。
			Seat:  (g.GetTrickLeader() + i) % domain.ShengJiPlayerCnt,
			Cards: cards,
		})
	}
	return out
}

// buildDeclarableSuits 人間がいま亮牌できるスートと強さを構築
func (p *ShengJiWebPresenter) buildDeclarableSuits(g interfaces.ShengJiGame) map[string]int {
	out := map[string]int{}
	if g.GetPhase() != domain.ShengJiPhaseDeclare {
		return out
	}
	seat := g.GetCurrentPlayerIdx()
	player := g.GetPlayer(seat)
	if player == nil || !player.GetIsHuman() {
		return out
	}
	for suit := domain.CardDesignSpade; suit <= domain.CardDesignDiamond; suit++ {
		if st := g.ShengJiDeclareStrength(seat, suit); st > 0 {
			out[strconv.Itoa(suit)] = st
		}
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *ShengJiWebPresenter) buildMessage(g interfaces.ShengJiGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		// **チーム戦なので勝敗は席ではなくチームで見る。**人間は席 0 = チーム 0。
		if g.GetWinnerTeam() == domain.ShengJiTeamOf(0) {
			return "your team wins", "shengji.result.humanWin", nil
		}
		return "the other team wins", "shengji.result.cpuWin", nil
	}
	switch g.GetPhase() {
	case domain.ShengJiPhaseDeclare:
		return "", "shengji.declarePhase", nil
	case domain.ShengJiPhaseKitty:
		return "", "shengji.kittyPhase", nil
	case domain.ShengJiPhasePlay:
		return "", "shengji.playPhase", nil
	case domain.ShengJiPhaseHandEnd:
		if r := g.GetLastResult(); r != nil && !r.DeclarerHeld {
			// **80 点で宣言側が交代する。**守りきった局とは別のメッセージにする。
			return "", "shengji.handTaken", nil
		}
		return "", "shengji.handHeld", nil
	}
	return "", "", nil
}

// ActionLogOutput 棋譜をJSON出力
func (p *ShengJiWebPresenter) ActionLogOutput(g interfaces.ShengJiGame) string {
	return actionLogOutputJSON(g)
}
