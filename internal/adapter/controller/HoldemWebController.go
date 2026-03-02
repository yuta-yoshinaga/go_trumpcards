package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"

	"github.com/ant0ine/go-json-rest/rest"
)

// HoldemWebInput テキサスホールデムWebインプット
type HoldemWebInput struct {
	Command         string `json:"command"`
	Amount          int    `json:"amount,omitempty"`
	SessionId       string `json:"sessionId"`
	SmallBlind      *int   `json:"smallBlind,omitempty"`
	BigBlind        *int   `json:"bigBlind,omitempty"`
	TournamentMode  *bool  `json:"tournamentMode,omitempty"`
	BlindLevelHands *int   `json:"blindLevelHands,omitempty"`
	BlindMultiplier *int   `json:"blindMultiplier,omitempty"`
}

// GetCommand returns the command string.
func (i HoldemWebInput) GetCommand() string { return i.Command }

// GetSessionID returns the session ID string.
func (i HoldemWebInput) GetSessionID() string { return i.SessionId }

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
	RoundResults    []*HoldemWebOutputResult    `json:"roundResults"`
	CpuActions      []*HoldemWebOutputCpuAction `json:"cpuActions"`
	Message         string                      `json:"message"`
	HandCount       int                         `json:"handCount"`
	SmallBlind      int                         `json:"smallBlind"`
	BigBlind        int                         `json:"bigBlind"`
	TournamentMode  bool                        `json:"tournamentMode"`
	BlindLevelHands int                         `json:"blindLevelHands"`
	BlindMultiplier int                         `json:"blindMultiplier"`
}

// HoldemWebController テキサスホールデムWebコントローラークラス
type HoldemWebController struct {
	baseController
	factory func() usecase.HoldemInteractorIF
	store   *SessionStore[usecase.HoldemInteractorIF]
}

// NewHoldemWebController コンストラクタ
func NewHoldemWebController(factory func() usecase.HoldemInteractorIF) *HoldemWebController {
	return &HoldemWebController{
		factory: factory,
		store:   NewSessionStore[usecase.HoldemInteractorIF](),
	}
}

// Exec ゲーム実行
func (hwc *HoldemWebController) Exec(w rest.ResponseWriter, r *rest.Request) {
	execWithSession(&hwc.baseController, w, r, hwc.store, hwc.factory,
		func(msg string) any { return hwc.newDefaultOutput(msg) },
		nil,
		func(w rest.ResponseWriter, hgi usecase.HoldemInteractorIF, param HoldemWebInput) bool {
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
					hwc.writeJsonResponse(w, http.StatusBadRequest, hwc.newDefaultOutput("param error: smallBlind must be less than bigBlind."))
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
				hwc.writePresenterResponse(w, hgi.ResetWithConfig(cfg))
			case "f", "fold":
				hwc.writePresenterResponse(w, hgi.Action(domain.HoldemActionFold, 0))
			case "ck", "check":
				hwc.writePresenterResponse(w, hgi.Action(domain.HoldemActionCheck, 0))
			case "c", "call":
				hwc.writePresenterResponse(w, hgi.Action(domain.HoldemActionCall, 0))
			case "b", "bet":
				hwc.writePresenterResponse(w, hgi.Action(domain.HoldemActionBet, param.Amount))
			case "ra", "raise":
				hwc.writePresenterResponse(w, hgi.Action(domain.HoldemActionRaise, param.Amount))
			case "a", "allin":
				hwc.writePresenterResponse(w, hgi.Action(domain.HoldemActionAllIn, 0))
			default:
				return false
			}
			return true
		})
}

// Stop stops the background cleanup goroutine of the session store.
func (hwc *HoldemWebController) Stop() {
	hwc.store.Stop()
}

// newDefaultOutput エラー・定型応答用のデフォルト出力を返す
func (hwc *HoldemWebController) newDefaultOutput(msg string) *HoldemWebOutput {
	return &HoldemWebOutput{
		Players:        make([]*HoldemWebOutputPlayer, 0),
		CommunityCards: make([]*WebOutputCard, 0),
		SidePots:       make([]*HoldemWebOutputSidePot, 0),
		RoundResults:   make([]*HoldemWebOutputResult, 0),
		CpuActions:     make([]*HoldemWebOutputCpuAction, 0),
		Message:        msg,
	}
}
