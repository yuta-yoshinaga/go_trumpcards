//go:build !js || !wasm || classic

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ReversisWebInput レヴェルシWebインプット
type ReversisWebInput struct {
	BaseWebInput
	CardIndex *int               `json:"cardIndex,omitempty"`
	Config    *ReversisWebConfig `json:"config,omitempty"`
}

// ReversisWebConfig レヴェルシWeb設定
type ReversisWebConfig struct {
	Rounds *int `json:"rounds,omitempty"`
}

// ReversisWebOutputPlayer レヴェルシWebアウトプットプレイヤー
type ReversisWebOutputPlayer struct {
	ID        int              `json:"id"`
	IsHuman   bool             `json:"isHuman"`
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	// Chips は持ちチップ。**多いほど良い**（勝敗はこれで決まる）。
	Chips int `json:"chips"`
	// RoundPenalty はこのラウンドの失点。**少ないほど良い**（プールを取れる）。
	RoundPenalty   int  `json:"roundPenalty"`
	TrickCount     int  `json:"trickCount"`
	TookQuinola    bool `json:"tookQuinola"`
	TookDiamondAce bool `json:"tookDiamondAce"`
}

// ReversisWebOutputHint ヒント出力
type ReversisWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	Reason    string `json:"reason"`
}

// ReversisWebOutput レヴェルシWebアウトプット
type ReversisWebOutput struct {
	Players          []*ReversisWebOutputPlayer `json:"players"`
	Phase            int                        `json:"phase"`
	RoundNumber      int                        `json:"roundNumber"`
	TrickNumber      int                        `json:"trickNumber"`
	Pool             int                        `json:"pool"`
	CurrentPlayerIdx int                        `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                        `json:"leadPlayerIdx"`
	DealerIdx        int                        `json:"dealerIdx"`
	CurrentTrick     []*WebOutputTrickCard      `json:"currentTrick"`
	ValidPlays       []int                      `json:"validPlays"`
	GameEndFlag      bool                       `json:"gameEndFlag"`
	WinnerIdx        int                        `json:"winnerIdx"`
	Hint             *ReversisWebOutputHint     `json:"hint,omitempty"`
	WebOutputBase
	Config ReversisWebOutputConfig `json:"config"`
}

// ReversisWebOutputConfig レヴェルシ設定アウトプット
type ReversisWebOutputConfig struct {
	Rounds int `json:"rounds"`
}

// ToConfig builds a ReversisConfig from the nested web config, applying bounds checking.
func (c *ReversisWebConfig) ToConfig() domain.ReversisConfig {
	cfg := domain.DefaultReversisConfig()
	cfg.Rounds = webutil.BoundedIntPtr(c.Rounds,
		domain.ReversisRoundsMin, domain.ReversisRoundsMax, cfg.Rounds)
	return cfg
}

// ToConfig builds a ReversisConfig from the web input.
func (p ReversisWebInput) ToConfig() domain.ReversisConfig {
	return configOrDefault(p.Config, (*ReversisWebConfig).ToConfig, domain.DefaultReversisConfig())
}

// ReversisWebController レヴェルシWebコントローラークラス
type ReversisWebController = GameWebController[usecase.ReversisInteractorIF, ReversisWebInput, *ReversisWebOutput]

// NewReversisWebController and NewReversisWebControllerWithProvider are
// the standard and provider-backed constructors for ReversisWebController.
var NewReversisWebController, NewReversisWebControllerWithProvider = webControllerPair[usecase.ReversisInteractorIF, ReversisWebInput, *ReversisWebOutput](
	newReversisDefaultOutput, reversisDispatch,
)

func newReversisDefaultOutput(msg string) *ReversisWebOutput {
	return &ReversisWebOutput{
		Players:       make([]*ReversisWebOutputPlayer, 0),
		CurrentTrick:  make([]*WebOutputTrickCard, 0),
		ValidPlays:    make([]int, 0),
		WinnerIdx:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func reversisDispatch(bc *baseController, w http.ResponseWriter, ri usecase.ReversisInteractorIF, param ReversisWebInput, newDefault func(string) *ReversisWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ri.ResetWithConfig(param.ToConfig()))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, ri.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, ri.NextRound())
	case "g", "giveup":
		bc.writePresenterResponse(w, ri.GiveUp())
	default:
		return dispatchHintAndLog(param.Command, bc, w, ri.Hint, ri.ActionLog)
	}
	return true
}
