//go:build !js || !wasm || extra3

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// LooWebConfig はルー (Loo) の Web 設定。
type LooWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	Ante          *int `json:"ante,omitempty"`
}

// ToConfig は LooWebConfig を domain.LooConfig に変換する (境界チェック付き)。
func (c *LooWebConfig) ToConfig() domain.LooConfig {
	cfg := domain.DefaultLooConfig()
	cfg.CpuDifficulty = domain.LooCpuDifficulty(webutil.BoundedIntPtr(
		c.CpuDifficulty, int(domain.LooCpuDifficultyEasy), int(domain.LooCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.Ante, c.Ante, 1, 1000000)
	return cfg
}

// LooWebInput はルー Web インプット。
type LooWebInput struct {
	BaseWebInput
	CardIndex *int          `json:"cardIndex,omitempty"`
	Play      *bool         `json:"play,omitempty"`
	Config    *LooWebConfig `json:"config,omitempty"`
}

// ToConfig は Web インプットから domain.LooConfig を構築する。
func (p LooWebInput) ToConfig() domain.LooConfig {
	return configOrDefault(p.Config, (*LooWebConfig).ToConfig, domain.DefaultLooConfig())
}

// LooWebOutputPlayer はルー Web アウトプットプレイヤー。
type LooWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	TrickCount int              `json:"trickCount"`
	Playing    bool             `json:"playing"`
	Chips      int              `json:"chips"`
}

// LooWebOutputDealDetail は 1 ディールの精算内訳。
type LooWebOutputDealDetail struct {
	PotStart  int         `json:"potStart"`
	TrumpSuit int         `json:"trumpSuit"`
	Playing   []bool      `json:"playing"`
	Tricks    map[int]int `json:"tricks"`
	Gained    map[int]int `json:"gained"`
	Looed     []int       `json:"looed"`
	PotCarry  int         `json:"potCarry"`
}

// LooWebOutputHint はヒント出力。
type LooWebOutputHint struct {
	CardIndices []int  `json:"cardIndices"`
	Decision    *bool  `json:"decision,omitempty"`
	Reason      string `json:"reason"`
}

// LooWebOutput はルー Web アウトプット。
type LooWebOutput struct {
	Players         []*LooWebOutputPlayer   `json:"players"`
	Phase           int                     `json:"phase"`
	RoundNumber     int                     `json:"roundNumber"`
	TrickNumber     int                     `json:"trickNumber"`
	TotalTricks     int                     `json:"totalTricks"`
	DealerIdx       int                     `json:"dealerIdx"`
	CurrentTurn     int                     `json:"currentTurn"`
	DecidePlayerIdx int                     `json:"decidePlayerIdx"`
	TrumpSuit       int                     `json:"trumpSuit"`
	TurnUp          *WebOutputCard          `json:"turnUp,omitempty"`
	Pot             int                     `json:"pot"`
	PotStart        int                     `json:"potStart"`
	CurrentTrick    []*WebOutputTrickCard   `json:"currentTrick"`
	LastTrick       []*WebOutputTrickCard   `json:"lastTrick"`
	LastTrickWinner int                     `json:"lastTrickWinner"`
	PlayableIndices []int                   `json:"playableIndices"`
	GameEndFlag     bool                    `json:"gameEndFlag"`
	LastDealDetail  *LooWebOutputDealDetail `json:"lastDealDetail"`
	IsHumanTurn     bool                    `json:"isHumanTurn"`
	Hint            *LooWebOutputHint       `json:"hint,omitempty"`
	WebOutputBase
	Config LooWebConfigOutput `json:"config"`
}

// LooWebConfigOutput は設定アウトプット。
type LooWebConfigOutput struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	Ante          int `json:"ante"`
}

// LooWebController はルー Web コントローラークラス。
type LooWebController = GameWebController[usecase.LooInteractorIF, LooWebInput, *LooWebOutput]

// NewLooWebController, NewLooWebControllerWithProvider are the standard and
// provider-backed constructors for LooWebController.
var NewLooWebController, NewLooWebControllerWithProvider = webControllerPair[usecase.LooInteractorIF, LooWebInput, *LooWebOutput](
	newLooDefaultOutput, looDispatch,
)

func newLooDefaultOutput(msg string) *LooWebOutput {
	return &LooWebOutput{
		Players:         make([]*LooWebOutputPlayer, 0),
		CurrentTrick:    make([]*WebOutputTrickCard, 0),
		LastTrick:       make([]*WebOutputTrickCard, 0),
		PlayableIndices: make([]int, 0),
		TotalTricks:     domain.LooTrickCount,
		LastTrickWinner: -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func looDispatch(bc *baseController, w http.ResponseWriter, li usecase.LooInteractorIF, param LooWebInput, newDefault func(string) *LooWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, li.ResetWithConfig(param.ToConfig()))
	case "d", "decide":
		if !requireParam(bc, w, newDefault, param.Play == nil, "param error: play is required.") {
			return true
		}
		bc.writePresenterResponse(w, li.Decide(*param.Play))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, li.Play(*param.CardIndex))
	case "nr", "nextround", "n", "next":
		bc.writePresenterResponse(w, li.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, li.Hint, li.ActionLog)
	}
	return true
}
