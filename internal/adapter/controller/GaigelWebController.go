//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// GaigelWebInput ガイゲルWebインプット
type GaigelWebInput struct {
	BaseWebInput
	CardIndex *int             `json:"cardIndex,omitempty"`
	Config    *GaigelWebConfig `json:"config,omitempty"`
}

// GaigelWebConfig ガイゲルWeb設定
type GaigelWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetScore   *int `json:"targetScore,omitempty"`
}

// GaigelWebOutputPlayer ガイゲルWebアウトプットプレイヤー
type GaigelWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	Team       int              `json:"team"`
	TrickCount int              `json:"trickCount"`
}

// GaigelWebOutputHint ヒント出力
type GaigelWebOutputHint struct {
	CardIndex  *int   `json:"cardIndex,omitempty"`
	Reason     string `json:"reason"`
	IsMarriage bool   `json:"isMarriage"`
}

// GaigelWebOutput ガイゲルWebアウトプット
type GaigelWebOutput struct {
	Players          []*GaigelWebOutputPlayer `json:"players"`
	Phase            int                      `json:"phase"`
	RoundNumber      int                      `json:"roundNumber"`
	TrickNumber      int                      `json:"trickNumber"`
	CurrentPlayerIdx int                      `json:"currentPlayerIdx"`
	DealerIdx        int                      `json:"dealerIdx"`
	TrumpSuit        int                      `json:"trumpSuit"`
	TrumpCard        *WebOutputCard           `json:"trumpCard,omitempty"`
	StockRemaining   int                      `json:"stockRemaining"`
	CurrentTrick     []*WebOutputTrickCard    `json:"currentTrick"`
	TeamScores       [2]int                   `json:"teamScores"`
	RoundPoints      [2]int                   `json:"roundPoints"`
	RoundMarriage    [2]int                   `json:"roundMarriage"`
	MarriageIndices  []int                    `json:"marriageIndices"`
	GameEndFlag      bool                     `json:"gameEndFlag"`
	WinnerTeam       int                      `json:"winnerTeam"`
	LeadPlayerIdx    int                      `json:"leadPlayerIdx"`
	Hint             *GaigelWebOutputHint     `json:"hint,omitempty"`
	WebOutputBase
	Config GaigelWebOutputConfig `json:"config"`
}

// GaigelWebOutputConfig ガイゲル設定アウトプット
type GaigelWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetScore   int `json:"targetScore"`
}

// ToConfig builds a GaigelConfig from the nested web config, applying bounds checking.
func (c *GaigelWebConfig) ToConfig() domain.GaigelConfig {
	cfg := domain.DefaultGaigelConfig()
	cfg.CpuDifficulty = domain.GaigelCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.GaigelCpuDifficultyEasy), int(domain.GaigelCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetScore, c.TargetScore, 1, 10000)
	return cfg
}

// ToConfig builds a GaigelConfig from the web input.
func (p GaigelWebInput) ToConfig() domain.GaigelConfig {
	return configOrDefault(p.Config, (*GaigelWebConfig).ToConfig, domain.DefaultGaigelConfig())
}

// GaigelWebController ガイゲルWebコントローラークラス
type GaigelWebController = GameWebController[usecase.GaigelInteractorIF, GaigelWebInput, *GaigelWebOutput]

// NewGaigelWebController and NewGaigelWebControllerWithProvider are
// the standard and provider-backed constructors for GaigelWebController.
var NewGaigelWebController, NewGaigelWebControllerWithProvider = webControllerPair[usecase.GaigelInteractorIF, GaigelWebInput, *GaigelWebOutput](
	newGaigelDefaultOutput, gaigelDispatch,
)

func newGaigelDefaultOutput(msg string) *GaigelWebOutput {
	return &GaigelWebOutput{
		Players:         make([]*GaigelWebOutputPlayer, 0),
		CurrentTrick:    make([]*WebOutputTrickCard, 0),
		MarriageIndices: make([]int, 0),
		WinnerTeam:      -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func gaigelDispatch(bc *baseController, w http.ResponseWriter, gi usecase.GaigelInteractorIF, param GaigelWebInput, newDefault func(string) *GaigelWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, gi.ResetWithConfig(param.ToConfig()))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, gi.Play(*param.CardIndex))
	case "m", "marriage":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, gi.DeclareMarriage(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, gi.NextTrick())
	case "nr", "nextround":
		bc.writePresenterResponse(w, gi.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, gi.Hint, gi.ActionLog)
	}
	return true
}
