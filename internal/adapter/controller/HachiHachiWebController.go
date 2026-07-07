//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// HachiHachiWebConfig は八八 (Hachi-Hachi) の Web 設定。
type HachiHachiWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetRounds  *int `json:"targetRounds,omitempty"`
}

// ToConfig は HachiHachiWebConfig を domain.HachiHachiConfig に変換する (境界チェック付き)。
func (c *HachiHachiWebConfig) ToConfig() domain.HachiHachiConfig {
	cfg := domain.DefaultHachiHachiConfig()
	cfg.CpuDifficulty = domain.HachiHachiCpuDifficulty(webutil.BoundedIntPtr(
		c.CpuDifficulty, int(domain.HachiHachiCpuDifficultyEasy), int(domain.HachiHachiCpuDifficultyHard), int(cfg.CpuDifficulty)))
	cfg.TargetRounds = webutil.BoundedIntPtr(
		c.TargetRounds, domain.HachiHachiTargetRoundsMin, domain.HachiHachiTargetRoundsMax, cfg.TargetRounds)
	return cfg
}

// HachiHachiWebInput は八八 Web インプット。
type HachiHachiWebInput struct {
	BaseWebInput
	CardIndex  *int                 `json:"cardIndex,omitempty"`
	FieldIndex *int                 `json:"fieldIndex,omitempty"`
	Config     *HachiHachiWebConfig `json:"config,omitempty"`
}

// ToConfig は Web インプットから domain.HachiHachiConfig を構築する。
func (p HachiHachiWebInput) ToConfig() domain.HachiHachiConfig {
	return configOrDefault(p.Config, (*HachiHachiWebConfig).ToConfig, domain.DefaultHachiHachiConfig())
}

// HachiHachiWebOutputYaku は成立出来役 1 件。
type HachiHachiWebOutputYaku struct {
	Key    string `json:"key"`
	Points int    `json:"points"`
}

// HachiHachiWebOutputPlayer は八八 Web アウトプットプレイヤー。
type HachiHachiWebOutputPlayer struct {
	ID            int                        `json:"id"`
	IsHuman       bool                       `json:"isHuman"`
	CardCount     int                        `json:"cardCount"`
	Cards         []*WebOutputCard           `json:"cards"`
	Captured      []*WebOutputCard           `json:"captured"`
	CapturedCount int                        `json:"capturedCount"`
	Score         int                        `json:"score"`
	RoundDelta    int                        `json:"roundDelta"`
	RawScore      int                        `json:"rawScore"`
	Yaku          []*HachiHachiWebOutputYaku `json:"yaku"`
}

// HachiHachiWebOutputPlayerScore は 1 プレイヤーのラウンド精算内訳。
type HachiHachiWebOutputPlayerScore struct {
	PlayerIdx int                        `json:"playerIdx"`
	RawScore  int                        `json:"rawScore"`
	Yaku      []*HachiHachiWebOutputYaku `json:"yaku"`
	Bonus     int                        `json:"bonus"`
	Delta     int                        `json:"delta"`
}

// HachiHachiWebOutputRoundResult は 1 ラウンドの結果。
type HachiHachiWebOutputRoundResult struct {
	Scores []*HachiHachiWebOutputPlayerScore `json:"scores"`
	Best   int                               `json:"best"`
}

// HachiHachiWebOutputHint はヒント出力。
type HachiHachiWebOutputHint struct {
	CardIndex  int    `json:"cardIndex"`
	FieldIndex int    `json:"fieldIndex"`
	Reason     string `json:"reason"`
}

// HachiHachiWebOutput は八八 Web アウトプット。
type HachiHachiWebOutput struct {
	Players         []*HachiHachiWebOutputPlayer    `json:"players"`
	Phase           int                             `json:"phase"`
	RoundNumber     int                             `json:"roundNumber"`
	CurrentTurn     int                             `json:"currentTurn"`
	FieldCards      []*WebOutputCard                `json:"fieldCards"`
	RemainingDeck   int                             `json:"remainingDeck"`
	PlayableIndices []int                           `json:"playableIndices"`
	CaptureOptions  map[int][]int                   `json:"captureOptions"`
	Winner          int                             `json:"winner"`
	Result          int                             `json:"result"`
	GameEndFlag     bool                            `json:"gameEndFlag"`
	IsHumanTurn     bool                            `json:"isHumanTurn"`
	LastRoundResult *HachiHachiWebOutputRoundResult `json:"lastRoundResult"`
	Hint            *HachiHachiWebOutputHint        `json:"hint,omitempty"`
	WebOutputBase
	Config HachiHachiWebConfigOutput `json:"config"`
}

// HachiHachiWebConfigOutput は設定アウトプット。
type HachiHachiWebConfigOutput struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetRounds  int `json:"targetRounds"`
}

// HachiHachiWebController は八八 Web コントローラークラス。
type HachiHachiWebController = GameWebController[usecase.HachiHachiInteractorIF, HachiHachiWebInput, *HachiHachiWebOutput]

// NewHachiHachiWebController, NewHachiHachiWebControllerWithProvider are the standard and
// provider-backed constructors for HachiHachiWebController.
var NewHachiHachiWebController, NewHachiHachiWebControllerWithProvider = webControllerPair[usecase.HachiHachiInteractorIF, HachiHachiWebInput, *HachiHachiWebOutput](
	newHachiHachiDefaultOutput, hachihachiDispatch,
)

func newHachiHachiDefaultOutput(msg string) *HachiHachiWebOutput {
	return &HachiHachiWebOutput{
		Players:         make([]*HachiHachiWebOutputPlayer, 0),
		FieldCards:      make([]*WebOutputCard, 0),
		PlayableIndices: make([]int, 0),
		CaptureOptions:  make(map[int][]int),
		Winner:          -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func hachihachiDispatch(bc *baseController, w http.ResponseWriter, ki usecase.HachiHachiInteractorIF, param HachiHachiWebInput, newDefault func(string) *HachiHachiWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ki.ResetWithConfig(param.ToConfig()))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		fieldIdx := -1
		if param.FieldIndex != nil {
			fieldIdx = *param.FieldIndex
		}
		bc.writePresenterResponse(w, ki.Play(*param.CardIndex, fieldIdx))
	case "nr", "nextround", "n", "next":
		bc.writePresenterResponse(w, ki.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, ki.Hint, ki.ActionLog)
	}
	return true
}
