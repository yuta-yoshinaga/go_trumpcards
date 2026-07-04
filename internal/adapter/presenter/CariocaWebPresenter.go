//go:build !js || !wasm || extra

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// CariocaWebPresenter カリオカ Web プレゼンター
type CariocaWebPresenter struct{}

// Output ゲーム状態を JSON 出力
func (p *CariocaWebPresenter) Output(g interfaces.CariocaGame, lastErr error) string {
	resObj := new(controller.CariocaWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TotalRounds = domain.CariocaTotalRounds
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.DrawPileCount = g.GetDrawPileCount()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()
	resObj.RoundWinnerIdx = g.GetRoundWinnerIdx()

	top := g.GetDiscardTop()
	if top != nil {
		resObj.DiscardTop = cardToOutput(top)
	}

	cfg := g.GetConfig()
	resObj.Config = controller.CariocaWebOutputConfig{
		PlayerCount:         cfg.PlayerCount,
		CpuDifficulty:       int(cfg.CpuDifficulty),
		FailContractPenalty: cfg.FailContractPenalty,
	}

	contract := g.GetCurrentContract()
	resObj.ContractSlots = make([]*controller.CariocaWebOutputContractSlot, 0, len(contract.Slots))
	for _, slot := range contract.Slots {
		resObj.ContractSlots = append(resObj.ContractSlots, &controller.CariocaWebOutputContractSlot{
			Kind: int(slot.Kind),
			Size: slot.Size,
		})
	}

	resObj.Players = p.buildPlayersOutput(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)

	return marshalOrError(resObj)
}

// buildPlayersOutput プレイヤー情報を構築
func (p *CariocaWebPresenter) buildPlayersOutput(g interfaces.CariocaGame) []*controller.CariocaWebOutputPlayer {
	out := make([]*controller.CariocaWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		showCards := player.GetIsHuman()
		phase := g.GetPhase()
		if phase == domain.CariocaPhaseRoundEnd || phase == domain.CariocaPhaseGameEnd {
			showCards = true
		}
		melds := make([]*controller.CariocaWebOutputMeld, 0, player.GetMeldCount())
		for mi := 0; mi < player.GetMeldCount(); mi++ {
			meld := player.GetMeld(mi)
			m := &controller.CariocaWebOutputMeld{
				Cards: make([]*controller.WebOutputCard, 0, len(meld)),
			}
			for _, c := range meld {
				m.Cards = append(m.Cards, cardToOutput(c))
			}
			melds = append(melds, m)
		}
		pObj := &controller.CariocaWebOutputPlayer{
			ID:              i,
			IsHuman:         player.GetIsHuman(),
			CardCount:       player.GetCardsSize(),
			Cards:           playerCardsToOutput(player, showCards),
			Melds:           melds,
			ContractMet:     player.IsContractMet(),
			RoundScore:      player.GetRoundScore(),
			CumulativeScore: player.GetCumulativeScore(),
		}
		out = append(out, pObj)
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *CariocaWebPresenter) buildMessage(g interfaces.CariocaGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		winnerIdx := g.GetWinnerIdx()
		player := g.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		return buildWinnerWebMessage("carioca", winnerIdx, isHuman)
	}
	switch g.GetPhase() {
	case domain.CariocaPhaseDraw:
		return "", "carioca.drawPhase", nil
	case domain.CariocaPhasePlay:
		return "", "carioca.playPhase", nil
	case domain.CariocaPhaseRoundEnd:
		return "", "carioca.roundEnd", nil
	}
	return "", "", nil
}

// ActionLogOutput 棋譜を JSON 出力
func (p *CariocaWebPresenter) ActionLogOutput(g interfaces.CariocaGame) string {
	return actionLogOutputJSON(g)
}
