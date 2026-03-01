package controller

import (
	"log"
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"

	"github.com/ant0ine/go-json-rest/rest"
)

// DaifugoWebInputConfig リセット時のローカルルール設定入力
type DaifugoWebInputConfig struct {
	JokerCount          int  `json:"jokerCount"`
	EightCutEnabled     bool `json:"eightCutEnabled"`
	SuitLockEnabled     bool `json:"suitLockEnabled"`
	ElevenBackEnabled   bool `json:"elevenBackEnabled"`
	SequenceEnabled     bool `json:"sequenceEnabled"`
	CardExchangeEnabled bool `json:"cardExchangeEnabled"`
	FiveSkipEnabled     bool `json:"fiveSkipEnabled"`
	SevenPassEnabled    bool `json:"sevenPassEnabled"`
	TenDiscardEnabled   bool `json:"tenDiscardEnabled"`
	SpadeThreeEnabled   bool `json:"spadeThreeEnabled"`
	CapitalFallEnabled  bool `json:"capitalFallEnabled"`
}

// DaifugoWebInput 大富豪Webインプット
type DaifugoWebInput struct {
	Command   string                 `json:"command"`
	Indices   []int                  `json:"indices"` // 出すカードのインデックス。play コマンド用。空の場合はパス。
	SessionId string                 `json:"sessionId"`
	Config    *DaifugoWebInputConfig `json:"config"` // リセット時のローカルルール設定 (省略可)
}

// DaifugoWebOutputPlayer 大富豪Webアウトプットプレイヤー
type DaifugoWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	IsFinished bool             `json:"isFinished"`
	Rank       int              `json:"rank"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
}

// DaifugoWebOutputAction 大富豪のプレイヤー行動記録
type DaifugoWebOutputAction struct {
	PlayerIdx   int              `json:"playerIdx"`
	PlayedCards []*WebOutputCard `json:"playedCards"` // nil = パス
}

// DaifugoWebOutputExchangeAction カード交換記録
type DaifugoWebOutputExchangeAction struct {
	FromPlayerIdx int              `json:"fromPlayerIdx"`
	ToPlayerIdx   int              `json:"toPlayerIdx"`
	Cards         []*WebOutputCard `json:"cards"`
}

// DaifugoWebOutputConfig ローカルルール設定
type DaifugoWebOutputConfig struct {
	JokerCount          int  `json:"jokerCount"`
	EightCutEnabled     bool `json:"eightCutEnabled"`
	SuitLockEnabled     bool `json:"suitLockEnabled"`
	ElevenBackEnabled   bool `json:"elevenBackEnabled"`
	SequenceEnabled     bool `json:"sequenceEnabled"`
	CardExchangeEnabled bool `json:"cardExchangeEnabled"`
	FiveSkipEnabled     bool `json:"fiveSkipEnabled"`
	SevenPassEnabled    bool `json:"sevenPassEnabled"`
	TenDiscardEnabled   bool `json:"tenDiscardEnabled"`
	SpadeThreeEnabled   bool `json:"spadeThreeEnabled"`
	CapitalFallEnabled  bool `json:"capitalFallEnabled"`
}

// DaifugoWebOutput 大富豪Webアウトプット
type DaifugoWebOutput struct {
	Players             []*DaifugoWebOutputPlayer         `json:"players"`
	CurrentTurn         int                               `json:"currentTurn"`
	TableCards          []*WebOutputCard                  `json:"tableCards"`
	LastPlayPlayerIdx   int                               `json:"lastPlayPlayerIdx"`
	GameEndFlag         bool                              `json:"gameEndFlag"`
	RevolutionActive    bool                              `json:"revolutionActive"`
	ElevenBackActive    bool                              `json:"elevenBackActive"`
	SuitLocked          bool                              `json:"suitLocked"`
	LockedSuit          string                            `json:"lockedSuit"`
	TableIsSequence     bool                              `json:"tableIsSequence"`
	Config              DaifugoWebOutputConfig            `json:"config"`
	ExchangeActions     []*DaifugoWebOutputExchangeAction `json:"exchangeActions"`
	CpuActions          []*DaifugoWebOutputAction         `json:"cpuActions"`
	HumanAction         *DaifugoWebOutputAction           `json:"humanAction"`
	Message             string                            `json:"message"`
	PendingAction       string                            `json:"pendingAction"`       // "none"|"sevenPass"|"tenDiscard"
	PendingActionTarget int                               `json:"pendingActionTarget"` // -1 if none
}

// DaifugoWebController 大富豪Webコントローラークラス
type DaifugoWebController struct {
	baseController
	factory func() usecase.DaifugoInteractorIF
	store   *SessionStore[usecase.DaifugoInteractorIF]
}

// NewDaifugoWebController コンストラクタ
func NewDaifugoWebController(factory func() usecase.DaifugoInteractorIF) *DaifugoWebController {
	return &DaifugoWebController{
		factory: factory,
		store:   NewSessionStore[usecase.DaifugoInteractorIF](),
	}
}

// Exec ゲーム実行
func (dwc *DaifugoWebController) Exec(w rest.ResponseWriter, r *rest.Request) {
	var param DaifugoWebInput
	err := r.DecodeJsonPayload(&param)
	if err != nil || param.Command == "" || param.SessionId == "" {
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
		if param.Config != nil {
			dgConfig := convertWebInputConfig(*param.Config)
			dwc.writePresenterResponse(w, dgi.ResetWithConfig(dgConfig), errOutput)
		} else {
			dwc.writePresenterResponse(w, dgi.Reset(), errOutput)
		}
	case "p", "play":
		indices := param.Indices
		if indices == nil {
			indices = []int{}
		}
		dwc.writePresenterResponse(w, dgi.Play(indices), errOutput)
	default:
		w.WriteHeader(http.StatusBadRequest)
		if err := w.WriteJson(dwc.newDefaultOutput("Unsupported command.")); err != nil {
			log.Printf("WriteJson error: %v", err)
		}
	}
}

// convertWebInputConfig DaifugoWebInputConfig を domain.DaifugoConfig に変換
func convertWebInputConfig(c DaifugoWebInputConfig) domain.DaifugoConfig {
	return domain.DaifugoConfig{
		JokerCount:          c.JokerCount,
		EightCutEnabled:     c.EightCutEnabled,
		SuitLockEnabled:     c.SuitLockEnabled,
		ElevenBackEnabled:   c.ElevenBackEnabled,
		SequenceEnabled:     c.SequenceEnabled,
		CardExchangeEnabled: c.CardExchangeEnabled,
		FiveSkipEnabled:     c.FiveSkipEnabled,
		SevenPassEnabled:    c.SevenPassEnabled,
		TenDiscardEnabled:   c.TenDiscardEnabled,
		SpadeThreeEnabled:   c.SpadeThreeEnabled,
		CapitalFallEnabled:  c.CapitalFallEnabled,
	}
}

// Stop stops the background cleanup goroutine of the session store.
func (dwc *DaifugoWebController) Stop() {
	dwc.store.Stop()
}

// newDefaultOutput エラー・定型応答用のデフォルト出力を返す
func (dwc *DaifugoWebController) newDefaultOutput(msg string) *DaifugoWebOutput {
	return &DaifugoWebOutput{
		Players:             make([]*DaifugoWebOutputPlayer, 0),
		TableCards:          make([]*WebOutputCard, 0),
		CpuActions:          make([]*DaifugoWebOutputAction, 0),
		ExchangeActions:     make([]*DaifugoWebOutputExchangeAction, 0),
		Message:             msg,
		PendingAction:       "none",
		PendingActionTarget: -1,
	}
}
