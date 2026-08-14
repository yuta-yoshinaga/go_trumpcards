//go:build !js || !wasm || extra2

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PolignacWebInput ポリニャックWebインプット
type PolignacWebInput struct {
	BaseWebInput
	CardIndex *int               `json:"cardIndex,omitempty"`
	Config    *PolignacWebConfig `json:"config,omitempty"`
}

// PolignacWebConfig ポリニャックWeb設定
type PolignacWebConfig struct {
	Rounds *int `json:"rounds,omitempty"`
}

// PolignacWebOutputPlayer ポリニャックWebアウトプットプレイヤー
type PolignacWebOutputPlayer struct {
	ID        int              `json:"id"`
	IsHuman   bool             `json:"isHuman"`
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	// Score は全ラウンドの累計**失点**。小さいほど良い。
	Score int `json:"score"`
	// RoundPenalty はこのラウンドで受けた失点。
	RoundPenalty  int  `json:"roundPenalty"`
	TrickCount    int  `json:"trickCount"`
	DeclaredCapot bool `json:"declaredCapot"`
}

// PolignacWebOutputHint ヒント出力
type PolignacWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	Reason    string `json:"reason"`
}

// PolignacWebOutput ポリニャックWebアウトプット
type PolignacWebOutput struct {
	Players          []*PolignacWebOutputPlayer `json:"players"`
	Phase            int                        `json:"phase"`
	RoundNumber      int                        `json:"roundNumber"`
	TrickNumber      int                        `json:"trickNumber"`
	CurrentPlayerIdx int                        `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                        `json:"leadPlayerIdx"`
	DealerIdx        int                        `json:"dealerIdx"`
	// CapotIdx は capot 宣言者 (-1: 宣言なし)、CapotTricks はその獲得トリック数。
	CapotIdx     int                    `json:"capotIdx"`
	CapotTricks  int                    `json:"capotTricks"`
	CurrentTrick []*WebOutputTrickCard  `json:"currentTrick"`
	ValidPlays   []int                  `json:"validPlays"`
	GameEndFlag  bool                   `json:"gameEndFlag"`
	WinnerIdx    int                    `json:"winnerIdx"`
	Hint         *PolignacWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
	Config PolignacWebOutputConfig `json:"config"`
}

// PolignacWebOutputConfig ポリニャック設定アウトプット
type PolignacWebOutputConfig struct {
	Rounds int `json:"rounds"`
}

// ToConfig builds a PolignacConfig from the nested web config, applying bounds checking.
func (c *PolignacWebConfig) ToConfig() domain.PolignacConfig {
	cfg := domain.DefaultPolignacConfig()
	cfg.Rounds = webutil.BoundedIntPtr(c.Rounds,
		domain.PolignacRoundsMin, domain.PolignacRoundsMax, cfg.Rounds)
	return cfg
}

// ToConfig builds a PolignacConfig from the web input.
func (p PolignacWebInput) ToConfig() domain.PolignacConfig {
	return configOrDefault(p.Config, (*PolignacWebConfig).ToConfig, domain.DefaultPolignacConfig())
}

// PolignacWebController ポリニャックWebコントローラークラス
type PolignacWebController = GameWebController[usecase.PolignacInteractorIF, PolignacWebInput, *PolignacWebOutput]

// NewPolignacWebController and NewPolignacWebControllerWithProvider are
// the standard and provider-backed constructors for PolignacWebController.
var NewPolignacWebController, NewPolignacWebControllerWithProvider = webControllerPair[usecase.PolignacInteractorIF, PolignacWebInput, *PolignacWebOutput](
	newPolignacDefaultOutput, polignacDispatch,
)

func newPolignacDefaultOutput(msg string) *PolignacWebOutput {
	return &PolignacWebOutput{
		Players:       make([]*PolignacWebOutputPlayer, 0),
		CurrentTrick:  make([]*WebOutputTrickCard, 0),
		ValidPlays:    make([]int, 0),
		CapotIdx:      -1,
		WinnerIdx:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func polignacDispatch(bc *baseController, w http.ResponseWriter, pi usecase.PolignacInteractorIF, param PolignacWebInput, newDefault func(string) *PolignacWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, pi.ResetWithConfig(param.ToConfig()))
	case "c", "capot":
		bc.writePresenterResponse(w, pi.DeclareCapot())
	case "pass":
		bc.writePresenterResponse(w, pi.Pass())
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, pi.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, pi.NextRound())
	case "g", "giveup":
		bc.writePresenterResponse(w, pi.GiveUp())
	default:
		return dispatchHintAndLog(param.Command, bc, w, pi.Hint, pi.ActionLog)
	}
	return true
}
