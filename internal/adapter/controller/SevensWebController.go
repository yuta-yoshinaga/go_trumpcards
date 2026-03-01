package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"

	"github.com/ant0ine/go-json-rest/rest"
)

// SevensWebInput 7並べWebインプット
type SevensWebInput struct {
	Command          string `json:"command"`
	Index            int    `json:"index"`                   // 出すカードのインデックス。play コマンド用。-1 でパス。
	JokerTargetSuit  int    `json:"jokerTargetSuit"`         // ジョーカー配置先スート
	JokerTargetValue int    `json:"jokerTargetValue"`        // ジョーカー配置先値
	TunnelEnabled    *bool  `json:"tunnelEnabled,omitempty"` // トンネルルール (reset時のみ)
	JokerCount       *int   `json:"jokerCount,omitempty"`    // ジョーカー枚数 (reset時のみ)
	CpuStrategy      *bool  `json:"cpuStrategy,omitempty"`   // CPU戦略 (reset時のみ)
	MaxPasses        *int   `json:"maxPasses,omitempty"`     // 最大パス回数 (reset時のみ, 0=無制限)
	SessionId        string `json:"sessionId"`
}

// GetCommand returns the command string.
func (i SevensWebInput) GetCommand() string { return i.Command }

// GetSessionID returns the session ID string.
func (i SevensWebInput) GetSessionID() string { return i.SessionId }

// SevensWebOutputPlayer 7並べWebアウトプットプレイヤー
type SevensWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	IsFinished bool             `json:"isFinished"`
	Rank       int              `json:"rank"`
	CardCount  int              `json:"cardCount"`
	PassesUsed int              `json:"passesUsed"`
	MaxPasses  int              `json:"maxPasses"`
	Cards      []*WebOutputCard `json:"cards"`
}

// SevensWebOutputAction 7並べのプレイヤー行動記録
type SevensWebOutputAction struct {
	PlayerIdx   int            `json:"playerIdx"`
	PlayedCard  *WebOutputCard `json:"playedCard"` // nil = パス
	TargetSuit  int            `json:"targetSuit"`
	TargetValue int            `json:"targetValue"`
	ForcedPass  bool           `json:"forcedPass"`
}

// SevensWebOutputConfig 7並べゲーム設定出力
type SevensWebOutputConfig struct {
	TunnelEnabled bool `json:"tunnelEnabled"`
	JokerCount    int  `json:"jokerCount"`
	CpuStrategy   bool `json:"cpuStrategy"`
	MaxPasses     int  `json:"maxPasses"`
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
	baseController
	factory func() usecase.SevensInteractorIF
	store   *SessionStore[usecase.SevensInteractorIF]
}

// NewSevensWebController コンストラクタ
func NewSevensWebController(factory func() usecase.SevensInteractorIF) *SevensWebController {
	return &SevensWebController{
		factory: factory,
		store:   NewSessionStore[usecase.SevensInteractorIF](),
	}
}

// Exec ゲーム実行
func (swc *SevensWebController) Exec(w rest.ResponseWriter, r *rest.Request) {
	execWithSession(&swc.baseController, w, r, swc.store, swc.factory,
		func(msg string) any { return swc.newDefaultOutput(msg) },
		nil,
		func(w rest.ResponseWriter, sgi usecase.SevensInteractorIF, param SevensWebInput) bool {
			switch param.Command {
			case "r", "reset":
				if param.TunnelEnabled != nil || param.JokerCount != nil || param.CpuStrategy != nil || param.MaxPasses != nil {
					jokerCount := derefInt(param.JokerCount)
					if jokerCount < 0 || jokerCount > domain.SevensMaxJokerCount {
						swc.writeJsonResponse(w, http.StatusBadRequest, swc.newDefaultOutput("param error: jokerCount must be between 0 and 2."))
						return true
					}
					swc.writePresenterResponse(w, sgi.ResetWithConfig(derefBool(param.TunnelEnabled), jokerCount, derefBool(param.CpuStrategy), derefIntDefault(param.MaxPasses, domain.SevensMaxPasses)))
				} else {
					swc.writePresenterResponse(w, sgi.Reset())
				}
			case "p", "play":
				swc.writePresenterResponse(w, sgi.Play(param.Index))
			case "j", "joker":
				swc.writePresenterResponse(w, sgi.PlayJoker(param.Index, param.JokerTargetSuit, param.JokerTargetValue))
			default:
				return false
			}
			return true
		})
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

func derefIntDefault(p *int, defaultVal int) int {
	if p == nil {
		return defaultVal
	}
	return *p
}

// Stop stops the background cleanup goroutine of the session store.
func (swc *SevensWebController) Stop() {
	swc.store.Stop()
}

// newDefaultOutput エラー・定型応答用のデフォルト出力を返す
func (swc *SevensWebController) newDefaultOutput(msg string) *SevensWebOutput {
	return &SevensWebOutput{
		Players:    make([]*SevensWebOutputPlayer, 0),
		CpuActions: make([]*SevensWebOutputAction, 0),
		Message:    msg,
	}
}
