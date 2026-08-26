//go:build !js || !wasm || casino

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// FollowTheQueenWebPresenter フォロー・ザ・クイーンWebプレゼンタークラス
type FollowTheQueenWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *FollowTheQueenWebPresenter) Output(s interfaces.FollowTheQueenGame, lastErr error) string {
	resObj := p.buildOutput(s, lastErr)
	return marshalOrError(resObj)
}

// buildOutput ゲーム状態をFollowTheQueenWebOutputに変換
func (p *FollowTheQueenWebPresenter) buildOutput(s interfaces.FollowTheQueenGame, lastErr error) *controller.FollowTheQueenWebOutput {
	resObj := new(controller.FollowTheQueenWebOutput)
	resObj.Phase = s.GetPhase()
	resObj.WildRank = s.GetWildRank()
	resObj.HumanHandRank = followTheQueenHumanHandRank(s)
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
		resObj.MetaAI = &controller.FollowTheQueenWebOutputMetaAI{
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
func (p *FollowTheQueenWebPresenter) buildSidePotsOutput(s interfaces.FollowTheQueenGame) []*controller.FollowTheQueenWebOutputSidePot {
	out := make([]*controller.FollowTheQueenWebOutputSidePot, 0)
	for _, sp := range s.GetSidePots() {
		out = append(out, &controller.FollowTheQueenWebOutputSidePot{
			Amount:          sp.Amount,
			EligiblePlayers: sp.EligiblePlayers,
		})
	}
	return out
}

// buildPlayersOutput プレイヤー情報を構築
func (p *FollowTheQueenWebPresenter) buildPlayersOutput(s interfaces.FollowTheQueenGame) []*controller.FollowTheQueenWebOutputPlayer {
	out := make([]*controller.FollowTheQueenWebOutputPlayer, 0)
	isShowdown := s.GetPhase() == domain.FollowTheQueenPhaseEnd || s.GetPhase() == domain.FollowTheQueenPhaseShowdown
	for i := 0; i < s.GetPlayerCnt(); i++ {
		player := s.GetPlayer(i)
		pObj := &controller.FollowTheQueenWebOutputPlayer{
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
func (p *FollowTheQueenWebPresenter) buildCpuActionsOutput(s interfaces.FollowTheQueenGame) []*controller.FollowTheQueenWebOutputCpuAction {
	out := make([]*controller.FollowTheQueenWebOutputCpuAction, 0)
	for _, action := range s.GetCpuActions() {
		out = append(out, &controller.FollowTheQueenWebOutputCpuAction{
			PlayerIdx: action.PlayerIdx,
			Action:    action.Action,
			Amount:    action.Amount,
		})
	}
	return out
}

// buildRoundResultsOutput ラウンド結果を構築
func (p *FollowTheQueenWebPresenter) buildRoundResultsOutput(s interfaces.FollowTheQueenGame) []*controller.FollowTheQueenWebOutputResult {
	out := make([]*controller.FollowTheQueenWebOutputResult, 0)
	for _, r := range s.GetRoundResults() {
		result := &controller.FollowTheQueenWebOutputResult{
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
func (p *FollowTheQueenWebPresenter) buildMessage(s interfaces.FollowTheQueenGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if s.IsMuckAvailable() {
		return "", "followthequeen.muck.prompt", nil
	}
	if s.GetGameEndFlag() {
		msg, code := p.buildResultMessage(s)
		return msg, code, nil
	}
	return "", "", nil
}

// buildResultMessage builds the end-of-round message and its i18n code
func (p *FollowTheQueenWebPresenter) buildResultMessage(s interfaces.FollowTheQueenGame) (string, string) {
	results := s.GetRoundResults()
	if len(results) == 0 {
		return "", "followthequeen.result.gameOver"
	}

	for _, r := range results {
		if s.GetPlayer(r.PlayerIdx).GetIsHuman() {
			if r.WonAmount > 0 {
				return "", "followthequeen.result.win"
			}
		}
	}

	// Human not in results (folded)
	for i := 0; i < s.GetPlayerCnt(); i++ {
		if s.GetPlayer(i).GetIsHuman() && s.GetPlayer(i).GetFolded() {
			return "", "followthequeen.result.folded"
		}
	}

	// Human mucked
	for _, r := range results {
		if s.GetPlayer(r.PlayerIdx).GetIsHuman() && r.Mucked {
			return "", "followthequeen.result.mucked"
		}
	}

	return "", "followthequeen.result.lose"
}

// ActionLogOutput 棋譜をJSON出力
func (p *FollowTheQueenWebPresenter) ActionLogOutput(s interfaces.FollowTheQueenGame) string {
	return actionLogOutputJSON(s)
}

// HintOutput はヒントを返す。Web ではクライアント側でヒントを算出するため、
// 状態出力にフォールバックする (CUI プレゼンターのみが専用ヒントを返す)。
func (p *FollowTheQueenWebPresenter) HintOutput(s interfaces.FollowTheQueenGame) string {
	return p.Output(s, nil)
}

// getHandName ハンドランクから名前を返す
func (p *FollowTheQueenWebPresenter) getHandName(rank int) string {
	if rank >= 0 && rank < len(domain.PokerHandNames) {
		return domain.PokerHandNames[rank]
	}
	return "Unknown"
}

// followTheQueenHumanHandRank は人間のいまの最善役位を返す (無ければ -1)。
//
// `PeekBestHand` は状態を変えないので描画から呼んで安全。ワイルドを見る評価は
// ここ 1 本だけで、ページ側に複製しない。
func followTheQueenHumanHandRank(s interfaces.FollowTheQueenGame) int {
	for i := 0; i < s.GetPlayerCnt(); i++ {
		player := s.GetPlayer(i)
		if player == nil || !player.GetIsHuman() || player.GetFolded() {
			continue
		}
		rank, best := player.PeekBestHand()
		if len(best) == 0 {
			return -1
		}
		return rank
	}
	return -1
}
