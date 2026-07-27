//go:build !js || !wasm || casino

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TwentyNineWebInput トゥエンティナイン (29) のWebインプット
type TwentyNineWebInput struct {
	BaseWebInput
	// Bid 入札種別 (0=Pass,16,20,24,28)
	Bid *int `json:"bid,omitempty"`
	// CardIndex プレイするカードのインデックス (play コマンド)
	CardIndex *int `json:"cardIndex,omitempty"`
	// Config ゲーム設定
	Config *TwentyNineWebConfig `json:"config,omitempty"`
}

// TwentyNineWebConfig トゥエンティナイン (29) のWeb設定
type TwentyNineWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetPoints  *int `json:"targetPoints,omitempty"`
}

// TwentyNineWebOutputPlayer トゥエンティナイン (29) のWebアウトプットプレイヤー
type TwentyNineWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	TrickCount int              `json:"trickCount"`
	TeamScore  int              `json:"teamScore"`
	IsDeclarer bool             `json:"isDeclarer"`
}

// TwentyNineWebOutput トゥエンティナイン (29) のWebアウトプット
type TwentyNineWebOutput struct {
	Players          []*TwentyNineWebOutputPlayer    `json:"players"`
	Phase            int                             `json:"phase"`
	RoundNumber      int                             `json:"roundNumber"`
	TrickNumber      int                             `json:"trickNumber"`
	CurrentPlayerIdx int                             `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                             `json:"leadPlayerIdx"`
	DealerIdx        int                             `json:"dealerIdx"`
	DeclarerIdx      int                             `json:"declarerIdx"`
	Contract         int                             `json:"contract"`
	TrumpSuit        int                             `json:"trumpSuit"`
	TrumpRevealed    bool                            `json:"trumpRevealed"`
	Bids             [domain.TwentyNinePlayerCnt]int `json:"bids"`
	CurrentTrick     []*WebOutputTrickCard           `json:"currentTrick"`
	TeamScores       [domain.TwentyNineTeamCnt]int   `json:"teamScores"`
	RoundTeamPoints  [domain.TwentyNineTeamCnt]int   `json:"roundTeamPoints"`
	PlayableIndices  []int                           `json:"playableIndices"`
	GameEndFlag      bool                            `json:"gameEndFlag"`
	WinnerTeam       int                             `json:"winnerTeam"`
	IsHumanTurn      bool                            `json:"isHumanTurn"`
	IsHumanBidTurn   bool                            `json:"isHumanBidTurn"`
	Hint             *WebOutputCardHint              `json:"hint,omitempty"`
	WebOutputBase
	Config TwentyNineWebOutputConfig `json:"config"`
}

// TwentyNineWebOutputConfig トゥエンティナイン (29) の設定アウトプット
type TwentyNineWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetPoints  int `json:"targetPoints"`
}

// ToConfig builds a TwentyNineConfig from the nested web config, applying bounds checking.
func (c *TwentyNineWebConfig) ToConfig() domain.TwentyNineConfig {
	cfg := domain.DefaultTwentyNineConfig()
	cfg.CpuDifficulty = domain.TwentyNineCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.TwentyNineCpuDifficultyEasy), int(domain.TwentyNineCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetPoints, c.TargetPoints, 1, 1000000)
	return cfg
}

// ToConfig builds a TwentyNineConfig from the web input.
func (p TwentyNineWebInput) ToConfig() domain.TwentyNineConfig {
	return configOrDefault(p.Config, (*TwentyNineWebConfig).ToConfig, domain.DefaultTwentyNineConfig())
}

// TwentyNineWebController トゥエンティナイン (29) のWebコントローラークラス
type TwentyNineWebController = GameWebController[usecase.TwentyNineInteractorIF, TwentyNineWebInput, *TwentyNineWebOutput]

// NewTwentyNineWebController and NewTwentyNineWebControllerWithProvider are
// the standard and provider-backed constructors for TwentyNineWebController.
var NewTwentyNineWebController, NewTwentyNineWebControllerWithProvider = webControllerPair[usecase.TwentyNineInteractorIF, TwentyNineWebInput, *TwentyNineWebOutput](
	newTwentyNineDefaultOutput, twentyNineDispatch,
)

func newTwentyNineDefaultOutput(msg string) *TwentyNineWebOutput {
	return &TwentyNineWebOutput{
		Players:         make([]*TwentyNineWebOutputPlayer, 0),
		CurrentTrick:    make([]*WebOutputTrickCard, 0),
		PlayableIndices: make([]int, 0),
		DeclarerIdx:     -1,
		WinnerTeam:      -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func twentyNineDispatch(bc *baseController, w http.ResponseWriter, di usecase.TwentyNineInteractorIF, param TwentyNineWebInput, newDefault func(string) *TwentyNineWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, di.ResetWithConfig(param.ToConfig()))
	case "b", "bid":
		if !requireParam(bc, w, newDefault, param.Bid == nil, "param error: bid is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.Bid(*param.Bid))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, di.NextTrick())
	case "nr", "nextround":
		bc.writePresenterResponse(w, di.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, di.Hint, di.ActionLog)
	}
	return true
}
