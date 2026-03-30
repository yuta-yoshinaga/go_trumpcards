package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// PineappleWebPresenter パイナップルポーカーWebプレゼンタークラス
type PineappleWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (pp *PineappleWebPresenter) Output(p interfaces.PineappleGame, lastErr error) string {
	resObj := pp.buildOutput(p, lastErr)
	return marshalOrError(resObj)
}

// buildOutput ゲーム状態をPineappleWebOutputに変換
func (pp *PineappleWebPresenter) buildOutput(p interfaces.PineappleGame, lastErr error) *controller.PineappleWebOutput {
	resObj := new(controller.PineappleWebOutput)
	resObj.Phase = p.GetPhase()
	resObj.Pot = p.GetPot()
	resObj.DealerIdx = p.GetDealerIdx()
	resObj.CurrentTurn = p.GetCurrentTurn()
	resObj.GameEndFlag = p.GetGameEndFlag()
	resObj.LastBet = p.GetLastBet()
	resObj.MinRaise = p.GetMinRaise()
	cfg := p.GetConfig()
	resObj.HandCount = p.GetHandCount()
	resObj.SmallBlind = cfg.SmallBlind
	resObj.BigBlind = cfg.BigBlind
	resObj.TournamentMode = cfg.TournamentMode
	resObj.BlindLevelHands = cfg.BlindLevelHands
	resObj.BlindMultiplier = cfg.BlindMultiplier
	resObj.BettingLimit = int(cfg.BettingLimit)
	resObj.TableSize = p.GetPlayerCnt()
	resObj.RaiseCount = p.GetRaiseCount()
	_, resObj.MaxBetAmount = domain.CalculateBettingLimits(cfg.BettingLimit, p.GetPot(), p.GetLastBet())
	resObj.RebuyAvailable = p.IsRebuyAvailable()
	resObj.AddonAvailable = p.IsAddonAvailable()
	resObj.RebuyCounts = p.GetRebuyCounts()
	resObj.AddonUsed = p.GetAddonUsed()
	resObj.RebuyEnabled = cfg.RebuyEnabled
	resObj.AddonEnabled = cfg.AddonEnabled
	resObj.RebuyMaxCount = cfg.RebuyMaxCount
	resObj.RebuyChips = cfg.RebuyChips
	resObj.AddonChips = cfg.AddonChips
	resObj.RebuyPeriodHands = cfg.RebuyPeriodHands
	resObj.AddonAfterHand = cfg.AddonAfterHand
	resObj.RebuyPhaseType = p.GetRebuyPhaseType()
	resObj.MuckAvailable = p.IsMuckAvailable()
	resObj.IsDiscardPhase = p.IsDiscardPhase()
	resObj.DiscardDone = p.GetDiscardDone()

	resObj.CommunityCards = cardsToOutput(p.GetCommunityCards())
	resObj.SidePots = pp.buildSidePotsOutput(p)
	resObj.Players = pp.buildPlayersOutput(p)
	resObj.CpuActions = pp.buildCpuActionsOutput(p)
	resObj.RoundResults = pp.buildRoundResultsOutput(p)

	// エクイティ情報
	eq := p.GetEquity()
	if eq != nil {
		handOdds := make([]*controller.HoldemWebOutputHandOdds, len(eq.HandOdds))
		for i, ho := range eq.HandOdds {
			handOdds[i] = &controller.HoldemWebOutputHandOdds{
				HandRank:    ho.HandRank,
				HandName:    ho.HandName,
				Probability: ho.Probability,
			}
		}
		resObj.Equity = &controller.PineappleWebOutputEquity{
			WinProbability: eq.Equity,
			HandOdds:       handOdds,
		}
		potOdds := p.GetPotOdds()
		resObj.PotOdds = &potOdds
	}

	resObj.Message, resObj.MessageCode = pp.buildMessage(p, lastErr)

	// メタAI情報
	if profile := p.GetHumanProfile(); profile != nil {
		resObj.MetaAI = &controller.PineappleWebOutputMetaAI{
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

// buildSidePotsOutput サイドポット情報を構築
func (pp *PineappleWebPresenter) buildSidePotsOutput(p interfaces.PineappleGame) []*controller.PineappleWebOutputSidePot {
	out := make([]*controller.PineappleWebOutputSidePot, 0)
	for _, sp := range p.GetSidePots() {
		out = append(out, &controller.PineappleWebOutputSidePot{
			Amount:          sp.Amount,
			EligiblePlayers: sp.EligiblePlayers,
		})
	}
	return out
}

// buildPlayersOutput プレイヤー情報を構築
func (pp *PineappleWebPresenter) buildPlayersOutput(p interfaces.PineappleGame) []*controller.PineappleWebOutputPlayer {
	out := make([]*controller.PineappleWebOutputPlayer, 0)
	isShowdown := p.GetPhase() == domain.PineapplePhaseEnd || p.GetPhase() == domain.PineapplePhaseShowdown
	for i := 0; i < p.GetPlayerCnt(); i++ {
		player := p.GetPlayer(i)
		pObj := &controller.PineappleWebOutputPlayer{
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

		// 人間のカードは常に表示、CPUのカードはショーダウン時のみ表示
		pObj.Cards = playerCardsToOutput(player, player.GetIsHuman() || (isShowdown && !player.GetFolded()))

		// ショーダウン時のハンド情報
		if isShowdown && !player.GetFolded() {
			pObj.HandRank = player.GetHandRank()
			pObj.HandName = pp.getHandName(player.GetHandRank())
			pObj.BestHand = cardsToOutput(player.GetBestHand())
		} else {
			pObj.BestHand = make([]*controller.WebOutputCard, 0)
		}

		out = append(out, pObj)
	}
	return out
}

// buildCpuActionsOutput CPU行動記録を構築
func (pp *PineappleWebPresenter) buildCpuActionsOutput(p interfaces.PineappleGame) []*controller.PineappleWebOutputCpuAction {
	out := make([]*controller.PineappleWebOutputCpuAction, 0)
	for _, action := range p.GetCpuActions() {
		out = append(out, &controller.PineappleWebOutputCpuAction{
			PlayerIdx: action.PlayerIdx,
			Action:    action.Action,
			Amount:    action.Amount,
		})
	}
	return out
}

// buildRoundResultsOutput ラウンド結果を構築
func (pp *PineappleWebPresenter) buildRoundResultsOutput(p interfaces.PineappleGame) []*controller.PineappleWebOutputResult {
	out := make([]*controller.PineappleWebOutputResult, 0)
	for _, r := range p.GetRoundResults() {
		result := &controller.PineappleWebOutputResult{
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

// buildMessage ゲーム結果メッセージを構築
func (pp *PineappleWebPresenter) buildMessage(p interfaces.PineappleGame, lastErr error) (string, string) {
	if lastErr != nil {
		return lastErr.Error(), ""
	}
	if p.IsDiscardPhase() {
		return "Select a card to discard.", "pineapple.discard.prompt"
	}
	if p.IsMuckAvailable() {
		return "Muck or show your hand.", "pineapple.muck.prompt"
	}
	if p.GetGameEndFlag() {
		return pp.buildResultMessage(p)
	}
	return "", ""
}

// buildResultMessage builds the end-of-round message and its i18n code
func (pp *PineappleWebPresenter) buildResultMessage(p interfaces.PineappleGame) (string, string) {
	results := p.GetRoundResults()
	if len(results) == 0 {
		return "Game over.", "pineapple.result.gameOver"
	}

	for _, r := range results {
		if p.GetPlayer(r.PlayerIdx).GetIsHuman() {
			if r.WonAmount > 0 {
				return "You are the winner.", "pineapple.result.win"
			}
		}
	}

	// Human not in results (folded)
	for i := 0; i < p.GetPlayerCnt(); i++ {
		if p.GetPlayer(i).GetIsHuman() && p.GetPlayer(i).GetFolded() {
			return "You folded.", "pineapple.result.folded"
		}
	}

	// Human mucked
	for _, r := range results {
		if p.GetPlayer(r.PlayerIdx).GetIsHuman() && r.Mucked {
			return "You mucked.", "pineapple.result.mucked"
		}
	}

	return "You lose.", "pineapple.result.lose"
}

// ActionLogOutput 棋譜をJSON出力
func (pp *PineappleWebPresenter) ActionLogOutput(p interfaces.PineappleGame) string {
	return actionLogOutputJSON(p)
}

// getHandName ハンドランクから名前を返す
func (pp *PineappleWebPresenter) getHandName(rank int) string {
	if rank >= 0 && rank < len(domain.PokerHandNames) {
		return domain.PokerHandNames[rank]
	}
	return "Unknown"
}
