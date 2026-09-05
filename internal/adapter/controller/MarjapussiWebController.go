//go:build !js || !wasm || extra5

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// MarjapussiWebInput マルヤプッシ (Marjapussi) のWebインプット
type MarjapussiWebInput struct {
	BaseWebInput
	// CardIndex プレイするカードのインデックス
	CardIndex *int `json:"cardIndex,omitempty"`
	// Config ゲーム設定
	Config *MarjapussiWebConfig `json:"config,omitempty"`
}

// MarjapussiWebConfig マルヤプッシのWeb設定
type MarjapussiWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetPoints  *int `json:"targetPoints,omitempty"`
}

// MarjapussiWebOutputPlayer マルヤプッシのWebアウトプットプレイヤー
type MarjapussiWebOutputPlayer struct {
	ID         int              `json:"id"`
	TeamID     int              `json:"teamId"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	TrickCount int              `json:"trickCount"`
	Score      int              `json:"score"`
}

// MarjapussiWebOutput マルヤプッシのWebアウトプット
type MarjapussiWebOutput struct {
	Players          []*MarjapussiWebOutputPlayer    `json:"players"`
	Phase            int                             `json:"phase"`
	RoundNumber      int                             `json:"roundNumber"`
	TrickNumber      int                             `json:"trickNumber"`
	CurrentPlayerIdx int                             `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                             `json:"leadPlayerIdx"`
	DealerIdx        int                             `json:"dealerIdx"`
	TrumpSuit        int                             `json:"trumpSuit"`
	CurrentTrick     []*WebOutputTrickCard           `json:"currentTrick"`
	TeamScores       [domain.MarjapussiTeamCnt]int   `json:"teamScores"`
	PlayerScores     [domain.MarjapussiPlayerCnt]int `json:"playerScores"`
	RoundCardPoints  [domain.MarjapussiTeamCnt]int   `json:"roundCardPoints"`
	RoundMarriage    [domain.MarjapussiTeamCnt]int   `json:"roundMarriage"`
	PussiCount       int                             `json:"pussiCount"`
	Pussi            []*WebOutputCard                `json:"pussi,omitempty"`
	PussiWinnerTeam  int                             `json:"pussiWinnerTeam"`
	LastTrickWinner  int                             `json:"lastTrickWinner"`
	PlayableIndices  []int                           `json:"playableIndices"`
	GameEndFlag      bool                            `json:"gameEndFlag"`
	WinnerPlayer     int                             `json:"winnerPlayer"`
	WinnerTeam       int                             `json:"winnerTeam"`
	IsHumanTurn      bool                            `json:"isHumanTurn"`
	Hint             *WebOutputCardHint              `json:"hint,omitempty"`
	WebOutputBase
	Config MarjapussiWebOutputConfig `json:"config"`
}

// MarjapussiWebOutputConfig マルヤプッシの設定アウトプット
type MarjapussiWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetPoints  int `json:"targetPoints"`
}

// ToConfig builds a MarjapussiConfig from the nested web config, applying bounds checking.
func (c *MarjapussiWebConfig) ToConfig() domain.MarjapussiConfig {
	cfg := domain.DefaultMarjapussiConfig()
	cfg.CpuDifficulty = domain.MarjapussiCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.MarjapussiCpuDifficultyEasy), int(domain.MarjapussiCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.PointLimit, c.TargetPoints, 1, 1000000)
	return cfg
}

// ToConfig builds a MarjapussiConfig from the web input.
func (p MarjapussiWebInput) ToConfig() domain.MarjapussiConfig {
	return configOrDefault(p.Config, (*MarjapussiWebConfig).ToConfig, domain.DefaultMarjapussiConfig())
}

// MarjapussiWebController マルヤプッシのWebコントローラークラス
type MarjapussiWebController = GameWebController[usecase.MarjapussiInteractorIF, MarjapussiWebInput, *MarjapussiWebOutput]

// NewMarjapussiWebController and NewMarjapussiWebControllerWithProvider are
// the standard and provider-backed constructors for MarjapussiWebController.
var NewMarjapussiWebController, NewMarjapussiWebControllerWithProvider = webControllerPair[usecase.MarjapussiInteractorIF, MarjapussiWebInput, *MarjapussiWebOutput](
	newMarjapussiDefaultOutput, marjapussiDispatch,
)

func newMarjapussiDefaultOutput(msg string) *MarjapussiWebOutput {
	return &MarjapussiWebOutput{
		Players:         make([]*MarjapussiWebOutputPlayer, 0),
		CurrentTrick:    make([]*WebOutputTrickCard, 0),
		PlayableIndices: make([]int, 0),
		PussiWinnerTeam: -1,
		LastTrickWinner: -1,
		WinnerPlayer:    -1,
		WinnerTeam:      -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func marjapussiDispatch(bc *baseController, w http.ResponseWriter, di usecase.MarjapussiInteractorIF, param MarjapussiWebInput, newDefault func(string) *MarjapussiWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, di.ResetWithConfig(param.ToConfig()))
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
