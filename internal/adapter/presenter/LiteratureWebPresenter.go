//go:build !js || !wasm || solo

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// LiteratureWebPresenter リテラチャー Webプレゼンタークラス
type LiteratureWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *LiteratureWebPresenter) Output(g interfaces.LiteratureGame, lastErr error) string {
	resObj := new(controller.LiteratureWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.WinThreshold = domain.LiteratureWinThreshold
	resObj.HalfSuitCnt = domain.LiteratureHalfSuitCnt
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerTeam = g.GetWinnerTeam()
	resObj.CancelledCount = g.LiteratureCancelledCount()
	resObj.OpenCount = g.LiteratureOpenCount()

	for team := range domain.LiteratureTeamCnt {
		resObj.TeamHalfSuits[team] = g.LiteratureTeamHalfSuits(team)
	}
	for half := range domain.LiteratureHalfSuitCnt {
		resObj.HalfSuits[half] = int(g.GetHalfSuitState(half))
		cards := domain.LiteratureHalfSuitCards(half)
		out := make([]*controller.WebOutputCard, 0, len(cards))
		for _, c := range cards {
			if wc := cardToOutput(c); wc != nil {
				out = append(out, wc)
			}
		}
		resObj.HalfSuitCards[half] = out
	}

	// **要求の履歴は公開情報。**推理の材料なので全部送る。
	asks := g.GetAsks()
	resObj.Asks = make([]*controller.LiteratureWebOutputAsk, 0, len(asks))
	for _, a := range asks {
		if a == nil {
			continue
		}
		resObj.Asks = append(resObj.Asks, literatureAskOut(a))
	}
	if a := g.GetLastAsk(); a != nil {
		resObj.LastAsk = literatureAskOut(a)
	}

	claims := g.GetClaims()
	resObj.Claims = make([]*controller.LiteratureWebOutputClaim, 0, len(claims))
	for _, c := range claims {
		if c == nil {
			continue
		}
		resObj.Claims = append(resObj.Claims, literatureClaimOut(c))
	}
	if c := g.GetLastClaim(); c != nil {
		resObj.LastClaim = literatureClaimOut(c)
	}

	cfg := g.GetConfig()
	resObj.Config = controller.LiteratureWebOutputConfig{CpuDifficulty: int(cfg.CpuDifficulty)}

	resObj.Players = p.buildPlayersOutput(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)

	return marshalOrError(resObj)
}

// literatureAskOut は 1 件の要求をワイヤ表現へ変換する。
func literatureAskOut(a *domain.LiteratureAsk) *controller.LiteratureWebOutputAsk {
	return &controller.LiteratureWebOutputAsk{
		From: a.From, To: a.To, Card: cardToOutput(a.Card), Success: a.Success,
	}
}

// literatureClaimOut は 1 件の宣言をワイヤ表現へ変換する。
func literatureClaimOut(c *domain.LiteratureClaimResult) *controller.LiteratureWebOutputClaim {
	return &controller.LiteratureWebOutputClaim{
		Player: c.Player, HalfSuit: c.HalfSuit, Outcome: int(c.Outcome), AwardedTeam: c.AwardedTeam,
	}
}

// buildPlayersOutput プレイヤー情報を構築
func (p *LiteratureWebPresenter) buildPlayersOutput(g interfaces.LiteratureGame) []*controller.LiteratureWebOutputPlayer {
	players := g.GetPlayers()
	out := make([]*controller.LiteratureWebOutputPlayer, 0, len(players))
	// **終局まで誰の手札も公開しない。**味方の手札が見えたら推理が成立しない。
	reveal := g.GetGameEndFlag()
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
		out = append(out, &controller.LiteratureWebOutputPlayer{
			ID:            i,
			IsHuman:       player.GetIsHuman(),
			Team:          domain.LiteratureTeamOf(i),
			CardCount:     player.GetCardsSize(),
			Cards:         cards,
			IsCurrentTurn: !g.GetGameEndFlag() && i == g.GetCurrentPlayerIdx(),
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *LiteratureWebPresenter) buildMessage(g interfaces.LiteratureGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		switch g.GetWinnerTeam() {
		case domain.LiteratureTeamOf(0):
			return "your team wins", "literature.result.humanWin", nil
		case -1:
			// **無効が絡むと同数で終わることがある。**
			return "nobody wins", "literature.result.draw", nil
		default:
			return "the other team wins", "literature.result.cpuWin", nil
		}
	}
	// **直前の宣言の結末は 3 通り。**「無効」は「相手に渡る」とは違う。
	if c := g.GetLastClaim(); c != nil {
		switch c.Outcome {
		case domain.LiteratureClaimCancelled:
			return "", "literature.claimCancelled", nil
		case domain.LiteratureClaimLost:
			return "", "literature.claimLost", nil
		case domain.LiteratureClaimWon:
			return "", "literature.claimWon", nil
		}
	}
	if a := g.GetLastAsk(); a != nil {
		if a.Success {
			return "", "literature.askHit", nil
		}
		return "", "literature.askMiss", nil
	}
	return "", "literature.playPhase", nil
}

// ActionLogOutput 棋譜をJSON出力
func (p *LiteratureWebPresenter) ActionLogOutput(g interfaces.LiteratureGame) string {
	return actionLogOutputJSON(g)
}
