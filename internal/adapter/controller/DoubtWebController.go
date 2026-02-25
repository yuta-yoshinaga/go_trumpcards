package controller

import (
	"fmt"
	"log"
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"

	"github.com/ant0ine/go-json-rest/rest"
)

// DoubtWebInput ダウトWebインプット
type DoubtWebInput struct {
	Command        string `json:"command"`
	CardIndices    []int  `json:"cardIndices,omitempty"`
	ClaimedValue   int    `json:"claimedValue,omitempty"`
	DoubterIndices []int  `json:"doubterIndices,omitempty"`
	SessionId      string `json:"sessionId"`
}

// DoubtWebOutputCard ダウトWebアウトプットカード
type DoubtWebOutputCard struct {
	Design string `json:"design"`
	Value  int    `json:"value"`
}

// DoubtWebOutputPlayer ダウトWebアウトプットプレイヤー
type DoubtWebOutputPlayer struct {
	ID         int                   `json:"id"`
	IsHuman    bool                  `json:"isHuman"`
	IsFinished bool                  `json:"isFinished"`
	CardCount  int                   `json:"cardCount"`
	Cards      []*DoubtWebOutputCard `json:"cards"`
}

// DoubtWebOutputAction ダウトのプレイヤー行動記録
type DoubtWebOutputAction struct {
	PlayerIdx    int  `json:"playerIdx"`
	ClaimedValue int  `json:"claimedValue"`
	CardCount    int  `json:"cardCount"`
	IsBluff      bool `json:"isBluff"`
}

// DoubtWebOutputDoubtResult ダウト解決結果
type DoubtWebOutputDoubtResult struct {
	DoubterIdx    int                   `json:"doubterIdx"`
	CardPlayerIdx int                   `json:"cardPlayerIdx"`
	WasLying      bool                  `json:"wasLying"`
	LoserIdx      int                   `json:"loserIdx"`
	CardCount     int                   `json:"cardCount"`
	RevealedCards []*DoubtWebOutputCard `json:"revealedCards"`
}

// DoubtWebOutput ダウトWebアウトプット
type DoubtWebOutput struct {
	Players         []*DoubtWebOutputPlayer    `json:"players"`
	CurrentTurn     int                        `json:"currentTurn"`
	Phase           int                        `json:"phase"`
	TableCardCount  int                        `json:"tableCardCount"`
	LastAction      *DoubtWebOutputAction      `json:"lastAction"`
	CpuDoubters     []int                      `json:"cpuDoubters"`
	CpuActions      []*DoubtWebOutputAction    `json:"cpuActions"`
	HumanAction     *DoubtWebOutputAction      `json:"humanAction"`
	LastDoubtResult *DoubtWebOutputDoubtResult `json:"lastDoubtResult"`
	GameEndFlag     bool                       `json:"gameEndFlag"`
	WinnerIdx       int                        `json:"winnerIdx"`
	Message         string                     `json:"message"`
}

// DoubtWebController ダウトWebコントローラークラス
type DoubtWebController struct {
	baseController
	factory func() usecase.DoubtInteractorIF
	store   *SessionStore[usecase.DoubtInteractorIF]
}

// NewDoubtWebController コンストラクタ
func NewDoubtWebController(factory func() usecase.DoubtInteractorIF) *DoubtWebController {
	return &DoubtWebController{
		factory: factory,
		store:   NewSessionStore[usecase.DoubtInteractorIF](),
	}
}

// MaxCardIndices カードインデックスの最大数 (52枚デッキ)
const MaxCardIndices = 52

// Exec ゲーム実行
func (dwc *DoubtWebController) Exec(w rest.ResponseWriter, r *rest.Request) {
	var param DoubtWebInput
	err := r.DecodeJsonPayload(&param)
	if err != nil || param.Command == "" || param.SessionId == "" || len(param.CardIndices) > MaxCardIndices {
		w.WriteHeader(http.StatusBadRequest)
		if err := w.WriteJson(dwc.newDefaultOutput("param error.")); err != nil {
			log.Printf("WriteJson error: %v", err)
		}
		return
	}
	if param.Command == "q" || param.Command == "quit" {
		w.WriteHeader(http.StatusOK)
		if err := w.WriteJson(dwc.newDefaultOutput("bye.")); err != nil {
			log.Printf("WriteJson error: %v", err)
		}
		return
	}
	dgi, mu, ok := dwc.store.GetWithLock(param.SessionId, dwc.factory)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		if err := w.WriteJson(dwc.newDefaultOutput("param error.")); err != nil {
			log.Printf("WriteJson error: %v", err)
		}
		return
	}
	mu.Lock()
	defer mu.Unlock()
	errOutput := dwc.newDefaultOutput("error.")
	switch param.Command {
	case "r", "reset":
		dwc.writePresenterResponse(w, dgi.Reset(), errOutput)
	case "p", "play":
		if param.ClaimedValue < domain.MinClaimedValue || param.ClaimedValue > domain.MaxClaimedValue {
			w.WriteHeader(http.StatusBadRequest)
			if err := w.WriteJson(dwc.newDefaultOutput(fmt.Sprintf("param error: claimedValue must be between %d and %d.", domain.MinClaimedValue, domain.MaxClaimedValue))); err != nil {
				log.Printf("WriteJson error: %v", err)
			}
			return
		}
		dwc.writePresenterResponse(w, dgi.Play(param.CardIndices, param.ClaimedValue), errOutput)
	case "d", "doubt":
		cpuDoubters := dgi.GetCpuDoubters()
		humanDoubts := false
		for _, idx := range param.DoubterIndices {
			if idx == 0 {
				humanDoubts = true
				break
			}
		}
		var doubters []int
		if humanDoubts {
			doubters = append([]int{0}, cpuDoubters...)
		} else {
			doubters = cpuDoubters
		}
		dwc.writePresenterResponse(w, dgi.ResolveDoubt(doubters), errOutput)
	case "s", "skip":
		cpuDoubters := dgi.GetCpuDoubters()
		if len(cpuDoubters) > 0 {
			dwc.writePresenterResponse(w, dgi.ResolveDoubt(cpuDoubters), errOutput)
		} else {
			dwc.writePresenterResponse(w, dgi.SkipDoubt(), errOutput)
		}
	default:
		w.WriteHeader(http.StatusOK)
		if err := w.WriteJson(dwc.newDefaultOutput("Unsupported command.")); err != nil {
			log.Printf("WriteJson error: %v", err)
		}
	}
}

// newDefaultOutput エラー・定型応答用のデフォルト出力を返す
func (dwc *DoubtWebController) newDefaultOutput(msg string) *DoubtWebOutput {
	return &DoubtWebOutput{
		Players:     make([]*DoubtWebOutputPlayer, 0),
		CpuDoubters: make([]int, 0),
		CpuActions:  make([]*DoubtWebOutputAction, 0),
		WinnerIdx:   -1,
		Message:     msg,
	}
}
