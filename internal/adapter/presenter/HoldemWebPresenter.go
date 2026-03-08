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

	// コミュニティカード
	resObj.CommunityCards = make([]*controller.WebOutputCard, 0)
	for _, card := range h.GetCommunityCards() {
		resObj.CommunityCards = append(resObj.CommunityCards, cardToOutput(card))
	}

	// サイドポット
	resObj.SidePots = make([]*controller.HoldemWebOutputSidePot, 0)
	for _, sp := range h.GetSidePots() {
		resObj.SidePots = append(resObj.SidePots, &controller.HoldemWebOutputSidePot{
			Amount:          sp.Amount,
			EligiblePlayers: sp.EligiblePlayers,
		})
	}

	// プレイヤー情報
	resObj.Players = make([]*controller.HoldemWebOutputPlayer, 0)
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

		resObj.Players = append(resObj.Players, pObj)
	}

	// CPU行動記録
	resObj.CpuActions = make([]*controller.HoldemWebOutputCpuAction, 0)
	for _, action := range h.GetCpuActions() {
		resObj.CpuActions = append(resObj.CpuActions, &controller.HoldemWebOutputCpuAction{
			PlayerIdx: action.PlayerIdx,
			Action:    action.Action,
			Amount:    action.Amount,
		})
	}

	// ラウンド結果
	resObj.RoundResults = make([]*controller.HoldemWebOutputResult, 0)
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
		resObj.RoundResults = append(resObj.RoundResults, result)
	}

	// メッセージ
	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if h.IsMuckAvailable() {
		resObj.Message = "Muck or show your hand."
		resObj.MessageCode = "holdem.muck.prompt"
	} else if h.GetGameEndFlag() {
		msg, code := hwp.buildResultMessage(h)
		resObj.Message = msg
		resObj.MessageCode = code
	}

	return marshalOrError(resObj)
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

// getHandName ハンドランクから名前を返す
func (hwp *HoldemWebPresenter) getHandName(rank int) string {
	if rank >= 0 && rank < len(domain.PokerHandNames) {
		return domain.PokerHandNames[rank]
	}
	return "Unknown"
}
