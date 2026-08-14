//go:build !js || !wasm || classic

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SlobberhannesWebInput スロバーハンネスWebインプット
type SlobberhannesWebInput struct {
	BaseWebInput
	CardIndex *int                    `json:"cardIndex,omitempty"`
	Config    *SlobberhannesWebConfig `json:"config,omitempty"`
}

// SlobberhannesWebConfig スロバーハンネスWeb設定
type SlobberhannesWebConfig struct {
	Rounds *int `json:"rounds,omitempty"`
}

// SlobberhannesWebOutputPlayer スロバーハンネスWebアウトプットプレイヤー
type SlobberhannesWebOutputPlayer struct {
	ID        int              `json:"id"`
	IsHuman   bool             `json:"isHuman"`
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	// Score は全ラウンドの累計。罰点が負、全回避ボーナスが正。
	Score      int `json:"score"`
	TrickCount int `json:"trickCount"`
	// このラウンドで受けた罰の内訳。3 つとも false なら +1。
	TookFirstTrick bool `json:"tookFirstTrick"`
	TookLastTrick  bool `json:"tookLastTrick"`
	TookQueen      bool `json:"tookQueen"`
}

// SlobberhannesWebOutputHint ヒント出力
type SlobberhannesWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	Reason    string `json:"reason"`
}

// SlobberhannesWebOutput スロバーハンネスWebアウトプット
type SlobberhannesWebOutput struct {
	Players          []*SlobberhannesWebOutputPlayer `json:"players"`
	Phase            int                             `json:"phase"`
	RoundNumber      int                             `json:"roundNumber"`
	TrickNumber      int                             `json:"trickNumber"`
	CurrentPlayerIdx int                             `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                             `json:"leadPlayerIdx"`
	DealerIdx        int                             `json:"dealerIdx"`
	CurrentTrick     []*WebOutputTrickCard           `json:"currentTrick"`
	ValidPlays       []int                           `json:"validPlays"`
	GameEndFlag      bool                            `json:"gameEndFlag"`
	WinnerIdx        int                             `json:"winnerIdx"`
	Hint             *SlobberhannesWebOutputHint     `json:"hint,omitempty"`
	WebOutputBase
	Config SlobberhannesWebOutputConfig `json:"config"`
}

// SlobberhannesWebOutputConfig スロバーハンネス設定アウトプット
type SlobberhannesWebOutputConfig struct {
	Rounds int `json:"rounds"`
}

// ToConfig builds a SlobberhannesConfig from the nested web config, applying bounds checking.
func (c *SlobberhannesWebConfig) ToConfig() domain.SlobberhannesConfig {
	cfg := domain.DefaultSlobberhannesConfig()
	cfg.Rounds = webutil.BoundedIntPtr(c.Rounds,
		domain.SlobberhannesRoundsMin, domain.SlobberhannesRoundsMax, cfg.Rounds)
	return cfg
}

// ToConfig builds a SlobberhannesConfig from the web input.
func (p SlobberhannesWebInput) ToConfig() domain.SlobberhannesConfig {
	return configOrDefault(p.Config, (*SlobberhannesWebConfig).ToConfig, domain.DefaultSlobberhannesConfig())
}

// SlobberhannesWebController スロバーハンネスWebコントローラークラス
type SlobberhannesWebController = GameWebController[usecase.SlobberhannesInteractorIF, SlobberhannesWebInput, *SlobberhannesWebOutput]

// NewSlobberhannesWebController and NewSlobberhannesWebControllerWithProvider are
// the standard and provider-backed constructors for SlobberhannesWebController.
var NewSlobberhannesWebController, NewSlobberhannesWebControllerWithProvider = webControllerPair[usecase.SlobberhannesInteractorIF, SlobberhannesWebInput, *SlobberhannesWebOutput](
	newSlobberhannesDefaultOutput, slobberhannesDispatch,
)

func newSlobberhannesDefaultOutput(msg string) *SlobberhannesWebOutput {
	return &SlobberhannesWebOutput{
		Players:       make([]*SlobberhannesWebOutputPlayer, 0),
		CurrentTrick:  make([]*WebOutputTrickCard, 0),
		ValidPlays:    make([]int, 0),
		WinnerIdx:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func slobberhannesDispatch(bc *baseController, w http.ResponseWriter, si usecase.SlobberhannesInteractorIF, param SlobberhannesWebInput, newDefault func(string) *SlobberhannesWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, si.ResetWithConfig(param.ToConfig()))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, si.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, si.NextRound())
	case "g", "giveup":
		bc.writePresenterResponse(w, si.GiveUp())
	default:
		return dispatchHintAndLog(param.Command, bc, w, si.Hint, si.ActionLog)
	}
	return true
}
