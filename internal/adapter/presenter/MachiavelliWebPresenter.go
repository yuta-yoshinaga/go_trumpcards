//go:build !js || !wasm || extra

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MachiavelliWebPresenter マキャヴェッリ Web プレゼンター
type MachiavelliWebPresenter struct{}

// Output ゲーム状態を JSON 出力
func (p *MachiavelliWebPresenter) Output(g interfaces.MachiavelliGame, lastErr error) string {
	resObj := new(controller.MachiavelliWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TargetRounds = g.GetTargetRounds()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.DrawPileCount = g.GetDrawPileCount()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()
	resObj.RoundWinnerIdx = g.GetRoundWinnerIdx()

	cfg := g.GetConfig()
	resObj.Config = controller.MachiavelliWebOutputConfig{
		PlayerCount:   cfg.PlayerCount,
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetRounds:  cfg.TargetRounds,
	}

	resObj.Table = machiavelliBuildTableOutput(g.GetTable())
	resObj.Players = p.buildPlayersOutput(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)

	return marshalOrError(resObj)
}

// machiavelliBuildTableOutput 共有テーブルのメルド群を出力用に変換する。
func machiavelliBuildTableOutput(table [][]*domain.Card) []*controller.MachiavelliWebOutputMeld {
	out := make([]*controller.MachiavelliWebOutputMeld, 0, len(table))
	for _, meld := range table {
		cards := make([]*controller.WebOutputCard, 0, len(meld))
		for _, c := range meld {
			cards = append(cards, cardToOutput(c))
		}
		out = append(out, &controller.MachiavelliWebOutputMeld{
			Cards: cards,
			Kind:  machiavelliMeldKind(meld),
		})
	}
	return out
}

// machiavelliMeldKind はメルドの種別を返す（0=set, 1=run）。
// 有効メルドでは、全カードが同ランクならセット、そうでなければランと判別できる。
func machiavelliMeldKind(meld []*domain.Card) int {
	if len(meld) == 0 {
		return 0
	}
	rank := meld[0].GetValue()
	for _, c := range meld {
		if c.GetValue() != rank {
			return 1 // run
		}
	}
	return 0 // set
}

// buildPlayersOutput プレイヤー情報を構築
func (p *MachiavelliWebPresenter) buildPlayersOutput(g interfaces.MachiavelliGame) []*controller.MachiavelliWebOutputPlayer {
	out := make([]*controller.MachiavelliWebOutputPlayer, 0)
	phase := g.GetPhase()
	revealAll := phase == domain.MachiavelliPhaseRoundEnd || phase == domain.MachiavelliPhaseGameEnd
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		showCards := player.GetIsHuman() || revealAll
		deadwood := 0
		if revealAll {
			deadwood = g.PlayerDeadwoodValue(i)
		}
		pObj := &controller.MachiavelliWebOutputPlayer{
			ID:              i,
			IsHuman:         player.GetIsHuman(),
			CardCount:       player.GetCardsSize(),
			Cards:           playerCardsToOutput(player, showCards),
			RoundScore:      player.GetRoundScore(),
			CumulativeScore: player.GetCumulativeScore(),
			Deadwood:        deadwood,
		}
		out = append(out, pObj)
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *MachiavelliWebPresenter) buildMessage(g interfaces.MachiavelliGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		winnerIdx := g.GetWinnerIdx()
		player := g.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		return buildWinnerWebMessage("machiavelli", winnerIdx, isHuman)
	}
	switch g.GetPhase() {
	case domain.MachiavelliPhaseTurn:
		return "", "machiavelli.turnPhase", nil
	case domain.MachiavelliPhaseRoundEnd:
		return "", "machiavelli.roundEnd", nil
	}
	return "", "", nil
}

// ActionLogOutput 棋譜を JSON 出力
func (p *MachiavelliWebPresenter) ActionLogOutput(g interfaces.MachiavelliGame) string {
	return actionLogOutputJSON(g)
}
