//go:build !js || !wasm || extra3

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// RamschWebInput Ramsch Web input.
type RamschWebInput struct {
	BaseWebInput
	Accept    *bool            `json:"accept,omitempty"`
	Pickup    *bool            `json:"pickup,omitempty"`
	DiscardA  *int             `json:"discardA,omitempty"`
	DiscardB  *int             `json:"discardB,omitempty"`
	GameType  *int             `json:"gameType,omitempty"`
	TrumpSuit *int             `json:"trumpSuit,omitempty"`
	CardIndex *int             `json:"cardIndex,omitempty"`
	Config    *RamschWebConfig `json:"config,omitempty"`
}

// RamschWebConfig Ramsch web configuration.
type RamschWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetScore   *int `json:"targetScore,omitempty"`
}

// RamschWebOutputPlayer Ramsch web output player.
type RamschWebOutputPlayer struct {
	ID        int              `json:"id"`
	IsHuman   bool             `json:"isHuman"`
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	// CardPoints はこのラウンドで集めた点。**多いほど不利。**
	CardPoints      int `json:"cardPoints"`
	RoundsWon       int `json:"roundsWon"`
	RoundsLost      int `json:"roundsLost"`
	RoundScore      int `json:"roundScore"`
	CumulativeScore int `json:"cumulativeScore"`
	TrickCount      int `json:"trickCount"`
}

// RamschWebOutputHint hint output. 助言する局面はプレイの手番だけ。
type RamschWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	Reason    string `json:"reason"`
}

// RamschWebOutput Ramsch web response payload.
type RamschWebOutput struct {
	Players          []*RamschWebOutputPlayer `json:"players"`
	Phase            int                      `json:"phase"`
	RoundNumber      int                      `json:"roundNumber"`
	TrickNumber      int                      `json:"trickNumber"`
	CurrentPlayerIdx int                      `json:"currentPlayerIdx"`
	CurrentTrick     []*WebOutputTrickCard    `json:"currentTrick"`
	ForehandIdx      int                      `json:"forehandIdx"`
	MiddlehandIdx    int                      `json:"middlehandIdx"`
	RearhandIdx      int                      `json:"rearhandIdx"`
	DealerIdx        int                      `json:"dealerIdx"`
	// Skat は伏せてある 2 枚。**最終トリックの獲得者が受け取る**ので、
	// ラウンド終了までは伏せたまま (中身は返さない)。
	Skat []*WebOutputCard `json:"skat,omitempty"`
	// LoserIdx は最も点を集めてしまったプレイヤー。同点・Durchmarsch・
	// ラウンド途中は -1。
	LoserIdx int `json:"loserIdx"`
	// Durchmarsch は 1 人が全トリックを取ったか（逆転勝ち）。
	Durchmarsch bool `json:"durchmarsch"`
	// DurchmarschIdx は総取りしたプレイヤー（無ければ -1）。
	DurchmarschIdx int                  `json:"durchmarschIdx"`
	GameEndFlag    bool                 `json:"gameEndFlag"`
	LeadPlayerIdx  int                  `json:"leadPlayerIdx"`
	Hint           *RamschWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
	Config RamschWebOutputConfig `json:"config"`
}

// RamschWebOutputConfig Ramsch configuration in the response.
type RamschWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetScore   int `json:"targetScore"`
}

// ToConfig builds a RamschConfig from the nested web config, applying bounds.
func (c *RamschWebConfig) ToConfig() domain.RamschConfig {
	cfg := domain.DefaultRamschConfig()
	cfg.CpuDifficulty = domain.RamschCpuDifficulty(webutil.BoundedIntPtr(
		c.CpuDifficulty,
		int(domain.RamschCpuDifficultyEasy),
		int(domain.RamschCpuDifficultyHard),
		int(cfg.CpuDifficulty),
	))
	webutil.ApplyBoundedInt(&cfg.TargetScore, c.TargetScore, 1, 10000)
	return cfg
}

// ToConfig builds a RamschConfig from the web input.
func (p RamschWebInput) ToConfig() domain.RamschConfig {
	return configOrDefault(p.Config, (*RamschWebConfig).ToConfig, domain.DefaultRamschConfig())
}

// RamschWebController Ramsch web controller.
type RamschWebController = GameWebController[usecase.RamschInteractorIF, RamschWebInput, *RamschWebOutput]

// NewRamschWebController, NewRamschWebControllerWithProvider standard constructors.
var NewRamschWebController, NewRamschWebControllerWithProvider = webControllerPair[usecase.RamschInteractorIF, RamschWebInput, *RamschWebOutput](
	newRamschDefaultOutput, ramschDispatch,
)

func newRamschDefaultOutput(msg string) *RamschWebOutput {
	return &RamschWebOutput{
		Players:        make([]*RamschWebOutputPlayer, 0),
		CurrentTrick:   make([]*WebOutputTrickCard, 0),
		LoserIdx:       -1,
		DurchmarschIdx: -1,
		WebOutputBase:  WebOutputBase{Message: msg},
	}
}

func ramschDispatch(bc *baseController, w http.ResponseWriter, si usecase.RamschInteractorIF, param RamschWebInput, newDefault func(string) *RamschWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, si.ResetWithConfig(param.ToConfig()))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, si.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, si.NextTrick())
	case "nr", "nextround":
		bc.writePresenterResponse(w, si.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, si.Hint, si.ActionLog)
	}
	return true
}
