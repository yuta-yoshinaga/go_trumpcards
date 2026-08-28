package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"

	"net/http"
)

// SevensWebInput 7並べWebインプット
type SevensWebInput struct {
	BaseWebInput
	Index                  int   `json:"index"`                            // 出すカードのインデックス。play コマンド用。-1 でパス。
	JokerTargetSuit        int   `json:"jokerTargetSuit"`                  // ジョーカー配置先スート
	JokerTargetValue       int   `json:"jokerTargetValue"`                 // ジョーカー配置先値
	TunnelEnabled          *bool `json:"tunnelEnabled,omitempty"`          // トンネルルール (reset時のみ)
	TunnelSkipWidth        *int  `json:"tunnelSkipWidth,omitempty"`        // カスタムトンネルスキップ幅 (reset時のみ)
	JokerCount             *int  `json:"jokerCount,omitempty"`             // ジョーカー枚数 (reset時のみ)
	CpuStrategy            *int  `json:"cpuStrategy,omitempty"`            // CPU戦略モード (reset時のみ, 0=シンプル, 1=戦略的, 2=嫌がらせ特化)
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
	// JokerReclaimed はこの手でジョーカーを回収したか。手札が黙って1枚増える
	// 唯一の経路なので、行動描写で説明できるように送る。
	JokerReclaimed bool `json:"jokerReclaimed"`
}

// SevensWebOutputConfig 7並べゲーム設定出力
type SevensWebOutputConfig struct {
	TunnelEnabled          bool `json:"tunnelEnabled"`
	TunnelSkipWidth        int  `json:"tunnelSkipWidth"`
	JokerCount             int  `json:"jokerCount"`
	CpuStrategy            int  `json:"cpuStrategy"`
	MaxPasses              int  `json:"maxPasses"`
	NoJokerFinish          bool `json:"noJokerFinish"`
	JokerReclaimEnabled    bool `json:"jokerReclaimEnabled"`
	EndStopEnabled         bool `json:"endStopEnabled"`
	JokerConsecutiveBanned bool `json:"jokerConsecutiveBanned"`
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
	WebOutputBase
}

// HasConfigParams reports whether any config parameter is set in the input.
func (p SevensWebInput) HasConfigParams() bool {
	return p.TunnelEnabled != nil || p.TunnelSkipWidth != nil || p.JokerCount != nil ||
		p.CpuStrategy != nil || p.MaxPasses != nil || p.NoJokerFinish != nil ||
		p.JokerReclaim != nil || p.EndStop != nil || p.JokerConsecutiveBanned != nil
}

// ToConfig builds a SevensConfig from the web input pointer fields.
func (p SevensWebInput) ToConfig() domain.SevensConfig {
	return domain.SevensConfig{
		TunnelEnabled:          deref(p.TunnelEnabled),
		TunnelSkipWidth:        deref(p.TunnelSkipWidth),
		JokerCount:             deref(p.JokerCount),
		CpuStrategy:            deref(p.CpuStrategy),
		MaxPasses:              derefDefault(p.MaxPasses, domain.SevensMaxPasses),
		NoJokerFinish:          deref(p.NoJokerFinish),
		JokerReclaimEnabled:    deref(p.JokerReclaim),
		EndStopEnabled:         deref(p.EndStop),
		JokerConsecutiveBanned: deref(p.JokerConsecutiveBanned),
	}
}

// SevensWebController 7並べWebコントローラークラス
type SevensWebController = GameWebController[usecase.SevensInteractorIF, SevensWebInput, *SevensWebOutput]

// NewSevensWebController and NewSevensWebControllerWithProvider are
// the standard and provider-backed constructors for SevensWebController.
var NewSevensWebController, NewSevensWebControllerWithProvider = webControllerPair[usecase.SevensInteractorIF, SevensWebInput, *SevensWebOutput](
	newSevensDefaultOutput, sevensDispatch,
)

func newSevensDefaultOutput(msg string) *SevensWebOutput {
	return &SevensWebOutput{
		Players:       make([]*SevensWebOutputPlayer, 0),
		CpuActions:    make([]*SevensWebOutputAction, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func sevensDispatch(bc *baseController, w http.ResponseWriter, sgi usecase.SevensInteractorIF, param SevensWebInput, _ func(string) *SevensWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		if param.HasConfigParams() {
			bc.writePresenterResponse(w, sgi.ResetWithConfig(param.ToConfig()))
		} else {
			bc.writePresenterResponse(w, sgi.Reset())
		}
	case "p", "play":
		bc.writePresenterResponse(w, sgi.Play(param.Index))
	case "j", "joker":
		bc.writePresenterResponse(w, sgi.PlayJoker(param.Index, param.JokerTargetSuit, param.JokerTargetValue))
	default:
		return dispatchLog(param.Command, bc, w, sgi.ActionLog)
	}
	return true
}
