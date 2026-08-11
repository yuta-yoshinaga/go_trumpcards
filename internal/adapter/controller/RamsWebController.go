//go:build !js || !wasm || extra2

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// RamsWebInput ラムスWebインプット
type RamsWebInput struct {
	BaseWebInput
	CardIndex *int           `json:"cardIndex,omitempty"`
	Config    *RamsWebConfig `json:"config,omitempty"`
}

// RamsWebConfig ラムスWeb設定
type RamsWebConfig struct {
	// PlayerCnt は 3〜5 人。**ラムスは可変人数が特徴。**
	PlayerCnt *int `json:"playerCnt,omitempty"`
	Rounds    *int `json:"rounds,omitempty"`
}

// RamsWebOutputPlayer ラムスWebアウトプットプレイヤー
type RamsWebOutputPlayer struct {
	ID        int              `json:"id"`
	IsHuman   bool             `json:"isHuman"`
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	// Chips は持ちチップ。**多いほど良い。**
	Chips int `json:"chips"`
	// InRound はこのラウンドに参加しているか、Decided は選び終えたか。
	InRound     bool `json:"inRound"`
	Decided     bool `json:"decided"`
	RoundTricks int  `json:"roundTricks"`
	TrickCount  int  `json:"trickCount"`
}

// RamsWebOutputHint ヒント出力
type RamsWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	Reason    string `json:"reason"`
}

// RamsWebOutput ラムスWebアウトプット
type RamsWebOutput struct {
	Players          []*RamsWebOutputPlayer `json:"players"`
	Phase            int                    `json:"phase"`
	RoundNumber      int                    `json:"roundNumber"`
	TrickNumber      int                    `json:"trickNumber"`
	Pot              int                    `json:"pot"`
	TrumpSuit        int                    `json:"trumpSuit"`
	UpCard           *WebOutputCard         `json:"upCard,omitempty"`
	CurrentPlayerIdx int                    `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                    `json:"leadPlayerIdx"`
	DealerIdx        int                    `json:"dealerIdx"`
	ActiveCount      int                    `json:"activeCount"`
	CurrentTrick     []*WebOutputTrickCard  `json:"currentTrick"`
	ValidPlays       []int                  `json:"validPlays"`
	GameEndFlag      bool                   `json:"gameEndFlag"`
	WinnerIdx        int                    `json:"winnerIdx"`
	Hint             *RamsWebOutputHint     `json:"hint,omitempty"`
	WebOutputBase
	Config RamsWebOutputConfig `json:"config"`
}

// RamsWebOutputConfig ラムス設定アウトプット
type RamsWebOutputConfig struct {
	PlayerCnt int `json:"playerCnt"`
	Rounds    int `json:"rounds"`
}

// ToConfig builds a RamsConfig from the nested web config, applying bounds checking.
func (c *RamsWebConfig) ToConfig() domain.RamsConfig {
	cfg := domain.DefaultRamsConfig()
	cfg.PlayerCnt = webutil.BoundedIntPtr(c.PlayerCnt,
		domain.RamsPlayerCntMin, domain.RamsPlayerCntMax, cfg.PlayerCnt)
	cfg.Rounds = webutil.BoundedIntPtr(c.Rounds,
		domain.RamsRoundsMin, domain.RamsRoundsMax, cfg.Rounds)
	return cfg
}

// ToConfig builds a RamsConfig from the web input.
func (p RamsWebInput) ToConfig() domain.RamsConfig {
	return configOrDefault(p.Config, (*RamsWebConfig).ToConfig, domain.DefaultRamsConfig())
}

// RamsWebController ラムスWebコントローラークラス
type RamsWebController = GameWebController[usecase.RamsInteractorIF, RamsWebInput, *RamsWebOutput]

// NewRamsWebController and NewRamsWebControllerWithProvider are
// the standard and provider-backed constructors for RamsWebController.
var NewRamsWebController, NewRamsWebControllerWithProvider = webControllerPair[usecase.RamsInteractorIF, RamsWebInput, *RamsWebOutput](
	newRamsDefaultOutput, ramsDispatch,
)

func newRamsDefaultOutput(msg string) *RamsWebOutput {
	return &RamsWebOutput{
		Players:       make([]*RamsWebOutputPlayer, 0),
		CurrentTrick:  make([]*WebOutputTrickCard, 0),
		ValidPlays:    make([]int, 0),
		WinnerIdx:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func ramsDispatch(bc *baseController, w http.ResponseWriter, ri usecase.RamsInteractorIF, param RamsWebInput, newDefault func(string) *RamsWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ri.ResetWithConfig(param.ToConfig()))
	case "in", "play":
		bc.writePresenterResponse(w, ri.Play())
	case "out", "pass":
		bc.writePresenterResponse(w, ri.Pass())
	case "c", "card":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, ri.PlayCard(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, ri.NextRound())
	case "g", "giveup":
		bc.writePresenterResponse(w, ri.GiveUp())
	default:
		return dispatchHintAndLog(param.Command, bc, w, ri.Hint, ri.ActionLog)
	}
	return true
}
