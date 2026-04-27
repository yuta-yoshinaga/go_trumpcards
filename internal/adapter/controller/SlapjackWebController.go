package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SlapjackWebInput スラップジャック Web 入力
type SlapjackWebInput struct {
	BaseWebInput
	PlayerIdx *int               `json:"playerIdx,omitempty"`
	Config    *SlapjackWebConfig `json:"config,omitempty"`
}

// SlapjackWebConfig スラップジャックの設定リクエスト
type SlapjackWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
}

// SlapjackWebPlayer プレイヤー出力
type SlapjackWebPlayer struct {
	Name      string `json:"name"`
	IsHuman   bool   `json:"isHuman"`
	StockSize int    `json:"stockSize"`
}

// SlapjackWebOutput スラップジャック Web 出力
type SlapjackWebOutput struct {
	Phase              int                  `json:"phase"`
	GameEndFlag        bool                 `json:"gameEndFlag"`
	WinnerIdx          int                  `json:"winnerIdx"`
	CurrentTurnIdx     int                  `json:"currentTurnIdx"`
	IsHumanTurn        bool                 `json:"isHumanTurn"`
	IsTopJack          bool                 `json:"isTopJack"`
	CenterPileSize     int                  `json:"centerPileSize"`
	TopCard            *WebOutputCard       `json:"topCard,omitempty"`
	Players            []*SlapjackWebPlayer `json:"players"`
	CpuDifficulty      int                  `json:"cpuDifficulty"`
	PendingKind        int                  `json:"pendingKind"`
	PendingDeadlineMs  int64                `json:"pendingDeadlineMs"`
	LastEventKind      int                  `json:"lastEventKind"`
	LastEventPlayerIdx int                  `json:"lastEventPlayerIdx"`
	WebOutputBase
}

// SlapjackWebController スラップジャック Web コントローラー
type SlapjackWebController = GameWebController[usecase.SlapjackInteractorIF, SlapjackWebInput, *SlapjackWebOutput]

// NewSlapjackWebController and NewSlapjackWebControllerWithProvider are the
// standard and provider-backed constructors for SlapjackWebController.
var NewSlapjackWebController, NewSlapjackWebControllerWithProvider = webControllerPair[usecase.SlapjackInteractorIF, SlapjackWebInput, *SlapjackWebOutput](
	newSlapjackDefaultOutput, slapjackDispatch,
)

func newSlapjackDefaultOutput(msg string) *SlapjackWebOutput {
	return &SlapjackWebOutput{
		WinnerIdx:     -1,
		CpuDifficulty: int(domain.SlapjackCpuNormal),
		Players:       make([]*SlapjackWebPlayer, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func slapjackDispatch(bc *baseController, w http.ResponseWriter, si usecase.SlapjackInteractorIF, param SlapjackWebInput, _ func(string) *SlapjackWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		if param.Config != nil {
			bc.writePresenterResponse(w, si.ResetWithConfig(slapjackConfigFromInput(si.GetConfig(), param.Config)))
		} else {
			bc.writePresenterResponse(w, si.Reset())
		}
		return true
	case "s", "step":
		bc.writePresenterResponse(w, si.Step())
		return true
	case "j", "slap":
		bc.writePresenterResponse(w, si.Slap(derefDefault(param.PlayerIdx, 0)))
		return true
	case "tick":
		bc.writePresenterResponse(w, si.Tick())
		return true
	case "log", "l":
		bc.writePresenterResponse(w, si.ActionLog())
		return true
	}
	return false
}

// slapjackConfigFromInput merges the partial Web config request into the current
// config so missing fields default to existing values rather than zero.
func slapjackConfigFromInput(current domain.SlapjackConfig, in *SlapjackWebConfig) domain.SlapjackConfig {
	out := current
	if in.CpuDifficulty != nil {
		out.CpuDifficulty = domain.SlapjackCpuDifficulty(*in.CpuDifficulty)
	}
	return out
}
