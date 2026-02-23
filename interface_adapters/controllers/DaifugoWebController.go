package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/usecases"

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
	factory func() usecases.DaifugoInteractorIF
	store   *SessionStore[usecases.DaifugoInteractorIF]
}

// NewDaifugoWebController コンストラクタ
func NewDaifugoWebController(factory func() usecases.DaifugoInteractorIF) *DaifugoWebController {
	return &DaifugoWebController{
		factory: factory,
		store:   NewSessionStore[usecases.DaifugoInteractorIF](),
	}
}

// Exec ゲーム実行
func (dwc *DaifugoWebController) Exec(w rest.ResponseWriter, r *rest.Request) {
	var param DaifugoWebInput
	status := http.StatusOK
	responseStr := ""
	err := r.DecodeJsonPayload(&param)
	if err != nil || param.Command == "" || param.SessionId == "" {
		status = http.StatusBadRequest
		responseStr = `{"message":"param error."}`
	} else if param.Command == "q" || param.Command == "quit" {
		responseStr = `{"message":"bye."}`
	} else {
		dgi, mu, ok := dwc.store.GetWithLock(param.SessionId, dwc.factory)
		if !ok {
			status = http.StatusBadRequest
			responseStr = `{"message":"param error."}`
		} else {
			mu.Lock()
			defer mu.Unlock()
			switch param.Command {
			case "r", "reset":
				responseStr = dgi.Reset()
			case "p", "play":
				indices := param.Indices
				if indices == nil {
					indices = []int{}
				}
				responseStr = dgi.Play(indices)
			default:
				responseStr = `{"message":"Unsupported command."}`
			}
		}
	}
	response := new(DaifugoWebOutput)
	err = json.Unmarshal([]byte(responseStr), &response)
	if err != nil || responseStr == "" {
		status = http.StatusBadRequest
		response.Message = "error."
	}
	// nil スライスは JSON で null になるので空スライスに統一する
	if response.Players == nil {
		response.Players = make([]*DaifugoWebOutputPlayer, 0)
	}
	if response.TableCards == nil {
		response.TableCards = make([]*DaifugoWebOutputCard, 0)
	}
	if response.CpuActions == nil {
		response.CpuActions = make([]*DaifugoWebOutputAction, 0)
	}
	if response.ExchangeActions == nil {
		response.ExchangeActions = make([]*DaifugoWebOutputExchangeAction, 0)
	}
	w.WriteHeader(status)
	_ = w.WriteJson(response)
}
