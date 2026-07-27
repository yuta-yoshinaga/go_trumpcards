//go:build !js || !wasm || casino

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// DoppelkopfWebInput ドッペルコップのWebインプット
type DoppelkopfWebInput struct {
	BaseWebInput
	// CardIndex プレイするカードのインデックス (play コマンド)
	CardIndex *int `json:"cardIndex,omitempty"`
	// Config ゲーム設定
	Config *DoppelkopfWebConfig `json:"config,omitempty"`
}

// DoppelkopfWebConfig ドッペルコップのWeb設定
type DoppelkopfWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	BaseChips     *int `json:"baseChips,omitempty"`
	StartChips    *int `json:"startChips,omitempty"`
	TargetChips   *int `json:"targetChips,omitempty"`
}

// DoppelkopfWebOutputPlayer ドッペルコップのWebアウトプットプレイヤー
type DoppelkopfWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	TrickCount int              `json:"trickCount"`
	Chips      int              `json:"chips"`
	// IsRe プレイヤーが Re チームかどうか (チーム公開後のみ true になりうる)
	IsRe bool `json:"isRe"`
}

// DoppelkopfWebOutput ドッペルコップのWebアウトプット
type DoppelkopfWebOutput struct {
	Players          []*DoppelkopfWebOutputPlayer `json:"players"`
	Phase            int                          `json:"phase"`
	RoundNumber      int                          `json:"roundNumber"`
	TrickNumber      int                          `json:"trickNumber"`
	CurrentPlayerIdx int                          `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                          `json:"leadPlayerIdx"`
	DealerIdx        int                          `json:"dealerIdx"`
	CurrentTrick     []*WebOutputTrickCard        `json:"currentTrick"`
	// ReTeam 各プレイヤーの Re チーム所属 (チーム公開後のみ true になりうる; 4要素)
	ReTeam          []bool             `json:"reTeam"`
	SoloRe          bool               `json:"soloRe"`
	TeamsRevealed   bool               `json:"teamsRevealed"`
	ReAnnounced     bool               `json:"reAnnounced"`
	KontraAnnounced bool               `json:"kontraAnnounced"`
	CanAnnounce     bool               `json:"canAnnounce"`
	YouAreRe        bool               `json:"youAreRe"`
	PlayableIndices []int              `json:"playableIndices"`
	RoundRePoints   int                `json:"roundRePoints"`
	RoundReWon      bool               `json:"roundReWon"`
	RoundGamePoints int                `json:"roundGamePoints"`
	GameEndFlag     bool               `json:"gameEndFlag"`
	WinnerIdx       int                `json:"winnerIdx"`
	Hint            *WebOutputCardHint `json:"hint,omitempty"`
	WebOutputBase
	Config DoppelkopfWebOutputConfig `json:"config"`
}

// DoppelkopfWebOutputConfig ドッペルコップの設定アウトプット
type DoppelkopfWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	BaseChips     int `json:"baseChips"`
	StartChips    int `json:"startChips"`
	TargetChips   int `json:"targetChips"`
}

// ToConfig builds a DoppelkopfConfig from the nested web config, applying bounds checking.
func (c *DoppelkopfWebConfig) ToConfig() domain.DoppelkopfConfig {
	cfg := domain.DefaultDoppelkopfConfig()
	cfg.CpuDifficulty = domain.DoppelkopfCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.DoppelkopfCpuDifficultyEasy), int(domain.DoppelkopfCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.BaseChips, c.BaseChips, 1, 100000)
	webutil.ApplyBoundedInt(&cfg.StartChips, c.StartChips, 1, 100000)
	webutil.ApplyBoundedInt(&cfg.TargetChips, c.TargetChips, cfg.StartChips+1, 1000000)
	return cfg
}

// ToConfig builds a DoppelkopfConfig from the web input.
func (p DoppelkopfWebInput) ToConfig() domain.DoppelkopfConfig {
	return configOrDefault(p.Config, (*DoppelkopfWebConfig).ToConfig, domain.DefaultDoppelkopfConfig())
}

// DoppelkopfWebController ドッペルコップのWebコントローラークラス
type DoppelkopfWebController = GameWebController[usecase.DoppelkopfInteractorIF, DoppelkopfWebInput, *DoppelkopfWebOutput]

// NewDoppelkopfWebController and NewDoppelkopfWebControllerWithProvider are
// the standard and provider-backed constructors for DoppelkopfWebController.
var NewDoppelkopfWebController, NewDoppelkopfWebControllerWithProvider = webControllerPair[usecase.DoppelkopfInteractorIF, DoppelkopfWebInput, *DoppelkopfWebOutput](
	newDoppelkopfDefaultOutput, doppelkopfDispatch,
)

func newDoppelkopfDefaultOutput(msg string) *DoppelkopfWebOutput {
	return &DoppelkopfWebOutput{
		Players:         make([]*DoppelkopfWebOutputPlayer, 0),
		CurrentTrick:    make([]*WebOutputTrickCard, 0),
		ReTeam:          make([]bool, domain.DoppelkopfPlayerCnt),
		PlayableIndices: make([]int, 0),
		WinnerIdx:       -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func doppelkopfDispatch(bc *baseController, w http.ResponseWriter, di usecase.DoppelkopfInteractorIF, param DoppelkopfWebInput, newDefault func(string) *DoppelkopfWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, di.ResetWithConfig(param.ToConfig()))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.Play(*param.CardIndex))
	case "a", "announce":
		bc.writePresenterResponse(w, di.Announce())
	case "n", "next":
		bc.writePresenterResponse(w, di.NextTrick())
	case "nr", "nextround":
		bc.writePresenterResponse(w, di.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, di.Hint, di.ActionLog)
	}
	return true
}
