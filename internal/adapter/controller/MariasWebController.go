//go:build !js || !wasm || classic

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// MariasWebInput マリアーシュのWebインプット
type MariasWebInput struct {
	BaseWebInput
	// CardIndex プレイするカードのインデックス (play コマンド)
	CardIndex *int `json:"cardIndex,omitempty"`
	// Config ゲーム設定
	Config *MariasWebConfig `json:"config,omitempty"`
}

// MariasWebConfig マリアーシュのWeb設定
type MariasWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetPoints  *int `json:"targetPoints,omitempty"`
}

// MariasWebOutputPlayer マリアーシュのWebアウトプットプレイヤー
type MariasWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	TrickCount int              `json:"trickCount"`
	Score      int              `json:"score"`
	IsSoloist  bool             `json:"isSoloist"`
}

// MariasWebOutput マリアーシュのWebアウトプット
type MariasWebOutput struct {
	Players          []*MariasWebOutputPlayer    `json:"players"`
	Phase            int                         `json:"phase"`
	RoundNumber      int                         `json:"roundNumber"`
	TrickNumber      int                         `json:"trickNumber"`
	CurrentPlayerIdx int                         `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                         `json:"leadPlayerIdx"`
	DealerIdx        int                         `json:"dealerIdx"`
	SoloistIdx       int                         `json:"soloistIdx"`
	TrumpSuit        int                         `json:"trumpSuit"`
	CurrentTrick     []*WebOutputTrickCard       `json:"currentTrick"`
	PlayerScores     [domain.MariasPlayerCnt]int `json:"playerScores"`
	RoundCardPoints  [domain.MariasPlayerCnt]int `json:"roundCardPoints"`
	RoundMarriage    [domain.MariasPlayerCnt]int `json:"roundMarriage"`
	LastTrickWinner  int                         `json:"lastTrickWinner"`
	PlayableIndices  []int                       `json:"playableIndices"`
	GameEndFlag      bool                        `json:"gameEndFlag"`
	WinnerPlayer     int                         `json:"winnerPlayer"`
	IsHumanTurn      bool                        `json:"isHumanTurn"`
	Hint             *WebOutputCardHint          `json:"hint,omitempty"`
	WebOutputBase
	Config MariasWebOutputConfig `json:"config"`
}

// MariasWebOutputConfig マリアーシュの設定アウトプット
type MariasWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetPoints  int `json:"targetPoints"`
}

// ToConfig builds a MariasConfig from the nested web config, applying bounds checking.
func (c *MariasWebConfig) ToConfig() domain.MariasConfig {
	cfg := domain.DefaultMariasConfig()
	cfg.CpuDifficulty = domain.MariasCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.MariasCpuDifficultyEasy), int(domain.MariasCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetPoints, c.TargetPoints, 1, 1000000)
	return cfg
}

// ToConfig builds a MariasConfig from the web input.
func (p MariasWebInput) ToConfig() domain.MariasConfig {
	return configOrDefault(p.Config, (*MariasWebConfig).ToConfig, domain.DefaultMariasConfig())
}

// MariasWebController マリアーシュのWebコントローラークラス
type MariasWebController = GameWebController[usecase.MariasInteractorIF, MariasWebInput, *MariasWebOutput]

// NewMariasWebController and NewMariasWebControllerWithProvider are
// the standard and provider-backed constructors for MariasWebController.
var NewMariasWebController, NewMariasWebControllerWithProvider = webControllerPair[usecase.MariasInteractorIF, MariasWebInput, *MariasWebOutput](
	newMariasDefaultOutput, mariasDispatch,
)

func newMariasDefaultOutput(msg string) *MariasWebOutput {
	return &MariasWebOutput{
		Players:         make([]*MariasWebOutputPlayer, 0),
		CurrentTrick:    make([]*WebOutputTrickCard, 0),
		PlayableIndices: make([]int, 0),
		LastTrickWinner: -1,
		WinnerPlayer:    -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func mariasDispatch(bc *baseController, w http.ResponseWriter, di usecase.MariasInteractorIF, param MariasWebInput, newDefault func(string) *MariasWebOutput) bool {
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
