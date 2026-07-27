//go:build !js || !wasm || solo

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SchnapsenWebInput シュナプセンWebインプット
type SchnapsenWebInput struct {
	BaseWebInput
	CardIndex *int                `json:"cardIndex,omitempty"`
	Config    *SchnapsenWebConfig `json:"config,omitempty"`
}

// SchnapsenWebConfig シュナプセンWeb設定
type SchnapsenWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
}

// SchnapsenWebOutputPlayer シュナプセンWebアウトプットプレイヤー
type SchnapsenWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	Points     int              `json:"points"`
	TrickCount int              `json:"trickCount"`
}

// SchnapsenWebOutputHint ヒント出力
type SchnapsenWebOutputHint struct {
	CardIndex  *int   `json:"cardIndex,omitempty"`
	Reason     string `json:"reason"`
	IsMarriage bool   `json:"isMarriage"`
}

// SchnapsenWebOutput シュナプセンWebアウトプット
type SchnapsenWebOutput struct {
	Players          []*SchnapsenWebOutputPlayer `json:"players"`
	Phase            int                         `json:"phase"`
	TrickNumber      int                         `json:"trickNumber"`
	CurrentPlayerIdx int                         `json:"currentPlayerIdx"`
	CurrentTrick     []*WebOutputTrickCard       `json:"currentTrick"`
	TrumpSuit        int                         `json:"trumpSuit"`
	TrumpCard        *WebOutputCard              `json:"trumpCard,omitempty"`
	DealerIdx        int                         `json:"dealerIdx"`
	LeadPlayerIdx    int                         `json:"leadPlayerIdx"`
	StockRemaining   int                         `json:"stockRemaining"`
	IsEndgame        bool                        `json:"isEndgame"`
	ValidPlays       []int                       `json:"validPlays"`
	MarriagePlays    []int                       `json:"marriagePlays"`
	GameEndFlag      bool                        `json:"gameEndFlag"`
	WinnerIdx        int                         `json:"winnerIdx"`
	Hint             *SchnapsenWebOutputHint     `json:"hint,omitempty"`
	WebOutputBase
	Config SchnapsenWebOutputConfig `json:"config"`
}

// SchnapsenWebOutputConfig シュナプセン設定アウトプット
type SchnapsenWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
}

// ToConfig builds a SchnapsenConfig from the nested web config, applying bounds checking.
func (c *SchnapsenWebConfig) ToConfig() domain.SchnapsenConfig {
	cfg := domain.DefaultSchnapsenConfig()
	cfg.CpuDifficulty = domain.SchnapsenCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty,
		int(domain.SchnapsenCpuDifficultyNormal), int(domain.SchnapsenCpuDifficultyNormal),
		int(cfg.CpuDifficulty)))
	return cfg
}

// ToConfig builds a SchnapsenConfig from the web input.
func (p SchnapsenWebInput) ToConfig() domain.SchnapsenConfig {
	return configOrDefault(p.Config, (*SchnapsenWebConfig).ToConfig, domain.DefaultSchnapsenConfig())
}

// SchnapsenWebController シュナプセンWebコントローラークラス
type SchnapsenWebController = GameWebController[usecase.SchnapsenInteractorIF, SchnapsenWebInput, *SchnapsenWebOutput]

// NewSchnapsenWebController and NewSchnapsenWebControllerWithProvider are
// the standard and provider-backed constructors for SchnapsenWebController.
var NewSchnapsenWebController, NewSchnapsenWebControllerWithProvider = webControllerPair[usecase.SchnapsenInteractorIF, SchnapsenWebInput, *SchnapsenWebOutput](
	newSchnapsenDefaultOutput, schnapsenDispatch,
)

func newSchnapsenDefaultOutput(msg string) *SchnapsenWebOutput {
	return &SchnapsenWebOutput{
		Players:       make([]*SchnapsenWebOutputPlayer, 0),
		CurrentTrick:  make([]*WebOutputTrickCard, 0),
		ValidPlays:    make([]int, 0),
		MarriagePlays: make([]int, 0),
		WinnerIdx:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func schnapsenDispatch(bc *baseController, w http.ResponseWriter, si usecase.SchnapsenInteractorIF, param SchnapsenWebInput, newDefault func(string) *SchnapsenWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, si.ResetWithConfig(param.ToConfig()))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, si.Play(*param.CardIndex))
	case "m", "marriage":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, si.DeclareMarriage(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, si.NextTrick())
	default:
		return dispatchHintAndLog(param.Command, bc, w, si.Hint, si.ActionLog)
	}
	return true
}
