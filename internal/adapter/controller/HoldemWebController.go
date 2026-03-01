package controller

import (
	"log"
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"

	"github.com/ant0ine/go-json-rest/rest"
)

// HoldemWebInput テキサスホールデムWebインプット
type HoldemWebInput struct {
	Command    string `json:"command"`
	Amount     int    `json:"amount,omitempty"`
	SessionId  string `json:"sessionId"`
	SmallBlind *int   `json:"smallBlind,omitempty"`
	BigBlind   *int   `json:"bigBlind,omitempty"`
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
	Players        []*HoldemWebOutputPlayer    `json:"players"`
	CommunityCards []*WebOutputCard            `json:"communityCards"`
	Pot            int                         `json:"pot"`
	SidePots       []*HoldemWebOutputSidePot   `json:"sidePots"`
	DealerIdx      int                         `json:"dealerIdx"`
	CurrentTurn    int                         `json:"currentTurn"`
	Phase          int                         `json:"phase"`
	GameEndFlag    bool                        `json:"gameEndFlag"`
	LastBet        int                         `json:"lastBet"`
	MinRaise       int                         `json:"minRaise"`
	RoundResults   []*HoldemWebOutputResult    `json:"roundResults"`
	CpuActions     []*HoldemWebOutputCpuAction `json:"cpuActions"`
	Message        string                      `json:"message"`
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
	var param HoldemWebInput
	err := r.DecodeJsonPayload(&param)
	if err != nil || param.Command == "" || param.SessionId == "" {
		w.WriteHeader(http.StatusBadRequest)
		if err := w.WriteJson(hwc.newDefaultOutput("param error.")); err != nil {
			log.Printf("WriteJson error: %v", err)
		}
		return
	}
	if param.Command == "q" || param.Command == "quit" {
		w.WriteHeader(http.StatusOK)
		if err := w.WriteJson(hwc.newDefaultOutput("bye.")); err != nil {
			log.Printf("WriteJson error: %v", err)
		}
		return
	}
	hgi, mu, ok := hwc.store.GetWithLock(param.SessionId, hwc.factory)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		if err := w.WriteJson(hwc.newDefaultOutput("param error.")); err != nil {
			log.Printf("WriteJson error: %v", err)
		}
		return
	}
	mu.Lock()
	defer mu.Unlock()
	errOutput := hwc.newDefaultOutput("error.")
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
			w.WriteHeader(http.StatusBadRequest)
			if err := w.WriteJson(hwc.newDefaultOutput("param error: smallBlind must be less than bigBlind.")); err != nil {
				log.Printf("WriteJson error: %v", err)
			}
			return
		}
		cfg.SmallBlind = sb
		cfg.BigBlind = bb
		hwc.writePresenterResponse(w, hgi.ResetWithConfig(cfg), errOutput)
	case "f", "fold":
		hwc.writePresenterResponse(w, hgi.Action(domain.HoldemActionFold, 0), errOutput)
	case "ck", "check":
		hwc.writePresenterResponse(w, hgi.Action(domain.HoldemActionCheck, 0), errOutput)
	case "c", "call":
		hwc.writePresenterResponse(w, hgi.Action(domain.HoldemActionCall, 0), errOutput)
	case "b", "bet":
		hwc.writePresenterResponse(w, hgi.Action(domain.HoldemActionBet, param.Amount), errOutput)
	case "ra", "raise":
		hwc.writePresenterResponse(w, hgi.Action(domain.HoldemActionRaise, param.Amount), errOutput)
	case "a", "allin":
		hwc.writePresenterResponse(w, hgi.Action(domain.HoldemActionAllIn, 0), errOutput)
	default:
		w.WriteHeader(http.StatusBadRequest)
		if err := w.WriteJson(hwc.newDefaultOutput("Unsupported command.")); err != nil {
			log.Printf("WriteJson error: %v", err)
		}
	}
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
