package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// WarWebInput 戦争Webインプット
type WarWebInput struct {
	BaseWebInput
	MaxRounds *int `json:"maxRounds,omitempty"`
}

// ToConfig インプットからWarConfigを生成する
func (in WarWebInput) ToConfig() domain.WarConfig {
	cfg := domain.DefaultWarConfig()
	if in.MaxRounds != nil {
		cfg.MaxRounds = webutil.BoundedIntPtr(in.MaxRounds, domain.WarMinMaxRounds, domain.WarMaxMaxRounds, cfg.MaxRounds)
	}
	return cfg
}

// WarWebOutputPlayer 戦争Webアウトプットプレイヤー
type WarWebOutputPlayer struct {
	ID              int  `json:"id"`
	IsHuman         bool `json:"isHuman"`
	DrawPileSize    int  `json:"drawPileSize"`
	DiscardPileSize int  `json:"discardPileSize"`
	TotalCards      int  `json:"totalCards"`
}

// WarWebOutputConfig 設定情報
type WarWebOutputConfig struct {
	MaxRounds int `json:"maxRounds"`
}

// WarWebOutput 戦争Webアウトプット
type WarWebOutput struct {
	Players         []*WarWebOutputPlayer `json:"players"`
	Phase           int                   `json:"phase"`
	GameEndFlag     bool                  `json:"gameEndFlag"`
	WinnerIdx       int                   `json:"winnerIdx"`
	PlayerRevealed  *WebOutputCard        `json:"playerRevealed"`
	CpuRevealed     *WebOutputCard        `json:"cpuRevealed"`
	WarPotSize      int                   `json:"warPotSize"`
	LastWinnerIdx   int                   `json:"lastWinnerIdx"`
	LastBurialCount int                   `json:"lastBurialCount"`
	RoundsPlayed    int                   `json:"roundsPlayed"`
	Config          WarWebOutputConfig    `json:"config"`
	WebOutputBase
}

// WarWebController 戦争Webコントローラー型
type WarWebController = GameWebController[usecase.WarInteractorIF, WarWebInput, *WarWebOutput]

// NewWarWebController and NewWarWebControllerWithProvider are
// the standard and provider-backed constructors for WarWebController.
var NewWarWebController, NewWarWebControllerWithProvider = webControllerPair[usecase.WarInteractorIF, WarWebInput, *WarWebOutput](
	newWarDefaultOutput, warDispatch,
)

func newWarDefaultOutput(msg string) *WarWebOutput {
	o := new(WarWebOutput)
	o.Message = msg
	return o
}

func warDispatch(
	bc *baseController,
	w http.ResponseWriter,
	wi usecase.WarInteractorIF,
	param WarWebInput,
	_ func(string) *WarWebOutput,
) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, wi.ResetWithConfig(param.ToConfig()))
	case "s", "step":
		bc.writePresenterResponse(w, wi.Step())
	case "log", "l":
		bc.writePresenterResponse(w, wi.ActionLog())
	default:
		return false
	}
	return true
}
