//go:build !js || !wasm || extra2

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CuckooWebInput Cuckoo Webインプット
type CuckooWebInput struct {
	BaseWebInput
	Config *CuckooWebConfig `json:"config,omitempty"`
}

// CuckooWebConfig Cuckoo Web設定
type CuckooWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	InitialLives  *int `json:"initialLives,omitempty"`
}

// CuckooWebOutputPlayer Cuckoo Webアウトプットプレイヤー
type CuckooWebOutputPlayer struct {
	ID            int            `json:"id"`
	IsHuman       bool           `json:"isHuman"`
	Card          *WebOutputCard `json:"card"`
	Lives         int            `json:"lives"`
	IsEliminated  bool           `json:"isEliminated"`
	KingRevealed  bool           `json:"kingRevealed"`
	IsCurrentTurn bool           `json:"isCurrentTurn"`
}

// CuckooWebOutput Cuckoo Webアウトプット
type CuckooWebOutput struct {
	Players          []*CuckooWebOutputPlayer `json:"players"`
	Phase            int                      `json:"phase"`
	RoundNumber      int                      `json:"roundNumber"`
	CurrentPlayerIdx int                      `json:"currentPlayerIdx"`
	DealerIdx        int                      `json:"dealerIdx"`
	StockCount       int                      `json:"stockCount"`
	GameEndFlag      bool                     `json:"gameEndFlag"`
	WinnerIdx        int                      `json:"winnerIdx"`
	PendingSwapFrom  int                      `json:"pendingSwapFrom"`
	PendingSwapTo    int                      `json:"pendingSwapTo"`
	RoundLowest      int                      `json:"roundLowest"`
	RoundLosers      []int                    `json:"roundLosers"`
	WebOutputBase
	Config CuckooWebOutputConfig `json:"config"`
}

// CuckooWebOutputConfig Cuckoo設定アウトプット
type CuckooWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	InitialLives  int `json:"initialLives"`
}

// ToConfig builds a CuckooConfig from the nested web config, applying bounds checking.
func (c *CuckooWebConfig) ToConfig() domain.CuckooConfig {
	cfg := domain.DefaultCuckooConfig()
	cfg.CpuDifficulty = domain.CuckooCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.CuckooCpuDifficultyEasy), int(domain.CuckooCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.InitialLives, c.InitialLives, domain.CuckooMinLives, domain.CuckooMaxLives)
	return cfg
}

// ToConfig builds a CuckooConfig from the web input.
func (p CuckooWebInput) ToConfig() domain.CuckooConfig {
	return configOrDefault(p.Config, (*CuckooWebConfig).ToConfig, domain.DefaultCuckooConfig())
}

// CuckooWebController Cuckoo Webコントローラークラス
type CuckooWebController = GameWebController[usecase.CuckooInteractorIF, CuckooWebInput, *CuckooWebOutput]

// NewCuckooWebController and NewCuckooWebControllerWithProvider are
// the standard and provider-backed constructors for CuckooWebController.
var NewCuckooWebController, NewCuckooWebControllerWithProvider = webControllerPair[usecase.CuckooInteractorIF, CuckooWebInput, *CuckooWebOutput](
	newCuckooDefaultOutput, cuckooDispatch,
)

func newCuckooDefaultOutput(msg string) *CuckooWebOutput {
	return &CuckooWebOutput{
		Players:         make([]*CuckooWebOutputPlayer, 0),
		WinnerIdx:       -1,
		PendingSwapFrom: -1,
		PendingSwapTo:   -1,
		RoundLowest:     -1,
		RoundLosers:     make([]int, 0),
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func cuckooDispatch(bc *baseController, w http.ResponseWriter, ci usecase.CuckooInteractorIF, param CuckooWebInput, _ func(string) *CuckooWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ci.ResetWithConfig(param.ToConfig()))
	case "k", "keep":
		bc.writePresenterResponse(w, ci.Keep())
	case "s", "swap":
		bc.writePresenterResponse(w, ci.Swap())
	case "rf", "refuse":
		bc.writePresenterResponse(w, ci.Refuse())
	case "ac", "accept":
		bc.writePresenterResponse(w, ci.AcceptSwap())
	case "nr", "nextround":
		bc.writePresenterResponse(w, ci.NextRound())
	default:
		return dispatchLog(param.Command, bc, w, ci.ActionLog)
	}
	return true
}
