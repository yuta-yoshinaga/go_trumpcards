//go:build !js || !wasm || extra

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// ContractRummyWebPresenter コントラクトラミー Web プレゼンター
type ContractRummyWebPresenter struct{}

// Output ゲーム状態を JSON 出力
func (p *ContractRummyWebPresenter) Output(g interfaces.ContractRummyGame, lastErr error) string {
	resObj := new(controller.ContractRummyWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TotalRounds = domain.ContractRummyTotalRounds
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
	resObj.Config = controller.ContractRummyWebOutputConfig{
		CpuDifficulty:       int(cfg.CpuDifficulty),
		FailContractPenalty: cfg.FailContractPenalty,
	}

	contract := g.GetCurrentContract()
	resObj.ContractSlots = make([]*controller.ContractRummyWebOutputContractSlot, 0, len(contract.Slots))
	for _, slot := range contract.Slots {
		resObj.ContractSlots = append(resObj.ContractSlots, &controller.ContractRummyWebOutputContractSlot{
			Kind: int(slot.Kind),
			Size: slot.Size,
		})
	}

	resObj.Players = p.buildPlayersOutput(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)

	return marshalOrError(resObj)
}

// buildPlayersOutput プレイヤー情報を構築
func (p *ContractRummyWebPresenter) buildPlayersOutput(g interfaces.ContractRummyGame) []*controller.ContractRummyWebOutputPlayer {
	out := make([]*controller.ContractRummyWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		showCards := player.GetIsHuman()
		phase := g.GetPhase()
		if phase == domain.ContractRummyPhaseRoundEnd || phase == domain.ContractRummyPhaseGameEnd {
			showCards = true
		}
		melds := make([]*controller.ContractRummyWebOutputMeld, 0, player.GetMeldCount())
		for mi := 0; mi < player.GetMeldCount(); mi++ {
			meld := player.GetMeld(mi)
			m := &controller.ContractRummyWebOutputMeld{
				Cards: make([]*controller.WebOutputCard, 0, len(meld)),
			}
			for _, c := range meld {
				m.Cards = append(m.Cards, cardToOutput(c))
			}
			melds = append(melds, m)
		}
		pObj := &controller.ContractRummyWebOutputPlayer{
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
func (p *ContractRummyWebPresenter) buildMessage(g interfaces.ContractRummyGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		winnerIdx := g.GetWinnerIdx()
		player := g.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		return buildWinnerWebMessage("contractrummy", winnerIdx, isHuman)
	}
	switch g.GetPhase() {
	case domain.ContractRummyPhaseDraw:
		return "", "contractrummy.drawPhase", nil
	case domain.ContractRummyPhasePlay:
		return "", "contractrummy.playPhase", nil
	case domain.ContractRummyPhaseRoundEnd:
		return "", "contractrummy.roundEnd", nil
	}
	return "", "", nil
}

// ActionLogOutput 棋譜を JSON 出力
func (p *ContractRummyWebPresenter) ActionLogOutput(g interfaces.ContractRummyGame) string {
	return actionLogOutputJSON(g)
}
