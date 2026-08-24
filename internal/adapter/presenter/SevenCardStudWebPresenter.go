//go:build !js || !wasm || casino

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// SevenCardStudWebPresenter セブンカードスタッドWebプレゼンタークラス
type SevenCardStudWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *SevenCardStudWebPresenter) Output(s interfaces.SevenCardStudGame, lastErr error) string {
	resObj := p.buildOutput(s, lastErr)
	return marshalOrError(resObj)
}

// buildOutput ゲーム状態をSevenCardStudWebOutputに変換
func (p *SevenCardStudWebPresenter) buildOutput(s interfaces.SevenCardStudGame, lastErr error) *controller.SevenCardStudWebOutput {
	resObj := new(controller.SevenCardStudWebOutput)
	resObj.Phase = s.GetPhase()
	resObj.IsHiLo = s.GetIsHiLo()
	resObj.IsChicago = s.GetIsChicago()
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
		resObj.MetaAI = &controller.SevenCardStudWebOutputMetaAI{
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
func (p *SevenCardStudWebPresenter) buildSidePotsOutput(s interfaces.SevenCardStudGame) []*controller.SevenCardStudWebOutputSidePot {
	out := make([]*controller.SevenCardStudWebOutputSidePot, 0)
	for _, sp := range s.GetSidePots() {
		out = append(out, &controller.SevenCardStudWebOutputSidePot{
			Amount:          sp.Amount,
			EligiblePlayers: sp.EligiblePlayers,
		})
	}
	return out
}

// buildPlayersOutput プレイヤー情報を構築
func (p *SevenCardStudWebPresenter) buildPlayersOutput(s interfaces.SevenCardStudGame) []*controller.SevenCardStudWebOutputPlayer {
	out := make([]*controller.SevenCardStudWebOutputPlayer, 0)
	isShowdown := s.GetPhase() == domain.SevenCardStudPhaseEnd || s.GetPhase() == domain.SevenCardStudPhaseShowdown
	for i := 0; i < s.GetPlayerCnt(); i++ {
		player := s.GetPlayer(i)
		pObj := &controller.SevenCardStudWebOutputPlayer{
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
func (p *SevenCardStudWebPresenter) buildCpuActionsOutput(s interfaces.SevenCardStudGame) []*controller.SevenCardStudWebOutputCpuAction {
	out := make([]*controller.SevenCardStudWebOutputCpuAction, 0)
	for _, action := range s.GetCpuActions() {
		out = append(out, &controller.SevenCardStudWebOutputCpuAction{
			PlayerIdx: action.PlayerIdx,
			Action:    action.Action,
			Amount:    action.Amount,
		})
	}
	return out
}

// buildRoundResultsOutput ラウンド結果を構築
func (p *SevenCardStudWebPresenter) buildRoundResultsOutput(s interfaces.SevenCardStudGame) []*controller.SevenCardStudWebOutputResult {
	out := make([]*controller.SevenCardStudWebOutputResult, 0)
	for _, r := range s.GetRoundResults() {
		result := &controller.SevenCardStudWebOutputResult{
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
			// ローの内訳も出す。これが無いと、なぜポットが半分になったのかが
			// 画面から読み取れない。
			result.LowQualifies = r.LowQualifies
			result.WonLow = r.WonLow
			if len(r.LowBestHand) > 0 {
				result.LowBestHand = cardsToOutput(r.LowBestHand)
			}
			// **スペード側の内訳も同じ理由で出す。** どの 1 枚で半分を取ったのかを
			// 出さないと、ポットが割れた理由が画面から読み取れない。
			result.WonSpade = r.WonSpade
			if r.SpadeCard != nil {
				result.SpadeCard = cardToOutput(r.SpadeCard)
			}
		}
		out = append(out, result)
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *SevenCardStudWebPresenter) buildMessage(s interfaces.SevenCardStudGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if s.IsMuckAvailable() {
		return "", "sevencardstud.muck.prompt", nil
	}
	if s.GetGameEndFlag() {
		msg, code := p.buildResultMessage(s)
		return msg, code, nil
	}
	return "", "", nil
}

// buildResultMessage builds the end-of-round message and its i18n code
func (p *SevenCardStudWebPresenter) buildResultMessage(s interfaces.SevenCardStudGame) (string, string) {
	results := s.GetRoundResults()
	if len(results) == 0 {
		return "", "sevencardstud.result.gameOver"
	}

	for _, r := range results {
		if s.GetPlayer(r.PlayerIdx).GetIsHuman() {
			if r.WonAmount > 0 {
				return "", "sevencardstud.result.win"
			}
		}
	}

	// Human not in results (folded)
	for i := 0; i < s.GetPlayerCnt(); i++ {
		if s.GetPlayer(i).GetIsHuman() && s.GetPlayer(i).GetFolded() {
			return "", "sevencardstud.result.folded"
		}
	}

	// Human mucked
	for _, r := range results {
		if s.GetPlayer(r.PlayerIdx).GetIsHuman() && r.Mucked {
			return "", "sevencardstud.result.mucked"
		}
	}

	return "", "sevencardstud.result.lose"
}

// ActionLogOutput 棋譜をJSON出力
func (p *SevenCardStudWebPresenter) ActionLogOutput(s interfaces.SevenCardStudGame) string {
	return actionLogOutputJSON(s)
}

// HintOutput はヒントを返す。Web ではクライアント側でヒントを算出するため、
// 状態出力にフォールバックする (CUI プレゼンターのみが専用ヒントを返す)。
func (p *SevenCardStudWebPresenter) HintOutput(s interfaces.SevenCardStudGame) string {
	return p.Output(s, nil)
}

// getHandName ハンドランクから名前を返す
func (p *SevenCardStudWebPresenter) getHandName(rank int) string {
	if rank >= 0 && rank < len(domain.PokerHandNames) {
		return domain.PokerHandNames[rank]
	}
	return "Unknown"
}
