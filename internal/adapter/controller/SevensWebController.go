package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"

	"github.com/ant0ine/go-json-rest/rest"
)

// SevensWebInput 7並べWebインプット
type SevensWebInput struct {
	BaseWebInput
	Index                  int   `json:"index"`                            // 出すカードのインデックス。play コマンド用。-1 でパス。
	JokerTargetSuit        int   `json:"jokerTargetSuit"`                  // ジョーカー配置先スート
	JokerTargetValue       int   `json:"jokerTargetValue"`                 // ジョーカー配置先値
	TunnelEnabled          *bool `json:"tunnelEnabled,omitempty"`          // トンネルルール (reset時のみ)
	JokerCount             *int  `json:"jokerCount,omitempty"`             // ジョーカー枚数 (reset時のみ)
	CpuStrategy            *bool `json:"cpuStrategy,omitempty"`            // CPU戦略 (reset時のみ)
	MaxPasses              *int  `json:"maxPasses,omitempty"`              // 最大パス回数 (reset時のみ, 0=無制限)
	NoJokerFinish          *bool `json:"noJokerFinish,omitempty"`          // ジョーカー上がり禁止 (reset時のみ)
	JokerReclaim           *bool `json:"jokerReclaim,omitempty"`           // ジョーカー回収 (reset時のみ)
	EndStop                *bool `json:"endStop,omitempty"`                // 片側ストップ (reset時のみ)
	JokerConsecutiveBanned *bool `json:"jokerConsecutiveBanned,omitempty"` // ジョーカー連続禁止 (reset時のみ)
}

// SevensWebOutputPlayer 7並べWebアウトプットプレイヤー
type SevensWebOutputPlayer struct {
	ID              int              `json:"id"`
	IsHuman         bool             `json:"isHuman"`
	IsFinished      bool             `json:"isFinished"`
	Rank            int              `json:"rank"`
	CardCount       int              `json:"cardCount"`
	PassesUsed      int              `json:"passesUsed"`
	MaxPasses       int              `json:"maxPasses"`
	Cards           []*WebOutputCard `json:"cards"`
	LastPlayedJoker bool             `json:"lastPlayedJoker"`
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
	TunnelEnabled          bool `json:"tunnelEnabled"`
	JokerCount             int  `json:"jokerCount"`
	CpuStrategy            bool `json:"cpuStrategy"`
	MaxPasses              int  `json:"maxPasses"`
	NoJokerFinish          bool `json:"noJokerFinish"`
	JokerReclaimEnabled    bool `json:"jokerReclaimEnabled"`
	EndStopEnabled         bool `json:"endStopEnabled"`
	JokerConsecutiveBanned bool `json:"jokerConsecutiveBanned"`
}

// SevensWebOutput 7並べWebアウトプット
type SevensWebOutput struct {
	Players       []*SevensWebOutputPlayer `json:"players"`
	CurrentTurn   int                      `json:"currentTurn"`
	TableMinVals  [5]int                   `json:"tableMinVals"`
	TableMaxVals  [5]int                   `json:"tableMaxVals"`
	TablePlaced   [5]int                   `json:"tablePlaced"`
	Config        SevensWebOutputConfig    `json:"config"`
	GameEndFlag   bool                     `json:"gameEndFlag"`
	CpuActions    []*SevensWebOutputAction `json:"cpuActions"`
	HumanAction   *SevensWebOutputAction   `json:"humanAction"`
	Message       string                   `json:"message"`
	MessageCode   string                   `json:"messageCode,omitempty"`
	MessageParams map[string]string        `json:"messageParams,omitempty"`
}

// SevensWebController 7並べWebコントローラークラス
type SevensWebController = GameWebController[usecase.SevensInteractorIF, SevensWebInput, *SevensWebOutput]

// NewSevensWebController コンストラクタ
func NewSevensWebController(factory func() usecase.SevensInteractorIF) *SevensWebController {
	return NewGameWebController(factory, newSevensDefaultOutput, nil, sevensDispatch)
}

func newSevensDefaultOutput(msg string) *SevensWebOutput {
	return &SevensWebOutput{
		Players:    make([]*SevensWebOutputPlayer, 0),
		CpuActions: make([]*SevensWebOutputAction, 0),
		Message:    msg,
	}
}

func sevensDispatch(bc *baseController, w rest.ResponseWriter, sgi usecase.SevensInteractorIF, param SevensWebInput, _ func(string) *SevensWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		if param.TunnelEnabled != nil || param.JokerCount != nil || param.CpuStrategy != nil || param.MaxPasses != nil || param.NoJokerFinish != nil || param.JokerReclaim != nil || param.EndStop != nil || param.JokerConsecutiveBanned != nil {
			cfg := domain.SevensConfig{
				TunnelEnabled:          derefBool(param.TunnelEnabled),
				JokerCount:             derefInt(param.JokerCount),
				CpuStrategy:            derefBool(param.CpuStrategy),
				MaxPasses:              derefIntDefault(param.MaxPasses, domain.SevensMaxPasses),
				NoJokerFinish:          derefBool(param.NoJokerFinish),
				JokerReclaimEnabled:    derefBool(param.JokerReclaim),
				EndStopEnabled:         derefBool(param.EndStop),
				JokerConsecutiveBanned: derefBool(param.JokerConsecutiveBanned),
			}
			bc.writePresenterResponse(w, sgi.ResetWithConfig(cfg))
		} else {
			bc.writePresenterResponse(w, sgi.Reset())
		}
	case "p", "play":
		bc.writePresenterResponse(w, sgi.Play(param.Index))
	case "j", "joker":
		bc.writePresenterResponse(w, sgi.PlayJoker(param.Index, param.JokerTargetSuit, param.JokerTargetValue))
	default:
		return false
	}
	return true
}
