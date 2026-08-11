//go:build !js || !wasm || extra2

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BalootWebInput バルートWebインプット
type BalootWebInput struct {
	BaseWebInput
	CardIndex *int             `json:"cardIndex,omitempty"`
	Suit      *int             `json:"suit,omitempty"`
	Config    *BalootWebConfig `json:"config,omitempty"`
}

// BalootWebConfig バルートWeb設定
type BalootWebConfig struct {
	Target *int `json:"target,omitempty"`
}

// BalootWebOutputPlayer バルートWebアウトプットプレイヤー
type BalootWebOutputPlayer struct {
	ID        int              `json:"id"`
	IsHuman   bool             `json:"isHuman"`
	Team      int              `json:"team"`
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	// HasBaloot は Hokom で切り札の K+Q を持っているか (20点)。
	HasBaloot  bool `json:"hasBaloot"`
	Declared   bool `json:"declared"`
	TrickCount int  `json:"trickCount"`
}

// BalootWebOutputHint ヒント出力
type BalootWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	Reason    string `json:"reason"`
	Suit      int    `json:"suit"`
}

// BalootWebOutput バルートWebアウトプット
type BalootWebOutput struct {
	Players     []*BalootWebOutputPlayer `json:"players"`
	Phase       int                      `json:"phase"`
	Mode        int                      `json:"mode"`
	RoundNumber int                      `json:"roundNumber"`
	TrickNumber int                      `json:"trickNumber"`
	// TrumpSuit は Hokom のときだけ意味を持つ (Sun では 0)。
	TrumpSuit        int `json:"trumpSuit"`
	DeclarerIdx      int `json:"declarerIdx"`
	CurrentPlayerIdx int `json:"currentPlayerIdx"`
	LeadPlayerIdx    int `json:"leadPlayerIdx"`
	DealerIdx        int `json:"dealerIdx"`
	// Scores は累計、RoundPoints は現ラウンド。どちらもチーム単位。
	Scores       []int                 `json:"scores"`
	RoundPoints  []int                 `json:"roundPoints"`
	CurrentTrick []*WebOutputTrickCard `json:"currentTrick"`
	ValidPlays   []int                 `json:"validPlays"`
	GameEndFlag  bool                  `json:"gameEndFlag"`
	WinnerTeam   int                   `json:"winnerTeam"`
	Hint         *BalootWebOutputHint  `json:"hint,omitempty"`
	WebOutputBase
	Config BalootWebOutputConfig `json:"config"`
}

// BalootWebOutputConfig バルート設定アウトプット
type BalootWebOutputConfig struct {
	Target int `json:"target"`
}

// ToConfig builds a BalootConfig from the nested web config, applying bounds checking.
func (c *BalootWebConfig) ToConfig() domain.BalootConfig {
	cfg := domain.DefaultBalootConfig()
	cfg.Target = webutil.BoundedIntPtr(c.Target,
		domain.BalootTargetMin, domain.BalootTargetMax, cfg.Target)
	return cfg
}

// ToConfig builds a BalootConfig from the web input.
func (p BalootWebInput) ToConfig() domain.BalootConfig {
	return configOrDefault(p.Config, (*BalootWebConfig).ToConfig, domain.DefaultBalootConfig())
}

// BalootWebController バルートWebコントローラークラス
type BalootWebController = GameWebController[usecase.BalootInteractorIF, BalootWebInput, *BalootWebOutput]

// NewBalootWebController and NewBalootWebControllerWithProvider are
// the standard and provider-backed constructors for BalootWebController.
var NewBalootWebController, NewBalootWebControllerWithProvider = webControllerPair[usecase.BalootInteractorIF, BalootWebInput, *BalootWebOutput](
	newBalootDefaultOutput, balootDispatch,
)

func newBalootDefaultOutput(msg string) *BalootWebOutput {
	return &BalootWebOutput{
		Players:       make([]*BalootWebOutputPlayer, 0),
		CurrentTrick:  make([]*WebOutputTrickCard, 0),
		ValidPlays:    make([]int, 0),
		Scores:        make([]int, 0),
		RoundPoints:   make([]int, 0),
		DeclarerIdx:   -1,
		WinnerTeam:    -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func balootDispatch(bc *baseController, w http.ResponseWriter, bi usecase.BalootInteractorIF, param BalootWebInput, newDefault func(string) *BalootWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, bi.ResetWithConfig(param.ToConfig()))
	case "sun":
		bc.writePresenterResponse(w, bi.DeclareSun())
	case "hokom":
		// **切り札のスートは Hokom のときだけ必要。** 既定値で埋めると
		// プレイヤーが選んでいないスートが切り札になってしまう。
		if !requireParam(bc, w, newDefault, param.Suit == nil, "param error: suit is required.") {
			return true
		}
		bc.writePresenterResponse(w, bi.DeclareHokom(*param.Suit))
	case "pass":
		bc.writePresenterResponse(w, bi.PassDeclaration())
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, bi.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, bi.NextRound())
	case "g", "giveup":
		bc.writePresenterResponse(w, bi.GiveUp())
	default:
		return dispatchHintAndLog(param.Command, bc, w, bi.Hint, bi.ActionLog)
	}
	return true
}
