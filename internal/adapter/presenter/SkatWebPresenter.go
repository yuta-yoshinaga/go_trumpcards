//go:build !js || !wasm || extra3

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// SkatWebPresenter Skat Web presenter.
type SkatWebPresenter struct{}

// Output renders the game state as a JSON string.
func (p *SkatWebPresenter) Output(s interfaces.SkatGame, lastErr error) string {
	resObj := p.buildBaseOutput(s)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(s, lastErr)
	return marshalOrError(resObj)
}

// buildBaseOutput builds the base output struct.
func (p *SkatWebPresenter) buildBaseOutput(s interfaces.SkatGame) *controller.SkatWebOutput {
	resObj := new(controller.SkatWebOutput)
	resObj.Phase = int(s.GetPhase())
	resObj.RoundNumber = s.GetRoundNumber()
	resObj.TrickNumber = s.GetTrickNumber()
	resObj.CurrentPlayerIdx = s.GetCurrentPlayerIdx()
	resObj.ForehandIdx = s.GetForehandIdx()
	resObj.MiddlehandIdx = s.GetMiddlehandIdx()
	resObj.RearhandIdx = s.GetRearhandIdx()
	resObj.DealerIdx = s.GetDealerIdx()
	resObj.DeclarerIdx = s.GetDeclarerIdx()
	resObj.CurrentBid = s.GetCurrentBid()
	resObj.ActiveBidActorIdx = s.GetActiveBidActorIdx()
	resObj.GameType = int(s.GetGameType())
	resObj.TrumpSuit = s.GetTrumpSuit()
	resObj.PickedSkat = s.PickedSkat()
	resObj.DeclarerCardPoints = s.GetDeclarerCardPoints()
	resObj.DefendersCardPoints = s.GetDefendersCardPoints()
	resObj.WinnerSide = s.GetWinnerSide()
	resObj.GameValue = s.GetGameValue()
	resObj.GameEndFlag = s.GetGameEndFlag()
	resObj.LeadPlayerIdx = s.GetLeadPlayerIdx()

	// Skat is exposed only when face-up to the declarer (post-pickup before
	// discard) or once revealed at round end.
	phase := s.GetPhase()
	if phase == domain.SkatPhaseRoundEnd || phase == domain.SkatPhaseGameEnd {
		// Reveal the skat at round end so the user can see the final cards.
		if orig := s.GetOriginalSkat(); len(orig) > 0 {
			resObj.OriginalSkat = make([]*controller.WebOutputCard, len(orig))
			for i, c := range orig {
				resObj.OriginalSkat[i] = cardToOutput(c)
			}
		}
		if cur := s.GetSkat(); len(cur) > 0 {
			resObj.Skat = make([]*controller.WebOutputCard, len(cur))
			for i, c := range cur {
				resObj.Skat[i] = cardToOutput(c)
			}
		}
	}

	cfg := s.GetConfig()
	resObj.Config = controller.SkatWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetScore:   cfg.TargetScore,
	}

	resObj.CurrentTrick = trickCardsToOutput(s.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(s)
	return resObj
}

// buildPlayersOutput builds the per-player output.
func (p *SkatWebPresenter) buildPlayersOutput(s interfaces.SkatGame) []*controller.SkatWebOutputPlayer {
	out := make([]*controller.SkatWebOutputPlayer, 0)
	for i := 0; i < s.GetPlayerCnt(); i++ {
		player := s.GetPlayer(i)
		out = append(out, &controller.SkatWebOutputPlayer{
			ID:              i,
			IsHuman:         player.GetIsHuman(),
			CardCount:       player.GetCardsSize(),
			Cards:           playerCardsToOutput(player, player.GetIsHuman()),
			Bid:             player.GetBid(),
			IsDeclarer:      player.GetIsDeclarer(),
			CardPoints:      player.GetCardPoints(),
			RoundsWon:       player.GetRoundsWon(),
			RoundsLost:      player.GetRoundsLost(),
			RoundScore:      player.GetRoundScore(),
			CumulativeScore: player.GetCumulativeScore(),
			TrickCount:      player.GetTrickCount(),
		})
	}
	return out
}

// buildMessage produces a message and i18n message code for the current state.
func (p *SkatWebPresenter) buildMessage(s interfaces.SkatGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if s.GetGameEndFlag() {
		return "", "skat.gameEnd", nil
	}
	switch s.GetPhase() {
	case domain.SkatPhaseBid:
		return "", "skat.bidPhase", nil
	case domain.SkatPhaseSkatPickup:
		return "", "skat.skatPickup", nil
	case domain.SkatPhaseDiscard:
		return "", "skat.discard", nil
	case domain.SkatPhaseGameDeclaration:
		return "", "skat.gameDeclaration", nil
	case domain.SkatPhasePlay:
		if len(s.GetCurrentTrick()) == 0 {
			return "", "skat.playPhase.lead", nil
		}
		return "", "skat.playPhase.follow", nil
	case domain.SkatPhaseTrickEnd:
		return "", "skat.trickEnd", nil
	case domain.SkatPhaseRoundEnd:
		return "", "skat.roundEnd", nil
	}
	return "", "", nil
}

// HintOutput renders the hint output.
func (p *SkatWebPresenter) HintOutput(s interfaces.SkatGame) string {
	hint := s.GetHint()
	resObj := p.buildBaseOutput(s)
	if hint != nil {
		resObj.Hint = &controller.SkatWebOutputHint{
			CardIndex:    hint.CardIndex,
			Bid:          hint.Bid,
			GameType:     hint.GameType,
			TrumpSuit:    hint.TrumpSuit,
			PickSkat:     hint.PickSkat,
			DiscardIndex: hint.DiscardIndex,
			Reason:       hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput returns the round's action log as JSON.
func (p *SkatWebPresenter) ActionLogOutput(s interfaces.SkatGame) string {
	return actionLogOutputJSON(s)
}
