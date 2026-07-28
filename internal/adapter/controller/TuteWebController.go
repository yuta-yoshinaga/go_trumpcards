//go:build !js || !wasm || casino

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TuteWebInput トゥーテのWebインプット
type TuteWebInput struct {
	BaseWebInput
	// CardIndex プレイするカードのインデックス (play コマンド)
	CardIndex *int `json:"cardIndex,omitempty"`
	// Suit 結婚宣言するスート (marriage コマンド: 1=♠ 2=♣ 3=♥ 4=♦)
	Suit *int `json:"suit,omitempty"`
	// Config ゲーム設定
	Config *TuteWebConfig `json:"config,omitempty"`
}

// TuteWebConfig トゥーテのWeb設定
type TuteWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetPoints  *int `json:"targetPoints,omitempty"`
}

// TuteWebOutputPlayer トゥーテのWebアウトプットプレイヤー
type TuteWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	TrickCount int              `json:"trickCount"`
	TeamScore  int              `json:"teamScore"`
}

// TuteWebOutputHint ヒント出力
type TuteWebOutputHint struct {
	CardIndices []int  `json:"cardIndices"`
	Marriage    int    `json:"marriage"`
	Reason      string `json:"reason"`
}

// TuteWebOutput トゥーテのWebアウトプット
type TuteWebOutput struct {
	Players          []*TuteWebOutputPlayer `json:"players"`
	Phase            int                    `json:"phase"`
	RoundNumber      int                    `json:"roundNumber"`
	TrickNumber      int                    `json:"trickNumber"`
	CurrentPlayerIdx int                    `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                    `json:"leadPlayerIdx"`
	DealerIdx        int                    `json:"dealerIdx"`
	TrumpSuit        int                    `json:"trumpSuit"`
	CurrentTrick     []*WebOutputTrickCard  `json:"currentTrick"`
	// DeclaredSuits 結婚宣言済みスート (インデックス1-4が有効; 0は未使用)
	DeclaredSuits      []bool                  `json:"declaredSuits"`
	TeamScores         [domain.TuteTeamCnt]int `json:"teamScores"`
	RoundTeamPoints    [domain.TuteTeamCnt]int `json:"roundTeamPoints"`
	CanDeclareMarriage bool                    `json:"canDeclareMarriage"`
	CanDeclareTute     bool                    `json:"canDeclareTute"`
	PlayableIndices    []int                   `json:"playableIndices"`
	GameEndFlag        bool                    `json:"gameEndFlag"`
	WinnerTeam         int                     `json:"winnerTeam"`
	IsHumanTurn        bool                    `json:"isHumanTurn"`
	Hint               *TuteWebOutputHint      `json:"hint,omitempty"`
	WebOutputBase
	Config TuteWebOutputConfig `json:"config"`
}

// TuteWebOutputConfig トゥーテの設定アウトプット
type TuteWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetPoints  int `json:"targetPoints"`
}

// ToConfig builds a TuteConfig from the nested web config, applying bounds checking.
func (c *TuteWebConfig) ToConfig() domain.TuteConfig {
	cfg := domain.DefaultTuteConfig()
	cfg.CpuDifficulty = domain.TuteCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.TuteCpuDifficultyEasy), int(domain.TuteCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetPoints, c.TargetPoints, 1, 1000000)
	return cfg
}

// ToConfig builds a TuteConfig from the web input.
func (p TuteWebInput) ToConfig() domain.TuteConfig {
	return configOrDefault(p.Config, (*TuteWebConfig).ToConfig, domain.DefaultTuteConfig())
}

// TuteWebController トゥーテのWebコントローラークラス
type TuteWebController = GameWebController[usecase.TuteInteractorIF, TuteWebInput, *TuteWebOutput]

// NewTuteWebController and NewTuteWebControllerWithProvider are
// the standard and provider-backed constructors for TuteWebController.
var NewTuteWebController, NewTuteWebControllerWithProvider = webControllerPair[usecase.TuteInteractorIF, TuteWebInput, *TuteWebOutput](
	newTuteDefaultOutput, tuteDispatch,
)

func newTuteDefaultOutput(msg string) *TuteWebOutput {
	return &TuteWebOutput{
		Players:         make([]*TuteWebOutputPlayer, 0),
		CurrentTrick:    make([]*WebOutputTrickCard, 0),
		DeclaredSuits:   make([]bool, domain.CardDesignMax+1),
		PlayableIndices: make([]int, 0),
		WinnerTeam:      -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func tuteDispatch(bc *baseController, w http.ResponseWriter, di usecase.TuteInteractorIF, param TuteWebInput, newDefault func(string) *TuteWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, di.ResetWithConfig(param.ToConfig()))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.Play(*param.CardIndex))
	case "m", "marriage":
		if !requireParam(bc, w, newDefault, param.Suit == nil, "param error: suit is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.DeclareMarriage(*param.Suit))
	case "tute":
		bc.writePresenterResponse(w, di.DeclareTute())
	case "n", "next":
		bc.writePresenterResponse(w, di.NextTrick())
	case "nr", "nextround":
		bc.writePresenterResponse(w, di.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, di.Hint, di.ActionLog)
	}
	return true
}
