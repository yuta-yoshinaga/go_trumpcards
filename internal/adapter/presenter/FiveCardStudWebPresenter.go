//go:build !js || !wasm || casino

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// FiveCardStudWebPresenter ファイブカードスタッドWebプレゼンタークラス
type FiveCardStudWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *FiveCardStudWebPresenter) Output(s interfaces.FiveCardStudGame, lastErr error) string {
	resObj := p.buildOutput(s, lastErr)
	return marshalOrError(resObj)
}

// buildOutput ゲーム状態をFiveCardStudWebOutputに変換
func (p *FiveCardStudWebPresenter) buildOutput(s interfaces.FiveCardStudGame, lastErr error) *controller.FiveCardStudWebOutput {
	resObj := new(controller.FiveCardStudWebOutput)
	resObj.Phase = s.GetPhase()
	resObj.Pot = s.GetPot()
	resObj.DealerIdx = s.GetDealerIdx()
	resObj.CurrentTurn = s.GetCurrentTurn()
	resObj.GameEndFlag = s.GetGameEndFlag()
	resObj.LastBet = s.GetLastBet()
	resObj.MinRaise = s.GetMinRaise()
	cfg := s.GetConfig()
	resObj.HandCount = s.GetHandCount()
	resObj.Ante = cfg.Ante
	resObj.BringIn = cfg.BringIn
	resObj.SmallBet = cfg.SmallBet
	resObj.BigBet = cfg.BigBet
	resObj.TournamentMode = cfg.TournamentMode
	resObj.AnteLevelHands = cfg.AnteLevelHands
	resObj.AnteMultiplier = cfg.AnteMultiplier
	resObj.BettingLimit = int(cfg.BettingLimit)
	resObj.TableSize = s.GetPlayerCnt()
	resObj.RaiseCount = s.GetRaiseCount()
	_, resObj.MaxBetAmount = domain.CalculateBettingLimits(cfg.BettingLimit, s.GetPot(), s.GetLastBet())
	resObj.BringInPlayerIdx = s.GetBringInPlayerIdx()
	resObj.RebuyAvailable = s.IsRebuyAvailable()
	resObj.AddonAvailable = s.IsAddonAvailable()
	resObj.RebuyCounts = s.GetRebuyCounts()
	resObj.AddonUsed = s.GetAddonUsed()
	resObj.RebuyEnabled = cfg.RebuyEnabled
	resObj.AddonEnabled = cfg.AddonEnabled
	resObj.RebuyMaxCount = cfg.RebuyMaxCount
	resObj.RebuyChips = cfg.RebuyChips
	resObj.AddonChips = cfg.AddonChips
	resObj.RebuyPeriodHands = cfg.RebuyPeriodHands
	resObj.AddonAfterHand = cfg.AddonAfterHand
	resObj.RebuyPhaseType = s.GetRebuyPhaseType()
	resObj.MuckAvailable = s.IsMuckAvailable()

	resObj.CommunityCard = cardToOutput(s.GetCommunityCard())
	resObj.SidePots = p.buildSidePotsOutput(s)
	resObj.Players = p.buildPlayersOutput(s)
	resObj.CpuActions = p.buildCpuActionsOutput(s)
	resObj.RoundResults = p.buildRoundResultsOutput(s)

	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(s, lastErr)

	// メタAI情報
	if profile := s.GetHumanProfile(); profile != nil {
		resObj.MetaAI = &controller.FiveCardStudWebOutputMetaAI{
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
func (p *FiveCardStudWebPresenter) buildSidePotsOutput(s interfaces.FiveCardStudGame) []*controller.FiveCardStudWebOutputSidePot {
	out := make([]*controller.FiveCardStudWebOutputSidePot, 0)
	for _, sp := range s.GetSidePots() {
		out = append(out, &controller.FiveCardStudWebOutputSidePot{
			Amount:          sp.Amount,
			EligiblePlayers: sp.EligiblePlayers,
		})
	}
	return out
}

// buildPlayersOutput プレイヤー情報を構築
func (p *FiveCardStudWebPresenter) buildPlayersOutput(s interfaces.FiveCardStudGame) []*controller.FiveCardStudWebOutputPlayer {
	out := make([]*controller.FiveCardStudWebOutputPlayer, 0)
	isShowdown := s.GetPhase() == domain.FiveCardStudPhaseEnd || s.GetPhase() == domain.FiveCardStudPhaseShowdown
	for i := 0; i < s.GetPlayerCnt(); i++ {
		player := s.GetPlayer(i)
		pObj := &controller.FiveCardStudWebOutputPlayer{
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

		// ドアカードは常に表示
		pObj.DoorCards = cardsToOutputOrEmpty(player.GetDoorCards())

		// ホールカード: 人間は常に表示、CPUはショーダウン時のみ表示
		showHole := player.GetIsHuman() || (isShowdown && !player.GetFolded())
		if showHole {
			pObj.HoleCards = cardsToOutputOrEmpty(player.GetHoleCards())
		} else {
			pObj.HoleCards = make([]*controller.WebOutputCard, 0)
		}

		// ショーダウン時のハンド情報
		if isShowdown && !player.GetFolded() {
			pObj.HandRank = player.GetHandRank()
			pObj.HandName = p.getHandName(player.GetHandRank())
			pObj.BestHand = cardsToOutput(player.GetBestHand())
		} else {
			pObj.BestHand = make([]*controller.WebOutputCard, 0)
		}

		out = append(out, pObj)
	}
	return out
}

// buildCpuActionsOutput CPU行動記録を構築
func (p *FiveCardStudWebPresenter) buildCpuActionsOutput(s interfaces.FiveCardStudGame) []*controller.FiveCardStudWebOutputCpuAction {
	out := make([]*controller.FiveCardStudWebOutputCpuAction, 0)
	for _, action := range s.GetCpuActions() {
		out = append(out, &controller.FiveCardStudWebOutputCpuAction{
			PlayerIdx: action.PlayerIdx,
			Action:    action.Action,
			Amount:    action.Amount,
		})
	}
	return out
}

// buildRoundResultsOutput ラウンド結果を構築
func (p *FiveCardStudWebPresenter) buildRoundResultsOutput(s interfaces.FiveCardStudGame) []*controller.FiveCardStudWebOutputResult {
	out := make([]*controller.FiveCardStudWebOutputResult, 0)
	for _, r := range s.GetRoundResults() {
		result := &controller.FiveCardStudWebOutputResult{
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
func (p *FiveCardStudWebPresenter) buildMessage(s interfaces.FiveCardStudGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if s.IsMuckAvailable() {
		return "", "fivecardstud.muck.prompt", nil
	}
	if s.GetGameEndFlag() {
		msg, code := p.buildResultMessage(s)
		return msg, code, nil
	}
	return "", "", nil
}

// buildResultMessage builds the end-of-round message and its i18n code
func (p *FiveCardStudWebPresenter) buildResultMessage(s interfaces.FiveCardStudGame) (string, string) {
	results := s.GetRoundResults()
	if len(results) == 0 {
		return "", "fivecardstud.result.gameOver"
	}

	for _, r := range results {
		if s.GetPlayer(r.PlayerIdx).GetIsHuman() {
			if r.WonAmount > 0 {
				return "", "fivecardstud.result.win"
			}
		}
	}

	// Human not in results (folded)
	for i := 0; i < s.GetPlayerCnt(); i++ {
		if s.GetPlayer(i).GetIsHuman() && s.GetPlayer(i).GetFolded() {
			return "", "fivecardstud.result.folded"
		}
	}

	// Human mucked
	for _, r := range results {
		if s.GetPlayer(r.PlayerIdx).GetIsHuman() && r.Mucked {
			return "", "fivecardstud.result.mucked"
		}
	}

	return "", "fivecardstud.result.lose"
}

// ActionLogOutput 棋譜をJSON出力
func (p *FiveCardStudWebPresenter) ActionLogOutput(s interfaces.FiveCardStudGame) string {
	return actionLogOutputJSON(s)
}

// getHandName ハンドランクから名前を返す
func (p *FiveCardStudWebPresenter) getHandName(rank int) string {
	if rank >= 0 && rank < len(domain.PokerHandNames) {
		return domain.PokerHandNames[rank]
	}
	return "Unknown"
}
