package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// OmahaWebPresenter オマハホールデムWebプレゼンタークラス
type OmahaWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (owp *OmahaWebPresenter) Output(o interfaces.OmahaGame, lastErr error) string {
	resObj := owp.buildOutput(o, lastErr)
	return marshalOrError(resObj)
}

// buildOutput ゲーム状態をOmahaWebOutputに変換
func (owp *OmahaWebPresenter) buildOutput(o interfaces.OmahaGame, lastErr error) *controller.OmahaWebOutput {
	resObj := new(controller.OmahaWebOutput)
	resObj.Phase = o.GetPhase()
	resObj.Pot = o.GetPot()
	resObj.DealerIdx = o.GetDealerIdx()
	resObj.CurrentTurn = o.GetCurrentTurn()
	resObj.GameEndFlag = o.GetGameEndFlag()
	resObj.LastBet = o.GetLastBet()
	resObj.MinRaise = o.GetMinRaise()
	cfg := o.GetConfig()
	resObj.HandCount = o.GetHandCount()
	resObj.SmallBlind = cfg.SmallBlind
	resObj.BigBlind = cfg.BigBlind
	resObj.TournamentMode = cfg.TournamentMode
	resObj.BlindLevelHands = cfg.BlindLevelHands
	resObj.BlindMultiplier = cfg.BlindMultiplier
	resObj.BettingLimit = int(cfg.BettingLimit)
	resObj.TableSize = o.GetPlayerCnt()
	resObj.RaiseCount = o.GetRaiseCount()
	_, resObj.MaxBetAmount = domain.CalculateBettingLimits(cfg.BettingLimit, o.GetPot(), o.GetLastBet())
	resObj.RebuyAvailable = o.IsRebuyAvailable()
	resObj.AddonAvailable = o.IsAddonAvailable()
	resObj.RebuyCounts = o.GetRebuyCounts()
	resObj.AddonUsed = o.GetAddonUsed()
	resObj.RebuyEnabled = cfg.RebuyEnabled
	resObj.AddonEnabled = cfg.AddonEnabled
	resObj.RebuyMaxCount = cfg.RebuyMaxCount
	resObj.RebuyChips = cfg.RebuyChips
	resObj.AddonChips = cfg.AddonChips
	resObj.RebuyPeriodHands = cfg.RebuyPeriodHands
	resObj.AddonAfterHand = cfg.AddonAfterHand
	resObj.RebuyPhaseType = o.GetRebuyPhaseType()
	resObj.MuckAvailable = o.IsMuckAvailable()

	resObj.CommunityCards = cardsToOutput(o.GetCommunityCards())
	resObj.SidePots = owp.buildSidePotsOutput(o)
	resObj.Players = owp.buildPlayersOutput(o)
	resObj.CpuActions = owp.buildCpuActionsOutput(o)
	resObj.RoundResults = owp.buildRoundResultsOutput(o)

	eq := o.GetEquity()
	if eq != nil {
		handOdds := make([]*controller.HoldemWebOutputHandOdds, len(eq.HandOdds))
		for i, ho := range eq.HandOdds {
			handOdds[i] = &controller.HoldemWebOutputHandOdds{
				HandRank:    ho.HandRank,
				HandName:    ho.HandName,
				Probability: ho.Probability,
			}
		}
		resObj.Equity = &controller.HoldemWebOutputEquity{
			WinProbability: eq.Equity,
			HandOdds:       handOdds,
		}
		potOdds := o.GetPotOdds()
		resObj.PotOdds = &potOdds
	}

	resObj.Message, resObj.MessageCode = owp.buildMessage(o, lastErr)

	// メタAI情報
	if profile := o.GetHumanProfile(); profile != nil {
		resObj.MetaAI = &controller.HoldemWebOutputMetaAI{
			Enabled:        true,
			GamesPlayed:    profile.GamesPlayed,
			BluffRate:      profile.BluffRate(1),
			FoldRate:       profile.FoldRate(),
			HesitationMean: profile.HesitationMean,
		}
		d := profile.Export()
		resObj.Profile = &d
	}

	return resObj
}

func (owp *OmahaWebPresenter) buildSidePotsOutput(o interfaces.OmahaGame) []*controller.HoldemWebOutputSidePot {
	out := make([]*controller.HoldemWebOutputSidePot, 0)
	for _, sp := range o.GetSidePots() {
		out = append(out, &controller.HoldemWebOutputSidePot{
			Amount:          sp.Amount,
			EligiblePlayers: sp.EligiblePlayers,
		})
	}
	return out
}

func (owp *OmahaWebPresenter) buildPlayersOutput(o interfaces.OmahaGame) []*controller.HoldemWebOutputPlayer {
	out := make([]*controller.HoldemWebOutputPlayer, 0)
	isShowdown := o.GetPhase() == domain.OmahaPhaseEnd || o.GetPhase() == domain.OmahaPhaseShowdown
	for i := 0; i < o.GetPlayerCnt(); i++ {
		player := o.GetPlayer(i)
		pObj := &controller.HoldemWebOutputPlayer{
			ID:            i,
			IsHuman:       player.GetIsHuman(),
			Chips:         player.GetChips(),
			CurrentBet:    player.GetCurrentBet(),
			Folded:        player.GetFolded(),
			AllIn:         player.GetAllIn(),
			PlayStyleName: player.GetPlayStyleName(),
			TotalHands:    player.GetTotalHands(),
			VPIP:          player.GetVPIP(),
			PFR:           player.GetPFR(),
			ThreeBet:      player.GetThreeBet(),
			AF:            player.GetAFDisplay(),
		}

		pObj.Cards = playerCardsToOutput(player, player.GetIsHuman() || (isShowdown && !player.GetFolded()))

		if isShowdown && !player.GetFolded() {
			pObj.HandRank = player.GetHandRank()
			pObj.HandName = owp.getHandName(player.GetHandRank())
			pObj.BestHand = cardsToOutput(player.GetBestHand())
		} else {
			pObj.BestHand = make([]*controller.WebOutputCard, 0)
		}

		out = append(out, pObj)
	}
	return out
}

func (owp *OmahaWebPresenter) buildCpuActionsOutput(o interfaces.OmahaGame) []*controller.HoldemWebOutputCpuAction {
	out := make([]*controller.HoldemWebOutputCpuAction, 0)
	for _, action := range o.GetCpuActions() {
		out = append(out, &controller.HoldemWebOutputCpuAction{
			PlayerIdx: action.PlayerIdx,
			Action:    action.Action,
			Amount:    action.Amount,
		})
	}
	return out
}

func (owp *OmahaWebPresenter) buildRoundResultsOutput(o interfaces.OmahaGame) []*controller.HoldemWebOutputResult {
	out := make([]*controller.HoldemWebOutputResult, 0)
	for _, r := range o.GetRoundResults() {
		result := &controller.HoldemWebOutputResult{
			PlayerIdx: r.PlayerIdx,
			HandRank:  r.HandRank,
			HandName:  r.HandName,
			Kickers:   domain.FormatKickers(r.Kickers),
			WonAmount: r.WonAmount,
			Mucked:    r.Mucked,
			BestHand:  make([]*controller.WebOutputCard, 0),
		}
		if r.Mucked {
			result.HandRank = 0
			result.HandName = ""
			result.Kickers = ""
		} else {
			result.BestHand = cardsToOutput(r.BestHand)
		}
		out = append(out, result)
	}
	return out
}

func (owp *OmahaWebPresenter) buildMessage(o interfaces.OmahaGame, lastErr error) (string, string) {
	if lastErr != nil {
		return lastErr.Error(), ""
	}
	if o.IsMuckAvailable() {
		return "Muck or show your hand.", "omaha.muck.prompt"
	}
	if o.GetGameEndFlag() {
		return owp.buildResultMessage(o)
	}
	return "", ""
}

func (owp *OmahaWebPresenter) buildResultMessage(o interfaces.OmahaGame) (string, string) {
	results := o.GetRoundResults()
	if len(results) == 0 {
		return "Game over.", "omaha.result.gameOver"
	}

	for _, r := range results {
		if o.GetPlayer(r.PlayerIdx).GetIsHuman() {
			if r.WonAmount > 0 {
				return "You are the winner.", "omaha.result.win"
			}
		}
	}

	for i := 0; i < o.GetPlayerCnt(); i++ {
		if o.GetPlayer(i).GetIsHuman() && o.GetPlayer(i).GetFolded() {
			return "You folded.", "omaha.result.folded"
		}
	}

	for _, r := range results {
		if o.GetPlayer(r.PlayerIdx).GetIsHuman() && r.Mucked {
			return "You mucked.", "omaha.result.mucked"
		}
	}

	return "You lose.", "omaha.result.lose"
}

// ActionLogOutput 棋譜をJSON出力
func (owp *OmahaWebPresenter) ActionLogOutput(o interfaces.OmahaGame) string {
	return actionLogOutputJSON(o)
}

func (owp *OmahaWebPresenter) getHandName(rank int) string {
	if rank >= 0 && rank < len(domain.PokerHandNames) {
		return domain.PokerHandNames[rank]
	}
	return "Unknown"
}
