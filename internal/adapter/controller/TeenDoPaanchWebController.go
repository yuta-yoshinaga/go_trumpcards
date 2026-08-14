//go:build !js || !wasm || solo

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TeenDoPaanchWebInput 3-2-5 Webインプット
type TeenDoPaanchWebInput struct {
	BaseWebInput
	CardIndex *int                   `json:"cardIndex,omitempty"`
	Suit      *int                   `json:"suit,omitempty"`
	Config    *TeenDoPaanchWebConfig `json:"config,omitempty"`
}

// TeenDoPaanchWebConfig 3-2-5 Web設定
type TeenDoPaanchWebConfig struct {
	Rounds *int `json:"rounds,omitempty"`
}

// TeenDoPaanchWebOutputPlayer 3-2-5 Webアウトプットプレイヤー
type TeenDoPaanchWebOutputPlayer struct {
	ID        int              `json:"id"`
	IsHuman   bool             `json:"isHuman"`
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	// Target はこのラウンドのノルマ (3 / 2 / 5)。**宣言ではなく割り当て。**
	Target     int `json:"target"`
	TrickCount int `json:"trickCount"`
	// Met はノルマを達成したラウンド数。**勝敗はこれで決まる。**
	Met int `json:"met"`
}

// TeenDoPaanchWebOutputHint ヒント出力
type TeenDoPaanchWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	Reason    string `json:"reason"`
	// Suit は切り札に勧めるスート（プレイ中は 0）。
	Suit int `json:"suit"`
}

// TeenDoPaanchWebOutput 3-2-5 Webアウトプット
type TeenDoPaanchWebOutput struct {
	Players     []*TeenDoPaanchWebOutputPlayer `json:"players"`
	Phase       int                            `json:"phase"`
	RoundNumber int                            `json:"roundNumber"`
	TrickNumber int                            `json:"trickNumber"`
	// TrumpSuit は 0 のあいだ未宣言。決めるのは FivePlayerIdx の席だけ。
	TrumpSuit     int `json:"trumpSuit"`
	FivePlayerIdx int `json:"fivePlayerIdx"`
	// LastExchange は直前のラウンド間で動いた札の枚数。
	LastExchange     int                        `json:"lastExchange"`
	CurrentPlayerIdx int                        `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                        `json:"leadPlayerIdx"`
	CurrentTrick     []*WebOutputTrickCard      `json:"currentTrick"`
	ValidPlays       []int                      `json:"validPlays"`
	GameEndFlag      bool                       `json:"gameEndFlag"`
	WinnerIdx        int                        `json:"winnerIdx"`
	Hint             *TeenDoPaanchWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
	Config TeenDoPaanchWebOutputConfig `json:"config"`
}

// TeenDoPaanchWebOutputConfig 3-2-5 設定アウトプット
type TeenDoPaanchWebOutputConfig struct {
	Rounds int `json:"rounds"`
}

// ToConfig builds a TeenDoPaanchConfig from the nested web config, applying bounds checking.
func (c *TeenDoPaanchWebConfig) ToConfig() domain.TeenDoPaanchConfig {
	cfg := domain.DefaultTeenDoPaanchConfig()
	cfg.Rounds = webutil.BoundedIntPtr(c.Rounds,
		domain.TeenDoPaanchRoundsMin, domain.TeenDoPaanchRoundsMax, cfg.Rounds)
	return cfg
}

// ToConfig builds a TeenDoPaanchConfig from the web input.
func (p TeenDoPaanchWebInput) ToConfig() domain.TeenDoPaanchConfig {
	return configOrDefault(p.Config, (*TeenDoPaanchWebConfig).ToConfig, domain.DefaultTeenDoPaanchConfig())
}

// TeenDoPaanchWebController 3-2-5 Webコントローラークラス
type TeenDoPaanchWebController = GameWebController[usecase.TeenDoPaanchInteractorIF, TeenDoPaanchWebInput, *TeenDoPaanchWebOutput]

// NewTeenDoPaanchWebController and NewTeenDoPaanchWebControllerWithProvider are
// the standard and provider-backed constructors for TeenDoPaanchWebController.
var NewTeenDoPaanchWebController, NewTeenDoPaanchWebControllerWithProvider = webControllerPair[usecase.TeenDoPaanchInteractorIF, TeenDoPaanchWebInput, *TeenDoPaanchWebOutput](
	newTeenDoPaanchDefaultOutput, teenDoPaanchDispatch,
)

func newTeenDoPaanchDefaultOutput(msg string) *TeenDoPaanchWebOutput {
	return &TeenDoPaanchWebOutput{
		Players:       make([]*TeenDoPaanchWebOutputPlayer, 0),
		CurrentTrick:  make([]*WebOutputTrickCard, 0),
		ValidPlays:    make([]int, 0),
		WinnerIdx:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func teenDoPaanchDispatch(bc *baseController, w http.ResponseWriter, ti usecase.TeenDoPaanchInteractorIF, param TeenDoPaanchWebInput, newDefault func(string) *TeenDoPaanchWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ti.ResetWithConfig(param.ToConfig()))
	case "t", "trump":
		// **スート無しの宣言は通さない。** 既定値で埋めると選んでいない
		// スートがそのラウンドの切り札になる。
		if !requireParam(bc, w, newDefault, param.Suit == nil, "param error: suit is required.") {
			return true
		}
		bc.writePresenterResponse(w, ti.DeclareTrump(*param.Suit))
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
