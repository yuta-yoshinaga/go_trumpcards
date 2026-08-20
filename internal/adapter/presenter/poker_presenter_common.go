//go:build !js || !wasm || casino

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// communityCardPresenterGame はコミュニティカード系ポーカー（Holdem, Omaha, ShortDeck, Pineapple）
// のプレゼンター出力に必要な共通メソッドを定義するインターフェース。
type communityCardPresenterGame interface {
	GetPhase() int
	GetPot() int
	GetDealerIdx() int
	GetCurrentTurn() int
	GetGameEndFlag() bool
	GetLastBet() int
	GetMinRaise() int
	GetConfig() domain.HoldemConfig
	GetHandCount() int
	GetPlayerCnt() int
	GetRaiseCount() int
	IsRebuyAvailable() bool
	IsAddonAvailable() bool
	GetRebuyCounts() []int
	GetAddonUsed() []bool
	GetRebuyPhaseType() int
	IsMuckAvailable() bool
	GetCommunityCards() []*domain.Card
	GetSidePots() []domain.SidePot
	GetCpuActions() []domain.HoldemCpuAction
	GetRoundResults() []domain.HoldemResult
	GetEquity() *domain.HoldemEquityResult
	GetPotOdds() float64
	GetHumanProfile() *domain.BettingHumanProfile
}

// communityCardPresenterPlayer はプレイヤー出力変換に必要な共通メソッドを定義するインターフェース。
type communityCardPresenterPlayer interface {
	cardHolder
	GetIsHuman() bool
	GetChips() int
	GetCurrentBet() int
	GetFolded() bool
	GetAllIn() bool
	GetPlayStyleName() string
	GetTotalHands() int
	GetVPIP() int
	GetPFR() int
	GetThreeBet() int
	GetAFDisplay() string
	GetHandRank() int
	GetBestHand() []*domain.Card
}

// lowHandPresenterPlayer はオプショナルなロー手札情報を提供するプレイヤー
// インターフェース (Omaha Hi-Lo 用)。OmahaPlayer のみ実装している。
type lowHandPresenterPlayer interface {
	GetLowBestHand() []*domain.Card
	GetLowQualifies() bool
}

// buildCommunityCardBaseOutput は HoldemWebOutput の共通フィールドを設定する。
// SidePots, CpuActions, RoundResults, Players も共通ロジックで構築する。
func buildCommunityCardBaseOutput(g communityCardPresenterGame) *controller.HoldemWebOutput {
	resObj := new(controller.HoldemWebOutput)
	resObj.Phase = g.GetPhase()
	resObj.Pot = g.GetPot()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.CurrentTurn = g.GetCurrentTurn()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.LastBet = g.GetLastBet()
	resObj.MinRaise = g.GetMinRaise()
	cfg := g.GetConfig()
	resObj.HandCount = g.GetHandCount()
	resObj.SmallBlind = cfg.SmallBlind
	resObj.BigBlind = cfg.BigBlind
	resObj.TournamentMode = cfg.TournamentMode
	resObj.BlindLevelHands = cfg.BlindLevelHands
	resObj.BlindMultiplier = cfg.BlindMultiplier
	resObj.BettingLimit = int(cfg.BettingLimit)
	resObj.TableSize = g.GetPlayerCnt()
	resObj.RaiseCount = g.GetRaiseCount()
	_, resObj.MaxBetAmount = domain.CalculateBettingLimits(cfg.BettingLimit, g.GetPot(), g.GetLastBet())
	resObj.RebuyAvailable = g.IsRebuyAvailable()
	resObj.AddonAvailable = g.IsAddonAvailable()
	resObj.RebuyCounts = g.GetRebuyCounts()
	resObj.AddonUsed = g.GetAddonUsed()
	resObj.RebuyEnabled = cfg.RebuyEnabled
	resObj.AddonEnabled = cfg.AddonEnabled
	resObj.RebuyMaxCount = cfg.RebuyMaxCount
	resObj.RebuyChips = cfg.RebuyChips
	resObj.AddonChips = cfg.AddonChips
	resObj.RebuyPeriodHands = cfg.RebuyPeriodHands
	resObj.AddonAfterHand = cfg.AddonAfterHand
	resObj.RebuyPhaseType = g.GetRebuyPhaseType()
	resObj.MuckAvailable = g.IsMuckAvailable()

	resObj.CommunityCards = cardsToOutput(g.GetCommunityCards())
	resObj.SidePots = buildPokerSidePots(g.GetSidePots())
	resObj.CpuActions = buildPokerCpuActions(g.GetCpuActions())
	resObj.RoundResults = buildPokerRoundResults(g.GetRoundResults())

	if eq := g.GetEquity(); eq != nil {
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
		potOdds := g.GetPotOdds()
		resObj.PotOdds = &potOdds
	}

	if profile := g.GetHumanProfile(); profile != nil {
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

// buildPokerPlayersOutput はプレイヤー情報を共通ロジックで構築する。
// handNameFn はハンドランクから名前を返す関数（ゲームごとに異なるハンド名テーブルに対応）。
func buildPokerPlayersOutput(phase, playerCnt int, getPlayer func(int) communityCardPresenterPlayer, showdownPhase, endPhase int, handNameFn func(int) string) []*controller.HoldemWebOutputPlayer {
	out := make([]*controller.HoldemWebOutputPlayer, 0)
	isShowdown := phase == endPhase || phase == showdownPhase
	for i := 0; i < playerCnt; i++ {
		player := getPlayer(i)
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
			pObj.HandName = handNameFn(player.GetHandRank())
			pObj.BestHand = cardsToOutput(player.GetBestHand())
			if lp, ok := player.(lowHandPresenterPlayer); ok && lp.GetLowQualifies() {
				pObj.LowQualifies = true
				pObj.LowBestHand = cardsToOutput(lp.GetLowBestHand())
			}
		} else {
			pObj.BestHand = make([]*controller.WebOutputCard, 0)
		}

		out = append(out, pObj)
	}
	return out
}

// buildPokerSidePots はサイドポット情報を出力用に変換する。
func buildPokerSidePots(sidePots []domain.SidePot) []*controller.HoldemWebOutputSidePot {
	out := make([]*controller.HoldemWebOutputSidePot, 0)
	for _, sp := range sidePots {
		out = append(out, &controller.HoldemWebOutputSidePot{
			Amount:          sp.Amount,
			EligiblePlayers: sp.EligiblePlayers,
		})
	}
	return out
}

// buildPokerCpuActions はCPU行動記録を出力用に変換する。
func buildPokerCpuActions(actions []domain.HoldemCpuAction) []*controller.HoldemWebOutputCpuAction {
	out := make([]*controller.HoldemWebOutputCpuAction, 0)
	for _, action := range actions {
		out = append(out, &controller.HoldemWebOutputCpuAction{
			PlayerIdx: action.PlayerIdx,
			Action:    action.Action,
			Amount:    action.Amount,
		})
	}
	return out
}

// buildPokerRoundResults はラウンド結果を出力用に変換する。
// Hi-Lo (Omaha 8 or Better) のラウンド結果には Low* / HiWonAmount /
// LowWonAmount フィールドが populated されるため、それらを Web 出力に
// マッピングする (omitempty で Hi-Lo 以外のゲームでは JSON に含まれない)。
func buildPokerRoundResults(results []domain.HoldemResult) []*controller.HoldemWebOutputResult {
	out := make([]*controller.HoldemWebOutputResult, 0)
	for _, r := range results {
		result := &controller.HoldemWebOutputResult{
			PlayerIdx:    r.PlayerIdx,
			HandRank:     r.HandRank,
			HandName:     r.HandName,
			Kickers:      domain.FormatKickers(r.Kickers),
			WonAmount:    r.WonAmount,
			Mucked:       r.Mucked,
			BestHand:     make([]*controller.WebOutputCard, 0),
			HiWonAmount:  r.HiWonAmount,
			LowWonAmount: r.LowWonAmount,
		}
		if r.Mucked {
			result.HandRank = 0
			result.HandName = ""
			result.Kickers = ""
		} else {
			result.BestHand = cardsToOutput(r.BestHand)
			if r.LowQualifies {
				result.LowQualifies = true
				result.LowBestHand = cardsToOutput(r.LowBestHand)
				result.LowKickers = domain.FormatKickers(r.LowKickers)
			}
		}
		out = append(out, result)
	}
	return out
}

// pokerHandKeys は役のロケール非依存キー。フロントの POKER_HAND_KEYS
// (frontend/src/utils/pokerSquaresUtils.ts) と同じ並びで、各ゲームの locale の
// `hand` ブロックがこのキーを引く。
var pokerHandKeys = []string{
	"highCard", "onePair", "twoPair", "threeOfAKind", "straight",
	"flush", "fullHouse", "fourOfAKind", "straightFlush", "royalFlush",
}

// pokerHandKey は役のキーを返す。範囲外は空 (呼び出し側が表示を省ける)。
//
// 表示名 (pokerHandName) はショーダウンの確定表示で使う英語固定の文字列で、
// クライアントで翻訳したい途中経過にはキーのほうを渡す。
func pokerHandKey(rank int) string {
	if rank >= 0 && rank < len(pokerHandKeys) {
		return pokerHandKeys[rank]
	}
	return ""
}

// pokerHandName はハンドランクから名前を返す。
func pokerHandName(rank int) string {
	if rank >= 0 && rank < len(domain.PokerHandNames) {
		return domain.PokerHandNames[rank]
	}
	return "Unknown"
}

// shortDeckHandName はショートデック用のハンドランクから名前を返す。
func shortDeckHandName(rank int) string {
	if rank >= 0 && rank < len(domain.ShortDeckHandNames) {
		return domain.ShortDeckHandNames[rank]
	}
	return "Unknown"
}
