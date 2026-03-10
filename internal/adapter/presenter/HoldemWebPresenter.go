package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// HoldemWebPresenter テキサスホールデムWebプレゼンタークラス
type HoldemWebPresenter struct{}

// NewHoldemWebPresenter コンストラクタ
func NewHoldemWebPresenter() *HoldemWebPresenter {
	return &HoldemWebPresenter{}
}

// Output ゲーム状態をJSON出力
func (hwp *HoldemWebPresenter) Output(h interfaces.HoldemGame, lastErr error) string {
	resObj := hwp.buildOutput(h, lastErr)
	return marshalOrError(resObj)
}

// buildOutput ゲーム状態をHoldemWebOutputに変換
func (hwp *HoldemWebPresenter) buildOutput(h interfaces.HoldemGame, lastErr error) *controller.HoldemWebOutput {
	resObj := new(controller.HoldemWebOutput)
	resObj.Phase = h.GetPhase()
	resObj.Pot = h.GetPot()
	resObj.DealerIdx = h.GetDealerIdx()
	resObj.CurrentTurn = h.GetCurrentTurn()
	resObj.GameEndFlag = h.GetGameEndFlag()
	resObj.LastBet = h.GetLastBet()
	resObj.MinRaise = h.GetMinRaise()
	cfg := h.GetConfig()
	resObj.HandCount = h.GetHandCount()
	resObj.SmallBlind = cfg.SmallBlind
	resObj.BigBlind = cfg.BigBlind
	resObj.TournamentMode = cfg.TournamentMode
	resObj.BlindLevelHands = cfg.BlindLevelHands
	resObj.BlindMultiplier = cfg.BlindMultiplier
	resObj.BettingLimit = int(cfg.BettingLimit)
	resObj.TableSize = h.GetPlayerCnt()
	resObj.RaiseCount = h.GetRaiseCount()
	resObj.MaxBetAmount = calcMaxBetAmount(cfg.BettingLimit, h.GetPot(), h.GetLastBet())
	resObj.RebuyAvailable = h.IsRebuyAvailable()
	resObj.AddonAvailable = h.IsAddonAvailable()
	resObj.RebuyCounts = h.GetRebuyCounts()
	resObj.AddonUsed = h.GetAddonUsed()
	resObj.RebuyEnabled = cfg.RebuyEnabled
	resObj.AddonEnabled = cfg.AddonEnabled
	resObj.RebuyMaxCount = cfg.RebuyMaxCount
	resObj.RebuyChips = cfg.RebuyChips
	resObj.AddonChips = cfg.AddonChips
	resObj.RebuyPeriodHands = cfg.RebuyPeriodHands
	resObj.AddonAfterHand = cfg.AddonAfterHand
	resObj.RebuyPhaseType = h.GetRebuyPhaseType()
	resObj.MuckAvailable = h.IsMuckAvailable()

	resObj.CommunityCards = hwp.buildCommunityCardsOutput(h)
	resObj.SidePots = hwp.buildSidePotsOutput(h)
	resObj.Players = hwp.buildPlayersOutput(h)
	resObj.CpuActions = hwp.buildCpuActionsOutput(h)
	resObj.RoundResults = hwp.buildRoundResultsOutput(h)

	resObj.Message, resObj.MessageCode = hwp.buildMessage(h, lastErr)

	return resObj
}

// buildCommunityCardsOutput コミュニティカード情報を構築
func (hwp *HoldemWebPresenter) buildCommunityCardsOutput(h interfaces.HoldemGame) []*controller.WebOutputCard {
	out := make([]*controller.WebOutputCard, 0)
	for _, card := range h.GetCommunityCards() {
		out = append(out, cardToOutput(card))
	}
	return out
}

// buildSidePotsOutput サイドポット情報を構築
func (hwp *HoldemWebPresenter) buildSidePotsOutput(h interfaces.HoldemGame) []*controller.HoldemWebOutputSidePot {
	out := make([]*controller.HoldemWebOutputSidePot, 0)
	for _, sp := range h.GetSidePots() {
		out = append(out, &controller.HoldemWebOutputSidePot{
			Amount:          sp.Amount,
			EligiblePlayers: sp.EligiblePlayers,
		})
	}
	return out
}

// buildPlayersOutput プレイヤー情報を構築
func (hwp *HoldemWebPresenter) buildPlayersOutput(h interfaces.HoldemGame) []*controller.HoldemWebOutputPlayer {
	out := make([]*controller.HoldemWebOutputPlayer, 0)
	isShowdown := h.GetPhase() == domain.HoldemPhaseEnd || h.GetPhase() == domain.HoldemPhaseShowdown
	for i := 0; i < h.GetPlayerCnt(); i++ {
		player := h.GetPlayer(i)
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
			Cards:         make([]*controller.WebOutputCard, 0),
			BestHand:      make([]*controller.WebOutputCard, 0),
		}

		// 人間のカードは常に表示、CPUのカードはショーダウン時のみ表示
		if player.GetIsHuman() || (isShowdown && !player.GetFolded()) {
			for j := 0; j < player.GetCardsSize(); j++ {
				pObj.Cards = append(pObj.Cards, cardToOutput(player.GetCard(j)))
			}
		}

		// ショーダウン時のハンド情報
		if isShowdown && !player.GetFolded() {
			pObj.HandRank = player.GetHandRank()
			pObj.HandName = hwp.getHandName(player.GetHandRank())
			for _, card := range player.GetBestHand() {
				pObj.BestHand = append(pObj.BestHand, cardToOutput(card))
			}
		}

		out = append(out, pObj)
	}
	return out
}

// buildCpuActionsOutput CPU行動記録を構築
func (hwp *HoldemWebPresenter) buildCpuActionsOutput(h interfaces.HoldemGame) []*controller.HoldemWebOutputCpuAction {
	out := make([]*controller.HoldemWebOutputCpuAction, 0)
	for _, action := range h.GetCpuActions() {
		out = append(out, &controller.HoldemWebOutputCpuAction{
			PlayerIdx: action.PlayerIdx,
			Action:    action.Action,
			Amount:    action.Amount,
		})
	}
	return out
}

// buildRoundResultsOutput ラウンド結果を構築
func (hwp *HoldemWebPresenter) buildRoundResultsOutput(h interfaces.HoldemGame) []*controller.HoldemWebOutputResult {
	out := make([]*controller.HoldemWebOutputResult, 0)
	for _, r := range h.GetRoundResults() {
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
			for _, card := range r.BestHand {
				result.BestHand = append(result.BestHand, cardToOutput(card))
			}
		}
		out = append(out, result)
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (hwp *HoldemWebPresenter) buildMessage(h interfaces.HoldemGame, lastErr error) (string, string) {
	if lastErr != nil {
		return lastErr.Error(), ""
	}
	if h.IsMuckAvailable() {
		return "Muck or show your hand.", "holdem.muck.prompt"
	}
	if h.GetGameEndFlag() {
		return hwp.buildResultMessage(h)
	}
	return "", ""
}

// buildResultMessage builds the end-of-round message and its i18n code
func (hwp *HoldemWebPresenter) buildResultMessage(h interfaces.HoldemGame) (string, string) {
	results := h.GetRoundResults()
	if len(results) == 0 {
		return "Game over.", "holdem.result.gameOver"
	}

	for _, r := range results {
		if h.GetPlayer(r.PlayerIdx).GetIsHuman() {
			if r.WonAmount > 0 {
				return "You are the winner.", "holdem.result.win"
			}
		}
	}

	// Human not in results (folded)
	for i := 0; i < h.GetPlayerCnt(); i++ {
		if h.GetPlayer(i).GetIsHuman() && h.GetPlayer(i).GetFolded() {
			return "You folded.", "holdem.result.folded"
		}
	}

	// Human mucked
	for _, r := range results {
		if h.GetPlayer(r.PlayerIdx).GetIsHuman() && r.Mucked {
			return "You mucked.", "holdem.result.mucked"
		}
	}

	return "You lose.", "holdem.result.lose"
}

// ActionLogOutput 棋譜をJSON出力
func (hwp *HoldemWebPresenter) ActionLogOutput(h interfaces.HoldemGame) string {
	return actionLogOutputJSON(h)
}

// getHandName ハンドランクから名前を返す
func (hwp *HoldemWebPresenter) getHandName(rank int) string {
	if rank >= 0 && rank < len(domain.PokerHandNames) {
		return domain.PokerHandNames[rank]
	}
	return "Unknown"
}
