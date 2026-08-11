//go:build !js || !wasm || extra3

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TarabishWebInput タラビッシュWebインプット
type TarabishWebInput struct {
	BaseWebInput
	CardIndex *int               `json:"cardIndex,omitempty"`
	Config    *TarabishWebConfig `json:"config,omitempty"`
}

// TarabishWebConfig タラビッシュWeb設定
type TarabishWebConfig struct {
	Target *int `json:"target,omitempty"`
}

// TarabishWebOutputPlayer タラビッシュWebアウトプットプレイヤー
type TarabishWebOutputPlayer struct {
	ID        int              `json:"id"`
	IsHuman   bool             `json:"isHuman"`
	Team      int              `json:"team"`
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	// MeldPoints は配り札から自動判定したメルド点、RunLen は最長ランの枚数。
	MeldPoints int  `json:"meldPoints"`
	RunLen     int  `json:"runLen"`
	HasBella   bool `json:"hasBella"`
	TrickCount int  `json:"trickCount"`
}

// TarabishWebOutputHint ヒント出力
type TarabishWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	Reason    string `json:"reason"`
}

// TarabishWebOutput タラビッシュWebアウトプット
type TarabishWebOutput struct {
	Players          []*TarabishWebOutputPlayer `json:"players"`
	Phase            int                        `json:"phase"`
	RoundNumber      int                        `json:"roundNumber"`
	TrickNumber      int                        `json:"trickNumber"`
	TrumpSuit        int                        `json:"trumpSuit"`
	UpCard           *WebOutputCard             `json:"upCard,omitempty"`
	TrumpTakerIdx    int                        `json:"trumpTakerIdx"`
	CurrentPlayerIdx int                        `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                        `json:"leadPlayerIdx"`
	DealerIdx        int                        `json:"dealerIdx"`
	// Scores は累計、RoundPoints は現ラウンド。どちらもチーム単位。
	Scores       []int                  `json:"scores"`
	RoundPoints  []int                  `json:"roundPoints"`
	CurrentTrick []*WebOutputTrickCard  `json:"currentTrick"`
	ValidPlays   []int                  `json:"validPlays"`
	GameEndFlag  bool                   `json:"gameEndFlag"`
	WinnerTeam   int                    `json:"winnerTeam"`
	Hint         *TarabishWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
	Config TarabishWebOutputConfig `json:"config"`
}

// TarabishWebOutputConfig タラビッシュ設定アウトプット
type TarabishWebOutputConfig struct {
	Target int `json:"target"`
}

// ToConfig builds a TarabishConfig from the nested web config, applying bounds checking.
func (c *TarabishWebConfig) ToConfig() domain.TarabishConfig {
	cfg := domain.DefaultTarabishConfig()
	cfg.Target = webutil.BoundedIntPtr(c.Target,
		domain.TarabishTargetMin, domain.TarabishTargetMax, cfg.Target)
	return cfg
}

// ToConfig builds a TarabishConfig from the web input.
func (p TarabishWebInput) ToConfig() domain.TarabishConfig {
	return configOrDefault(p.Config, (*TarabishWebConfig).ToConfig, domain.DefaultTarabishConfig())
}

// TarabishWebController タラビッシュWebコントローラークラス
type TarabishWebController = GameWebController[usecase.TarabishInteractorIF, TarabishWebInput, *TarabishWebOutput]

// NewTarabishWebController and NewTarabishWebControllerWithProvider are
// the standard and provider-backed constructors for TarabishWebController.
var NewTarabishWebController, NewTarabishWebControllerWithProvider = webControllerPair[usecase.TarabishInteractorIF, TarabishWebInput, *TarabishWebOutput](
	newTarabishDefaultOutput, tarabishDispatch,
)

func newTarabishDefaultOutput(msg string) *TarabishWebOutput {
	return &TarabishWebOutput{
		Players:       make([]*TarabishWebOutputPlayer, 0),
		CurrentTrick:  make([]*WebOutputTrickCard, 0),
		ValidPlays:    make([]int, 0),
		Scores:        make([]int, 0),
		RoundPoints:   make([]int, 0),
		TrumpTakerIdx: -1,
		WinnerTeam:    -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func tarabishDispatch(bc *baseController, w http.ResponseWriter, ti usecase.TarabishInteractorIF, param TarabishWebInput, newDefault func(string) *TarabishWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ti.ResetWithConfig(param.ToConfig()))
	case "t", "take":
		bc.writePresenterResponse(w, ti.TakeTrump())
	case "pass":
		bc.writePresenterResponse(w, ti.PassTrump())
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, ti.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, ti.NextRound())
	case "g", "giveup":
		bc.writePresenterResponse(w, ti.GiveUp())
	default:
		return dispatchHintAndLog(param.Command, bc, w, ti.Hint, ti.ActionLog)
	}
	return true
}
