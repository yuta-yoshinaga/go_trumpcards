//go:build !js || !wasm || extra3

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PopeJoanWebInput ポープ・ジョーンWebインプット
type PopeJoanWebInput struct {
	BaseWebInput
	CardIndex *int               `json:"cardIndex,omitempty"`
	Config    *PopeJoanWebConfig `json:"config,omitempty"`
}

// PopeJoanWebConfig ポープ・ジョーンWeb設定
type PopeJoanWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetDeals   *int `json:"targetDeals,omitempty"`
}

// PopeJoanWebOutputCompartment は盤の 1 区画。
type PopeJoanWebOutputCompartment struct {
	// Name は "ace" / "game" / "pope" / "matrimony" / "intrigue" など。
	Name string `json:"name"`
	// Chips は残高。**取られなかった区画は持ち越される**ので貯まっていく。
	Chips int `json:"chips"`
}

// PopeJoanWebOutputAward は区画が誰にいくら渡ったかの記録。
type PopeJoanWebOutputAward struct {
	Compartment string `json:"compartment"`
	Player      int    `json:"player"`
	Chips       int    `json:"chips"`
	// ByTurnUp はめくり札でディーラーが即座に取ったか。
	ByTurnUp bool `json:"byTurnUp"`
}

// PopeJoanWebOutputPlayer ポープ・ジョーンWebアウトプットプレイヤー
type PopeJoanWebOutputPlayer struct {
	ID      int  `json:"id"`
	IsHuman bool `json:"isHuman"`
	// CardCount は手札の枚数。**残り 1 枚につき 1 チップ払う**ので公開する。
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	Chips     int              `json:"chips"`
	// HoldsPope は Pope を抱えているか。**支払いを免除される**ので、伏せると
	// 精算が理解できなくなる。
	HoldsPope bool `json:"holdsPope"`
	Hidden    bool `json:"hidden"`
}

// PopeJoanWebOutputHint ヒント出力
type PopeJoanWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	Reason    string `json:"reason"`
}

// PopeJoanWebOutput ポープ・ジョーンWebアウトプット
type PopeJoanWebOutput struct {
	Players          []*PopeJoanWebOutputPlayer      `json:"players"`
	Phase            int                             `json:"phase"`
	CurrentPlayerIdx int                             `json:"currentPlayerIdx"`
	Compartments     []*PopeJoanWebOutputCompartment `json:"compartments"`
	// TrumpSuit は dead hand の最後の 1 枚で決まる。区画はこのスートでしか払わない。
	TrumpSuit int `json:"trumpSuit"`
	// TurnUp はそのめくり札。Pope/A/K/Q/J ならディーラーが即座に取っている。
	TurnUp *WebOutputCard `json:"turnUp,omitempty"`
	// Awards はこのディールで区画が動いた記録。
	Awards     []*PopeJoanWebOutputAward `json:"awards"`
	PlayedPile []*WebOutputCard          `json:"playedPile"`
	// RunSuit は今の並びのスート (-1: 好きな札で始められる)。
	RunSuit int `json:"runSuit"`
	// RunRank は今の並びの最高ランク (A は 14 として数える)。
	RunRank     int                    `json:"runRank"`
	DealNo      int                    `json:"dealNo"`
	TargetDeals int                    `json:"targetDeals"`
	DealWinner  int                    `json:"dealWinner"`
	GameEndFlag bool                   `json:"gameEndFlag"`
	WinnerIdx   int                    `json:"winnerIdx"`
	Hint        *PopeJoanWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
	Config PopeJoanWebOutputConfig `json:"config"`
}

// PopeJoanWebOutputConfig ポープ・ジョーン設定アウトプット
type PopeJoanWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetDeals   int `json:"targetDeals"`
}

// ToConfig builds a PopeJoanConfig from the nested web config, applying bounds checking.
func (c *PopeJoanWebConfig) ToConfig() domain.PopeJoanConfig {
	cfg := domain.DefaultPopeJoanConfig()
	cfg.CpuDifficulty = domain.PopeJoanCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty,
		int(domain.PopeJoanCpuDifficultyNormal), int(domain.PopeJoanCpuDifficultyNormal),
		int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetDeals, c.TargetDeals, 1, 100)
	return cfg
}

// ToConfig builds a PopeJoanConfig from the input, falling back to defaults when absent.
//
// Must go through configOrDefault: `config` is optional on the wire, so a plain
// reset arrives with a nil *PopeJoanWebConfig and calling the method on it would
// dereference nil.
func (i PopeJoanWebInput) ToConfig() domain.PopeJoanConfig {
	return configOrDefault(i.Config, (*PopeJoanWebConfig).ToConfig, domain.DefaultPopeJoanConfig())
}

// PopeJoanWebController ポープ・ジョーンWebコントローラ
type PopeJoanWebController = GameWebController[usecase.PopeJoanInteractorIF, PopeJoanWebInput, *PopeJoanWebOutput]

// NewPopeJoanWebController and NewPopeJoanWebControllerWithProvider are the
// standard and provider-backed constructors for PopeJoanWebController.
var NewPopeJoanWebController, NewPopeJoanWebControllerWithProvider = webControllerPair[usecase.PopeJoanInteractorIF, PopeJoanWebInput, *PopeJoanWebOutput](
	newPopeJoanDefaultOutput, popeJoanDispatch,
)

func newPopeJoanDefaultOutput(msg string) *PopeJoanWebOutput {
	return &PopeJoanWebOutput{
		Players:       make([]*PopeJoanWebOutputPlayer, 0),
		Compartments:  make([]*PopeJoanWebOutputCompartment, 0),
		Awards:        make([]*PopeJoanWebOutputAward, 0),
		PlayedPile:    make([]*WebOutputCard, 0),
		TargetDeals:   domain.DefaultPopeJoanConfig().TargetDeals,
		TrumpSuit:     -1,
		RunSuit:       -1,
		DealWinner:    -1,
		WinnerIdx:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func popeJoanDispatch(bc *baseController, w http.ResponseWriter, pi usecase.PopeJoanInteractorIF, param PopeJoanWebInput, newDefault func(string) *PopeJoanWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, pi.ResetWithConfig(param.ToConfig()))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, pi.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, pi.NextDeal())
	default:
		return dispatchHintAndLog(param.Command, bc, w, pi.Hint, pi.ActionLog)
	}
	return true
}

// NewPopeJoanDefaultOutputForTest exposes the default-output builder to the
// external controller_test package.
func NewPopeJoanDefaultOutputForTest(msg string) *PopeJoanWebOutput {
	return newPopeJoanDefaultOutput(msg)
}
