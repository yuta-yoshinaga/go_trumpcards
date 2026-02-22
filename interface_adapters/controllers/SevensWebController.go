package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/usecases"

	"github.com/ant0ine/go-json-rest/rest"
)

// SevensWebInput 7並べWebインプット
type SevensWebInput struct {
	Command          string `json:"command"`
	Index            int    `json:"index"`                                      // 出すカードのインデックス。play コマンド用。-1 でパス。
	JokerTargetSuit  int    `json:"jokerTargetSuit"`                            // ジョーカー配置先スート
	JokerTargetValue int    `json:"jokerTargetValue"`                           // ジョーカー配置先値
	TunnelEnabled    *bool  `json:"tunnelEnabled,omitempty"`                    // トンネルルール (reset時のみ)
	JokerCount       *int   `json:"jokerCount,omitempty"`                       // ジョーカー枚数 (reset時のみ)
	CpuStrategy      *bool  `json:"cpuStrategy,omitempty"`                      // CPU戦略 (reset時のみ)
	SessionId        string `json:"sessionId"`
}

// SevensWebOutputCard 7並べWebアウトプットカード
type SevensWebOutputCard struct {
	Design string `json:"design"`
	Value  int    `json:"value"`
}

// SevensWebOutputPlayer 7並べWebアウトプットプレイヤー
type SevensWebOutputPlayer struct {
	ID         int                    `json:"id"`
	IsHuman    bool                   `json:"isHuman"`
	IsFinished bool                   `json:"isFinished"`
	Rank       int                    `json:"rank"`
	CardCount  int                    `json:"cardCount"`
	PassesUsed int                    `json:"passesUsed"`
	MaxPasses  int                    `json:"maxPasses"`
	Cards      []*SevensWebOutputCard `json:"cards"`
}

// SevensWebOutputAction 7並べのプレイヤー行動記録
type SevensWebOutputAction struct {
	PlayerIdx   int                  `json:"playerIdx"`
	PlayedCard  *SevensWebOutputCard `json:"playedCard"` // nil = パス
	TargetSuit  int                  `json:"targetSuit"`
	TargetValue int                  `json:"targetValue"`
}

// SevensWebOutputConfig 7並べゲーム設定出力
type SevensWebOutputConfig struct {
	TunnelEnabled bool `json:"tunnelEnabled"`
	JokerCount    int  `json:"jokerCount"`
	CpuStrategy   bool `json:"cpuStrategy"`
}

// SevensWebOutput 7並べWebアウトプット
type SevensWebOutput struct {
	Players      []*SevensWebOutputPlayer `json:"players"`
	CurrentTurn  int                      `json:"currentTurn"`
	TableMinVals [5]int                   `json:"tableMinVals"`
	TableMaxVals [5]int                   `json:"tableMaxVals"`
	TablePlaced  [5]int                   `json:"tablePlaced"`
	Config       SevensWebOutputConfig    `json:"config"`
	GameEndFlag  bool                     `json:"gameEndFlag"`
	CpuActions   []*SevensWebOutputAction `json:"cpuActions"`
	HumanAction  *SevensWebOutputAction   `json:"humanAction"`
	Message      string                   `json:"message"`
}

// SevensWebController 7並べWebコントローラークラス
type SevensWebController struct {
	factory func() usecases.SevensInteractorIF
	store   *SessionStore[usecases.SevensInteractorIF]
}

// NewSevensWebController コンストラクタ
func NewSevensWebController(factory func() usecases.SevensInteractorIF) *SevensWebController {
	return &SevensWebController{
		factory: factory,
		store:   NewSessionStore[usecases.SevensInteractorIF](),
	}
}

// Exec ゲーム実行
func (swc *SevensWebController) Exec(w rest.ResponseWriter, r *rest.Request) {
	var param SevensWebInput
	err := r.DecodeJsonPayload(&param)
	if err != nil || param.Command == "" || param.SessionId == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = w.WriteJson(swc.newDefaultOutput("param error."))
		return
	}
	if param.Command == "q" || param.Command == "quit" {
		w.WriteHeader(http.StatusOK)
		_ = w.WriteJson(swc.newDefaultOutput("bye."))
		return
	}
	sgi, ok := swc.store.Get(param.SessionId, swc.factory)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		_ = w.WriteJson(swc.newDefaultOutput("param error."))
		return
	}
	switch param.Command {
	case "r", "reset":
		if param.TunnelEnabled != nil || param.JokerCount != nil || param.CpuStrategy != nil {
			swc.writePresenterResponse(w, sgi.ResetWithConfig(derefBool(param.TunnelEnabled), derefInt(param.JokerCount), derefBool(param.CpuStrategy)))
		} else {
			swc.writePresenterResponse(w, sgi.Reset())
		}
	case "p", "play":
		swc.writePresenterResponse(w, sgi.Play(param.Index))
	case "j", "joker":
		swc.writePresenterResponse(w, sgi.PlayJoker(param.Index, param.JokerTargetSuit, param.JokerTargetValue))
	default:
		w.WriteHeader(http.StatusOK)
		_ = w.WriteJson(swc.newDefaultOutput("Unsupported command."))
	}
}

// writePresenterResponse プレゼンターの出力を再エンコードせず直接書き込む
func (swc *SevensWebController) writePresenterResponse(w rest.ResponseWriter, responseStr string) {
	if responseStr == "" || !json.Valid([]byte(responseStr)) {
		w.WriteHeader(http.StatusBadRequest)
		_ = w.WriteJson(swc.newDefaultOutput("error."))
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = w.WriteJson(json.RawMessage(responseStr))
}

func derefBool(p *bool) bool {
	if p == nil {
		return false
	}
	return *p
}

func derefInt(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}

// newDefaultOutput エラー・定型応答用のデフォルト出力を返す
func (swc *SevensWebController) newDefaultOutput(msg string) *SevensWebOutput {
	return &SevensWebOutput{
		Players:    make([]*SevensWebOutputPlayer, 0),
		CpuActions: make([]*SevensWebOutputAction, 0),
		Message:    msg,
	}
}
