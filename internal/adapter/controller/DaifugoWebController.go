package controller

import (
	"encoding/json"
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"

	"github.com/ant0ine/go-json-rest/rest"
)

// DaifugoWebInput 大富豪Webインプット
type DaifugoWebInput struct {
	Command   string `json:"command"`
	Indices   []int  `json:"indices"` // 出すカードのインデックス。play コマンド用。空の場合はパス。
	SessionId string `json:"sessionId"`
}

// DaifugoWebOutputCard 大富豪Webアウトプットカード
type DaifugoWebOutputCard struct {
	Design string `json:"design"`
	Value  int    `json:"value"`
}

// DaifugoWebOutputPlayer 大富豪Webアウトプットプレイヤー
type DaifugoWebOutputPlayer struct {
	ID         int                     `json:"id"`
	IsHuman    bool                    `json:"isHuman"`
	IsFinished bool                    `json:"isFinished"`
	Rank       int                     `json:"rank"`
	CardCount  int                     `json:"cardCount"`
	Cards      []*DaifugoWebOutputCard `json:"cards"`
}

// DaifugoWebOutputAction 大富豪のプレイヤー行動記録
type DaifugoWebOutputAction struct {
	PlayerIdx   int                     `json:"playerIdx"`
	PlayedCards []*DaifugoWebOutputCard `json:"playedCards"` // nil = パス
}

// DaifugoWebOutputExchangeAction カード交換記録
type DaifugoWebOutputExchangeAction struct {
	FromPlayerIdx int                     `json:"fromPlayerIdx"`
	ToPlayerIdx   int                     `json:"toPlayerIdx"`
	Cards         []*DaifugoWebOutputCard `json:"cards"`
}

// DaifugoWebOutputConfig ローカルルール設定
type DaifugoWebOutputConfig struct {
	JokerCount          int  `json:"jokerCount"`
	EightCutEnabled     bool `json:"eightCutEnabled"`
	SuitLockEnabled     bool `json:"suitLockEnabled"`
	ElevenBackEnabled   bool `json:"elevenBackEnabled"`
	SequenceEnabled     bool `json:"sequenceEnabled"`
	CardExchangeEnabled bool `json:"cardExchangeEnabled"`
}

// DaifugoWebOutput 大富豪Webアウトプット
type DaifugoWebOutput struct {
	Players           []*DaifugoWebOutputPlayer        `json:"players"`
	CurrentTurn       int                              `json:"currentTurn"`
	TableCards        []*DaifugoWebOutputCard          `json:"tableCards"`
	LastPlayPlayerIdx int                              `json:"lastPlayPlayerIdx"`
	GameEndFlag       bool                             `json:"gameEndFlag"`
	RevolutionActive  bool                             `json:"revolutionActive"`
	ElevenBackActive  bool                             `json:"elevenBackActive"`
	SuitLocked        bool                             `json:"suitLocked"`
	LockedSuit        string                           `json:"lockedSuit"`
	TableIsSequence   bool                             `json:"tableIsSequence"`
	Config            DaifugoWebOutputConfig           `json:"config"`
	ExchangeActions   []*DaifugoWebOutputExchangeAction `json:"exchangeActions"`
	CpuActions        []*DaifugoWebOutputAction        `json:"cpuActions"`
	HumanAction       *DaifugoWebOutputAction          `json:"humanAction"`
	Message           string                           `json:"message"`
}

// DaifugoWebController 大富豪Webコントローラークラス
type DaifugoWebController struct {
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
		_ = w.WriteJson(dwc.newDefaultOutput("param error."))
		return
	}
	if param.Command == "q" || param.Command == "quit" {
		w.WriteHeader(http.StatusOK)
		_ = w.WriteJson(dwc.newDefaultOutput("bye."))
		return
	}
	dgi, mu, ok := dwc.store.GetWithLock(param.SessionId, dwc.factory)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		_ = w.WriteJson(dwc.newDefaultOutput("param error."))
		return
	}
	mu.Lock()
	defer mu.Unlock()
	switch param.Command {
	case "r", "reset":
		dwc.writePresenterResponse(w, dgi.Reset())
	case "p", "play":
		indices := param.Indices
		if indices == nil {
			indices = []int{}
		}
		dwc.writePresenterResponse(w, dgi.Play(indices))
	default:
		w.WriteHeader(http.StatusOK)
		_ = w.WriteJson(dwc.newDefaultOutput("Unsupported command."))
	}
}

// writePresenterResponse プレゼンターの出力を再エンコードせず直接書き込む
func (dwc *DaifugoWebController) writePresenterResponse(w rest.ResponseWriter, responseStr string) {
	if responseStr == "" || !json.Valid([]byte(responseStr)) {
		w.WriteHeader(http.StatusBadRequest)
		_ = w.WriteJson(dwc.newDefaultOutput("error."))
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = w.WriteJson(json.RawMessage(responseStr))
}

// newDefaultOutput エラー・定型応答用のデフォルト出力を返す
func (dwc *DaifugoWebController) newDefaultOutput(msg string) *DaifugoWebOutput {
	return &DaifugoWebOutput{
		Players:         make([]*DaifugoWebOutputPlayer, 0),
		TableCards:      make([]*DaifugoWebOutputCard, 0),
		CpuActions:      make([]*DaifugoWebOutputAction, 0),
		ExchangeActions: make([]*DaifugoWebOutputExchangeAction, 0),
		Message:         msg,
	}
}
