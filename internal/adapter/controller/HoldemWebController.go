package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"

	"github.com/ant0ine/go-json-rest/rest"
)

// HoldemWebInput テキサスホールデムWebインプット
type HoldemWebInput struct {
	BaseWebInput
	Amount          int   `json:"amount,omitempty"`
	SmallBlind      *int  `json:"smallBlind,omitempty"`
	BigBlind        *int  `json:"bigBlind,omitempty"`
	TournamentMode  *bool `json:"tournamentMode,omitempty"`
	BlindLevelHands *int  `json:"blindLevelHands,omitempty"`
	BlindMultiplier *int  `json:"blindMultiplier,omitempty"`
	BettingLimit    *int  `json:"bettingLimit,omitempty"`
	TableSize       *int  `json:"tableSize,omitempty"`
}

// HoldemWebOutputPlayer テキサスホールデムWebアウトプットプレイヤー
type HoldemWebOutputPlayer struct {
	ID            int              `json:"id"`
	IsHuman       bool             `json:"isHuman"`
	Cards         []*WebOutputCard `json:"cards"`
	Chips         int              `json:"chips"`
	CurrentBet    int              `json:"currentBet"`
	Folded        bool             `json:"folded"`
	AllIn         bool             `json:"allIn"`
	HandRank      int              `json:"handRank"`
	HandName      string           `json:"handName"`
	BestHand      []*WebOutputCard `json:"bestHand"`
	PlayStyleName string           `json:"playStyleName"`
	TotalHands    int              `json:"totalHands"`
	VPIP          int              `json:"vpip"`
	PFR           int              `json:"pfr"`
	ThreeBet      int              `json:"threeBet"`
	AF            string           `json:"af"`
}

// HoldemWebOutputCpuAction テキサスホールデムCPU行動記録
type HoldemWebOutputCpuAction struct {
	PlayerIdx int `json:"playerIdx"`
	Action    int `json:"action"`
	Amount    int `json:"amount"`
}

// HoldemWebOutputResult テキサスホールデムショーダウン結果
type HoldemWebOutputResult struct {
	PlayerIdx int              `json:"playerIdx"`
	HandRank  int              `json:"handRank"`
	HandName  string           `json:"handName"`
	Kickers   string           `json:"kickers"`
	BestHand  []*WebOutputCard `json:"bestHand"`
	WonAmount int              `json:"wonAmount"`
}

// HoldemWebOutputSidePot テキサスホールデムサイドポット
type HoldemWebOutputSidePot struct {
	Amount          int   `json:"amount"`
	EligiblePlayers []int `json:"eligiblePlayers"`
}

// HoldemWebOutput テキサスホールデムWebアウトプット
type HoldemWebOutput struct {
	Players         []*HoldemWebOutputPlayer    `json:"players"`
	CommunityCards  []*WebOutputCard            `json:"communityCards"`
	Pot             int                         `json:"pot"`
	SidePots        []*HoldemWebOutputSidePot   `json:"sidePots"`
	DealerIdx       int                         `json:"dealerIdx"`
	CurrentTurn     int                         `json:"currentTurn"`
	Phase           int                         `json:"phase"`
	GameEndFlag     bool                        `json:"gameEndFlag"`
	LastBet         int                         `json:"lastBet"`
	MinRaise        int                         `json:"minRaise"`
	BettingLimit    int                         `json:"bettingLimit"`
	RaiseCount      int                         `json:"raiseCount"`
	MaxBetAmount    int                         `json:"maxBetAmount"`
	RoundResults    []*HoldemWebOutputResult    `json:"roundResults"`
	CpuActions      []*HoldemWebOutputCpuAction `json:"cpuActions"`
	Message         string                      `json:"message"`
	MessageCode     string                      `json:"messageCode,omitempty"`
	MessageParams   map[string]string           `json:"messageParams,omitempty"`
	HandCount       int                         `json:"handCount"`
	SmallBlind      int                         `json:"smallBlind"`
	BigBlind        int                         `json:"bigBlind"`
	TournamentMode  bool                        `json:"tournamentMode"`
	BlindLevelHands int                         `json:"blindLevelHands"`
	BlindMultiplier int                         `json:"blindMultiplier"`
	TableSize       int                         `json:"tableSize"`
}

// HoldemWebController テキサスホールデムWebコントローラークラス
type HoldemWebController = GameWebController[usecase.HoldemInteractorIF, HoldemWebInput, *HoldemWebOutput]

// NewHoldemWebController コンストラクタ
func NewHoldemWebController(factory func() usecase.HoldemInteractorIF) *HoldemWebController {
	return NewGameWebController(factory, newHoldemDefaultOutput, holdemDispatch)
}

func newHoldemDefaultOutput(msg string) *HoldemWebOutput {
	return &HoldemWebOutput{
		Players:        make([]*HoldemWebOutputPlayer, 0),
		CommunityCards: make([]*WebOutputCard, 0),
		SidePots:       make([]*HoldemWebOutputSidePot, 0),
		RoundResults:   make([]*HoldemWebOutputResult, 0),
		CpuActions:     make([]*HoldemWebOutputCpuAction, 0),
		Message:        msg,
	}
}

func holdemDispatch(bc *baseController, w rest.ResponseWriter, hgi usecase.HoldemInteractorIF, param HoldemWebInput, newDefault func(string) *HoldemWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		cfg := domain.DefaultHoldemConfig()
		sb, bb := cfg.SmallBlind, cfg.BigBlind
		sbProvided := param.SmallBlind != nil && *param.SmallBlind >= 1
		bbProvided := param.BigBlind != nil && *param.BigBlind >= 1
		if sbProvided {
			sb = *param.SmallBlind
		}
		if bbProvided {
			bb = *param.BigBlind
		}
		// 片方のみ指定された場合、もう片方を自動調整
		if sbProvided && !bbProvided && sb >= cfg.BigBlind {
			bb = sb * 2
		} else if bbProvided && !sbProvided && bb > 1 {
			sb = bb / 2
			if sb < 1 {
				sb = 1
			}
		}
		if sb >= bb {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: smallBlind must be less than bigBlind."))
			return true
		}
		cfg.SmallBlind = sb
		cfg.BigBlind = bb
		// トーナメントモード設定
		if param.TournamentMode != nil {
			cfg.TournamentMode = *param.TournamentMode
		}
		if param.BlindLevelHands != nil && *param.BlindLevelHands >= 1 {
			cfg.BlindLevelHands = *param.BlindLevelHands
		}
		if param.BlindMultiplier != nil && *param.BlindMultiplier >= 101 {
			cfg.BlindMultiplier = *param.BlindMultiplier
		}
		if param.BettingLimit != nil {
			bl := *param.BettingLimit
			if bl < 0 {
				bl = 0
			} else if bl > 2 {
				bl = 2
			}
			cfg.BettingLimit = domain.BettingLimitType(bl)
		}
		if param.TableSize != nil {
			ts := *param.TableSize
			if !domain.IsValidHoldemTableSize(ts) {
				bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: tableSize must be 4, 6, or 9."))
				return true
			}
			cfg.TableSize = ts
		}
		bc.writePresenterResponse(w, hgi.ResetWithConfig(cfg))
	case "f", "fold":
		bc.writePresenterResponse(w, hgi.Action(domain.HoldemActionFold, 0))
	case "ck", "check":
		bc.writePresenterResponse(w, hgi.Action(domain.HoldemActionCheck, 0))
	case "c", "call":
		bc.writePresenterResponse(w, hgi.Action(domain.HoldemActionCall, 0))
	case "b", "bet":
		bc.writePresenterResponse(w, hgi.Action(domain.HoldemActionBet, param.Amount))
	case "ra", "raise":
		bc.writePresenterResponse(w, hgi.Action(domain.HoldemActionRaise, param.Amount))
	case "a", "allin":
		bc.writePresenterResponse(w, hgi.Action(domain.HoldemActionAllIn, 0))
	default:
		return false
	}
	return true
}
