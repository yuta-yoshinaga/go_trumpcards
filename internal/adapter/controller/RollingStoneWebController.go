//go:build !js || !wasm || extra3

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// RollingStoneWebInput ローリングストーンWebインプット
type RollingStoneWebInput struct {
	BaseWebInput
	CardIndex *int                   `json:"cardIndex,omitempty"`
	Config    *RollingStoneWebConfig `json:"config,omitempty"`
}

// RollingStoneWebConfig ローリングストーンWeb設定
type RollingStoneWebConfig struct {
	PlayerCnt *int `json:"playerCnt,omitempty"`
}

// RollingStoneWebOutputPlayer ローリングストーンWebアウトプットプレイヤー
type RollingStoneWebOutputPlayer struct {
	ID        int              `json:"id"`
	IsHuman   bool             `json:"isHuman"`
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	// Pickups は引き取った回数。**多いほど不利。**
	Pickups int `json:"pickups"`
	// FinishedAt は上がった順位（0 = まだ）。
	FinishedAt int `json:"finishedAt"`
}

// RollingStoneWebOutputHint ヒント出力
type RollingStoneWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	Reason    string `json:"reason"`
}

// RollingStoneWebOutput ローリングストーンWebアウトプット
type RollingStoneWebOutput struct {
	Players []*RollingStoneWebOutputPlayer `json:"players"`
	Phase   int                            `json:"phase"`
	// MustPickUp は人間がフォローできず引き取るしかないか。
	//
	// **これが真なら出せる札は無い。** ページは validPlays を見る前にこれを見ます。
	MustPickUp       bool                       `json:"mustPickUp"`
	ValidPlays       []int                      `json:"validPlays"`
	CurrentTrick     []*WebOutputTrickCard      `json:"currentTrick"`
	CurrentPlayerIdx int                        `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                        `json:"leadPlayerIdx"`
	TrickNumber      int                        `json:"trickNumber"`
	LastPickupIdx    int                        `json:"lastPickupIdx"`
	FinishedCnt      int                        `json:"finishedCnt"`
	DeckSize         int                        `json:"deckSize"`
	Discarded        int                        `json:"discarded"`
	GameEndFlag      bool                       `json:"gameEndFlag"`
	WinnerIdx        int                        `json:"winnerIdx"`
	Hint             *RollingStoneWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
	Config RollingStoneWebOutputConfig `json:"config"`
}

// RollingStoneWebOutputConfig ローリングストーン設定アウトプット
type RollingStoneWebOutputConfig struct {
	PlayerCnt int `json:"playerCnt"`
}

// ToConfig builds a RollingStoneConfig from the nested web config, applying bounds checking.
func (c *RollingStoneWebConfig) ToConfig() domain.RollingStoneConfig {
	cfg := domain.DefaultRollingStoneConfig()
	cfg.PlayerCnt = webutil.BoundedIntPtr(c.PlayerCnt,
		domain.RollingStonePlayerCntMin, domain.RollingStonePlayerCntMax, cfg.PlayerCnt)
	return cfg
}

// ToConfig builds a RollingStoneConfig from the web input.
func (p RollingStoneWebInput) ToConfig() domain.RollingStoneConfig {
	return configOrDefault(p.Config, (*RollingStoneWebConfig).ToConfig, domain.DefaultRollingStoneConfig())
}

// RollingStoneWebController ローリングストーンWebコントローラークラス
type RollingStoneWebController = GameWebController[usecase.RollingStoneInteractorIF, RollingStoneWebInput, *RollingStoneWebOutput]

// NewRollingStoneWebController and NewRollingStoneWebControllerWithProvider are
// the standard and provider-backed constructors for RollingStoneWebController.
var NewRollingStoneWebController, NewRollingStoneWebControllerWithProvider = webControllerPair[usecase.RollingStoneInteractorIF, RollingStoneWebInput, *RollingStoneWebOutput](
	newRollingStoneDefaultOutput, rollingStoneDispatch,
)

func newRollingStoneDefaultOutput(msg string) *RollingStoneWebOutput {
	return &RollingStoneWebOutput{
		Players:       make([]*RollingStoneWebOutputPlayer, 0),
		ValidPlays:    make([]int, 0),
		CurrentTrick:  make([]*WebOutputTrickCard, 0),
		LastPickupIdx: -1,
		WinnerIdx:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func rollingStoneDispatch(bc *baseController, w http.ResponseWriter, ri usecase.RollingStoneInteractorIF, param RollingStoneWebInput, newDefault func(string) *RollingStoneWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ri.ResetWithConfig(param.ToConfig()))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, ri.Play(*param.CardIndex))
	case "u", "pickup":
		// **引き取りは別のコマンド。** カード指定が無いのは省略ではなく、
		// 「出せる札が無い」という別の行動だからです。
		bc.writePresenterResponse(w, ri.PickUp())
	case "g", "giveup":
		bc.writePresenterResponse(w, ri.GiveUp())
	default:
		return dispatchHintAndLog(param.Command, bc, w, ri.Hint, ri.ActionLog)
	}
	return true
}
