//go:build !js || !wasm || casino

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// FortyFivesWebInput オークション・フォーティファイブズのWebインプット
type FortyFivesWebInput struct {
	BaseWebInput
	// Bid 入札種別 (0=Pass,15,20,25)
	Bid *int `json:"bid,omitempty"`
	// CardIndex プレイするカードのインデックス (play コマンド)
	CardIndex *int `json:"cardIndex,omitempty"`
	// Config ゲーム設定
	Config *FortyFivesWebConfig `json:"config,omitempty"`
}

// FortyFivesWebConfig オークション・フォーティファイブズのWeb設定
type FortyFivesWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetPoints  *int `json:"targetPoints,omitempty"`
}

// FortyFivesWebOutputPlayer オークション・フォーティファイブズのWebアウトプットプレイヤー
type FortyFivesWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	TrickCount int              `json:"trickCount"`
	TeamScore  int              `json:"teamScore"`
	IsDeclarer bool             `json:"isDeclarer"`
}

// FortyFivesWebOutput オークション・フォーティファイブズのWebアウトプット
type FortyFivesWebOutput struct {
	Players          []*FortyFivesWebOutputPlayer    `json:"players"`
	Phase            int                             `json:"phase"`
	RoundNumber      int                             `json:"roundNumber"`
	TrickNumber      int                             `json:"trickNumber"`
	CurrentPlayerIdx int                             `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                             `json:"leadPlayerIdx"`
	DealerIdx        int                             `json:"dealerIdx"`
	DeclarerIdx      int                             `json:"declarerIdx"`
	Contract         int                             `json:"contract"`
	TrumpSuit        int                             `json:"trumpSuit"`
	Bids             [domain.FortyFivesPlayerCnt]int `json:"bids"`
	CurrentTrick     []*WebOutputTrickCard           `json:"currentTrick"`
	TeamScores       [domain.FortyFivesTeamCnt]int   `json:"teamScores"`
	RoundTeamPoints  [domain.FortyFivesTeamCnt]int   `json:"roundTeamPoints"`
	PlayableIndices  []int                           `json:"playableIndices"`
	GameEndFlag      bool                            `json:"gameEndFlag"`
	WinnerTeam       int                             `json:"winnerTeam"`
	IsHumanTurn      bool                            `json:"isHumanTurn"`
	IsHumanBidTurn   bool                            `json:"isHumanBidTurn"`
	Hint             *WebOutputCardHint              `json:"hint,omitempty"`
	WebOutputBase
	Config FortyFivesWebOutputConfig `json:"config"`
}

// FortyFivesWebOutputConfig オークション・フォーティファイブズの設定アウトプット
type FortyFivesWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetPoints  int `json:"targetPoints"`
}

// ToConfig builds a FortyFivesConfig from the nested web config, applying bounds checking.
func (c *FortyFivesWebConfig) ToConfig() domain.FortyFivesConfig {
	cfg := domain.DefaultFortyFivesConfig()
	cfg.CpuDifficulty = domain.FortyFivesCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.FortyFivesCpuDifficultyEasy), int(domain.FortyFivesCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetPoints, c.TargetPoints, 1, 1000000)
	return cfg
}

// ToConfig builds a FortyFivesConfig from the web input.
func (p FortyFivesWebInput) ToConfig() domain.FortyFivesConfig {
	return configOrDefault(p.Config, (*FortyFivesWebConfig).ToConfig, domain.DefaultFortyFivesConfig())
}

// FortyFivesWebController オークション・フォーティファイブズのWebコントローラークラス
type FortyFivesWebController = GameWebController[usecase.FortyFivesInteractorIF, FortyFivesWebInput, *FortyFivesWebOutput]

// NewFortyFivesWebController and NewFortyFivesWebControllerWithProvider are
// the standard and provider-backed constructors for FortyFivesWebController.
var NewFortyFivesWebController, NewFortyFivesWebControllerWithProvider = webControllerPair[usecase.FortyFivesInteractorIF, FortyFivesWebInput, *FortyFivesWebOutput](
	newFortyFivesDefaultOutput, fortyFivesDispatch,
)

func newFortyFivesDefaultOutput(msg string) *FortyFivesWebOutput {
	return &FortyFivesWebOutput{
		Players:         make([]*FortyFivesWebOutputPlayer, 0),
		CurrentTrick:    make([]*WebOutputTrickCard, 0),
		PlayableIndices: make([]int, 0),
		DeclarerIdx:     -1,
		WinnerTeam:      -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func fortyFivesDispatch(bc *baseController, w http.ResponseWriter, di usecase.FortyFivesInteractorIF, param FortyFivesWebInput, newDefault func(string) *FortyFivesWebOutput) bool {
	return dispatchBidTrickPlay(param.Command, bc, w, bidTrickPlayFns{
		resetWithConfig: func() string { return di.ResetWithConfig(param.ToConfig()) },
		bid:             di.Bid,
		play:            di.Play,
		nextTrick:       di.NextTrick,
		nextRound:       di.NextRound,
		hint:            di.Hint,
		actionLog:       di.ActionLog,
	}, param.Bid, param.CardIndex, newDefault)
}
